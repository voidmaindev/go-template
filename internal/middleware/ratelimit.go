package middleware

import (
	"context"
	"log/slog"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/voidmaindev/go-template/internal/common"
	"github.com/voidmaindev/go-template/internal/config"
	"github.com/voidmaindev/go-template/internal/redis"
)

const (
	// RateLimitHeaderLimit is the header for the rate limit maximum.
	RateLimitHeaderLimit = "X-RateLimit-Limit"
	// RateLimitHeaderRemaining is the header for remaining requests.
	RateLimitHeaderRemaining = "X-RateLimit-Remaining"
	// RateLimitHeaderReset is the header for when the rate limit resets.
	RateLimitHeaderReset = "X-RateLimit-Reset"
	// RetryAfterHeader is sent on 429 responses.
	RetryAfterHeader = "Retry-After"
)

// Rate limit tiers
const (
	TierAuth      = "auth"       // Public auth endpoints (login, register, refresh)
	TierAuthUser  = "auth_user"  // Authenticated auth ops (logout, password change)
	TierRBACAdmin = "rbac_admin" // RBAC admin operations
	TierAPIWrite  = "api_write"  // POST, PUT, DELETE operations
	TierAPIRead   = "api_read"   // GET operations
	TierGlobal    = "global"     // Fallback catch-all
)

// RateLimiterFactoryKey is the container key for the rate limiter factory
const RateLimiterFactoryKey = "middleware.rateLimiterFactory"

// userIDKey is the key used to store user ID in Fiber locals (set by JWT middleware)
const userIDKey = "user_id"

// RateLimiterFactory creates rate limiters with shared Redis connection and config
type RateLimiterFactory struct {
	redis  *redis.Client
	cfg    *config.RateLimitConfig
	window time.Duration
}

// NewRateLimiterFactory creates a new rate limiter factory
func NewRateLimiterFactory(redisClient *redis.Client, cfg *config.RateLimitConfig) *RateLimiterFactory {
	return &RateLimiterFactory{
		redis:  redisClient,
		cfg:    cfg,
		window: time.Duration(cfg.WindowSeconds) * time.Second,
	}
}

// ForTier creates a rate limiter middleware for the specified tier
func (f *RateLimiterFactory) ForTier(tier string) fiber.Handler {
	if !f.cfg.Enabled {
		return func(c *fiber.Ctx) error {
			return c.Next()
		}
	}

	limit := f.getLimit(tier)

	return func(c *fiber.Ctx) error {
		key := f.generateKey(c, tier)

		ctx, cancel := context.WithTimeout(c.Context(), 100*time.Millisecond)
		defer cancel()

		result, err := f.redis.RateLimitCheck(ctx, key, limit, f.window)
		if err != nil {
			// Fail open - allow request but log warning
			slog.Warn("rate limit check failed, allowing request",
				"error", err,
				"tier", tier,
				"key", key,
			)
			return c.Next()
		}

		// Set rate limit headers
		c.Set(RateLimitHeaderLimit, strconv.Itoa(limit))
		c.Set(RateLimitHeaderRemaining, strconv.Itoa(result.Remaining))
		c.Set(RateLimitHeaderReset, strconv.FormatInt(result.ResetAt, 10))

		if !result.Allowed {
			retryAfter := result.ResetAt - time.Now().Unix()
			if retryAfter < 1 {
				retryAfter = 1
			}
			c.Set(RetryAfterHeader, strconv.FormatInt(retryAfter, 10))

			return c.Status(fiber.StatusTooManyRequests).JSON(common.Response{
				Success: false,
				Error:   "too many requests, please try again later",
			})
		}

		return c.Next()
	}
}

// getLimit returns the rate limit for a tier
func (f *RateLimiterFactory) getLimit(tier string) int {
	switch tier {
	case TierAuth:
		return f.cfg.AuthLimit
	case TierAuthUser:
		return f.cfg.AuthUserLimit
	case TierRBACAdmin:
		return f.cfg.RBACAdminLimit
	case TierAPIWrite:
		return f.cfg.APIWriteLimit
	case TierAPIRead:
		return f.cfg.APIReadLimit
	case TierGlobal:
		return f.cfg.GlobalLimit
	default:
		return f.cfg.GlobalLimit
	}
}

// generateKey creates a rate limit key based on tier and request context.
// Public endpoints: ratelimit:{tier}:{ip}
// Authenticated:    ratelimit:{tier}:{user_id}:{ip}
func (f *RateLimiterFactory) generateKey(c *fiber.Ctx, tier string) string {
	ip := getClientIP(c)

	// For public auth endpoints, use IP only
	if tier == TierAuth {
		return "ratelimit:" + tier + ":" + ip
	}

	// For authenticated endpoints, include user ID if available
	userID := c.Locals(userIDKey)
	if userID != nil {
		if uid, ok := userID.(uint); ok && uid > 0 {
			return "ratelimit:" + tier + ":" + strconv.FormatUint(uint64(uid), 10) + ":" + ip
		}
	}

	// Fallback to IP only
	return "ratelimit:" + tier + ":" + ip
}

// getClientIP returns the client IP address.
// Fiber's c.IP() already handles proxy headers when configured.
func getClientIP(c *fiber.Ctx) string {
	return c.IP()
}

// AuthRateLimiter creates a rate limiter for public auth endpoints (login, register, refresh).
// This is a convenience function for backward compatibility.
// Deprecated: Use RateLimiterFactory.ForTier(TierAuth) instead.
func AuthRateLimiter() fiber.Handler {
	// Return a passthrough if no factory is configured
	// The factory-based approach should be used instead
	return func(c *fiber.Ctx) error {
		slog.Warn("AuthRateLimiter called without factory, rate limiting disabled")
		return c.Next()
	}
}

// GeneralRateLimiter creates a rate limiter for general API endpoints.
// Deprecated: Use RateLimiterFactory.ForTier(TierAPIRead) or ForTier(TierAPIWrite) instead.
func GeneralRateLimiter() fiber.Handler {
	return func(c *fiber.Ctx) error {
		slog.Warn("GeneralRateLimiter called without factory, rate limiting disabled")
		return c.Next()
	}
}
