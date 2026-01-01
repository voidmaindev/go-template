# GoTemplate

A professional, scalable Go backend template using Fiber v2, GORM, PostgreSQL, and Redis.

## Features

- **Multi-App Architecture**: Run multiple apps from the same codebase with different domain sets
- **Self-Registering Domains**: Each domain is self-contained with its own registration logic
- **Dependency Container**: Automatic dependency injection and cross-domain dependencies
- **Generic Repository Pattern**: Maximum database abstraction using Go generics
- **JWT Authentication**: Access and refresh tokens with Redis-based blacklisting
- **CLI with Cobra**: Professional CLI with subcommands (serve, migrate)
- **Docker Ready**: Multi-stage Dockerfile with docker-compose
- **Validation**: Request validation using go-playground/validator
- **Pagination**: Built-in pagination support for all list endpoints

## Project Structure

```
.
├── cmd/api/
│   ├── main.go                      # Application entry point
│   └── cmd/                         # Cobra CLI commands
│       ├── root.go
│       ├── serve.go                 # serve [app] command
│       └── migrate.go               # migrate [app] command
├── internal/
│   ├── app/                         # App definitions
│   │   ├── app.go                   # App struct & registry
│   │   ├── main.go                  # main app (all domains)
│   │   └── geography.go             # geography app (country, city)
│   ├── container/                   # Dependency container
│   ├── config/                      # Configuration management
│   ├── database/                    # Database connection & migrations
│   ├── redis/                       # Redis client
│   ├── middleware/                  # JWT, CORS, logging, recovery
│   ├── common/                      # Generic repository, base model, responses
│   └── domain/
│       ├── user/                    # User domain with auth
│       ├── item/                    # Item domain (template)
│       ├── country/                 # Country domain (template)
│       ├── city/                    # City domain (depends on country)
│       └── document/                # Document with line items
├── pkg/
│   ├── utils/                       # Hash, JWT, money utilities
│   └── validator/                   # Validation helpers
├── docker-compose.yml
├── Dockerfile
└── Makefile
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
   # Edit .env with your settings
   ```

2. **Start with Docker Compose**
   ```bash
   make docker-up
   ```

3. **Or run locally**
   ```bash
   # Start PostgreSQL and Redis
   make docker-up

   # Run migrations for main app
   go run . migrate main

   # Run the main app
   go run . serve main
   ```

### CLI Commands

```bash
# Run an app
go run . serve main           # Run main app (all domains)
go run . serve geography      # Run geography app (country, city only)

# Run migrations
go run . migrate main         # Migrate main app tables
go run . migrate geography    # Migrate geography app tables

# With flags
go run . serve main -p 8080   # Custom port
go run . serve main -H 127.0.0.1  # Custom host
```

### Make Commands

```bash
make build          # Build the application
make run            # Run the application
make test           # Run tests
make docker-up      # Start Docker containers
make docker-down    # Stop Docker containers
make docker-rebuild # Rebuild and start
make docker-logs    # Start with logs
make help           # Show all commands
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
| GET | `/api/v1/users` | List all users |
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

func (d *domain) Models() []interface{} {
    return []interface{}{&Product{}}
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

2. Run with: `go run . serve myapp`

## Removing Template Domains

To use this template for your project:

1. Delete `internal/domain/item/`, `country/`, `city/`, `document/`
2. Update `internal/app/main.go` to remove those domains
3. Delete `internal/app/geography.go` if not needed
4. Create your own domains

## Environment Variables

See `.env.example` for all available configuration options.

## License

MIT
