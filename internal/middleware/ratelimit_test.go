package middleware

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
)

func TestDefaultRateLimitConfig(t *testing.T) {
	cfg := DefaultRateLimitConfig()

	if cfg.Max != 5 {
		t.Errorf("Max = %d, want 5", cfg.Max)
	}
	if cfg.Window != time.Minute {
		t.Errorf("Window = %v, want %v", cfg.Window, time.Minute)
	}
	if cfg.KeyGenerator == nil {
		t.Error("KeyGenerator should not be nil")
	}
}

func TestAuthRateLimitConfig(t *testing.T) {
	cfg := AuthRateLimitConfig()

	if cfg.Max != 5 {
		t.Errorf("Max = %d, want 5", cfg.Max)
	}
	if cfg.Window != time.Minute {
		t.Errorf("Window = %v, want %v", cfg.Window, time.Minute)
	}
}

func TestRateLimitMiddleware_AllowsRequests(t *testing.T) {
	app := fiber.New()

	app.Use(RateLimitMiddleware(&RateLimitConfig{
		Max:    5,
		Window: time.Minute,
	}))

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
}

func TestRateLimitMiddleware_BlocksAfterLimit(t *testing.T) {
	app := fiber.New()

	// Very strict limit for testing
	app.Use(RateLimitMiddleware(&RateLimitConfig{
		Max:    2,
		Window: time.Minute,
	}))

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

	// Check response body
	body, _ := io.ReadAll(resp.Body)
	if len(body) == 0 {
		t.Error("Response body should not be empty for rate limit error")
	}
}

func TestRateLimitMiddleware_NilConfig(t *testing.T) {
	app := fiber.New()

	// Should not panic with nil config
	app.Use(RateLimitMiddleware(nil))

	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusOK)
	}
}

func TestAuthRateLimiter(t *testing.T) {
	app := fiber.New()

	app.Use(AuthRateLimiter())

	app.Post("/login", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	// Should allow first request
	req := httptest.NewRequest("POST", "/login", nil)
	resp, _ := app.Test(req)

	if resp.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusOK)
	}
}

func TestGeneralRateLimiter(t *testing.T) {
	app := fiber.New()

	app.Use(GeneralRateLimiter())

	app.Get("/api/data", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	// General limiter allows 100 req/min, so first request should pass
	req := httptest.NewRequest("GET", "/api/data", nil)
	resp, _ := app.Test(req)

	if resp.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusOK)
	}
}

func TestRateLimitMiddleware_KeyGeneratorIsCalled(t *testing.T) {
	// Test that custom key generator is called
	keyGeneratorCalled := false

	app := fiber.New()

	app.Use(RateLimitMiddleware(&RateLimitConfig{
		Max:    10,
		Window: time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string {
			keyGeneratorCalled = true
			return c.Get("X-Client-ID", "default")
		},
	}))

	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Client-ID", "test-client")
	resp, _ := app.Test(req)

	if resp.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	if !keyGeneratorCalled {
		t.Error("KeyGenerator should have been called")
	}
}

func TestRateLimitConfig_CustomKeyGenerator(t *testing.T) {
	// Test that custom key generator works for rate limiting
	// Note: Testing key isolation would require integration tests with separate clients
	// Fiber's test client shares state, so we only test that the same key gets blocked

	app := fiber.New()

	app.Use(RateLimitMiddleware(&RateLimitConfig{
		Max:    2,
		Window: time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.Get("X-API-Key", "anonymous")
		},
	}))

	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	// Requests with API key 1
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("X-API-Key", "key1")
		resp, _ := app.Test(req)
		if resp.StatusCode != http.StatusOK {
			t.Errorf("key1 request %d: StatusCode = %d, want %d", i+1, resp.StatusCode, http.StatusOK)
		}
	}

	// Third request with key1 should be blocked
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-API-Key", "key1")
	resp, _ := app.Test(req)

	if resp.StatusCode != http.StatusTooManyRequests {
		t.Errorf("key1 third request: StatusCode = %d, want %d", resp.StatusCode, http.StatusTooManyRequests)
	}
}
