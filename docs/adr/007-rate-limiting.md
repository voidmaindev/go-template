# ADR-007: Rate Limiting Strategy

## Status

Accepted

## Context

The application needs protection against abuse and DoS attacks. Different endpoints have different risk profiles:
- Authentication endpoints (login, register) are high-risk for brute force attacks
- Write operations (POST, PUT, DELETE) consume more resources
- Read operations (GET) are generally less expensive but still need limits

We need a rate limiting solution that:
- Protects against abuse without impacting legitimate users
- Scales horizontally with multiple API instances
- Provides different limits for different operation types
- Supports both anonymous and authenticated users

## Decision

### 6-Tier Rate Limiting Architecture

Implement a distributed rate limiting system with 6 tiers, each with configurable limits:

| Tier | Purpose | Key Format | Default Limit |
|------|---------|------------|---------------|
| `auth` | Public auth endpoints (login, register, refresh) | `ratelimit:auth:{ip}` | 10/min |
| `auth_user` | Authenticated auth operations (logout, password change) | `ratelimit:auth_user:{user_id}:{ip}` | 30/min |
| `rbac_admin` | RBAC administration operations | `ratelimit:rbac_admin:{user_id}:{ip}` | 60/min |
| `api_write` | POST, PUT, DELETE operations | `ratelimit:api_write:{user_id}:{ip}` | 120/min |
| `api_read` | GET operations | `ratelimit:api_read:{user_id}:{ip}` | 300/min |
| `global` | Fallback catch-all | `ratelimit:global:{ip}` | 1000/min |

### Redis Sliding Window Algorithm

Use Redis sorted sets for accurate sliding window rate limiting:

```go
// Lua script for atomic rate limiting
luaScript := `
    local key = KEYS[1]
    local limit = tonumber(ARGV[1])
    local window = tonumber(ARGV[2])
    local now = tonumber(ARGV[3])
    local windowStart = now - window

    -- Remove old entries outside the window
    redis.call('ZREMRANGEBYSCORE', key, '-inf', windowStart)

    -- Count current entries
    local count = redis.call('ZCARD', key)

    if count < limit then
        -- Add new entry with current timestamp
        redis.call('ZADD', key, now, now .. ':' .. math.random())
        redis.call('EXPIRE', key, window)
        return {1, limit - count - 1, now + window}  -- allowed, remaining, reset
    else
        return {0, 0, now + window}  -- denied, remaining, reset
    end
`
```

**Why Sliding Window?**
- More accurate than fixed window (no burst at window boundaries)
- More efficient than leaky bucket (no need to track per-request timing)
- Natural fit for Redis sorted sets

### RateLimiterFactory Pattern

A factory creates tier-specific middleware:

```go
type RateLimiterFactory struct {
    redis  *redis.Client
    config *config.RateLimitConfig
    window time.Duration
}

func (f *RateLimiterFactory) ForTier(tier string) fiber.Handler {
    return func(c *fiber.Ctx) error {
        key := f.buildKey(tier, c)
        limit := f.getLimit(tier)

        result, err := f.redis.RateLimitCheck(ctx, key, limit, f.window)
        if err != nil {
            // Fail-open: allow request if Redis check fails
            slog.Warn("rate limit check failed", "error", err)
            return c.Next()
        }

        // Set rate limit headers
        c.Set("X-RateLimit-Limit", strconv.Itoa(limit))
        c.Set("X-RateLimit-Remaining", strconv.Itoa(result.Remaining))
        c.Set("X-RateLimit-Reset", strconv.FormatInt(result.ResetAt, 10))

        if !result.Allowed {
            c.Set("Retry-After", strconv.FormatInt(result.ResetAt-time.Now().Unix(), 10))
            return c.Status(429).JSON(fiber.Map{
                "error": "rate limit exceeded",
                "retry_after": result.ResetAt - time.Now().Unix(),
            })
        }

        return c.Next()
    }
}
```

### Key Construction

- **Public endpoints**: Key by IP only (`ratelimit:{tier}:{ip}`)
- **Authenticated endpoints**: Key by user ID + IP (`ratelimit:{tier}:{user_id}:{ip}`)

Using both user ID and IP for authenticated endpoints:
- Prevents a compromised account from being used to attack specific endpoints
- Allows different users on the same IP (corporate NAT) to have separate limits

### Response Headers

Standard rate limit headers are included in all responses:

| Header | Description |
|--------|-------------|
| `X-RateLimit-Limit` | Maximum requests allowed in the window |
| `X-RateLimit-Remaining` | Requests remaining in current window |
| `X-RateLimit-Reset` | Unix timestamp when the window resets |
| `Retry-After` | Seconds until retry (only on 429) |

### Fail-Open Strategy

When Redis is unavailable:
1. Log a warning with the error
2. Allow the request to proceed
3. Continue normal operation

**Rationale**: Availability is preferred over strict rate limiting. Brief Redis outages shouldn't cause service disruptions. The fail-open window is typically short, and other security measures (authentication, RBAC) remain active.

## Configuration

```yaml
# Rate limit configuration (per minute)
RATE_LIMIT_AUTH=10
RATE_LIMIT_AUTH_USER=30
RATE_LIMIT_RBAC_ADMIN=60
RATE_LIMIT_API_WRITE=120
RATE_LIMIT_API_READ=300
RATE_LIMIT_GLOBAL=1000
RATE_LIMIT_WINDOW_SECONDS=60
```

## Usage

```go
// In domain routes
func (d *domain) Routes(api fiber.Router, c *container.Container) {
    handler := HandlerKey.MustGet(c)
    rateLimiter := middleware.RateLimiterFactoryKey.MustGet(c)

    products := api.Group("/products")

    // Read endpoints use api_read tier
    products.Get("/", rateLimiter.ForTier(middleware.TierAPIRead), handler.List)
    products.Get("/:id", rateLimiter.ForTier(middleware.TierAPIRead), handler.GetByID)

    // Write endpoints use api_write tier
    products.Post("/", rateLimiter.ForTier(middleware.TierAPIWrite), handler.Create)
    products.Put("/:id", rateLimiter.ForTier(middleware.TierAPIWrite), handler.Update)
    products.Delete("/:id", rateLimiter.ForTier(middleware.TierAPIWrite), handler.Delete)
}
```

## Consequences

### Positive

- **Abuse Protection**: Effective against brute force and DoS attacks
- **Granular Control**: Different limits for different operation types
- **Horizontal Scaling**: Redis-backed, works across multiple instances
- **Transparent**: Standard headers inform clients of limits
- **Resilient**: Fail-open prevents rate limiting from causing outages

### Negative

- **Redis Dependency**: Requires Redis for distributed rate limiting
- **Memory Usage**: Sorted sets consume more memory than simple counters
- **Complexity**: 6 tiers add configuration complexity

### Neutral

- **IP Spoofing**: Attackers can rotate IPs, but this raises the cost of attacks
- **False Positives**: Shared IPs (corporate NAT) may hit limits faster

## Related

- ADR-005: Type-Safe Dependency Injection (factory pattern)
- ADR-006: Self-Registration and OAuth (auth endpoint protection)
