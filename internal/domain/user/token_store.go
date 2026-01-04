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
	for i := 0; i < maxRetries; i++ {
		if err := s.Blacklist(ctx, token, expiry); err != nil {
			lastErr = err
			// Brief pause before retry
			time.Sleep(time.Millisecond * 50 * time.Duration(i+1))
			continue
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
func (s *TokenStore) BlacklistMultiple(ctx context.Context, tokens []string, expiry time.Duration) error {
	for _, token := range tokens {
		if err := s.Blacklist(ctx, token, expiry); err != nil {
			return err
		}
	}
	return nil
}

// GetTTL returns the remaining time-to-live of a blacklisted token
func (s *TokenStore) GetTTL(ctx context.Context, token string) (time.Duration, error) {
	key := s.getKey(token)
	return s.redis.GetTTL(ctx, key)
}
