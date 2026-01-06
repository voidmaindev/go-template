# go-template

A production-ready Go backend template using Fiber v2, GORM, PostgreSQL, and Redis.

## Features

- **Multi-App Architecture**: Run multiple apps from the same codebase with different domain sets
- **Self-Registering Domains**: Each domain is self-contained with its own registration logic
- **Dependency Container**: Type-safe dependency injection with generics
- **Generic Repository Pattern**: Maximum database abstraction using Go generics
- **JWT Authentication**: Access and refresh tokens with Redis-based blacklisting
- **CLI with Cobra**: Professional CLI with subcommands (serve, migrate, seed, version)
- **Database Seeders**: Separate seeder infrastructure with tracking and idempotency
- **Docker Ready**: Multi-stage Dockerfile with docker-compose
- **Validation**: Request validation using go-playground/validator
- **Filtering & Sorting**: Django-style query syntax with full operator support
- **Pagination**: Built-in pagination for all list endpoints
- **Structured Logging**: Uses Go 1.21+ slog for consistent, structured logging
- **Graceful Shutdown**: Configurable shutdown timeout with connection draining
- **Health Checks**: Kubernetes-ready probes (`/healthz`, `/readyz`, `/health`)
- **Security Hardened**: Error sanitization, info leakage prevention, secure defaults
- **Comprehensive Tests**: Unit tests for services, middleware, handlers, and utilities
- **OpenAPI 3.0 Documentation**: Spec-first API with Scalar UI
- **GORM Go Code Migrations**: Versioned database migrations with rollback support
- **OpenTelemetry Tracing**: Distributed tracing with OTLP export
- **Prometheus Metrics**: HTTP, database, and business metrics at `/metrics`
- **Request Timeouts**: Configurable per-operation timeout middleware
- **Rate Limiting**: With standard headers (`X-RateLimit-*`)
- **Request Correlation**: Request ID in all error responses

## API Documentation

This project uses **OpenAPI 3.0** with code generation via [oapi-codegen](https://github.com/oapi-codegen/oapi-codegen) and **Scalar** for interactive documentation.

### Access Documentation

| URL | Description |
|-----|-------------|
| `http://localhost:3000/docs` | Scalar interactive API documentation |
| `http://localhost:3000/openapi.json` | OpenAPI 3.0 specification (JSON) |

### Regenerating API Code

After modifying `api/openapi.yaml`:

```bash
# Install oapi-codegen (if not installed)
go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest

# Regenerate code
oapi-codegen --config oapi-codegen.yaml api/openapi.yaml
```

Generated files:
- `internal/api/api.gen.go` - Types, interfaces, and router

## Project Structure

```
.
├── api/
│   └── openapi.yaml                 # OpenAPI 3.0 specification
├── cmd/api/
│   ├── main.go                      # Application entry point
│   └── cmd/                         # Cobra CLI commands
│       ├── root.go
│       ├── serve.go                 # serve [app] command
│       ├── migrate.go               # migrate [up|down|status|reset|refresh]
│       ├── seed.go                  # seed [status|reset] --fresh
│       └── version.go               # version command
├── internal/
│   ├── api/                         # Generated API code
│   │   ├── api.gen.go               # Generated types & interfaces
│   │   ├── server.go                # StrictServerInterface impl
│   │   └── converters.go            # Type conversion helpers
│   ├── docs/                        # API documentation
│   │   └── scalar.go                # Scalar UI handler
│   ├── app/                         # App definitions
│   │   ├── app.go                   # App struct & registry
│   │   ├── main.go                  # main app (all domains)
│   │   └── geography.go             # geography app (country, city)
│   ├── container/                   # Dependency container
│   ├── config/                      # Configuration management
│   ├── database/                    # Database connection
│   │   ├── migrations/              # GORM Go code migrations
│   │   │   ├── migrations.go        # Migration registry & runner
│   │   │   └── 000001_*.go          # Individual migrations
│   │   └── seeders/                 # Database seeders
│   │       ├── seeders.go           # Seeder registry & runner
│   │       └── admin_user.go        # Default admin user seeder
│   ├── redis/                       # Redis client
│   ├── logger/                      # Structured logging
│   ├── health/                      # Kubernetes health probes
│   │   └── health.go                # /healthz, /readyz, /health
│   ├── telemetry/                   # Observability
│   │   ├── telemetry.go             # OpenTelemetry initialization
│   │   ├── tracer.go                # Span helpers
│   │   └── metrics.go               # Prometheus metrics
│   ├── middleware/                  # HTTP middleware
│   │   ├── jwt.go                   # JWT authentication
│   │   ├── ratelimit.go             # Rate limiting with headers
│   │   ├── timeout.go               # Request timeouts
│   │   └── ...                      # CORS, logging, recovery
│   ├── testutil/                    # Test utilities
│   │   ├── testutil.go              # Assertions & helpers
│   │   ├── containers.go            # Testcontainers (Postgres/Redis)
│   │   └── fixtures.go              # Test data factories
│   ├── integration/                 # Integration tests
│   │   └── setup_test.go            # Test suite setup
│   ├── common/                      # Shared components
│   │   ├── filter/                  # Filtering & sorting system
│   │   ├── errors/                  # Typed domain errors with codes
│   │   │   ├── codes.go             # ErrorCode type with HTTP mapping
│   │   │   ├── domain_error.go      # DomainError struct
│   │   │   └── helpers.go           # NotFound, Unauthorized, etc.
│   │   ├── validation/              # Centralized validation
│   │   │   ├── result.go            # FieldError and Result types
│   │   │   ├── validator.go         # Validator wrapper
│   │   │   └── builder.go           # Composable validation builder
│   │   ├── ctxutil/                 # Context utilities
│   │   │   └── context.go           # Request/User context helpers
│   │   ├── logging/                 # Structured logging
│   │   │   └── logger.go            # Domain/Operation logger
│   │   ├── base_model.go            # Base model with timestamps
│   │   ├── base_repository.go       # Generic repository implementation
│   │   ├── repository.go            # Repository interface
│   │   ├── specification.go         # Query specification pattern
│   │   ├── query_options.go         # Composable query options
│   │   ├── pagination.go            # Pagination utilities
│   │   ├── response.go              # HTTP response helpers (with request_id)
│   │   └── errors.go                # Legacy errors (backward compat)
│   └── domain/
│       ├── user/                    # User domain with auth
│       │   ├── errors.go            # Domain-specific errors
│       │   ├── specs.go             # Query specifications
│       │   ├── validation.go        # Domain validator
│       │   └── ...                  # model, dto, service, handler
│       ├── item/                    # Item domain (example)
│       ├── country/                 # Country domain (example)
│       ├── city/                    # City domain (depends on country)
│       └── document/                # Document with line items (example)
├── pkg/
│   ├── utils/                       # Hash, JWT, money utilities
│   └── validator/                   # Validation helpers
├── grafana/
│   └── provisioning/                # Grafana auto-provisioning
│       ├── datasources/             # Prometheus datasource config
│       └── dashboards/              # Pre-built dashboard JSON files
├── .env.example                     # Environment variables template
├── config.yaml.example              # Configuration file template
├── oapi-codegen.yaml                # OpenAPI code generator config
├── docker-compose.yml               # Docker services
└── Dockerfile                       # Multi-stage build
```

## Quick Start

### Prerequisites

- Go 1.24+
- Docker & Docker Compose
- PostgreSQL 16+
- Redis 7+

### Development Setup

1. **Clone and setup environment**
   ```bash
   cp .env.example .env
   cp config.yaml.example config.yaml
   # Edit .env and config.yaml with your settings
   ```

2. **Start infrastructure with Docker Compose**
   ```bash
   # Core services only (Postgres, Redis, API)
   docker compose up -d

   # With observability stack (Jaeger, Prometheus, Grafana)
   docker compose --profile observability up -d
   ```

3. **Run migrations, seed data, and start the server**
   ```bash
   # Run migrations for main app
   go run ./cmd/api migrate main

   # Run seeders (creates default admin user)
   go run ./cmd/api seed

   # Start the main app
   go run ./cmd/api serve main
   ```

### CLI Commands

```bash
# Run an app
go run ./cmd/api serve main           # Run main app (all domains)
go run ./cmd/api serve geography      # Run geography app (country, city only)

# Database migrations
go run ./cmd/api migrate up           # Apply all pending migrations
go run ./cmd/api migrate up --to=000003    # Migrate up to specific version
go run ./cmd/api migrate down         # Rollback last migration
go run ./cmd/api migrate down 3       # Rollback last 3 migrations
go run ./cmd/api migrate status       # Show migration status
go run ./cmd/api migrate reset        # Rollback all migrations
go run ./cmd/api migrate refresh      # Reset + re-apply all migrations

# Database seeders
go run ./cmd/api seed                 # Run all pending seeders
go run ./cmd/api seed --fresh         # Reset + run all seeders
go run ./cmd/api seed status          # Show seeder status
go run ./cmd/api seed reset           # Clear seeder records (allows re-run)

# With flags
go run ./cmd/api serve main -p 8080   # Custom port
go run ./cmd/api serve main -H 127.0.0.1  # Custom host

# Version info
go run ./cmd/api version
```

### Build and Run

```bash
# Build the application
go build -o api ./cmd/api

# Run the built binary
./api serve main
./api migrate main
./api version
```

## API Endpoints

### Authentication

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/auth/register` | Register new user |
| POST | `/api/v1/auth/login` | Login and get tokens |
| POST | `/api/v1/auth/logout` | Logout (requires auth) |
| POST | `/api/v1/auth/refresh` | Refresh access token |

### Users

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/users/me` | Get current user |
| PUT | `/api/v1/users/me` | Update current user |
| PUT | `/api/v1/users/me/password` | Change password |
| GET | `/api/v1/users` | List all users (admin) |
| GET | `/api/v1/users/:id` | Get user by ID |
| DELETE | `/api/v1/users/:id` | Delete user |

### Items, Countries, Cities

Full CRUD for each: `GET /`, `GET /:id`, `POST /`, `PUT /:id`, `DELETE /:id`

- `/api/v1/items`
- `/api/v1/countries`
- `/api/v1/cities`
- `/api/v1/countries/:id/cities` - Get cities by country

### Documents

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/documents` | List documents |
| GET | `/api/v1/documents/:id` | Get document with items |
| POST | `/api/v1/documents` | Create document with items |
| PUT | `/api/v1/documents/:id` | Update document |
| DELETE | `/api/v1/documents/:id` | Delete document |
| POST | `/api/v1/documents/:id/items` | Add item to document |
| PUT | `/api/v1/documents/:id/items/:itemId` | Update document item |
| DELETE | `/api/v1/documents/:id/items/:itemId` | Remove item |

## Filtering & Sorting

All list endpoints support Django-style filtering and sorting via query parameters.

> **Interactive Documentation**: For a complete interactive reference with examples, visit the [Scalar API docs](http://localhost:3000/docs) when the server is running.

### Query Syntax

```
?field=value              # Equality (default operator)
?field__operator=value    # Specific operator
```

### Supported Operators

| Operator | Description | Example |
|----------|-------------|---------|
| `eq` | Equals (default) | `?name=Berlin` or `?name__eq=Berlin` |
| `gt` | Greater than | `?price__gt=1000` |
| `lt` | Less than | `?price__lt=5000` |
| `gte` | Greater than or equal | `?created_at__gte=2024-01-01` |
| `lte` | Less than or equal | `?price__lte=9999` |
| `contains` | Contains substring (case-insensitive) | `?name__contains=new` |
| `starts_with` | Starts with (case-insensitive) | `?name__starts_with=New` |
| `ends_with` | Ends with (case-insensitive) | `?email__ends_with=@gmail.com` |
| `in` | In list (comma-separated) | `?status__in=active,pending` |
| `is_null` | Is null | `?deleted_at__is_null=true` |
| `is_not_null` | Is not null | `?email__is_not_null=true` |

### Relation Filtering

Filter on related entity fields using dot notation (one level deep):

```
?country.name__contains=Germany    # Filter cities by country name
?city.country.code=DEU             # Filter documents by city's country code
```

### Range Filtering (Multiple Filters on Same Field)

You can apply multiple filters on the same field to create range queries:

```bash
# Date range: items created in 2024
GET /api/v1/items?created_at__gte=2024-01-01&created_at__lte=2024-12-31

# Price range: items between $10 and $100
GET /api/v1/items?price__gte=1000&price__lte=10000

# ID range with exclusion
GET /api/v1/items?id__gt=10&id__lt=100
```

Each filter is applied with AND logic, so all conditions must be satisfied.

### Sorting

```
?sort=field&order=asc     # Sort ascending (default)
?sort=field&order=desc    # Sort descending
```

### Pagination

```
?page=1&page_size=20      # Page 1, 20 items per page
```

- Default page: 1
- Default page_size: 10
- Maximum page_size: 100

### Combined Examples

```bash
# Get cities containing "New", sorted by name
GET /api/v1/cities?name__contains=New&sort=name&order=asc

# Get items priced between 1000 and 5000
GET /api/v1/items?price__gte=1000&price__lte=5000

# Get cities in Germany, page 2, 25 per page
GET /api/v1/cities?country.name=Germany&page=2&page_size=25

# Get users with email ending in @company.com, sorted by creation date
GET /api/v1/users?email__ends_with=@company.com&sort=created_at&order=desc

# Get documents created after 2024-01-01 with non-null notes
GET /api/v1/documents?created_at__gte=2024-01-01&notes__is_not_null=true
```

### Response Format

List endpoints return paginated results:

```json
{
  "success": true,
  "data": {
    "data": [...],
    "total": 100,
    "page": 1,
    "page_size": 10,
    "total_pages": 10,
    "has_more": true
  }
}
```

## Request Examples

### Register

```bash
curl -X POST http://localhost:3000/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email": "user@example.com", "password": "password123", "name": "John Doe"}'
```

### Login

```bash
curl -X POST http://localhost:3000/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email": "user@example.com", "password": "password123"}'
```

### Create Item (with auth)

```bash
curl -X POST http://localhost:3000/api/v1/items \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <access_token>" \
  -d '{"name": "Product", "description": "A product", "price": 1999}'
```

## Error Handling

The template uses a typed domain error system with automatic HTTP status mapping.

### Error Codes

| Code | HTTP Status | Description |
|------|-------------|-------------|
| `NOT_FOUND` | 404 | Resource not found |
| `ALREADY_EXISTS` | 409 | Duplicate resource |
| `VALIDATION_ERROR` | 400 | Invalid input |
| `UNAUTHORIZED` | 401 | Authentication required |
| `FORBIDDEN` | 403 | Permission denied |
| `CONFLICT` | 409 | State conflict |
| `INTERNAL_ERROR` | 500 | Internal server error |
| `BAD_REQUEST` | 400 | Malformed request |

### Defining Domain Errors

Each domain defines its own errors in `errors.go`:

```go
package item

import "github.com/voidmaindev/go-template/internal/common/errors"

const domainName = "item"

var (
    ErrItemNotFound = errors.NotFound(domainName, "item")
    ErrItemExists   = errors.AlreadyExists(domainName, "item", "name")
)
```

### Using Errors in Services

```go
func (s *service) GetByID(ctx context.Context, id uint) (*Item, error) {
    item, err := s.repo.FindByID(ctx, id)
    if err != nil {
        if errors.IsNotFound(err) {
            return nil, ErrItemNotFound
        }
        return nil, errors.Internal(domainName, err).WithOperation("GetByID")
    }
    return item, nil
}
```

### Error Response Format

```json
{
  "success": false,
  "error": {
    "code": "NOT_FOUND",
    "message": "item not found",
    "domain": "item"
  },
  "request_id": "abc-123"
}
```

## Validation

The template provides a composable validation system that combines struct tags with business rules.

### Struct Tag Validation

```go
type CreateItemRequest struct {
    Name        string `json:"name" validate:"required,min=1,max=255"`
    Description string `json:"description" validate:"max=1000"`
    Price       int64  `json:"price" validate:"required,gte=0"`
}
```

### Custom Business Rules

```go
func (v *Validator) ValidateCreate(ctx context.Context, req *CreateItemRequest) *validation.Result {
    return validation.NewBuilder(ctx).
        Struct(v.validator, req).                    // Struct tag validation
        CustomWithCode("name", "DUPLICATE", func(ctx context.Context) error {
            exists, _ := v.repo.ExistsByName(ctx, req.Name)
            if exists {
                return fmt.Errorf("item with this name already exists")
            }
            return nil
        }).
        Condition(req.Price < 0, "price", "INVALID", "price cannot be negative").
        Result()
}
```

### Validation Error Response

```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "validation failed",
    "details": [
      {"field": "name", "code": "REQUIRED", "message": "field is required"},
      {"field": "price", "code": "TOO_SMALL", "message": "value is too small"}
    ]
  },
  "request_id": "abc-123"
}
```

## Query Specifications

The specification pattern allows composable, reusable query conditions.

### Built-in Specifications

```go
// By field value
spec := common.ByField("status", "active")

// Combine with AND
spec := common.And(
    common.ByField("status", "active"),
    common.ByField("type", "product"),
)

// Combine with OR
spec := common.Or(
    common.ByField("role", "admin"),
    common.ByField("role", "moderator"),
)
```

### Domain-Specific Specifications

Each domain can define its own specs in `specs.go`:

```go
package user

// ByEmail finds user by email
type EmailSpec struct {
    Email string
}

func ByEmail(email string) EmailSpec {
    return EmailSpec{Email: email}
}

func (s EmailSpec) ApplyGorm(db *gorm.DB) *gorm.DB {
    return db.Where("email = ?", s.Email)
}
```

### Using Specifications

```go
// Find one by spec
user, err := repo.FindOne(ctx, user.ByEmail("user@example.com"))

// Find many with options
users, total, err := repo.FindBySpec(ctx,
    common.And(user.ByRole("admin"), user.ActiveAfter(lastMonth)),
    common.WithPagination(pagination),
    common.WithPreload("Profile"),
)
```

## Query Options

Composable query options replace method explosion with a flexible API.

### Available Options

```go
// Eager loading
common.WithPreload("Country", "Items")

// Pagination
common.WithPagination(&common.Pagination{Page: 1, PageSize: 20})

// Sorting
common.WithSort("created_at", "desc")

// Dynamic filtering
common.WithFilter(filterParams)

// Limit/Offset (non-paginated)
common.WithLimit(10)
common.WithOffset(5)
```

### Usage Examples

```go
// Simple list with pagination
items, total, _ := repo.FindAll(ctx,
    common.WithPagination(pagination),
    common.WithSort("name", "asc"),
)

// With preloading and filtering
cities, total, _ := repo.FindAll(ctx,
    common.WithPreload("Country"),
    common.WithFilter(filterParams),
    common.WithPagination(pagination),
)

// Using builder pattern
opts := common.NewQueryBuilder().
    Preload("Country", "Items").
    Paginate(1, 20).
    Sort("created_at", true).
    Build()

items, total, _ := repo.FindAll(ctx, opts)
```

## Context Propagation

Request context (request ID, user info) flows through all layers for logging and tracing.

### Context Values

| Key | Type | Description |
|-----|------|-------------|
| `request_id` | string | Unique request identifier |
| `trace_id` | string | OpenTelemetry trace ID |
| `user_id` | uint | Authenticated user ID |
| `user_email` | string | Authenticated user email |
| `user_role` | string | Authenticated user role |

### Accessing Context

```go
import "github.com/voidmaindev/go-template/internal/common/ctxutil"

func (s *service) Create(ctx context.Context, req *CreateRequest) error {
    requestID := ctxutil.GetRequestID(ctx)
    userID := ctxutil.GetUserID(ctx)

    // Context values are automatically included in logs
    s.logger.WithOperation("create").Info(ctx, "creating item")
}
```

## Structured Logging

Domain-aware logging with automatic context enrichment.

### Basic Usage

```go
import "github.com/voidmaindev/go-template/internal/common/logging"

type service struct {
    logger *logging.Logger
}

func NewService() Service {
    return &service{
        logger: logging.New("item"),
    }
}
```

### Operation Logging

```go
func (s *service) Create(ctx context.Context, req *CreateRequest) (*Item, error) {
    log := s.logger.WithOperation("create")

    log.Info(ctx, "creating item", "name", req.Name)

    // ... create item ...

    log.WithEntity("item", item.ID).Info(ctx, "item created")
    return item, nil
}
```

### Log Output (JSON)

```json
{
  "time": "2024-01-15T10:30:00Z",
  "level": "INFO",
  "msg": "item created",
  "domain": "item",
  "operation": "create",
  "request_id": "abc-123",
  "trace_id": "xyz-789",
  "user_id": 42,
  "entity_type": "item",
  "entity_id": 156
}
```

## Creating New Domains

To add a new domain:

1. Create folder `internal/domain/<name>/`
2. Create files:
   - `model.go` - Domain model
   - `dto.go` - Request/Response types
   - `errors.go` - Domain-specific errors
   - `repository.go` - Repository interface
   - `service.go` - Business logic
   - `handler.go` - HTTP handlers
   - `register.go` - DI and route registration
   - `specs.go` - Query specifications (optional)
   - `validation.go` - Domain validation (optional)
3. In `register.go`, implement the `container.Domain` interface:

```go
package product

// Component keys for this domain
const (
    RepositoryKey = "product.repository"
    ServiceKey    = "product.service"
    HandlerKey    = "product.handler"
)

type domain struct{}

func NewDomain() container.Domain { return &domain{} }

func (d *domain) Name() string { return "product" }

func (d *domain) Models() []any {
    return []any{&Product{}}
}

func (d *domain) Register(c *container.Container) {
    repo := NewRepository(c.DB)
    c.Set(RepositoryKey, repo)
    // ... service, handler
}

func (d *domain) Routes(api fiber.Router, c *container.Container) {
    // ... register routes
}
```

4. Add to relevant app(s) in `internal/app/`:

```go
// internal/app/main.go
Domains: func() []container.Domain {
    return []container.Domain{
        // ...existing domains
        product.NewDomain(),
    }
}
```

## Creating New Apps

To add a new app:

1. Create `internal/app/<appname>.go`:

```go
package app

func init() {
    Register(&App{
        Name:        "myapp",
        Description: "My custom application",
        Domains: func() []container.Domain {
            return []container.Domain{
                user.NewDomain(),  // Required for auth
                // Add your domains here
            }
        },
    })
}
```

2. Run with: `go run ./cmd/api serve myapp`

## Removing Template Domains

To use this template for your project:

1. Delete `internal/domain/item/`, `country/`, `city/`, `document/`
2. Update `internal/app/main.go` to remove those domains
3. Delete `internal/app/geography.go` if not needed
4. Create your own domains

## Configuration

Configuration can be set via environment variables or `config.yaml` file. Environment variables take precedence.

### Application

| Variable | Default | Description |
|----------|---------|-------------|
| `APP_NAME` | go-template | Application name |
| `APP_ENV` | development | Environment (development, staging, production) |
| `APP_DEBUG` | true | Enable debug mode |

### Server

| Variable | Default | Description |
|----------|---------|-------------|
| `SERVER_HOST` | 0.0.0.0 | Server host |
| `SERVER_PORT` | 3000 | Server port |
| `SERVER_READ_TIMEOUT` | 10s | HTTP read timeout |
| `SERVER_WRITE_TIMEOUT` | 10s | HTTP write timeout |
| `SERVER_IDLE_TIMEOUT` | 120s | HTTP idle timeout |
| `SERVER_SHUTDOWN_TIMEOUT` | 30s | Graceful shutdown timeout |

### Database

| Variable | Default | Description |
|----------|---------|-------------|
| `DB_HOST` | localhost | PostgreSQL host |
| `DB_PORT` | 5432 | PostgreSQL port |
| `DB_USER` | postgres | Database user |
| `DB_PASSWORD` | postgres | Database password |
| `DB_NAME` | go-template | Database name |
| `DB_SSL_MODE` | disable | SSL mode |
| `DB_MAX_IDLE_CONNS` | 10 | Max idle connections |
| `DB_MAX_OPEN_CONNS` | 100 | Max open connections |
| `DB_MAX_LIFETIME` | 1h | Connection max lifetime |
| `DB_SLOW_QUERY_THRESHOLD` | 200ms | Slow query log threshold |
| `DB_RETRY_ATTEMPTS` | 5 | Connection retry attempts |
| `DB_RETRY_DELAY` | 5s | Delay between retries |

### Redis

| Variable | Default | Description |
|----------|---------|-------------|
| `REDIS_HOST` | localhost | Redis host |
| `REDIS_PORT` | 6379 | Redis port |
| `REDIS_PASSWORD` | | Redis password |
| `REDIS_DB` | 0 | Redis database number |
| `REDIS_RETRY_ATTEMPTS` | 5 | Connection retry attempts |
| `REDIS_RETRY_DELAY` | 5s | Delay between retries |

### JWT

| Variable | Default | Description |
|----------|---------|-------------|
| `JWT_SECRET` | (required in prod) | JWT secret key (min 32 chars) |
| `JWT_ACCESS_EXPIRY` | 15m | Access token expiry |
| `JWT_REFRESH_EXPIRY` | 168h | Refresh token expiry (7 days) |
| `JWT_ISSUER` | go-template | JWT issuer |

### Seeding

| Variable | Default | Description |
|----------|---------|-------------|
| `SEED_ADMIN_EMAIL` | admin@admin.com | Admin user email |
| `SEED_ADMIN_PASSWORD` | (required in prod) | Admin user password |
| `SEED_ADMIN_NAME` | Administrator | Admin user display name |

> **Note**: In production, `SEED_ADMIN_PASSWORD` must be set via environment variable or secrets. The application will fail validation if not provided.

### CORS

| Variable | Default | Description |
|----------|---------|-------------|
| `CORS_ALLOWED_ORIGINS` | localhost:3000,localhost:5173 | Allowed CORS origins |

### Telemetry (OpenTelemetry)

| Variable | Default | Description |
|----------|---------|-------------|
| `TELEMETRY_ENABLED` | false | Enable OpenTelemetry tracing |
| `TELEMETRY_SERVICE_NAME` | go-template | Service name for traces |
| `TELEMETRY_SERVICE_VERSION` | 1.0.0 | Service version for traces |
| `OTLP_ENDPOINT` | localhost:4318 | OTLP collector endpoint |
| `OTLP_INSECURE` | true | Use insecure connection |
| `TELEMETRY_SAMPLING_RATIO` | 1.0 | Trace sampling ratio (0.0-1.0) |

## Observability

### Prometheus Metrics

When the server is running, Prometheus metrics are available at `/metrics`:

```bash
curl http://localhost:3000/metrics
```

Available metrics:
- `http_requests_total` - Total HTTP requests by method, path, status
- `http_request_duration_seconds` - Request latency histogram
- `db_queries_total` - Database query count
- `db_query_duration_seconds` - Database query latency
- `redis_operations_total` - Redis operation count
- Custom business metrics

### Distributed Tracing

Enable OpenTelemetry tracing by setting:

```bash
TELEMETRY_ENABLED=true
OTLP_ENDPOINT=localhost:4318  # Jaeger, Tempo, or any OTLP collector
```

Traces include:
- HTTP request spans with method, path, status
- Database query spans
- Redis operation spans
- Custom spans via `telemetry.StartSpan()`

### Local Observability Stack

Start the full observability stack with Docker:

```bash
docker compose --profile observability up -d
```

This starts:
| Service | URL | Description |
|---------|-----|-------------|
| Jaeger | http://localhost:16686 | Distributed tracing UI |
| Prometheus | http://localhost:9090 | Metrics database & query |
| Grafana | http://localhost:3001 | Dashboards (admin/admin) |

Enable telemetry in your `.env`:
```bash
TELEMETRY_ENABLED=true
OTLP_ENDPOINT=jaeger:4318
```

### Pre-Built Grafana Dashboards

Grafana comes with auto-provisioned dashboards ready out-of-the-box. No manual configuration required.

**Available Dashboards** (in "Go Application" folder):

| Dashboard | Description |
|-----------|-------------|
| **Go Runtime Metrics** | Goroutines, memory allocation, heap usage, GC duration, OS threads |
| **HTTP Metrics** | Request rate, p95 latency, error rate, requests in flight, top endpoints |
| **Business & Infrastructure** | Users registered, logins, documents created, DB queries, Redis operations |

All dashboards use the `go_template_*` metric prefix and auto-refresh every 30 seconds.

**Provisioning Location:**
```
grafana/provisioning/
├── datasources/
│   └── prometheus.yml    # Auto-configured Prometheus connection
└── dashboards/
    ├── dashboards.yml    # Dashboard provider config
    ├── go-runtime.json   # Go runtime metrics
    ├── http-metrics.json # HTTP request metrics
    └── business-metrics.json # Business & DB metrics
```

## Testing

```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test ./... -v

# Run tests with coverage
go test ./... -cover

# Run specific package tests
go test ./internal/domain/user/... -v
go test ./internal/common/filter/... -v
```

## Kubernetes Deployment

### Health Probes

The application provides Kubernetes-ready health endpoints:

| Endpoint | Purpose | Use For |
|----------|---------|---------|
| `GET /healthz` | Liveness probe | `livenessProbe` - is the app alive? |
| `GET /readyz` | Readiness probe | `readinessProbe` - are dependencies ready? |
| `GET /health` | Combined check | Backward compatible, detailed status |

Example Kubernetes deployment:

```yaml
livenessProbe:
  httpGet:
    path: /healthz
    port: 3000
  initialDelaySeconds: 5
  periodSeconds: 10

readinessProbe:
  httpGet:
    path: /readyz
    port: 3000
  initialDelaySeconds: 5
  periodSeconds: 5
```

### Rate Limiting

All API endpoints include rate limit headers:

| Header | Description |
|--------|-------------|
| `X-RateLimit-Limit` | Max requests per window |
| `X-RateLimit-Remaining` | Remaining requests |
| `X-RateLimit-Reset` | Unix timestamp when limit resets |

Default limits:
- General API: 100 requests/minute
- Auth endpoints: 5 requests/minute

## Security Features

- **Error Sanitization**: Internal errors are logged but not exposed to clients
- **Password Security**: Bcrypt hashing with validation for complexity requirements
- **Token Blacklisting**: Logout invalidates tokens with retry logic
- **Timing Attack Prevention**: Constant-time responses for sensitive operations
- **Info Leakage Prevention**: Generic error messages for authentication failures
- **SQL Injection Prevention**: Parameterized queries with field whitelisting
- **Health Checks**: Kubernetes-ready liveness and readiness probes
- **Graceful Shutdown**: Configurable timeout for connection draining
- **Request Correlation**: All error responses include `request_id` for debugging
- **Rate Limiting**: Standard headers for client-side rate limit awareness

## License

MIT
