# Architecture Overview

This document describes the high-level architecture of the go-template backend application.

## System Layers

```
┌─────────────────────────────────────────────────────────────────┐
│                         CLI (Cobra)                              │
│                   cmd/api (serve, migrate, seed)                 │
└─────────────────────────────────────────────────────────────────┘
                                 │
                                 ▼
┌─────────────────────────────────────────────────────────────────┐
│                      HTTP Layer (Fiber v2)                       │
│  ┌─────────────────────────────────────────────────────────────┐│
│  │ Middleware Chain: RequestID → CORS → Logger → Recovery →    ││
│  │                   RateLimit → JWT → RBAC                    ││
│  └─────────────────────────────────────────────────────────────┘│
└─────────────────────────────────────────────────────────────────┘
                                 │
                                 ▼
┌─────────────────────────────────────────────────────────────────┐
│                     Domain Layer                                 │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐           │
│  │   User   │ │   RBAC   │ │   Auth   │ │   Item   │ ...       │
│  │ Domain   │ │ Domain   │ │ Domain   │ │ Domain   │           │
│  └──────────┘ └──────────┘ └──────────┘ └──────────┘           │
│      │             │             │             │                │
│  Handler → Service → Repository                                 │
└─────────────────────────────────────────────────────────────────┘
                                 │
                                 ▼
┌─────────────────────────────────────────────────────────────────┐
│                    Infrastructure Layer                          │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐          │
│  │  PostgreSQL  │  │    Redis     │  │   Casbin     │          │
│  │   (GORM)     │  │  (Sessions)  │  │   (RBAC)     │          │
│  └──────────────┘  └──────────────┘  └──────────────┘          │
└─────────────────────────────────────────────────────────────────┘
```

## Request Flow

1. **HTTP Request** → Fiber router receives the request
2. **Middleware Chain** → Executes in order:
   - `RequestID`: Generates unique request ID
   - `CORS`: Handles cross-origin requests
   - `Logger`: Logs request/response
   - `Recovery`: Catches panics
   - `RateLimit`: Enforces rate limits by tier
   - `JWT`: Validates access token (if protected route)
   - `RBAC`: Checks permissions (if protected route)
3. **Handler** → Parses request, calls service
4. **Service** → Executes business logic, calls repository
5. **Repository** → Performs database operations via GORM
6. **Response** → Formatted JSON response with request_id

## Domain Structure

Each domain is self-contained and follows this structure:

```
internal/domain/<name>/
├── model.go        # GORM model with BaseModel embedding
├── dto.go          # Request/Response DTOs
├── errors.go       # Domain-specific errors
├── repository.go   # Repository interface (extends common.Repository[T])
├── service.go      # Service interface and implementation
├── handler.go      # HTTP handlers
├── register.go     # Domain registration (implements container.Domain)
├── specs.go        # Query specifications (optional)
└── validation.go   # Custom validation rules (optional)
```

### Domain Interface

Every domain implements `container.Domain`:

```go
type Domain interface {
    Name() string                              // Domain identifier
    Models() []any                             // GORM models for migration
    Register(c *Container)                     // Register repos, services, handlers
    Routes(api fiber.Router, c *Container)     // Register HTTP routes
}
```

### Registration Pattern

Domains use explicit registration via typed keys:

```go
// Component keys (typed for compile-time safety)
var (
    RepositoryKey = container.Key[Repository]("item.repository")
    ServiceKey    = container.Key[Service]("item.service")
    HandlerKey    = container.Key[*Handler]("item.handler")
)

func (d *domain) Register(c *container.Container) {
    repo := NewRepository(c.DB)
    RepositoryKey.Set(c, repo)

    service := NewService(repo)
    ServiceKey.Set(c, service)

    handler := NewHandler(service)
    HandlerKey.Set(c, handler)
}
```

## Dependency Injection Container

The DI container (`internal/container/`) provides:

- **Type-safe keys**: `container.Key[T]` prevents string typos
- **Core dependencies**: DB, Redis, Config
- **Domain registration**: `AddDomain()` for explicit ordering
- **Component access**: `Key.Get()`, `Key.MustGet()`, `Key.Set()`

### Dependency Order

Domains are registered in order - dependencies must come first:

```go
func mainDomains() []container.Domain {
    return []container.Domain{
        rbac.NewDomain(),     // first (user depends on rbac)
        user.NewDomain(),     // depends on: rbac
        email.NewDomain(),    // standalone
        audit.NewDomain(),    // depends on: user (tokenStore)
        auth.NewDomain(),     // depends on: user, email, rbac, audit
        item.NewDomain(),     // standalone
        country.NewDomain(),  // standalone
        city.NewDomain(),     // depends on: country
        document.NewDomain(), // depends on: city, item
    }
}
```

## Multi-App Architecture

The application supports multiple app configurations (`internal/app/`):

```go
// internal/app/main.go - All domains
func MainApp() *App {
    return &App{
        Name:        "main",
        Description: "Full application with all domains",
        Domains:     mainDomains,
    }
}

// internal/app/geography.go - Subset of domains
func GeographyApp() *App {
    return &App{
        Name:        "geography",
        Description: "Geography-only app",
        Domains:     geographyDomains,
    }
}
```

Run specific app: `go run ./cmd/api serve main` or `go run ./cmd/api serve geography`

## Generic Repository Pattern

The base repository (`internal/common/base_repository.go`) provides:

```go
type BaseRepository[T any] struct {
    db *gorm.DB
}

// Implements common.Repository[T] interface:
// Create, FindByID, FindAll, Update, Delete, etc.
```

Domains extend with custom methods:

```go
type Repository interface {
    common.Repository[User]           // Embed generic interface
    FindByEmail(ctx, email) (*User, error)  // Custom method
}

type repository struct {
    *common.BaseRepository[User]      // Embed implementation
}
```

## Error Handling

The error system (`internal/common/errors/`) provides:

- **DomainError**: Typed errors with code, message, domain
- **Error codes**: `NOT_FOUND`, `VALIDATION_ERROR`, `FORBIDDEN`, etc.
- **HTTP mapping**: Each code maps to appropriate status
- **Stack traces**: In development mode
- **Predicate functions**: `errors.IsNotFound()`, `errors.IsForbidden()`, etc.

```go
// Domain errors (errors.go)
var ErrItemNotFound = errors.NotFound(domainName, "item")

// Service usage
if errors.IsNotFound(err) {
    return nil, ErrItemNotFound
}
return nil, errors.Internal(domainName, err).WithOperation("GetByID")
```

## Authentication & Authorization

### JWT Flow
1. Login returns access + refresh tokens
2. Access token in `Authorization: Bearer <token>`
3. Refresh token for getting new access token
4. Token blacklisting via Redis on logout

### RBAC (Casbin)
- Multi-role users
- Domain-based permissions (user:read, item:write)
- System roles: admin, full_reader, full_writer, self_registered
- Permission middleware: `RequirePermission(enforcer, domain, action)`

## Observability

- **Structured Logging**: slog with domain/operation context
- **Distributed Tracing**: OpenTelemetry with OTLP export
- **Metrics**: Prometheus at `/metrics`
- **Health Checks**: `/healthz` (liveness), `/readyz` (readiness)

## Rate Limiting

The application uses a 6-tier rate limiting system backed by Redis sliding window:

```
┌─────────────────────────────────────────────────────────────────┐
│                     Rate Limiting Tiers                          │
├──────────────┬──────────────────────────────────────────────────┤
│ auth         │ Public auth endpoints (login, register)          │
│ auth_user    │ Authenticated auth ops (logout, password)        │
│ rbac_admin   │ RBAC administration operations                   │
│ api_write    │ POST, PUT, DELETE operations                     │
│ api_read     │ GET operations                                   │
│ global       │ Fallback catch-all                               │
└──────────────┴──────────────────────────────────────────────────┘
```

Key features:
- **Redis Sliding Window**: Accurate rate limiting using sorted sets
- **Per-Tier Limits**: Different limits for different operation types
- **Key Strategy**: Public endpoints by IP, authenticated by user+IP
- **Fail-Open**: Allows requests if Redis check fails (availability over strictness)
- **Standard Headers**: `X-RateLimit-Limit`, `X-RateLimit-Remaining`, `X-RateLimit-Reset`

See [ADR-007: Rate Limiting Strategy](adr/007-rate-limiting.md) for details.

## Token Invalidation

Dual strategy for JWT invalidation:

```
┌─────────────────────────────────────────────────────────────────┐
│                    Token Validation Flow                         │
├─────────────────────────────────────────────────────────────────┤
│  1. Parse JWT (signature, expiry, claims)                        │
│  2. Check blacklist (single token) → token:blacklist:{hash}     │
│  3. Check invalidation timestamp → auth:token:invalidated:{uid} │
│  4. Compare token issued_at > invalidation timestamp            │
└─────────────────────────────────────────────────────────────────┘
```

- **Token Blacklist**: For single-token invalidation (logout)
- **Timestamp-Based**: For bulk invalidation (password change, compromise)
- **Atomic Operations**: Critical paths use Redis atomic operations

See [ADR-011: Token Invalidation Strategy](adr/011-token-invalidation.md) for details.

## Domain Lifecycle (Shutdowner)

Domains can implement graceful shutdown via the `Shutdowner` interface:

```go
type Shutdowner interface {
    Shutdown(ctx context.Context) error
}
```

Shutdown flow:
1. Signal received (SIGINT/SIGTERM)
2. Stop accepting new requests
3. Wait for in-flight requests
4. Call `container.Shutdown()` - domains shut down in **reverse order (LIFO)**
5. Close Redis and database connections

See [ADR-010: Domain Lifecycle Management](adr/010-domain-lifecycle.md) for details.

## Filtering System

Django-style filtering with operator syntax:

```
?field__operator=value
```

Supported operators:
| Operator | Example | SQL |
|----------|---------|-----|
| `eq` | `?status__eq=active` | `= value` |
| `gt`, `lt`, `gte`, `lte` | `?price__gt=100` | `> value` |
| `contains` | `?name__contains=smith` | `ILIKE %value%` |
| `in` | `?status__in=a,b,c` | `IN (...)` |
| `is_null`, `is_not_null` | `?deleted_at__is_null=true` | `IS NULL` |

Features:
- **Type-Safe Config**: `FilterConfig` defines allowed fields and operators
- **Relation Filtering**: `?country.name__contains=United`
- **SQL Injection Prevention**: Whitelist approach for columns and relations

See [ADR-012: Django-Style Filtering](adr/012-filtering-system.md) for details.

## Key Patterns

| Pattern | Location | Purpose |
|---------|----------|---------|
| Generic Repository | `common/base_repository.go` | DRY database operations |
| Typed DI Keys | `container/keys.go` | Compile-time type safety |
| Domain Interface | `container/container.go` | Plugin-like extensibility |
| ServiceConfig | Domain services | Clean multi-dependency injection |
| Query Specifications | `common/specification.go` | Composable query conditions |
| Query Options | `common/query_options.go` | Flexible query building |

## Related Documentation

### Architecture Decision Records
- [ADR-001: Domain-Driven Architecture](adr/001-domain-driven-architecture.md)
- [ADR-002: Generic Repository Pattern](adr/002-generic-repository-pattern.md)
- [ADR-003: RBAC with Casbin](adr/003-rbac-with-casbin.md)
- [ADR-004: OpenAPI-First Design](adr/004-openapi-first-design.md)
- [ADR-005: Type-Safe Dependency Injection](adr/005-type-safe-dependency-injection.md)
- [ADR-006: Self-Registration and OAuth](adr/006-self-registration-oauth.md)
- [ADR-007: Rate Limiting Strategy](adr/007-rate-limiting.md)
- [ADR-008: SendGrid Email Provider](adr/008-pluggable-email-provider.md)
- [ADR-009: OAuth Security Hardening](adr/009-oauth-security-hardening.md)
- [ADR-010: Domain Lifecycle Management](adr/010-domain-lifecycle.md)
- [ADR-011: Token Invalidation Strategy](adr/011-token-invalidation.md)
- [ADR-012: Django-Style Filtering](adr/012-filtering-system.md)

### Developer Guides
- [Adding a New Domain](guides/adding-new-domain.md)
- [Adding a New App](guides/adding-new-app.md)
- [Testing Guide](guides/testing.md)
- [Migrations Guide](guides/migrations.md)
- [Seeders Guide](guides/seeders.md)
- [Removing Example Domains](guides/removing-domains.md)
