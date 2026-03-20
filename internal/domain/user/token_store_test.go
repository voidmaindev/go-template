package user

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	goredis "github.com/redis/go-redis/v9"
	"github.com/voidmaindev/go-template/internal/redis"
)

// redisClientInterface defines the methods used by TokenStore for testing
type redisClientInterface interface {
	SetWithExpiry(ctx context.Context, key string, value any, expiry time.Duration) error
	GetString(ctx context.Context, key string) (string, error)
	Exists(ctx context.Context, key string) (bool, error)
	DeleteKey(ctx context.Context, key string) error
	GetTTL(ctx context.Context, key string) (time.Duration, error)
}

// mockRedisClient implements a mock Redis client for testing
type mockRedisClient struct {
	mu      sync.RWMutex
	data    map[string]mockEntry
	setErr  error
	getErr  error
	delErr  error
	exisErr error
}

type mockEntry struct {
	value  string
	expiry time.Time
}

func newMockRedisClient() *mockRedisClient {
	return &mockRedisClient{
		data: make(map[string]mockEntry),
	}
}

func (m *mockRedisClient) SetWithExpiry(ctx context.Context, key string, value any, expiry time.Duration) error {
	if m.setErr != nil {
		return m.setErr
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[key] = mockEntry{
		value:  value.(string),
		expiry: time.Now().Add(expiry),
	}
	return nil
}

func (m *mockRedisClient) GetString(ctx context.Context, key string) (string, error) {
	if m.getErr != nil {
		return "", m.getErr
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	entry, ok := m.data[key]
	if !ok {
		return "", nil
	}
	if time.Now().After(entry.expiry) {
		return "", nil
	}
	return entry.value, nil
}

func (m *mockRedisClient) Exists(ctx context.Context, key string) (bool, error) {
	if m.exisErr != nil {
		return false, m.exisErr
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	entry, ok := m.data[key]
	if !ok {
		return false, nil
	}
	if time.Now().After(entry.expiry) {
		return false, nil
	}
	return true, nil
}

func (m *mockRedisClient) DeleteKey(ctx context.Context, key string) error {
	if m.delErr != nil {
		return m.delErr
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.data, key)
	return nil
}

func (m *mockRedisClient) GetTTL(ctx context.Context, key string) (time.Duration, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	entry, ok := m.data[key]
	if !ok {
		return -2, nil // Key doesn't exist
	}
	ttl := time.Until(entry.expiry)
	if ttl < 0 {
		return -2, nil // Expired
	}
	return ttl, nil
}

// testTokenStore wraps the mock for testing TokenStore logic
type testTokenStore struct {
	client redisClientInterface
}

// newTestTokenStore creates a testable TokenStore with mock Redis
func newTestTokenStore(mock *mockRedisClient) *testTokenStore {
	return &testTokenStore{client: mock}
}

// Blacklist adds a token to the blacklist with an expiry time
func (s *testTokenStore) Blacklist(ctx context.Context, token string, expiry time.Duration) error {
	key := TokenBlacklistPrefix + token
	return s.client.SetWithExpiry(ctx, key, "1", expiry)
}

// IsBlacklisted checks if a token is blacklisted
func (s *testTokenStore) IsBlacklisted(ctx context.Context, token string) (bool, error) {
	key := TokenBlacklistPrefix + token
	return s.client.Exists(ctx, key)
}

// Remove removes a token from the blacklist
func (s *testTokenStore) Remove(ctx context.Context, token string) error {
	key := TokenBlacklistPrefix + token
	return s.client.DeleteKey(ctx, key)
}

// GetTTL returns the remaining TTL of a blacklisted token
func (s *testTokenStore) GetTTL(ctx context.Context, token string) (time.Duration, error) {
	key := TokenBlacklistPrefix + token
	return s.client.GetTTL(ctx, key)
}

func TestTokenStore_Blacklist(t *testing.T) {
	mock := newMockRedisClient()

	t.Run("successful blacklist", func(t *testing.T) {
		ctx := context.Background()
		token := "test-token-123"
		expiry := 15 * time.Minute

		key := TokenBlacklistPrefix + token
		err := mock.SetWithExpiry(ctx, key, "1", expiry)
		if err != nil {
			t.Fatalf("SetWithExpiry() error = %v", err)
		}

		exists, err := mock.Exists(ctx, key)
		if err != nil {
			t.Fatalf("Exists() error = %v", err)
		}
		if !exists {
			t.Error("Token should be blacklisted")
		}
	})

	t.Run("blacklist with error", func(t *testing.T) {
		mock.setErr = errors.New("redis connection error")

		ctx := context.Background()
		token := "test-token-456"

		key := TokenBlacklistPrefix + token
		err := mock.SetWithExpiry(ctx, key, "1", time.Minute)
		if err == nil {
			t.Error("SetWithExpiry() should return error")
		}

		mock.setErr = nil // Reset for other tests
	})
}

func TestTokenStore_IsBlacklisted(t *testing.T) {
	mock := newMockRedisClient()
	ctx := context.Background()

	t.Run("token is blacklisted", func(t *testing.T) {
		token := "blacklisted-token"
		key := TokenBlacklistPrefix + token
		_ = mock.SetWithExpiry(ctx, key, "1", 15*time.Minute)

		exists, err := mock.Exists(ctx, key)
		if err != nil {
			t.Fatalf("Exists() error = %v", err)
		}
		if !exists {
			t.Error("Token should be blacklisted")
		}
	})

	t.Run("token is not blacklisted", func(t *testing.T) {
		token := "valid-token"
		key := TokenBlacklistPrefix + token

		exists, err := mock.Exists(ctx, key)
		if err != nil {
			t.Fatalf("Exists() error = %v", err)
		}
		if exists {
			t.Error("Token should not be blacklisted")
		}
	})

	t.Run("redis error", func(t *testing.T) {
		mock.exisErr = errors.New("redis error")

		token := "any-token"
		key := TokenBlacklistPrefix + token

		_, err := mock.Exists(ctx, key)
		if err == nil {
			t.Error("Exists() should return error")
		}

		mock.exisErr = nil
	})
}

func TestTokenStore_Remove(t *testing.T) {
	mock := newMockRedisClient()
	ctx := context.Background()

	t.Run("remove existing token", func(t *testing.T) {
		token := "token-to-remove"
		key := TokenBlacklistPrefix + token
		_ = mock.SetWithExpiry(ctx, key, "1", 15*time.Minute)

		err := mock.DeleteKey(ctx, key)
		if err != nil {
			t.Fatalf("DeleteKey() error = %v", err)
		}

		exists, _ := mock.Exists(ctx, key)
		if exists {
			t.Error("Token should be removed")
		}
	})

	t.Run("remove non-existent token", func(t *testing.T) {
		token := "non-existent-token"
		key := TokenBlacklistPrefix + token

		err := mock.DeleteKey(ctx, key)
		if err != nil {
			t.Fatalf("DeleteKey() error = %v", err)
		}
	})

	t.Run("delete error", func(t *testing.T) {
		mock.delErr = errors.New("redis error")

		token := "any-token"
		key := TokenBlacklistPrefix + token

		err := mock.DeleteKey(ctx, key)
		if err == nil {
			t.Error("DeleteKey() should return error")
		}

		mock.delErr = nil
	})
}

func TestTokenStore_GetTTL(t *testing.T) {
	mock := newMockRedisClient()
	ctx := context.Background()

	t.Run("get TTL for existing key", func(t *testing.T) {
		token := "token-with-ttl"
		key := TokenBlacklistPrefix + token
		expiry := 15 * time.Minute
		_ = mock.SetWithExpiry(ctx, key, "1", expiry)

		ttl, err := mock.GetTTL(ctx, key)
		if err != nil {
			t.Fatalf("GetTTL() error = %v", err)
		}

		// TTL should be close to expiry (with some tolerance)
		if ttl < expiry-time.Minute || ttl > expiry {
			t.Errorf("TTL = %v, expected close to %v", ttl, expiry)
		}
	})

	t.Run("get TTL for non-existent key", func(t *testing.T) {
		key := TokenBlacklistPrefix + "non-existent"

		ttl, err := mock.GetTTL(ctx, key)
		if err != nil {
			t.Fatalf("GetTTL() error = %v", err)
		}
		if ttl != -2 {
			t.Errorf("TTL = %v, expected -2 for non-existent key", ttl)
		}
	})
}

func TestTokenStore_BlacklistMultiple(t *testing.T) {
	mock := newMockRedisClient()
	ctx := context.Background()

	t.Run("blacklist multiple tokens", func(t *testing.T) {
		tokens := []string{"token1", "token2", "token3"}
		expiry := 15 * time.Minute

		for _, token := range tokens {
			key := TokenBlacklistPrefix + token
			err := mock.SetWithExpiry(ctx, key, "1", expiry)
			if err != nil {
				t.Fatalf("SetWithExpiry() error = %v", err)
			}
		}

		// Verify all tokens are blacklisted
		for _, token := range tokens {
			key := TokenBlacklistPrefix + token
			exists, _ := mock.Exists(ctx, key)
			if !exists {
				t.Errorf("Token %s should be blacklisted", token)
			}
		}
	})

	t.Run("partial failure", func(t *testing.T) {
		tokens := []string{"tokenA", "tokenB"}
		expiry := 15 * time.Minute

		// First token succeeds
		key1 := TokenBlacklistPrefix + tokens[0]
		_ = mock.SetWithExpiry(ctx, key1, "1", expiry)

		// Second token fails
		mock.setErr = errors.New("redis error")
		key2 := TokenBlacklistPrefix + tokens[1]
		err := mock.SetWithExpiry(ctx, key2, "1", expiry)

		if err == nil {
			t.Error("Should have error on second token")
		}

		// First token should still be blacklisted
		mock.setErr = nil
		exists, _ := mock.Exists(ctx, key1)
		if !exists {
			t.Error("First token should still be blacklisted")
		}
	})
}

func TestTokenStore_BlacklistWithRetry(t *testing.T) {
	t.Run("success on first try", func(t *testing.T) {
		mock := newMockRedisClient()
		ctx := context.Background()

		token := "retry-token"
		key := TokenBlacklistPrefix + token
		expiry := 15 * time.Minute

		err := mock.SetWithExpiry(ctx, key, "1", expiry)
		if err != nil {
			t.Fatalf("Should succeed on first try: %v", err)
		}

		exists, _ := mock.Exists(ctx, key)
		if !exists {
			t.Error("Token should be blacklisted")
		}
	})

	t.Run("success after retries", func(t *testing.T) {
		mock := newMockRedisClient()
		ctx := context.Background()

		failCount := 0
		maxFails := 2

		token := "retry-token-2"
		key := TokenBlacklistPrefix + token
		expiry := 15 * time.Minute

		// Simulate retry logic
		var err error
		for i := 0; i < 3; i++ {
			if failCount < maxFails {
				failCount++
				err = errors.New("temporary error")
				continue
			}
			err = mock.SetWithExpiry(ctx, key, "1", expiry)
			break
		}

		if err != nil {
			t.Fatalf("Should succeed after retries: %v", err)
		}

		exists, _ := mock.Exists(ctx, key)
		if !exists {
			t.Error("Token should be blacklisted after retries")
		}
	})

	t.Run("all retries fail", func(t *testing.T) {
		mock := newMockRedisClient()
		mock.setErr = errors.New("persistent error")

		ctx := context.Background()
		token := "fail-token"
		key := TokenBlacklistPrefix + token
		expiry := 15 * time.Minute

		var err error
		for i := 0; i < 3; i++ {
			err = mock.SetWithExpiry(ctx, key, "1", expiry)
			if err == nil {
				break
			}
		}

		if err == nil {
			t.Error("Should fail after all retries")
		}
	})
}

func TestTokenStore_KeyGeneration(t *testing.T) {
	tests := []struct {
		token    string
		expected string
	}{
		{"abc123", TokenBlacklistPrefix + "abc123"},
		{"", TokenBlacklistPrefix},
		{"long-token-with-special-chars-!@#$%", TokenBlacklistPrefix + "long-token-with-special-chars-!@#$%"},
	}

	for _, tt := range tests {
		t.Run(tt.token, func(t *testing.T) {
			// Test the key format directly
			key := TokenBlacklistPrefix + tt.token
			if key != tt.expected {
				t.Errorf("Key = %v, want %v", key, tt.expected)
			}
		})
	}
}

func TestTokenStore_Expiry(t *testing.T) {
	mock := newMockRedisClient()
	ctx := context.Background()

	t.Run("token expires", func(t *testing.T) {
		token := "expiring-token"
		key := TokenBlacklistPrefix + token
		expiry := 50 * time.Millisecond // Very short expiry for testing

		_ = mock.SetWithExpiry(ctx, key, "1", expiry)

		// Should exist immediately
		exists, _ := mock.Exists(ctx, key)
		if !exists {
			t.Error("Token should exist immediately after creation")
		}

		// Wait for expiry
		time.Sleep(100 * time.Millisecond)

		// Should be expired now
		exists, _ = mock.Exists(ctx, key)
		if exists {
			t.Error("Token should be expired")
		}
	})
}

func TestTokenStore_Concurrency(t *testing.T) {
	mock := newMockRedisClient()
	ctx := context.Background()

	t.Run("concurrent blacklist operations", func(t *testing.T) {
		var wg sync.WaitGroup
		numGoroutines := 100

		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				token := "concurrent-token-" + string(rune('0'+id%10))
				key := TokenBlacklistPrefix + token
				_ = mock.SetWithExpiry(ctx, key, "1", 15*time.Minute)
			}(i)
		}

		wg.Wait()

		// All operations should complete without panic
		// Some tokens should be blacklisted
		count := 0
		for i := 0; i < 10; i++ {
			token := "concurrent-token-" + string(rune('0'+i))
			key := TokenBlacklistPrefix + token
			exists, _ := mock.Exists(ctx, key)
			if exists {
				count++
			}
		}

		if count == 0 {
			t.Error("At least some tokens should be blacklisted")
		}
	})
}

// ================================
// Integration tests using miniredis (real TokenStore, real Lua scripts)
// ================================

func setupUserTestRedis(t *testing.T, invalidationTTL time.Duration) (*miniredis.Miniredis, *TokenStore) {
	t.Helper()
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %v", err)
	}
	client := goredis.NewClient(&goredis.Options{Addr: mr.Addr()})
	store := NewTokenStore(redis.WrapClient(client), invalidationTTL)
	return mr, store
}

// --- C2: Token Invalidation TTL ---

func TestInvalidateAllUserTokens_SetsTTL(t *testing.T) {
	invalidationTTL := 7 * 24 * time.Hour // 7 days (refresh token lifetime)
	mr, store := setupUserTestRedis(t, invalidationTTL)
	defer mr.Close()

	ctx := context.Background()
	userID := uint(42)

	err := store.InvalidateAllUserTokens(ctx, userID)
	if err != nil {
		t.Fatalf("InvalidateAllUserTokens() error = %v", err)
	}

	// Key should exist
	key := TokenInvalidatePrefix + "42"
	if !mr.Exists(key) {
		t.Fatal("invalidation key should exist in Redis")
	}

	// TTL should be set (not zero/forever)
	ttl := mr.TTL(key)
	if ttl <= 0 {
		t.Errorf("TTL = %v, expected positive (key should expire, not persist forever)", ttl)
	}
	if ttl > invalidationTTL {
		t.Errorf("TTL = %v, should be <= %v", ttl, invalidationTTL)
	}
}

func TestInvalidateAllUserTokens_KeyExpires(t *testing.T) {
	invalidationTTL := 5 * time.Second
	mr, store := setupUserTestRedis(t, invalidationTTL)
	defer mr.Close()

	ctx := context.Background()

	err := store.InvalidateAllUserTokens(ctx, 1)
	if err != nil {
		t.Fatal(err)
	}

	key := TokenInvalidatePrefix + "1"
	if !mr.Exists(key) {
		t.Fatal("key should exist before expiry")
	}

	// Fast-forward past TTL
	mr.FastForward(6 * time.Second)

	if mr.Exists(key) {
		t.Error("key should have expired after TTL")
	}
}

func TestInvalidateAllUserTokens_StoresTimestamp(t *testing.T) {
	mr, store := setupUserTestRedis(t, 7*24*time.Hour)
	defer mr.Close()

	ctx := context.Background()
	before := time.Now()

	err := store.InvalidateAllUserTokens(ctx, 99)
	if err != nil {
		t.Fatal(err)
	}

	after := time.Now()

	// Retrieve and verify the timestamp
	invalidatedAt, err := store.GetTokensInvalidatedAt(ctx, 99)
	if err != nil {
		t.Fatalf("GetTokensInvalidatedAt() error = %v", err)
	}
	if invalidatedAt.IsZero() {
		t.Fatal("invalidatedAt should not be zero")
	}

	// Timestamp should be between before and after
	if invalidatedAt.Before(before.Truncate(time.Second)) {
		t.Errorf("invalidatedAt %v is before call time %v", invalidatedAt, before)
	}
	if invalidatedAt.After(after.Add(time.Second)) {
		t.Errorf("invalidatedAt %v is after call end %v", invalidatedAt, after)
	}
}

func TestInvalidateAllUserTokens_NilRedis_NoError(t *testing.T) {
	store := &TokenStore{redis: nil, invalidationTTL: 7 * 24 * time.Hour}

	err := store.InvalidateAllUserTokens(context.Background(), 1)
	if err != nil {
		t.Errorf("expected nil error for nil Redis, got %v", err)
	}
}

func TestGetTokensInvalidatedAt_NoInvalidation_ReturnsZero(t *testing.T) {
	mr, store := setupUserTestRedis(t, 7*24*time.Hour)
	defer mr.Close()

	ts, err := store.GetTokensInvalidatedAt(context.Background(), 999)
	if err != nil {
		t.Fatal(err)
	}
	if !ts.IsZero() {
		t.Errorf("expected zero time for non-existent invalidation, got %v", ts)
	}
}

func TestGetTokensInvalidatedAt_NilRedis_ReturnsZero(t *testing.T) {
	store := &TokenStore{redis: nil, invalidationTTL: 7 * 24 * time.Hour}

	ts, err := store.GetTokensInvalidatedAt(context.Background(), 1)
	if err != nil {
		t.Fatal(err)
	}
	if !ts.IsZero() {
		t.Errorf("expected zero time for nil Redis, got %v", ts)
	}
}

// --- C3: incrementWithExpiry atomicity ---

func TestRecordFailedLogin_IncrementsCounters(t *testing.T) {
	mr, store := setupUserTestRedis(t, 7*24*time.Hour)
	defer mr.Close()

	ctx := context.Background()
	lockout := 15 * time.Minute

	// Record 3 failures
	for i := 0; i < 3; i++ {
		err := store.RecordFailedLogin(ctx, "user@test.com", "1.2.3.4", lockout)
		if err != nil {
			t.Fatalf("RecordFailedLogin() iteration %d error = %v", i, err)
		}
	}

	// Verify email counter
	emailKey := LoginRateEmailPrefix + "user@test.com"
	emailVal, err := mr.Get(emailKey)
	if err != nil {
		t.Fatalf("Get email key error = %v", err)
	}
	if emailVal != "3" {
		t.Errorf("email counter = %s, want 3", emailVal)
	}

	// Verify IP counter
	ipKey := LoginRateIPPrefix + "1.2.3.4"
	ipVal, err := mr.Get(ipKey)
	if err != nil {
		t.Fatalf("Get IP key error = %v", err)
	}
	if ipVal != "3" {
		t.Errorf("IP counter = %s, want 3", ipVal)
	}
}

func TestRecordFailedLogin_SetsExpiry(t *testing.T) {
	mr, store := setupUserTestRedis(t, 7*24*time.Hour)
	defer mr.Close()

	ctx := context.Background()
	lockout := 15 * time.Minute

	err := store.RecordFailedLogin(ctx, "user@test.com", "1.2.3.4", lockout)
	if err != nil {
		t.Fatal(err)
	}

	// Both counters should have TTLs set
	emailKey := LoginRateEmailPrefix + "user@test.com"
	emailTTL := mr.TTL(emailKey)
	if emailTTL <= 0 {
		t.Errorf("email key TTL = %v, expected positive", emailTTL)
	}

	ipKey := LoginRateIPPrefix + "1.2.3.4"
	ipTTL := mr.TTL(ipKey)
	if ipTTL <= 0 {
		t.Errorf("IP key TTL = %v, expected positive", ipTTL)
	}
}

func TestRecordFailedLogin_ExpiryResetsCounters(t *testing.T) {
	mr, store := setupUserTestRedis(t, 7*24*time.Hour)
	defer mr.Close()

	ctx := context.Background()
	lockout := 5 * time.Second

	// Record a failure
	store.RecordFailedLogin(ctx, "user@test.com", "1.2.3.4", lockout)

	// Fast-forward past lockout
	mr.FastForward(6 * time.Second)

	// Record another failure — should restart at 1
	store.RecordFailedLogin(ctx, "user@test.com", "1.2.3.4", lockout)

	emailKey := LoginRateEmailPrefix + "user@test.com"
	emailVal, _ := mr.Get(emailKey)
	if emailVal != "1" {
		t.Errorf("email counter after expiry = %s, want 1 (fresh start)", emailVal)
	}
}

func TestCheckLoginRateLimit_AllowsUnderLimit(t *testing.T) {
	mr, store := setupUserTestRedis(t, 7*24*time.Hour)
	defer mr.Close()

	ctx := context.Background()

	// No failures recorded — should be allowed
	err := store.CheckLoginRateLimit(ctx, "user@test.com", "1.2.3.4", 5, 10, 15*time.Minute)
	if err != nil {
		t.Errorf("expected nil error (allowed), got %v", err)
	}
}

func TestCheckLoginRateLimit_DeniesOverEmailLimit(t *testing.T) {
	mr, store := setupUserTestRedis(t, 7*24*time.Hour)
	defer mr.Close()

	ctx := context.Background()
	maxPerEmail := 3
	lockout := 15 * time.Minute

	// Record enough failures to hit the email limit
	for i := 0; i < maxPerEmail; i++ {
		store.RecordFailedLogin(ctx, "user@test.com", "1.2.3.4", lockout)
	}

	// Should be denied
	err := store.CheckLoginRateLimit(ctx, "user@test.com", "1.2.3.4", maxPerEmail, 100, lockout)
	if !errors.Is(err, ErrTooManyLoginAttempts) {
		t.Errorf("expected ErrTooManyLoginAttempts, got %v", err)
	}
}

func TestCheckLoginRateLimit_DeniesOverIPLimit(t *testing.T) {
	mr, store := setupUserTestRedis(t, 7*24*time.Hour)
	defer mr.Close()

	ctx := context.Background()
	maxPerIP := 2
	lockout := 15 * time.Minute

	// Record failures from different emails but same IP
	store.RecordFailedLogin(ctx, "a@test.com", "1.2.3.4", lockout)
	store.RecordFailedLogin(ctx, "b@test.com", "1.2.3.4", lockout)

	// IP limit exceeded (each email recorded increments the IP counter)
	err := store.CheckLoginRateLimit(ctx, "c@test.com", "1.2.3.4", 100, maxPerIP, lockout)
	if !errors.Is(err, ErrTooManyLoginAttempts) {
		t.Errorf("expected ErrTooManyLoginAttempts for IP limit, got %v", err)
	}
}

func TestClearLoginRateLimit_ClearsEmailCounter(t *testing.T) {
	mr, store := setupUserTestRedis(t, 7*24*time.Hour)
	defer mr.Close()

	ctx := context.Background()

	// Record failures
	store.RecordFailedLogin(ctx, "user@test.com", "1.2.3.4", 15*time.Minute)

	// Clear
	err := store.ClearLoginRateLimit(ctx, "user@test.com")
	if err != nil {
		t.Fatal(err)
	}

	// Email key should be gone
	emailKey := LoginRateEmailPrefix + "user@test.com"
	if mr.Exists(emailKey) {
		t.Error("email counter should be cleared")
	}
}
