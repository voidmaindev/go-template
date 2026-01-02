package middleware

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/voidmaindev/GoTemplate/internal/common"
)

// RateLimitConfig holds rate limiting configuration
type RateLimitConfig struct {
	// Max number of requests allowed in the window
	Max int
	// Window duration for rate limiting
	Window time.Duration
	// KeyGenerator generates a unique key for each client (default: IP-based)
	KeyGenerator func(*fiber.Ctx) string
}

// DefaultRateLimitConfig returns the default rate limit configuration
func DefaultRateLimitConfig() *RateLimitConfig {
	return &RateLimitConfig{
		Max:    5,
		Window: time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},
	}
}

// AuthRateLimitConfig returns a stricter rate limit for auth endpoints
func AuthRateLimitConfig() *RateLimitConfig {
	return &RateLimitConfig{
		Max:    5,               // 5 attempts
		Window: time.Minute,     // per minute
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},
	}
}

// RateLimitMiddleware creates a rate limiting middleware with the given config
func RateLimitMiddleware(cfg *RateLimitConfig) fiber.Handler {
	if cfg == nil {
		cfg = DefaultRateLimitConfig()
	}

	return limiter.New(limiter.Config{
		Max:        cfg.Max,
		Expiration: cfg.Window,
		KeyGenerator: func(c *fiber.Ctx) string {
			if cfg.KeyGenerator != nil {
				return cfg.KeyGenerator(c)
			}
			return c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(common.Response{
				Success: false,
				Error:   "too many requests, please try again later",
			})
		},
		SkipFailedRequests:     false,
		SkipSuccessfulRequests: false,
	})
}

// AuthRateLimiter creates a rate limiter specifically for authentication endpoints
// More strict: 5 requests per minute per IP
func AuthRateLimiter() fiber.Handler {
	return RateLimitMiddleware(AuthRateLimitConfig())
}

// GeneralRateLimiter creates a rate limiter for general API endpoints
// Less strict: 100 requests per minute per IP
func GeneralRateLimiter() fiber.Handler {
	return RateLimitMiddleware(&RateLimitConfig{
		Max:    100,
		Window: time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},
	})
}
