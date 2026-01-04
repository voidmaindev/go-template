# go-template

A production-ready Go backend template using Fiber v2, GORM, PostgreSQL, and Redis.

## Features

- **Multi-App Architecture**: Run multiple apps from the same codebase with different domain sets
- **Self-Registering Domains**: Each domain is self-contained with its own registration logic
- **Dependency Container**: Type-safe dependency injection with generics
- **Generic Repository Pattern**: Maximum database abstraction using Go generics
- **JWT Authentication**: Access and refresh tokens with Redis-based blacklisting
- **CLI with Cobra**: Professional CLI with subcommands (serve, migrate, version)
- **Docker Ready**: Multi-stage Dockerfile with docker-compose
- **Validation**: Request validation using go-playground/validator
- **Filtering & Sorting**: Django-style query syntax with full operator support
- **Pagination**: Built-in pagination for all list endpoints
- **Structured Logging**: Uses Go 1.21+ slog for consistent, structured logging
- **Graceful Shutdown**: Configurable shutdown timeout with connection draining
- **Health Checks**: Comprehensive health endpoint with DB/Redis verification
- **Security Hardened**: Error sanitization, info leakage prevention, secure defaults
- **Comprehensive Tests**: Unit tests for services, middleware, handlers, and utilities

## Project Structure

```
.
├── cmd/api/
│   ├── main.go                      # Application entry point
│   └── cmd/                         # Cobra CLI commands
│       ├── root.go
│       ├── serve.go                 # serve [app] command
│       ├── migrate.go               # migrate [app] command
│       └── version.go               # version command
├── internal/
│   ├── app/                         # App definitions
│   │   ├── app.go                   # App struct & registry
│   │   ├── main.go                  # main app (all domains)
│   │   └── geography.go             # geography app (country, city)
│   ├── container/                   # Dependency container
│   ├── config/                      # Configuration management
│   ├── database/                    # Database connection & migrations
│   ├── redis/                       # Redis client
│   ├── logger/                      # Structured logging
│   ├── middleware/                  # JWT, CORS, logging, recovery
│   ├── common/                      # Shared components
│   │   ├── filter/                  # Filtering & sorting system
│   │   ├── base_model.go            # Base model with timestamps
│   │   ├── base_repository.go       # Generic repository implementation
│   │   ├── repository.go            # Repository interface
│   │   ├── pagination.go            # Pagination utilities
│   │   ├── response.go              # HTTP response helpers
│   │   └── errors.go                # Common errors
│   └── domain/
│       ├── user/                    # User domain with auth
│       ├── item/                    # Item domain (example)
│       ├── country/                 # Country domain (example)
│       ├── city/                    # City domain (depends on country)
│       └── document/                # Document with line items (example)
├── pkg/
│   ├── utils/                       # Hash, JWT, money utilities
│   └── validator/                   # Validation helpers
├── .env.example                     # Environment variables template
├── config.yaml.example              # Configuration file template
├── docker-compose.yml               # Docker services
└── Dockerfile                       # Multi-stage build
```

## Quick Start

### Prerequisites

- Go 1.22+
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
   docker compose up -d
   ```

3. **Run migrations and start the server**
   ```bash
   # Run migrations for main app
   go run ./cmd/api migrate main

   # Start the main app
   go run ./cmd/api serve main
   ```

### CLI Commands

```bash
# Run an app
go run ./cmd/api serve main           # Run main app (all domains)
go run ./cmd/api serve geography      # Run geography app (country, city only)

# Run migrations
go run ./cmd/api migrate main         # Migrate main app tables
go run ./cmd/api migrate geography    # Migrate geography app tables

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

## Creating New Domains

To add a new domain:

1. Create folder `internal/domain/<name>/`
2. Create files: `model.go`, `dto.go`, `repository.go`, `service.go`, `handler.go`, `register.go`
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

### CORS

| Variable | Default | Description |
|----------|---------|-------------|
| `CORS_ALLOWED_ORIGINS` | localhost:3000,localhost:5173 | Allowed CORS origins |

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

## Security Features

- **Error Sanitization**: Internal errors are logged but not exposed to clients
- **Password Security**: Bcrypt hashing with validation for complexity requirements
- **Token Blacklisting**: Logout invalidates tokens with retry logic
- **Timing Attack Prevention**: Constant-time responses for sensitive operations
- **Info Leakage Prevention**: Generic error messages for authentication failures
- **SQL Injection Prevention**: Parameterized queries with field whitelisting
- **Health Checks**: Verify both DB and Redis connectivity
- **Graceful Shutdown**: Configurable timeout for connection draining

## License

MIT
