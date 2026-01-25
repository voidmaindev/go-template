# ADR-001: Domain-Driven Architecture

## Status

Accepted

## Context

We need a modular architecture that:
- Allows domains (user, auth, etc.) to be developed independently
- Supports plugin-like extensibility
- Minimizes coupling between domains
- Enables easy addition of new domains without modifying core code

## Decision

Implement a self-registering Domain interface pattern:

```go
type Domain interface {
    Name() string
    Routes(r chi.Router, c *container.Container)
    Migrate(db *gorm.DB) error
}
```

Each domain registers itself via `init()` function:

```go
func init() {
    domains.Register(&UserDomain{})
}
```

The application iterates over registered domains at startup to:
1. Run database migrations
2. Register HTTP routes
3. Initialize domain-specific services

## Consequences

### Positive
- **Modularity**: Domains are self-contained packages
- **Extensibility**: New domains are added by importing the package
- **Encapsulation**: Each domain owns its models, handlers, and services
- **Testability**: Domains can be tested in isolation

### Negative
- **Registration order**: Domain initialization order depends on import order
- **Hidden dependencies**: `init()` functions make dependencies implicit
- **Discovery**: Must scan code to find all registered domains

### Mitigations
- Document domain dependencies explicitly
- Use container for cross-domain dependencies
- Keep domain count manageable
