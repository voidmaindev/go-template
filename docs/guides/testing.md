# Testing Guide

This guide covers testing patterns and best practices for the go-template application.

## Test Types

| Type | Location | Purpose |
|------|----------|---------|
| Unit Tests | `*_test.go` in domain packages | Test isolated components with mocks |
| Integration Tests | `internal/integration/` | Test full stack with real databases |

## Running Tests

```bash
# Run all tests
go test ./...

# Run unit tests only (skip integration)
go test -short ./...

# Run specific package tests
go test ./internal/domain/example_item/...

# Run with verbose output
go test -v ./...

# Run with coverage
go test -cover ./...
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## Unit Testing

### Mock Repository Pattern

Create in-memory mock repositories for unit tests:

```go
// internal/domain/product/repository_mock_test.go
type mockRepository struct {
    items     map[uint]*Product
    nextID    uint
    findError error  // Inject errors for testing
}

func newMockRepository() *mockRepository {
    return &mockRepository{
        items:  make(map[uint]*Product),
        nextID: 1,
    }
}

func (m *mockRepository) Create(ctx context.Context, product *Product) error {
    if m.findError != nil {
        return m.findError
    }
    product.ID = m.nextID
    m.nextID++
    m.items[product.ID] = product
    return nil
}

func (m *mockRepository) FindByID(ctx context.Context, id uint) (*Product, error) {
    if m.findError != nil {
        return nil, m.findError
    }
    if p, ok := m.items[id]; ok {
        return p, nil
    }
    return nil, gorm.ErrRecordNotFound
}

func (m *mockRepository) FindAll(ctx context.Context, opts ...common.QueryOption) ([]*Product, int64, error) {
    results := make([]*Product, 0, len(m.items))
    for _, p := range m.items {
        results = append(results, p)
    }
    return results, int64(len(results)), nil
}

// Implement other Repository methods...
```

### Service Unit Tests

Test service logic in isolation:

```go
// internal/domain/product/service_test.go
func TestProductService_Create(t *testing.T) {
    repo := newMockRepository()
    svc := NewService(repo)

    req := &CreateProductRequest{
        Name:  "Test Product",
        SKU:   "TEST-001",
        Price: 1000,
    }

    product, err := svc.Create(context.Background(), req)

    testutil.RequireNoError(t, err)
    testutil.RequireEqual(t, "Test Product", product.Name)
    testutil.RequireEqual(t, uint(1), product.ID)
}

func TestProductService_Create_DuplicateSKU(t *testing.T) {
    repo := newMockRepository()
    svc := NewService(repo)

    // Create first product
    req := &CreateProductRequest{Name: "First", SKU: "SKU-001", Price: 100}
    _, err := svc.Create(context.Background(), req)
    testutil.RequireNoError(t, err)

    // Attempt duplicate SKU
    req2 := &CreateProductRequest{Name: "Second", SKU: "SKU-001", Price: 200}
    _, err = svc.Create(context.Background(), req2)

    testutil.RequireError(t, err)
    testutil.RequireTrue(t, errors.IsAlreadyExists(err))
}

func TestProductService_GetByID_NotFound(t *testing.T) {
    repo := newMockRepository()
    svc := NewService(repo)

    _, err := svc.GetByID(context.Background(), 999)

    testutil.RequireError(t, err)
    testutil.RequireTrue(t, errors.IsNotFound(err))
}
```

### Error Injection

Test error handling by injecting errors:

```go
func TestProductService_Create_RepositoryError(t *testing.T) {
    repo := newMockRepository()
    repo.findError = fmt.Errorf("database connection failed")
    svc := NewService(repo)

    req := &CreateProductRequest{Name: "Test", SKU: "TEST", Price: 100}
    _, err := svc.Create(context.Background(), req)

    testutil.RequireError(t, err)
}
```

## Integration Testing

### TestSuite Setup

Use the integration test suite for full-stack tests:

```go
// internal/integration/product_test.go
package integration

import (
    "testing"
    "github.com/voidmaindev/go-template/internal/domain/product"
)

func TestProductIntegration(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test in short mode")
    }

    suite := SetupTestSuite(t)
    defer suite.Cleanup()

    t.Run("Create and retrieve product", func(t *testing.T) {
        svc := product.ServiceKey.MustGet(suite.Container)

        created, err := svc.Create(suite.Ctx, &product.CreateProductRequest{
            Name:  "Integration Test Product",
            SKU:   "INT-TEST-001",
            Price: 5000,
        })
        testutil.RequireNoError(t, err)

        retrieved, err := svc.GetByID(suite.Ctx, created.ID)
        testutil.RequireNoError(t, err)
        testutil.RequireEqual(t, created.ID, retrieved.ID)
        testutil.RequireEqual(t, "Integration Test Product", retrieved.Name)
    })
}
```

### Testcontainers

The suite automatically provisions PostgreSQL and Redis containers:

```go
// internal/testutil/containers.go

// SetupIntegrationTest creates both PostgreSQL and Redis containers
func SetupIntegrationTest(t *testing.T) (*gorm.DB, *redis.Client, func()) {
    db, dbCleanup := SetupPostgresOnly(t)
    redisClient, redisCleanup := SetupRedisOnly(t)

    cleanup := func() {
        redisCleanup()
        dbCleanup()
    }

    return db, redisClient, cleanup
}

// SetupPostgresOnly creates a PostgreSQL container
func SetupPostgresOnly(t *testing.T) (*gorm.DB, func()) {
    ctx := context.Background()

    req := testcontainers.ContainerRequest{
        Image:        "postgres:16-alpine",
        ExposedPorts: []string{"5432/tcp"},
        Env: map[string]string{
            "POSTGRES_USER":     "test",
            "POSTGRES_PASSWORD": "test",
            "POSTGRES_DB":       "test",
        },
        WaitingFor: wait.ForLog("database system is ready to accept connections").
            WithOccurrence(2).
            WithStartupTimeout(60 * time.Second),
    }

    container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
        ContainerRequest: req,
        Started:          true,
    })
    require.NoError(t, err)

    // Get connection details and create GORM connection...

    t.Cleanup(func() {
        container.Terminate(ctx)
    })

    return db, cleanup
}
```

### Test Fixtures

Use fixtures for consistent test data:

```go
// internal/testutil/fixtures.go

// FixtureBuilder provides fluent API for creating test data
type FixtureBuilder struct {
    db  *gorm.DB
    ctx context.Context
}

func NewFixtureBuilder(db *gorm.DB) *FixtureBuilder {
    return &FixtureBuilder{db: db, ctx: context.Background()}
}

func (f *FixtureBuilder) CreateUser(overrides ...func(*user.User)) *user.User {
    u := &user.User{
        Email:    fmt.Sprintf("test-%s@example.com", uuid.New().String()[:8]),
        Password: "$2a$10$...",  // Pre-hashed "password123"
        Name:     "Test User",
    }
    for _, override := range overrides {
        override(u)
    }
    f.db.Create(u)
    return u
}

func (f *FixtureBuilder) CreateProduct(overrides ...func(*product.Product)) *product.Product {
    p := &product.Product{
        Name:  fmt.Sprintf("Product-%s", uuid.New().String()[:8]),
        SKU:   fmt.Sprintf("SKU-%s", uuid.New().String()[:8]),
        Price: 1000,
    }
    for _, override := range overrides {
        override(p)
    }
    f.db.Create(p)
    return p
}
```

### Table Cleanup

Clean tables between tests:

```go
func (s *TestSuite) CleanupTables(tables ...string) {
    for _, table := range tables {
        s.DB.Exec(fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table))
    }
}

// Usage in test
func TestSomething(t *testing.T) {
    suite := SetupTestSuite(t)
    defer suite.Cleanup()

    suite.CleanupTables("products", "orders")

    // Test with clean slate...
}
```

## Test Utilities

### Custom Assertions

Use `testutil` for consistent assertions:

```go
import "github.com/voidmaindev/go-template/internal/testutil"

// Require (fails immediately)
testutil.RequireNoError(t, err)
testutil.RequireError(t, err)
testutil.RequireEqual(t, expected, actual)
testutil.RequireNotEqual(t, unexpected, actual)
testutil.RequireNil(t, value)
testutil.RequireNotNil(t, value)
testutil.RequireTrue(t, condition)
testutil.RequireFalse(t, condition)
testutil.RequireLen(t, slice, expectedLen)
testutil.RequireContains(t, slice, item)

// Assert (continues on failure)
testutil.AssertNoError(t, err)
testutil.AssertEqual(t, expected, actual)
testutil.AssertTrue(t, condition)
```

### Test Context

Create test contexts with timeouts:

```go
// Default test context (5 second timeout)
ctx := testutil.TestContext()

// Custom timeout
ctx := testutil.TestContextWithTimeout(10 * time.Second)
```

### Waiting for Async Operations

```go
// Wait for a condition with polling
err := testutil.WaitForCondition(func() bool {
    // Check condition
    result, _ := svc.GetStatus(ctx, id)
    return result.Status == "completed"
}, 5*time.Second, 100*time.Millisecond)
testutil.RequireNoError(t, err)
```

### Test JWT Tokens

Generate tokens for authenticated endpoint tests:

```go
cfg := testutil.TestJWTConfig()

// Generate access token
accessToken, err := testutil.GenerateTestAccessToken(cfg, userID, []string{"admin"})

// Generate refresh token
refreshToken, err := testutil.GenerateTestRefreshToken(cfg, userID)
```

## Best Practices

### Test Organization

```
internal/domain/product/
├── service.go
├── service_test.go      # Unit tests with mocks
├── repository.go
└── repository_mock_test.go  # Mock implementation

internal/integration/
├── setup_test.go        # TestSuite and helpers
├── product_test.go      # Product integration tests
└── user_test.go         # User integration tests
```

### Naming Conventions

```go
// Test function naming
func TestServiceName_MethodName(t *testing.T) { }
func TestServiceName_MethodName_SpecificScenario(t *testing.T) { }

// Examples
func TestProductService_Create(t *testing.T) { }
func TestProductService_Create_DuplicateSKU(t *testing.T) { }
func TestProductService_GetByID_NotFound(t *testing.T) { }
```

### Subtests for Related Cases

```go
func TestProductService_Create(t *testing.T) {
    t.Run("success", func(t *testing.T) {
        // Happy path
    })

    t.Run("duplicate SKU", func(t *testing.T) {
        // Error case
    })

    t.Run("invalid input", func(t *testing.T) {
        // Validation error
    })
}
```

### Test Isolation

- Each test should be independent
- Use unique identifiers (UUIDs) for test data
- Clean up state between tests
- Don't rely on test execution order

## Checklist

- [ ] Write unit tests for service methods
- [ ] Create mock repository implementing full interface
- [ ] Test error paths and edge cases
- [ ] Write integration tests for critical flows
- [ ] Use fixtures for consistent test data
- [ ] Clean tables between integration tests
- [ ] Use `-short` flag to skip integration tests in CI
