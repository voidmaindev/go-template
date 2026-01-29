package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strconv"
	"time"

	goredis "github.com/redis/go-redis/v9"
	"github.com/voidmaindev/go-template/internal/redis"
)

// Redis key prefixes for auth tokens
const (
	keyPrefixVerify     = "auth:verify:"
	keyPrefixReset      = "auth:reset:"
	keyPrefixOAuthState = "auth:oauth:state:"
	keyPrefixRateLimit  = "auth:rate:"
)

// TokenStore handles storage of verification and password reset tokens in Redis
type TokenStore struct {
	redis *redis.Client
}

// NewTokenStore creates a new token store
func NewTokenStore(redisClient *redis.Client) *TokenStore {
	return &TokenStore{
		redis: redisClient,
	}
}

// GenerateToken generates a cryptographically secure random token
func GenerateToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// StoreVerificationToken stores a verification token for a user
func (s *TokenStore) StoreVerificationToken(ctx context.Context, token string, userID uint, expiry time.Duration) error {
	key := keyPrefixVerify + token
	return s.redis.Set(ctx, key, strconv.FormatUint(uint64(userID), 10), expiry).Err()
}

// GetVerificationToken retrieves the user ID associated with a verification token
func (s *TokenStore) GetVerificationToken(ctx context.Context, token string) (uint, error) {
	key := keyPrefixVerify + token
	result, err := s.redis.Get(ctx, key).Result()
	if err != nil {
		if err == goredis.Nil {
			return 0, ErrInvalidToken
		}
		return 0, err
	}

	userID, err := strconv.ParseUint(result, 10, 64)
	if err != nil {
		return 0, err
	}
	return uint(userID), nil
}

// DeleteVerificationToken deletes a verification token (single-use)
func (s *TokenStore) DeleteVerificationToken(ctx context.Context, token string) error {
	key := keyPrefixVerify + token
	return s.redis.Del(ctx, key).Err()
}

// StorePasswordResetToken stores a password reset token for a user
func (s *TokenStore) StorePasswordResetToken(ctx context.Context, token string, userID uint, expiry time.Duration) error {
	key := keyPrefixReset + token
	return s.redis.Set(ctx, key, strconv.FormatUint(uint64(userID), 10), expiry).Err()
}

// GetPasswordResetToken retrieves the user ID associated with a password reset token
func (s *TokenStore) GetPasswordResetToken(ctx context.Context, token string) (uint, error) {
	key := keyPrefixReset + token
	result, err := s.redis.Get(ctx, key).Result()
	if err != nil {
		if err == goredis.Nil {
			return 0, ErrInvalidToken
		}
		return 0, err
	}

	userID, err := strconv.ParseUint(result, 10, 64)
	if err != nil {
		return 0, err
	}
	return uint(userID), nil
}

// DeletePasswordResetToken deletes a password reset token (single-use)
func (s *TokenStore) DeletePasswordResetToken(ctx context.Context, token string) error {
	key := keyPrefixReset + token
	return s.redis.Del(ctx, key).Err()
}

// OAuthStateData holds data for OAuth state validation
type OAuthStateData struct {
	Provider    string `json:"provider"`
	RedirectURL string `json:"redirect_url,omitempty"`
	UserID      uint   `json:"user_id,omitempty"` // For linking flow
}

// StoreOAuthState stores OAuth state for CSRF protection
func (s *TokenStore) StoreOAuthState(ctx context.Context, state string, data *OAuthStateData, expiry time.Duration) error {
	key := keyPrefixOAuthState + state
	// Store as simple format: provider:userID:redirectURL
	value := fmt.Sprintf("%s:%d:%s", data.Provider, data.UserID, data.RedirectURL)
	return s.redis.Set(ctx, key, value, expiry).Err()
}

// GetOAuthState retrieves OAuth state data
func (s *TokenStore) GetOAuthState(ctx context.Context, state string) (*OAuthStateData, error) {
	key := keyPrefixOAuthState + state
	result, err := s.redis.Get(ctx, key).Result()
	if err != nil {
		if err == goredis.Nil {
			return nil, ErrOAuthStateMismatch
		}
		return nil, err
	}

	// Parse: provider:userID:redirectURL
	var provider, redirectURL string
	var userID uint64
	_, err = fmt.Sscanf(result, "%s:%d:%s", &provider, &userID, &redirectURL)
	if err != nil {
		// Fallback to simple split
		parts := splitOAuthState(result)
		provider = parts[0]
		if len(parts) > 1 {
			userID, _ = strconv.ParseUint(parts[1], 10, 64)
		}
		if len(parts) > 2 {
			redirectURL = parts[2]
		}
	}

	return &OAuthStateData{
		Provider:    provider,
		UserID:      uint(userID),
		RedirectURL: redirectURL,
	}, nil
}

// DeleteOAuthState deletes OAuth state (single-use)
func (s *TokenStore) DeleteOAuthState(ctx context.Context, state string) error {
	key := keyPrefixOAuthState + state
	return s.redis.Del(ctx, key).Err()
}

// CheckRateLimit checks if an action is rate limited
// Returns true if allowed, false if rate limited
func (s *TokenStore) CheckRateLimit(ctx context.Context, identifier, action string, limit int, window time.Duration) (bool, error) {
	key := fmt.Sprintf("%s%s:%s", keyPrefixRateLimit, action, identifier)

	// Increment counter
	count, err := s.redis.Incr(ctx, key).Result()
	if err != nil {
		return false, err
	}

	// Set expiry on first request
	if count == 1 {
		s.redis.Expire(ctx, key, window)
	}

	return count <= int64(limit), nil
}

// splitOAuthState splits OAuth state string by colon
func splitOAuthState(s string) []string {
	result := make([]string, 0, 3)
	current := ""
	for _, c := range s {
		if c == ':' {
			result = append(result, current)
			current = ""
		} else {
			current += string(c)
		}
	}
	result = append(result, current)
	return result
}
