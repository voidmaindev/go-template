package middleware

import (
	"context"
	"log/slog"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/voidmaindev/go-template/internal/common"
	"github.com/voidmaindev/go-template/internal/config"
	"github.com/voidmaindev/go-template/internal/container"
	"github.com/voidmaindev/go-template/internal/redis"
	"github.com/voidmaindev/go-template/internal/telemetry"
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

// Tier is a typed rate limit tier to prevent typos at compile time.
type Tier string

// Rate limit tiers
const (
	TierAuth      Tier = "auth"       // Public auth endpoints (login, register, refresh)
	TierAuthUser  Tier = "auth_user"  // Authenticated auth ops (logout, password change)
	TierRBACAdmin Tier = "rbac_admin" // RBAC admin operations
	TierAPIWrite  Tier = "api_write"  // POST, PUT, DELETE operations
	TierAPIRead   Tier = "api_read"   // GET operations
	TierGlobal    Tier = "global"     // Fallback catch-all
)

// RateLimiterFactoryKey is the typed container key for the rate limiter factory
var RateLimiterFactoryKey = container.Key[*RateLimiterFactory]("middleware.rateLimiterFactory")

// contextKey is a typed key for context/locals values to avoid collisions
type contextKey string

// UserIDKey is the typed key for storing user ID in Fiber locals (set by JWT middleware)
const UserIDKey contextKey = "user_id"

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
func (f *RateLimiterFactory) ForTier(tier Tier) fiber.Handler {
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

			// Record rate limit hit metric
			telemetry.IncrementRateLimitHits(string(tier))

			return c.Status(fiber.StatusTooManyRequests).JSON(common.Response{
				Success: false,
				Error:   "too many requests, please try again later",
			})
		}

		return c.Next()
	}
}

// getLimit returns the rate limit for a tier
func (f *RateLimiterFactory) getLimit(tier Tier) int {
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
func (f *RateLimiterFactory) generateKey(c *fiber.Ctx, tier Tier) string {
	ip := getClientIP(c)
	tierStr := string(tier)

	// For public auth endpoints, use IP only
	if tier == TierAuth {
		return "ratelimit:" + tierStr + ":" + ip
	}

	// For authenticated endpoints, include user ID if available
	userID := c.Locals(UserIDKey)
	if userID != nil {
		if uid, ok := userID.(uint); ok && uid > 0 {
			return "ratelimit:" + tierStr + ":" + strconv.FormatUint(uint64(uid), 10) + ":" + ip
		}
	}

	// Fallback to IP only
	return "ratelimit:" + tierStr + ":" + ip
}

// getClientIP returns the client IP address.
// Fiber's c.IP() already handles proxy headers when configured.
func getClientIP(c *fiber.Ctx) string {
	return c.IP()
}
