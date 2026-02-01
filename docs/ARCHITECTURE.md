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

- [ADR-001: Domain-Driven Architecture](adr/001-domain-driven-architecture.md)
- [ADR-002: Generic Repository Pattern](adr/002-generic-repository-pattern.md)
- [ADR-003: RBAC with Casbin](adr/003-rbac-with-casbin.md)
- [ADR-005: Type-Safe Dependency Injection](adr/005-type-safe-dependency-injection.md)
- [Guide: Adding a New Domain](guides/adding-new-domain.md)
- [Guide: Adding a New App](guides/adding-new-app.md)
