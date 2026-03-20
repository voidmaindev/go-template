package redis

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	goredis "github.com/redis/go-redis/v9"
)

func setupMiniredis(t *testing.T) (*miniredis.Miniredis, *Client) {
	t.Helper()
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %v", err)
	}
	client := goredis.NewClient(&goredis.Options{Addr: mr.Addr()})
	return mr, WrapClient(client)
}

// ================================
// IncrWithExpiry tests
// ================================

func TestIncrWithExpiry_FirstIncrement_SetsExpiry(t *testing.T) {
	mr, client := setupMiniredis(t)
	defer mr.Close()
	defer client.Close()

	ctx := context.Background()
	key := "test:counter"
	expiry := 10 * time.Second

	count, err := client.IncrWithExpiry(ctx, key, expiry)
	if err != nil {
		t.Fatalf("IncrWithExpiry() error = %v", err)
	}
	if count != 1 {
		t.Errorf("count = %d, want 1", count)
	}

	// Verify TTL was set
	ttl := mr.TTL(key)
	if ttl <= 0 {
		t.Errorf("TTL = %v, expected positive duration", ttl)
	}
	if ttl > expiry {
		t.Errorf("TTL = %v, should be <= %v", ttl, expiry)
	}
}

func TestIncrWithExpiry_SubsequentIncrements_PreserveExpiry(t *testing.T) {
	mr, client := setupMiniredis(t)
	defer mr.Close()
	defer client.Close()

	ctx := context.Background()
	key := "test:counter:multi"
	expiry := 30 * time.Second

	// First call — sets expiry
	_, err := client.IncrWithExpiry(ctx, key, expiry)
	if err != nil {
		t.Fatalf("first IncrWithExpiry() error = %v", err)
	}

	originalTTL := mr.TTL(key)

	// Fast-forward a bit
	mr.FastForward(2 * time.Second)

	// Second call — should NOT reset expiry
	count, err := client.IncrWithExpiry(ctx, key, expiry)
	if err != nil {
		t.Fatalf("second IncrWithExpiry() error = %v", err)
	}
	if count != 2 {
		t.Errorf("count = %d, want 2", count)
	}

	// TTL should have decreased (not been reset)
	newTTL := mr.TTL(key)
	if newTTL >= originalTTL {
		t.Errorf("TTL should have decreased: was %v, now %v", originalTTL, newTTL)
	}
}

func TestIncrWithExpiry_CountsCorrectly(t *testing.T) {
	mr, client := setupMiniredis(t)
	defer mr.Close()
	defer client.Close()

	ctx := context.Background()
	key := "test:counter:sequence"
	expiry := 10 * time.Second

	for i := 1; i <= 5; i++ {
		count, err := client.IncrWithExpiry(ctx, key, expiry)
		if err != nil {
			t.Fatalf("IncrWithExpiry() iteration %d error = %v", i, err)
		}
		if count != int64(i) {
			t.Errorf("iteration %d: count = %d, want %d", i, count, i)
		}
	}
}

func TestIncrWithExpiry_KeyExpires_ResetsCounter(t *testing.T) {
	mr, client := setupMiniredis(t)
	defer mr.Close()
	defer client.Close()

	ctx := context.Background()
	key := "test:counter:expiry"
	expiry := 5 * time.Second

	// Increment to 3
	for i := 0; i < 3; i++ {
		_, err := client.IncrWithExpiry(ctx, key, expiry)
		if err != nil {
			t.Fatal(err)
		}
	}

	// Fast-forward past expiry
	mr.FastForward(6 * time.Second)

	// Should start fresh at 1
	count, err := client.IncrWithExpiry(ctx, key, expiry)
	if err != nil {
		t.Fatalf("IncrWithExpiry() after expiry error = %v", err)
	}
	if count != 1 {
		t.Errorf("count after expiry = %d, want 1 (fresh start)", count)
	}

	// Should have a new TTL
	ttl := mr.TTL(key)
	if ttl <= 0 {
		t.Error("TTL should be set after fresh start")
	}
}

func TestIncrWithExpiry_DifferentKeys_Independent(t *testing.T) {
	mr, client := setupMiniredis(t)
	defer mr.Close()
	defer client.Close()

	ctx := context.Background()
	expiry := 10 * time.Second

	// Increment key A 3 times
	for i := 0; i < 3; i++ {
		client.IncrWithExpiry(ctx, "key:a", expiry)
	}

	// Increment key B once
	countB, _ := client.IncrWithExpiry(ctx, "key:b", expiry)

	if countB != 1 {
		t.Errorf("key B count = %d, want 1 (independent of key A)", countB)
	}
}

func TestIncrWithExpiry_ClosedConnection_ReturnsError(t *testing.T) {
	mr, client := setupMiniredis(t)
	mr.Close() // Close Redis before the call

	ctx := context.Background()
	_, err := client.IncrWithExpiry(ctx, "test:key", 10*time.Second)
	if err == nil {
		t.Error("expected error for closed connection")
	}
}

func TestIncrWithExpiry_CancelledContext(t *testing.T) {
	mr, client := setupMiniredis(t)
	defer mr.Close()
	defer client.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := client.IncrWithExpiry(ctx, "test:key", 10*time.Second)
	if err == nil {
		t.Error("expected error for cancelled context")
	}
}

func TestIncrWithExpiry_ConcurrentAccess(t *testing.T) {
	mr, client := setupMiniredis(t)
	defer mr.Close()
	defer client.Close()

	ctx := context.Background()
	key := "test:counter:concurrent"
	expiry := 30 * time.Second
	goroutines := 50

	errs := make(chan error, goroutines)
	counts := make(chan int64, goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			count, err := client.IncrWithExpiry(ctx, key, expiry)
			errs <- err
			counts <- count
		}()
	}

	for i := 0; i < goroutines; i++ {
		if err := <-errs; err != nil {
			t.Errorf("concurrent IncrWithExpiry() error: %v", err)
		}
	}

	// Final value should equal number of goroutines
	val, err := client.Get(ctx, key).Int64()
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if val != int64(goroutines) {
		t.Errorf("final count = %d, want %d", val, goroutines)
	}

	// TTL should be set
	ttl := mr.TTL(key)
	if ttl <= 0 {
		t.Error("TTL should be set after concurrent increments")
	}
}

func TestIncrWithExpiry_ShortExpiry(t *testing.T) {
	mr, client := setupMiniredis(t)
	defer mr.Close()
	defer client.Close()

	ctx := context.Background()
	key := "test:counter:shortttl"
	expiry := 100 * time.Millisecond

	count, err := client.IncrWithExpiry(ctx, key, expiry)
	if err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Errorf("count = %d, want 1", count)
	}

	// TTL should be set (miniredis uses PEXPIRE so millisecond precision)
	ttl := mr.TTL(key)
	if ttl <= 0 {
		t.Errorf("TTL = %v, expected positive", ttl)
	}
}
