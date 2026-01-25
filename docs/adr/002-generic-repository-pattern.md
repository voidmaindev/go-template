# ADR-002: Generic Repository Pattern

## Status

Accepted

## Context

We need a data access layer that:
- Reduces boilerplate for CRUD operations
- Provides type-safe database operations
- Maintains consistent error handling across domains
- Supports GORM while abstracting database details

## Decision

Implement a generic `BaseRepository[T]` using Go generics:

```go
type BaseRepository[T any] struct {
    db        *gorm.DB
    modelName string
}

func (r *BaseRepository[T]) FindByID(id uint) (*T, error)
func (r *BaseRepository[T]) Create(entity *T) error
func (r *BaseRepository[T]) Update(entity *T) error
func (r *BaseRepository[T]) Delete(id uint) error
func (r *BaseRepository[T]) FindAll(opts ...QueryOption) ([]T, error)
```

Domain repositories embed `BaseRepository` and add domain-specific methods:

```go
type UserRepository struct {
    *BaseRepository[models.User]
}

func (r *UserRepository) FindByEmail(email string) (*models.User, error)
```

## Consequences

### Positive
- **DRY**: Common CRUD operations defined once
- **Type safety**: Compile-time type checking for all operations
- **Consistency**: Uniform API across all repositories
- **Extensibility**: Domain repos add specialized methods

### Negative
- **Reflection**: Uses `reflect` to determine model names for error messages
- **Learning curve**: Generics syntax may be unfamiliar to some developers
- **Abstraction overhead**: Additional layer between handlers and GORM

### Mitigations
- Clear documentation with examples
- Model name caching to minimize reflection overhead
- Keep domain methods simple and delegate to GORM when needed
