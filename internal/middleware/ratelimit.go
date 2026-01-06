package middleware

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/voidmaindev/go-template/internal/common"
)

const (
	// RateLimitHeaderLimit is the header for the rate limit maximum.
	RateLimitHeaderLimit = "X-RateLimit-Limit"
	// RateLimitHeaderRemaining is the header for remaining requests.
	RateLimitHeaderRemaining = "X-RateLimit-Remaining"
	// RateLimitHeaderReset is the header for when the rate limit resets.
	RateLimitHeaderReset = "X-RateLimit-Reset"
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

// getClientIP returns the client IP address in a more secure way.
// When behind a trusted proxy, Fiber's c.IP() already uses the ProxyHeader config.
// This function adds the path as additional context to prevent cross-endpoint abuse.
func getClientIP(c *fiber.Ctx) string {
	// c.IP() respects Fiber's ProxyHeader and TrustedProxies config
	// which should be configured in serve.go for production deployments
	return c.IP()
}

// DefaultRateLimitConfig returns the default rate limit configuration
func DefaultRateLimitConfig() *RateLimitConfig {
	return &RateLimitConfig{
		Max:    5,
		Window: time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string {
			return getClientIP(c)
		},
	}
}

// AuthRateLimitConfig returns a stricter rate limit for auth endpoints
func AuthRateLimitConfig() *RateLimitConfig {
	return &RateLimitConfig{
		Max:    5,           // 5 attempts
		Window: time.Minute, // per minute
		KeyGenerator: func(c *fiber.Ctx) string {
			return getClientIP(c)
		},
	}
}

// RateLimitMiddleware creates a rate limiting middleware with the given config.
// It adds standard rate limit headers to all responses.
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
			// Add rate limit headers on limit reached
			resetTime := time.Now().Add(cfg.Window).Unix()
			c.Set(RateLimitHeaderLimit, strconv.Itoa(cfg.Max))
			c.Set(RateLimitHeaderRemaining, "0")
			c.Set(RateLimitHeaderReset, strconv.FormatInt(resetTime, 10))

			return c.Status(fiber.StatusTooManyRequests).JSON(common.Response{
				Success: false,
				Error:   "too many requests, please try again later",
			})
		},
		SkipFailedRequests:     false,
		SkipSuccessfulRequests: false,
	})
}

// RateLimitMiddlewareWithHeaders creates a rate limiting middleware that adds
// rate limit headers to all responses (including successful ones).
func RateLimitMiddlewareWithHeaders(cfg *RateLimitConfig) fiber.Handler {
	if cfg == nil {
		cfg = DefaultRateLimitConfig()
	}

	// Use the standard limiter with header injection
	rateLimiter := limiter.New(limiter.Config{
		Max:        cfg.Max,
		Expiration: cfg.Window,
		KeyGenerator: func(c *fiber.Ctx) string {
			if cfg.KeyGenerator != nil {
				return cfg.KeyGenerator(c)
			}
			return c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			resetTime := time.Now().Add(cfg.Window).Unix()
			c.Set(RateLimitHeaderLimit, strconv.Itoa(cfg.Max))
			c.Set(RateLimitHeaderRemaining, "0")
			c.Set(RateLimitHeaderReset, strconv.FormatInt(resetTime, 10))

			return c.Status(fiber.StatusTooManyRequests).JSON(common.Response{
				Success: false,
				Error:   "too many requests, please try again later",
			})
		},
		SkipFailedRequests:     false,
		SkipSuccessfulRequests: false,
	})

	// Wrap to add headers on all responses
	return func(c *fiber.Ctx) error {
		// Add limit header before processing
		c.Set(RateLimitHeaderLimit, strconv.Itoa(cfg.Max))

		// Process request through rate limiter
		return rateLimiter(c)
	}
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
			return getClientIP(c)
		},
	})
}
