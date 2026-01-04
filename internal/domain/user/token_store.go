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
