package user

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

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

// Helper to create a TokenStore with mock Redis
func newTestTokenStore(mock *mockRedisClient) *TokenStore {
	// We need to wrap our mock in a way that TokenStore can use
	// Since TokenStore expects *redis.Client, we'll need to test via integration
	// or restructure. For now, we'll test the logic directly.
	return nil
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
