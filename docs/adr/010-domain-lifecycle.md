# ADR-010: Domain Lifecycle Management

## Status

Accepted

## Context

Domains may hold resources that need cleanup on shutdown:
- Background workers (e.g., email queue processors)
- Open connections (e.g., external service clients)
- Cached data that should be flushed
- In-flight operations that should complete gracefully

Without proper lifecycle management:
- Resources leak on shutdown
- In-flight requests may fail unexpectedly
- Dependent services may not be notified of shutdown
- Data may be lost or corrupted

## Decision

### Shutdowner Interface

Domains that need cleanup implement an optional `Shutdowner` interface:

```go
// Shutdowner is an optional interface for domains that need cleanup on shutdown.
type Shutdowner interface {
    Shutdown(ctx context.Context) error
}
```

**Optional Implementation**: Not all domains need shutdown logic. Domains that only register routes and services can skip implementing `Shutdowner`.

### LIFO Shutdown Order

Domains are shut down in **reverse registration order** (Last-In-First-Out):

```go
func (c *Container) Shutdown(ctx context.Context) error {
    var errs []error

    // Shutdown in reverse order (last registered, first shutdown)
    for i := len(c.domains) - 1; i >= 0; i-- {
        d := c.domains[i]
        if shutdowner, ok := d.(Shutdowner); ok {
            if err := shutdowner.Shutdown(ctx); err != nil {
                errs = append(errs, fmt.Errorf("domain %q shutdown failed: %w", d.Name(), err))
            }
        }
    }

    return errors.Join(errs...)
}
```

**Why LIFO?**
- Domains registered later may depend on domains registered earlier
- Example: `auth` depends on `user`, so `auth` should shut down before `user`
- Mirrors constructor/destructor patterns in object-oriented systems

### Shutdown Flow

```
                 Signal (SIGINT/SIGTERM)
                          │
                          ▼
              ┌───────────────────────┐
              │  Create shutdown ctx  │
              │  (with timeout)       │
              └───────────────────────┘
                          │
                          ▼
              ┌───────────────────────┐
              │  Stop accepting new   │
              │  HTTP requests        │
              └───────────────────────┘
                          │
                          ▼
              ┌───────────────────────┐
              │  Wait for in-flight   │
              │  requests to complete │
              └───────────────────────┘
                          │
                          ▼
              ┌───────────────────────┐
              │  container.Shutdown() │
              │  (domains LIFO)       │
              └───────────────────────┘
                          │
                          ▼
              ┌───────────────────────┐
              │  Close Redis conn     │
              └───────────────────────┘
                          │
                          ▼
              ┌───────────────────────┐
              │  Close DB connection  │
              └───────────────────────┘
                          │
                          ▼
                        Exit
```

### Error Aggregation

All shutdown errors are collected and returned as a single combined error using `errors.Join()`:

```go
// Multiple errors are joined, not short-circuited
return errors.Join(errs...)
```

This ensures:
- All domains attempt shutdown even if some fail
- Operators see all failures in logs
- Cleanup proceeds as far as possible

### Shutdown Timeout

A context with timeout is passed to `Shutdown()`:

```go
shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

if err := container.Shutdown(shutdownCtx); err != nil {
    slog.Error("domain shutdown error", "error", err)
}
```

Domains should respect the context deadline:
- Cancel long-running operations when context expires
- Use `select` with `ctx.Done()` for blocking operations
- Return early if deadline exceeded

### Implementation Example

```go
// Domain with a background worker
type domain struct {
    worker     *Worker
    workerDone chan struct{}
}

func (d *domain) Register(c *container.Container) {
    d.workerDone = make(chan struct{})
    d.worker = NewWorker(...)
    go func() {
        defer close(d.workerDone)
        d.worker.Run()
    }()
}

// Implement Shutdowner for cleanup
func (d *domain) Shutdown(ctx context.Context) error {
    // Signal worker to stop
    d.worker.Stop()

    // Wait for worker to finish or timeout
    select {
    case <-d.workerDone:
        return nil
    case <-ctx.Done():
        return fmt.Errorf("worker shutdown timeout: %w", ctx.Err())
    }
}
```

## Consequences

### Positive

- **Graceful Shutdown**: In-flight operations can complete
- **Resource Cleanup**: No resource leaks on shutdown
- **Dependency Ordering**: LIFO ensures correct shutdown sequence
- **Comprehensive Errors**: All failures are reported
- **Optional**: Simple domains don't need implementation

### Negative

- **Shutdown Delay**: Graceful shutdown takes time (up to timeout)
- **Interface Check**: Runtime type assertion to detect Shutdowner
- **Timeout Risk**: Hung operations may block until timeout

### Neutral

- **Force Kill**: Users can still SIGKILL if graceful shutdown hangs
- **Complexity**: Only domains with resources need to implement

## Related

- ADR-001: Domain-Driven Architecture (domain structure)
- ADR-005: Type-Safe Dependency Injection (container pattern)
