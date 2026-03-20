package auth

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	goredis "github.com/redis/go-redis/v9"
	"github.com/voidmaindev/go-template/internal/redis"
)

func setupAuthTestRedis(t *testing.T) (*miniredis.Miniredis, *TokenStore) {
	t.Helper()
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %v", err)
	}
	client := goredis.NewClient(&goredis.Options{Addr: mr.Addr()})
	store := NewTokenStore(redis.WrapClient(client))
	return mr, store
}

// ================================
// CheckRateLimit tests
// ================================

func TestCheckRateLimit_AllowsUnderLimit(t *testing.T) {
	mr, store := setupAuthTestRedis(t)
	defer mr.Close()

	ctx := context.Background()
	limit := 5
	window := 60 * time.Second

	for i := 1; i <= limit; i++ {
		allowed, err := store.CheckRateLimit(ctx, "user@test.com", "login", limit, window)
		if err != nil {
			t.Fatalf("CheckRateLimit() iteration %d error = %v", i, err)
		}
		if !allowed {
			t.Errorf("iteration %d: should be allowed (count=%d, limit=%d)", i, i, limit)
		}
	}
}

func TestCheckRateLimit_DeniesOverLimit(t *testing.T) {
	mr, store := setupAuthTestRedis(t)
	defer mr.Close()

	ctx := context.Background()
	limit := 3
	window := 60 * time.Second

	// Use up the limit
	for i := 0; i < limit; i++ {
		store.CheckRateLimit(ctx, "user@test.com", "login", limit, window)
	}

	// Next request should be denied
	allowed, err := store.CheckRateLimit(ctx, "user@test.com", "login", limit, window)
	if err != nil {
		t.Fatalf("CheckRateLimit() error = %v", err)
	}
	if allowed {
		t.Error("should be denied over limit")
	}
}

func TestCheckRateLimit_SetsExpiry(t *testing.T) {
	mr, store := setupAuthTestRedis(t)
	defer mr.Close()

	ctx := context.Background()
	window := 30 * time.Second

	_, err := store.CheckRateLimit(ctx, "user@test.com", "verify", 10, window)
	if err != nil {
		t.Fatalf("CheckRateLimit() error = %v", err)
	}

	// Verify key has a TTL (not persisting forever)
	key := keyPrefixRateLimit + "verify:user@test.com"
	ttl := mr.TTL(key)
	if ttl <= 0 {
		t.Errorf("TTL = %v, expected positive (key should expire)", ttl)
	}
	if ttl > window {
		t.Errorf("TTL = %v, should be <= %v", ttl, window)
	}
}

func TestCheckRateLimit_ExpiryResetsCounter(t *testing.T) {
	mr, store := setupAuthTestRedis(t)
	defer mr.Close()

	ctx := context.Background()
	limit := 2
	window := 5 * time.Second

	// Use up the limit
	store.CheckRateLimit(ctx, "user@test.com", "login", limit, window)
	store.CheckRateLimit(ctx, "user@test.com", "login", limit, window)

	// Should be denied
	allowed, _ := store.CheckRateLimit(ctx, "user@test.com", "login", limit, window)
	if allowed {
		t.Error("should be denied after limit reached")
	}

	// Fast-forward past window
	mr.FastForward(6 * time.Second)

	// Should be allowed again
	allowed, err := store.CheckRateLimit(ctx, "user@test.com", "login", limit, window)
	if err != nil {
		t.Fatalf("CheckRateLimit() after expiry error = %v", err)
	}
	if !allowed {
		t.Error("should be allowed after window expires")
	}
}

func TestCheckRateLimit_DifferentIdentifiers_Independent(t *testing.T) {
	mr, store := setupAuthTestRedis(t)
	defer mr.Close()

	ctx := context.Background()
	limit := 1
	window := 60 * time.Second

	// Use up limit for user A
	store.CheckRateLimit(ctx, "userA@test.com", "login", limit, window)

	// User B should still be allowed
	allowed, err := store.CheckRateLimit(ctx, "userB@test.com", "login", limit, window)
	if err != nil {
		t.Fatal(err)
	}
	if !allowed {
		t.Error("user B should be allowed (independent of user A)")
	}
}

func TestCheckRateLimit_DifferentActions_Independent(t *testing.T) {
	mr, store := setupAuthTestRedis(t)
	defer mr.Close()

	ctx := context.Background()
	limit := 1
	window := 60 * time.Second

	// Use up limit for "login" action
	store.CheckRateLimit(ctx, "user@test.com", "login", limit, window)

	// "verify" action should still be allowed
	allowed, err := store.CheckRateLimit(ctx, "user@test.com", "verify", limit, window)
	if err != nil {
		t.Fatal(err)
	}
	if !allowed {
		t.Error("different action should be allowed (independent)")
	}
}

func TestCheckRateLimit_RedisDown_ReturnsError(t *testing.T) {
	mr, store := setupAuthTestRedis(t)
	mr.Close() // Kill Redis

	ctx := context.Background()
	_, err := store.CheckRateLimit(ctx, "user@test.com", "login", 5, 60*time.Second)
	if err == nil {
		t.Error("expected error when Redis is down")
	}
}
