# ADR-011: Token Invalidation Strategy

## Status

Accepted

## Context

JWT tokens are stateless by design, but certain scenarios require invalidation:
- User logout (invalidate specific token)
- Password change (invalidate all user tokens)
- Account compromise (immediate revocation of all access)
- Admin force-logout (security incident response)

The challenge: How to invalidate JWTs efficiently without checking a database on every request?

## Decision

### Dual Invalidation Strategy

Implement two complementary mechanisms:

1. **Token Blacklist**: For single-token invalidation (logout)
2. **Timestamp-Based Invalidation**: For bulk invalidation (password change, compromise)

### 1. Token Blacklist (Single Token)

Store individual invalidated tokens in Redis with TTL matching token expiry:

```go
// Key format: token:blacklist:{token_hash}
// Value: "1"
// TTL: matches token expiry time

func (s *TokenStore) BlacklistToken(ctx context.Context, token string, expiresAt time.Time) error {
    key := fmt.Sprintf("%s%s", TokenBlacklistPrefix, hashToken(token))
    ttl := time.Until(expiresAt)
    if ttl <= 0 {
        return nil  // Already expired, no need to blacklist
    }
    return s.redis.SetValue(ctx, key, "1", ttl)
}

func (s *TokenStore) IsBlacklisted(ctx context.Context, token string) (bool, error) {
    key := fmt.Sprintf("%s%s", TokenBlacklistPrefix, hashToken(token))
    exists, err := s.redis.Exists(ctx, key)
    return exists, err
}
```

**Why hash the token?**
- Tokens can be long (1KB+), hashes are fixed-size
- Tokens in Redis keys could be leaked via KEYS command
- SHA-256 hash is sufficient for unique identification

### 2. Timestamp-Based Invalidation (All User Tokens)

Store a Unix timestamp indicating when all user tokens were invalidated:

```go
// Key format: auth:token:invalidated:{user_id}
// Value: Unix timestamp
// TTL: None (persists until explicitly deleted)

func (s *TokenStore) InvalidateAllUserTokens(ctx context.Context, userID uint) error {
    key := fmt.Sprintf("%s%d", TokenInvalidatePrefix, userID)
    return s.redis.SetValue(ctx, key, time.Now().Unix(), 0)  // No expiry
}

func (s *TokenStore) GetTokensInvalidatedAt(ctx context.Context, userID uint) (time.Time, error) {
    key := fmt.Sprintf("%s%d", TokenInvalidatePrefix, userID)
    timestamp, err := s.redis.GetInt64(ctx, key)
    if err != nil {
        return time.Time{}, nil  // No invalidation timestamp = all tokens valid
    }
    return time.Unix(timestamp, 0), nil
}
```

### Validation Flow

```
              Token Received
                    │
                    ▼
        ┌───────────────────────┐
        │ Parse JWT (signature, │
        │ expiry, claims)       │
        └───────────────────────┘
                    │
                    ▼
        ┌───────────────────────┐
        │ Check blacklist       │
        │ (single token)        │──── Found ───▶ REJECT (401)
        └───────────────────────┘
                    │ Not found
                    ▼
        ┌───────────────────────┐
        │ Get invalidation      │
        │ timestamp for user    │
        └───────────────────────┘
                    │
                    ▼
        ┌───────────────────────┐
        │ Token issued_at >     │
        │ invalidation time?    │──── No ───▶ REJECT (401)
        └───────────────────────┘
                    │ Yes
                    ▼
               ALLOW REQUEST
```

### JWT Claims

Tokens include `iat` (issued at) claim for timestamp comparison:

```go
type Claims struct {
    jwt.RegisteredClaims
    UserID uint   `json:"user_id"`
    // iat is automatically included by jwt-go
}
```

### Atomic Operations

For critical operations, use atomic Redis operations with retries:

```go
func (s *TokenStore) BlacklistAtomicWithRetry(ctx context.Context, token string, exp time.Time) error {
    maxRetries := 3
    for i := 0; i < maxRetries; i++ {
        err := s.BlacklistAtomic(ctx, token, exp)
        if err == nil {
            return nil
        }
        if i < maxRetries-1 {
            time.Sleep(time.Duration(i+1) * 10 * time.Millisecond)
        }
    }
    return fmt.Errorf("blacklist failed after %d retries", maxRetries)
}
```

### Batch Operations

For multiple tokens (e.g., logging out all sessions), use Redis pipeline:

```go
func (s *TokenStore) BlacklistMultiple(ctx context.Context, tokens []TokenExpiry) error {
    pipe := s.redis.Pipeline()
    for _, t := range tokens {
        key := fmt.Sprintf("%s%s", TokenBlacklistPrefix, hashToken(t.Token))
        ttl := time.Until(t.ExpiresAt)
        if ttl > 0 {
            pipe.Set(ctx, key, "1", ttl)
        }
    }
    _, err := pipe.Exec(ctx)
    return err
}
```

### Login Rate Limiting (Co-located)

The token store also manages login rate limiting:

```go
// Per-email rate limiting
Key: auth:login:rate:email:{email}
Value: attempt count
TTL: 15 minutes (resets after window)

// Per-IP rate limiting
Key: auth:login:rate:ip:{ip}
Value: attempt count
TTL: 15 minutes
```

Methods:
- `CheckLoginRateLimit(email, ip)` - Check if limits exceeded
- `RecordFailedLogin(email, ip)` - Increment counters
- `ClearLoginRateLimit(email, ip)` - Clear on successful login

## Consequences

### Positive

- **Fast Invalidation**: O(1) Redis lookups vs O(n) database queries
- **Granular Control**: Single-token or all-tokens invalidation
- **Memory Efficient**: Blacklisted tokens auto-expire via TTL
- **No Token Proliferation**: Timestamp approach doesn't grow with token count
- **Reliable**: Atomic operations prevent race conditions

### Negative

- **Redis Dependency**: Token validation requires Redis
- **Clock Skew**: Timestamp comparison assumes synchronized clocks
- **Persistence**: Redis must persist data for token security to survive restarts

### Neutral

- **Two Lookups**: Each request requires blacklist + timestamp check
- **Trade-off**: Slight latency increase for strong security guarantees

## Security Considerations

1. **Hash Storage**: Tokens are hashed before storage to prevent leakage
2. **No Expiry on Timestamps**: Invalidation timestamps persist until explicitly cleared
3. **Atomic Operations**: Critical paths use atomic Redis operations
4. **Fail-Closed**: If Redis is unavailable, authentication fails (unlike rate limiting which fails open)

## Related

- ADR-003: RBAC with Casbin (authorization after authentication)
- ADR-006: Self-Registration and OAuth (token generation)
- ADR-007: Rate Limiting Strategy (login rate limiting)
