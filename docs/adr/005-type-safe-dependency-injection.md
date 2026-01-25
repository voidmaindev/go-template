# ADR-005: Type-Safe Dependency Injection

## Status

Accepted

## Context

We need a dependency injection system that:
- Provides compile-time type safety
- Supports lazy initialization
- Allows singleton and transient lifetimes
- Is simple and requires minimal boilerplate

## Decision

Implement a custom container using `container.Key[T]` pattern:

```go
// Key definition (compile-time type safety)
var UserServiceKey = container.NewKey[*UserService]("user.service")

// Registration
container.Register(UserServiceKey, func(c *Container) *UserService {
    return NewUserService(c.MustGet(UserRepoKey))
})

// Retrieval (type-safe, no casting)
userService := container.MustGet(UserServiceKey)
```

**Key Features**:
- Generic `Key[T]` captures type at definition site
- Factory functions for lazy initialization
- Singleton by default (cached after first creation)
- `MustGet` panics on missing dependency (fail-fast)

## Consequences

### Positive
- **Type safety**: No type assertions at retrieval
- **Compile-time checks**: Wrong type usage caught at compile time
- **Lazy init**: Dependencies created on first access
- **Simple API**: Register and Get are the only operations

### Negative
- **String keys**: Internal key storage uses strings
- **Runtime errors**: Missing dependencies panic at runtime
- **No interfaces**: Keys are tied to concrete types

### Mitigations
- Define all keys in central location (`internal/container/keys.go`)
- Use `Get` (returns error) for optional dependencies
- Document dependency graph in architecture docs
