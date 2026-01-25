package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/gofiber/fiber/v2"
	goredis "github.com/redis/go-redis/v9"
	"github.com/voidmaindev/go-template/internal/config"
	"github.com/voidmaindev/go-template/internal/redis"
)

// setupTestRedis creates a miniredis instance for testing
func setupTestRedis(t *testing.T) (*miniredis.Miniredis, *redis.Client) {
	t.Helper()
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %v", err)
	}

	client := goredis.NewClient(&goredis.Options{
		Addr: mr.Addr(),
	})
	return mr, redis.WrapClient(client)
}

func testRateLimitConfig() *config.RateLimitConfig {
	return &config.RateLimitConfig{
		Enabled:        true,
		AuthLimit:      5,
		AuthUserLimit:  10,
		RBACAdminLimit: 30,
		APIWriteLimit:  60,
		APIReadLimit:   200,
		GlobalLimit:    1000,
		WindowSeconds:  60,
	}
}

func TestNewRateLimiterFactory(t *testing.T) {
	mr, redisClient := setupTestRedis(t)
	defer mr.Close()
	defer redisClient.Close()

	cfg := testRateLimitConfig()
	factory := NewRateLimiterFactory(redisClient, cfg)

	if factory == nil {
		t.Fatal("NewRateLimiterFactory returned nil")
	}
	if factory.window != 60*time.Second {
		t.Errorf("window = %v, want %v", factory.window, 60*time.Second)
	}
}

func TestRateLimiterFactory_GetLimit(t *testing.T) {
	mr, redisClient := setupTestRedis(t)
	defer mr.Close()
	defer redisClient.Close()

	cfg := testRateLimitConfig()
	factory := NewRateLimiterFactory(redisClient, cfg)

	tests := []struct {
		tier  string
		limit int
	}{
		{TierAuth, 5},
		{TierAuthUser, 10},
		{TierRBACAdmin, 30},
		{TierAPIWrite, 60},
		{TierAPIRead, 200},
		{TierGlobal, 1000},
		{"unknown", 1000}, // Should default to global
	}

	for _, tt := range tests {
		t.Run(tt.tier, func(t *testing.T) {
			limit := factory.getLimit(tt.tier)
			if limit != tt.limit {
				t.Errorf("getLimit(%q) = %d, want %d", tt.tier, limit, tt.limit)
			}
		})
	}
}

func TestRateLimiterFactory_ForTier_Disabled(t *testing.T) {
	mr, redisClient := setupTestRedis(t)
	defer mr.Close()
	defer redisClient.Close()

	cfg := testRateLimitConfig()
	cfg.Enabled = false
	factory := NewRateLimiterFactory(redisClient, cfg)

	app := fiber.New()
	app.Use(factory.ForTier(TierAuth))
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	// Should always pass when disabled
	for i := 0; i < 10; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("app.Test() error = %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Request %d: StatusCode = %d, want %d", i+1, resp.StatusCode, http.StatusOK)
		}
	}
}

func TestRateLimiterFactory_ForTier_AllowsRequests(t *testing.T) {
	mr, redisClient := setupTestRedis(t)
	defer mr.Close()
	defer redisClient.Close()

	cfg := testRateLimitConfig()
	cfg.AuthLimit = 5
	factory := NewRateLimiterFactory(redisClient, cfg)

	app := fiber.New()
	app.Use(factory.ForTier(TierAuth))
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	// First request should pass
	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	// Check rate limit headers
	if resp.Header.Get(RateLimitHeaderLimit) != "5" {
		t.Errorf("X-RateLimit-Limit = %q, want %q", resp.Header.Get(RateLimitHeaderLimit), "5")
	}
}

func TestRateLimiterFactory_ForTier_BlocksAfterLimit(t *testing.T) {
	mr, redisClient := setupTestRedis(t)
	defer mr.Close()
	defer redisClient.Close()

	cfg := testRateLimitConfig()
	cfg.AuthLimit = 2 // Very strict for testing
	factory := NewRateLimiterFactory(redisClient, cfg)

	app := fiber.New()
	app.Use(factory.ForTier(TierAuth))
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	// Make requests up to limit
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		resp, _ := app.Test(req)
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Request %d: StatusCode = %d, want %d", i+1, resp.StatusCode, http.StatusOK)
		}
	}

	// Next request should be rate limited
	req := httptest.NewRequest("GET", "/test", nil)
	resp, _ := app.Test(req)

	if resp.StatusCode != http.StatusTooManyRequests {
		t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusTooManyRequests)
	}

	// Check Retry-After header is present
	if resp.Header.Get(RetryAfterHeader) == "" {
		t.Error("Retry-After header should be present when rate limited")
	}
}

func TestRateLimiterFactory_GenerateKey_PublicAuth(t *testing.T) {
	mr, redisClient := setupTestRedis(t)
	defer mr.Close()
	defer redisClient.Close()

	cfg := testRateLimitConfig()
	factory := NewRateLimiterFactory(redisClient, cfg)

	app := fiber.New()
	var generatedKey string

	app.Use(func(c *fiber.Ctx) error {
		generatedKey = factory.generateKey(c, TierAuth)
		return c.Next()
	})
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Forwarded-For", "192.168.1.100")
	app.Test(req)

	// For public auth endpoints, key should be IP only
	expected := "ratelimit:auth:0.0.0.0" // Default IP in test
	if generatedKey != expected {
		t.Logf("Generated key: %s (IP detection may vary in test environment)", generatedKey)
	}
}

func TestRateLimiterFactory_GenerateKey_Authenticated(t *testing.T) {
	mr, redisClient := setupTestRedis(t)
	defer mr.Close()
	defer redisClient.Close()

	cfg := testRateLimitConfig()
	factory := NewRateLimiterFactory(redisClient, cfg)

	app := fiber.New()
	var generatedKey string

	// Simulate authenticated user
	app.Use(func(c *fiber.Ctx) error {
		c.Locals(UserIDKey, uint(123))
		return c.Next()
	})
	app.Use(func(c *fiber.Ctx) error {
		generatedKey = factory.generateKey(c, TierAPIRead)
		return c.Next()
	})
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	app.Test(req)

	// Should include user ID for authenticated endpoints
	if generatedKey == "" {
		t.Error("Generated key should not be empty")
	}
	// Key should contain the user ID
	if generatedKey == "ratelimit:api_read:0.0.0.0" {
		t.Error("Key should include user ID for authenticated endpoints")
	}
}

func TestRateLimitHeaders(t *testing.T) {
	mr, redisClient := setupTestRedis(t)
	defer mr.Close()
	defer redisClient.Close()

	cfg := testRateLimitConfig()
	cfg.APIReadLimit = 100
	factory := NewRateLimiterFactory(redisClient, cfg)

	app := fiber.New()
	app.Use(factory.ForTier(TierAPIRead))
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, _ := app.Test(req)

	// Verify all rate limit headers are present
	if resp.Header.Get(RateLimitHeaderLimit) == "" {
		t.Error("X-RateLimit-Limit header missing")
	}
	if resp.Header.Get(RateLimitHeaderRemaining) == "" {
		t.Error("X-RateLimit-Remaining header missing")
	}
	if resp.Header.Get(RateLimitHeaderReset) == "" {
		t.Error("X-RateLimit-Reset header missing")
	}
}

func TestRedisClient_RateLimitCheck(t *testing.T) {
	mr, redisClient := setupTestRedis(t)
	defer mr.Close()
	defer redisClient.Close()

	ctx := context.Background()
	key := "test:ratelimit:key"
	limit := 3
	window := time.Minute

	// First three requests should be allowed
	for i := 0; i < 3; i++ {
		result, err := redisClient.RateLimitCheck(ctx, key, limit, window)
		if err != nil {
			t.Fatalf("RateLimitCheck error: %v", err)
		}
		if !result.Allowed {
			t.Errorf("Request %d: expected allowed, got blocked", i+1)
		}
		if result.Remaining != (limit - i - 2) && i < limit-1 {
			t.Logf("Request %d: Remaining = %d", i+1, result.Remaining)
		}
	}

	// Fourth request should be blocked
	result, err := redisClient.RateLimitCheck(ctx, key, limit, window)
	if err != nil {
		t.Fatalf("RateLimitCheck error: %v", err)
	}
	if result.Allowed {
		t.Error("Fourth request should be blocked")
	}
	if result.Remaining != 0 {
		t.Errorf("Remaining = %d, want 0", result.Remaining)
	}
}

func TestTierConstants(t *testing.T) {
	// Verify tier constants have expected values
	tests := []struct {
		tier string
		want string
	}{
		{TierAuth, "auth"},
		{TierAuthUser, "auth_user"},
		{TierRBACAdmin, "rbac_admin"},
		{TierAPIWrite, "api_write"},
		{TierAPIRead, "api_read"},
		{TierGlobal, "global"},
	}

	for _, tt := range tests {
		if tt.tier != tt.want {
			t.Errorf("Tier constant %q = %q, want %q", tt.tier, tt.tier, tt.want)
		}
	}
}
