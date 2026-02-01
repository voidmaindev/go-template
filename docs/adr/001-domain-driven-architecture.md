# ADR-001: Domain-Driven Architecture

## Status

Accepted (Updated: 2025)

## Context

We need a modular architecture that:
- Allows domains (user, auth, etc.) to be developed independently
- Supports plugin-like extensibility
- Minimizes coupling between domains
- Enables easy addition of new domains without modifying core code
- Provides compile-time type safety for dependency injection

## Decision

Implement an explicit Domain interface pattern with typed dependency injection:

```go
// container/container.go
type Domain interface {
    Name() string                              // Domain identifier (e.g., "user", "item")
    Models() []any                             // GORM models for auto-migration
    Register(c *Container)                     // Register repos, services, handlers
    Routes(api fiber.Router, c *Container)     // Register HTTP routes
}
```

Each domain defines typed keys for compile-time safety:

```go
// user/register.go
var (
    RepositoryKey = container.Key[Repository]("user.repository")
    ServiceKey    = container.Key[Service]("user.service")
    HandlerKey    = container.Key[*Handler]("user.handler")
)

type domain struct{}

func NewDomain() container.Domain { return &domain{} }

func (d *domain) Name() string { return "user" }

func (d *domain) Models() []any {
    return []any{&User{}, &ExternalIdentity{}}
}

func (d *domain) Register(c *container.Container) {
    repo := NewRepository(c.DB)
    RepositoryKey.Set(c, repo)

    service := NewService(repo)
    ServiceKey.Set(c, service)

    handler := NewHandler(service)
    HandlerKey.Set(c, handler)
}

func (d *domain) Routes(api fiber.Router, c *container.Container) {
    handler := HandlerKey.MustGet(c)
    // Register routes with Fiber...
}
```

Apps explicitly register domains in dependency order:

```go
// app/main.go
func MainApp() *App {
    return &App{
        Name:        "main",
        Description: "Full application with all domains",
        Domains:     mainDomains,
    }
}

func mainDomains() []container.Domain {
    return []container.Domain{
        rbac.NewDomain(),     // Must be first (user depends on rbac)
        user.NewDomain(),     // Depends on: rbac
        email.NewDomain(),    // Standalone
        auth.NewDomain(),     // Depends on: user, email, rbac
        item.NewDomain(),     // Standalone
        // ...
    }
}
```

The application startup:
1. Creates container with core dependencies (DB, Redis, Config)
2. Adds domains via `container.AddDomain(d)` in explicit order
3. Auto-migrates all domain models via `container.GetAllModels()`
4. Registers all components via `container.RegisterAll()`
5. Registers all routes via `container.RegisterRoutes(api)`

## Consequences

### Positive
- **Modularity**: Domains are self-contained packages
- **Type Safety**: `container.Key[T]` prevents string typos at compile time
- **Explicit Dependencies**: Registration order documents dependencies clearly
- **Extensibility**: New domains added by creating `NewDomain()` factory
- **Encapsulation**: Each domain owns its models, handlers, and services
- **Testability**: Domains can be tested in isolation with mock container
- **Multi-App**: Different apps can compose different domain sets

### Negative
- **Manual ordering**: Developer must ensure correct domain registration order
- **Verbose keys**: Each domain must define its typed keys

### Mitigations
- Comments in app files document dependency relationships
- Panic recovery in `TryRegisterAll()` provides clear error messages
- ADR-005 documents the typed key pattern in detail

## See Also

- [ADR-005: Type-Safe Dependency Injection](005-type-safe-dependency-injection.md)
- [ARCHITECTURE.md](../ARCHITECTURE.md) for system overview
