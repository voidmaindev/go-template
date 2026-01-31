package user

import (
	"context"
	"fmt"
	"time"

	"github.com/voidmaindev/go-template/internal/redis"
)

const (
	// TokenBlacklistPrefix is the Redis key prefix for blacklisted tokens
	TokenBlacklistPrefix = "token:blacklist:"
	// LoginRateEmailPrefix is the Redis key prefix for login rate limiting by email
	LoginRateEmailPrefix = "auth:login:rate:email:"
	// LoginRateIPPrefix is the Redis key prefix for login rate limiting by IP
	LoginRateIPPrefix = "auth:login:rate:ip:"
	// TokenInvalidatePrefix is the Redis key prefix for user token invalidation timestamps
	TokenInvalidatePrefix = "auth:token:invalidated:"
)

// TokenStore handles token blacklisting using Redis
type TokenStore struct {
	redis *redis.Client
}

// NewTokenStore creates a new TokenStore
func NewTokenStore(redisClient *redis.Client) *TokenStore {
	return &TokenStore{
		redis: redisClient,
	}
}

// Blacklist adds a token to the blacklist with an expiry time
func (s *TokenStore) Blacklist(ctx context.Context, token string, expiry time.Duration) error {
	key := s.getKey(token)
	return s.redis.SetWithExpiry(ctx, key, "1", expiry)
}

// BlacklistWithRetry adds a token to the blacklist with retry logic
func (s *TokenStore) BlacklistWithRetry(ctx context.Context, token string, expiry time.Duration, maxRetries int) error {
	var lastErr error
	backoff := time.Millisecond * 50

	for i := 0; i < maxRetries; i++ {
		if err := s.Blacklist(ctx, token, expiry); err != nil {
			lastErr = err
			// Context-aware backoff
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(backoff * time.Duration(i+1)):
				continue
			}
		}
		return nil
	}
	return lastErr
}

// IsBlacklisted checks if a token is blacklisted
func (s *TokenStore) IsBlacklisted(ctx context.Context, token string) (bool, error) {
	key := s.getKey(token)
	return s.redis.Exists(ctx, key)
}

// Remove removes a token from the blacklist
func (s *TokenStore) Remove(ctx context.Context, token string) error {
	key := s.getKey(token)
	return s.redis.DeleteKey(ctx, key)
}

// getKey generates the Redis key for a token
func (s *TokenStore) getKey(token string) string {
	return fmt.Sprintf("%s%s", TokenBlacklistPrefix, token)
}

// BlacklistMultiple adds multiple tokens to the blacklist
// Uses Redis pipeline for efficiency and collects all errors instead of stopping on first
func (s *TokenStore) BlacklistMultiple(ctx context.Context, tokens []string, expiry time.Duration) error {
	if len(tokens) == 0 {
		return nil
	}

	// Prepare key-value pairs for pipeline
	keys := make([]string, len(tokens))
	for i, token := range tokens {
		keys[i] = s.getKey(token)
	}

	// Use pipeline for atomic batch operation
	return s.redis.SetMultipleWithExpiry(ctx, keys, "1", expiry)
}

// GetTTL returns the remaining time-to-live of a blacklisted token
func (s *TokenStore) GetTTL(ctx context.Context, token string) (time.Duration, error) {
	key := s.getKey(token)
	return s.redis.GetTTL(ctx, key)
}

// BlacklistAtomic adds a token to the blacklist atomically using SETNX.
// Returns true if the token was newly blacklisted, false if it was already blacklisted.
// This prevents race conditions in token refresh where the same refresh token
// could be used multiple times before blacklisting completes.
func (s *TokenStore) BlacklistAtomic(ctx context.Context, token string, expiry time.Duration) (bool, error) {
	key := s.getKey(token)
	return s.redis.SetNX(ctx, key, "1", expiry)
}

// BlacklistAtomicWithRetry adds a token to the blacklist atomically with retry logic.
// Returns true if newly blacklisted, false if already blacklisted or failed after retries.
func (s *TokenStore) BlacklistAtomicWithRetry(ctx context.Context, token string, expiry time.Duration, maxRetries int) (bool, error) {
	var lastErr error
	backoff := time.Millisecond * 50

	for i := 0; i < maxRetries; i++ {
		wasSet, err := s.BlacklistAtomic(ctx, token, expiry)
		if err != nil {
			lastErr = err
			// Context-aware backoff
			select {
			case <-ctx.Done():
				return false, ctx.Err()
			case <-time.After(backoff * time.Duration(i+1)):
				continue
			}
		}
		return wasSet, nil
	}
	return false, lastErr
}

// ================================
// Login Rate Limiting
// ================================

// CheckLoginRateLimit checks if login is allowed for the given email and IP
// Returns nil if allowed, error if rate limited
func (s *TokenStore) CheckLoginRateLimit(ctx context.Context, email, ip string, maxPerEmail, maxPerIP int, lockoutDuration time.Duration) error {
	// Check per-email limit
	emailKey := LoginRateEmailPrefix + email
	emailCount, err := s.redis.GetInt(ctx, emailKey)
	if err != nil && !s.redis.IsNil(err) {
		return err
	}
	if emailCount >= maxPerEmail {
		return ErrTooManyLoginAttempts
	}

	// Check per-IP limit
	ipKey := LoginRateIPPrefix + ip
	ipCount, err := s.redis.GetInt(ctx, ipKey)
	if err != nil && !s.redis.IsNil(err) {
		return err
	}
	if ipCount >= maxPerIP {
		return ErrTooManyLoginAttempts
	}

	return nil
}

// RecordFailedLogin increments the failed login counters for email and IP
func (s *TokenStore) RecordFailedLogin(ctx context.Context, email, ip string, lockoutDuration time.Duration) error {
	// Increment email counter
	emailKey := LoginRateEmailPrefix + email
	if err := s.incrementWithExpiry(ctx, emailKey, lockoutDuration); err != nil {
		return err
	}

	// Increment IP counter
	ipKey := LoginRateIPPrefix + ip
	return s.incrementWithExpiry(ctx, ipKey, lockoutDuration)
}

// ClearLoginRateLimit clears the login rate limit for an email (after successful login)
func (s *TokenStore) ClearLoginRateLimit(ctx context.Context, email string) error {
	key := LoginRateEmailPrefix + email
	return s.redis.DeleteKey(ctx, key)
}

// incrementWithExpiry atomically increments a counter and sets expiry
func (s *TokenStore) incrementWithExpiry(ctx context.Context, key string, expiry time.Duration) error {
	count, err := s.redis.Incr(ctx, key)
	if err != nil {
		return err
	}
	// Set expiry only on first increment
	if count == 1 {
		return s.redis.SetExpiry(ctx, key, expiry)
	}
	return nil
}

// ================================
// Token Invalidation
// ================================

// InvalidateAllUserTokens stores the current timestamp to invalidate all tokens issued before it
func (s *TokenStore) InvalidateAllUserTokens(ctx context.Context, userID uint) error {
	if s.redis == nil {
		return nil // No Redis client (e.g., in tests)
	}
	key := fmt.Sprintf("%s%d", TokenInvalidatePrefix, userID)
	// Store current timestamp with no expiry (persists until explicitly cleared)
	return s.redis.SetValue(ctx, key, time.Now().Unix(), 0)
}

// GetTokensInvalidatedAt returns the timestamp when a user's tokens were invalidated
// Returns zero time if no invalidation has occurred
func (s *TokenStore) GetTokensInvalidatedAt(ctx context.Context, userID uint) (time.Time, error) {
	if s.redis == nil {
		return time.Time{}, nil // No Redis client (e.g., in tests)
	}
	key := fmt.Sprintf("%s%d", TokenInvalidatePrefix, userID)
	timestamp, err := s.redis.GetInt64(ctx, key)
	if err != nil {
		if s.redis.IsNil(err) {
			return time.Time{}, nil // No invalidation recorded
		}
		return time.Time{}, err
	}
	return time.Unix(timestamp, 0), nil
}
