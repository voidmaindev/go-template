# GoTemplate

A professional, scalable Go backend template using Fiber v2, GORM, PostgreSQL, and Redis.

## Features

- **Entity-Based Architecture**: Each entity has its own folder with model, repository, service, handler, and routes
- **Generic Repository Pattern**: Maximum database abstraction using Go generics
- **JWT Authentication**: Access and refresh tokens with Redis-based blacklisting
- **User Management**: Register, login, logout, refresh token, change password
- **Template Entities**: Item, Country, City, Document (with line items)
- **Docker Ready**: Multi-stage Dockerfile with docker-compose
- **Validation**: Request validation using go-playground/validator
- **Pagination**: Built-in pagination support for all list endpoints

## Project Structure

```
.
├── cmd/api/main.go                  # Application entry point
├── internal/
│   ├── config/                      # Configuration management
│   ├── database/                    # Database connection & migrations
│   ├── redis/                       # Redis client
│   ├── middleware/                  # JWT, CORS, logging, recovery
│   ├── common/                      # Generic repository, base model, responses
│   └── entity/
│       ├── user/                    # User entity with auth
│       ├── item/                    # Item entity (template)
│       ├── country/                 # Country entity (template)
│       ├── city/                    # City entity (belongs to Country)
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

   # Run the API
   make run
   ```

### Available Commands

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

## Creating New Entities

To add a new entity:

1. Create folder `internal/domain/<name>/`
2. Create files: `model.go`, `dto.go`, `repository.go`, `service.go`, `handler.go`, `routes.go`
3. Add model to migrations in `cmd/api/main.go`
4. Register routes in `cmd/api/main.go`

Use the existing entities (item, country, city) as templates.

## Removing Template Entities

To use this template for your project:

1. Delete `internal/domain/item/`, `country/`, `city/`, `document/`
2. Remove imports and registrations from `cmd/api/main.go`
3. Create your own entities

## Environment Variables

See `.env.example` for all available configuration options.

## License

MIT
