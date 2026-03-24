# Adding a New Domain

A domain is a self-contained module with its own model, repository, service, handler, and routes. This guide walks through creating a complete domain with CRUD operations.

## File Structure

```
internal/domain/<name>/
├── model.go       # GORM model
├── dto.go         # Request/Response types
├── errors.go      # Domain-specific errors
├── repository.go  # Data access layer
├── service.go     # Business logic
├── handler.go     # HTTP handlers
├── register.go    # DI and route registration
├── specs.go       # Query specifications (optional)
└── validation.go  # Custom validation (optional)
```

## What Lives Inside vs. Outside the Domain Folder

Almost everything happens inside `internal/domain/<name>/`. There is exactly **one mandatory external change**:

| File | Change | Why |
|------|--------|-----|
| `internal/app/main.go` | Add import + `yourDomain.NewDomain()` to `mainDomains()` | Go has no package auto-discovery; this 2-line change registers the domain |

Everything else is self-contained:
- **Routes** self-register via the `Routes()` method
- **Tables** auto-create via `Models()` + GORM AutoMigrate on startup
- **DI wiring** happens in `Register()` using typed container keys
- **Cross-domain deps** resolved via container keys (e.g., `country.RepositoryKey.MustGet(c)`)

### Optional external changes (advanced cases only)

| File | When needed |
|------|-------------|
| `internal/app/geography.go` (or other app files) | Domain should also run in a subset app |
| `internal/database/migrations/` | Schema *evolution* (column renames, data migrations) — NOT needed for initial table creation |
| `internal/database/seeders/` | Domain needs seed data (e.g., default roles, test fixtures) |

## Step-by-Step Guide

### 1. Create the Model (`model.go`)

```go
package product

import (
    "github.com/voidmaindev/go-template/internal/common"
    "github.com/voidmaindev/go-template/internal/common/filter"
)

type Product struct {
    common.BaseModel
    Name        string `gorm:"size:200;not null" json:"name"`
    Description string `gorm:"type:text" json:"description"`
    Price       int64  `gorm:"not null;default:0" json:"price"`
    SKU         string `gorm:"size:50;uniqueIndex" json:"sku"`
}

func (Product) TableName() string {
    return "products"
}

// FilterConfig enables filtering/sorting on list endpoints
func (Product) FilterConfig() filter.Config {
    return filter.Config{
        TableName: "products",
        Fields: map[string]filter.FieldConfig{
            "id":         {DBColumn: "id", Type: filter.TypeNumber, Operators: filter.NumberOps, Sortable: true},
            "name":       {DBColumn: "name", Type: filter.TypeString, Operators: filter.StringOps, Sortable: true},
            "price":      {DBColumn: "price", Type: filter.TypeNumber, Operators: filter.NumberOps, Sortable: true},
            "sku":        {DBColumn: "sku", Type: filter.TypeString, Operators: filter.StringOps, Sortable: true},
            "created_at": {DBColumn: "created_at", Type: filter.TypeDate, Operators: filter.DateOps, Sortable: true},
            "updated_at": {DBColumn: "updated_at", Type: filter.TypeDate, Operators: filter.DateOps, Sortable: true},
        },
    }
}
```

### 2. Create DTOs (`dto.go`)

```go
package product

import "time"

type CreateProductRequest struct {
    Name        string `json:"name" validate:"required,min=1,max=200"`
    Description string `json:"description" validate:"omitempty,max=5000"`
    Price       int64  `json:"price" validate:"gte=0"`
    SKU         string `json:"sku" validate:"required,min=1,max=50"`
}

type UpdateProductRequest struct {
    Name        *string `json:"name" validate:"omitempty,min=1,max=200"`
    Description *string `json:"description" validate:"omitempty,max=5000"`
    Price       *int64  `json:"price" validate:"omitempty,gte=0"`
}

type ProductResponse struct {
    ID          uint      `json:"id"`
    Name        string    `json:"name"`
    Description string    `json:"description"`
    Price       int64     `json:"price"`
    SKU         string    `json:"sku"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}

func (p *Product) ToResponse() *ProductResponse {
    return &ProductResponse{
        ID:          p.ID,
        Name:        p.Name,
        Description: p.Description,
        Price:       p.Price,
        SKU:         p.SKU,
        CreatedAt:   p.CreatedAt,
        UpdatedAt:   p.UpdatedAt,
    }
}
```

### 3. Create Domain Errors (`errors.go`)

```go
package product

import "github.com/voidmaindev/go-template/internal/common/errors"

const domainName = "product"

var (
    ErrProductNotFound  = errors.NotFound(domainName, "product")
    ErrProductSKUExists = errors.AlreadyExists(domainName, "product", "sku")
)
```

### 4. Create Repository (`repository.go`)

```go
package product

import (
    "context"

    "github.com/voidmaindev/go-template/internal/common"
    "gorm.io/gorm"
)

type Repository interface {
    common.Repository[Product]
    FindBySKU(ctx context.Context, sku string) (*Product, error)
}

type repository struct {
    *common.BaseRepository[Product]
}

func NewRepository(db *gorm.DB) Repository {
    return &repository{
        BaseRepository: common.NewBaseRepository[Product](db, "product"),
    }
}

func (r *repository) FindBySKU(ctx context.Context, sku string) (*Product, error) {
    return r.FindOne(ctx, map[string]any{"sku": sku})
}
```

### 5. Create Service (`service.go`)

For simple services with 1-3 dependencies, use direct constructor parameters:

```go
package product

import (
    "context"

    "github.com/voidmaindev/go-template/internal/common"
    "github.com/voidmaindev/go-template/internal/common/errors"
    "github.com/voidmaindev/go-template/internal/common/filter"
    "github.com/voidmaindev/go-template/pkg/utils"
)

type Service interface {
    GetByID(ctx context.Context, id uint) (*Product, error)
    Create(ctx context.Context, req *CreateProductRequest) (*Product, error)
    Update(ctx context.Context, id uint, req *UpdateProductRequest) (*Product, error)
    Delete(ctx context.Context, id uint) error
    ListFiltered(ctx context.Context, params *filter.Params) (*common.PaginatedResult[Product], error)
}

type service struct {
    repo Repository
}

func NewService(repo Repository) Service {
    return &service{repo: repo}
}
```

For services with 4+ dependencies, use a config struct pattern for readability:

```go
// ServiceConfig holds all dependencies for the service.
type ServiceConfig struct {
    Repo           Repository
    Cache          CacheService
    EventPublisher EventPublisher
    Config         *config.ProductConfig
}

type service struct {
    repo           Repository
    cache          CacheService
    eventPublisher EventPublisher
    config         *config.ProductConfig
}

func NewService(cfg ServiceConfig) Service {
    return &service{
        repo:           cfg.Repo,
        cache:          cfg.Cache,
        eventPublisher: cfg.EventPublisher,
        config:         cfg.Config,
    }
}
```

Service method implementations:

```go
func (s *service) ListFiltered(ctx context.Context, params *filter.Params) (*common.PaginatedResult[Product], error) {
    products, total, err := s.repo.FindAllFiltered(ctx, params)
    if err != nil {
        return nil, errors.Internal(domainName, err).WithOperation("ListFiltered")
    }
    return common.NewPaginatedResultFromFilter(products, total, params), nil
}

func (s *service) GetByID(ctx context.Context, id uint) (*Product, error) {
    product, err := s.repo.FindByID(ctx, id)
    if err != nil {
        if errors.IsNotFound(err) {
            return nil, ErrProductNotFound
        }
        return nil, errors.Internal(domainName, err).WithOperation("GetByID")
    }
    return product, nil
}

func (s *service) Create(ctx context.Context, req *CreateProductRequest) (*Product, error) {
    // Check SKU uniqueness
    existing, _ := s.repo.FindBySKU(ctx, req.SKU)
    if existing != nil {
        return nil, ErrProductSKUExists
    }

    product := &Product{
        Name:        req.Name,
        Description: req.Description,
        Price:       req.Price,
        SKU:         req.SKU,
    }

    if err := s.repo.Create(ctx, product); err != nil {
        return nil, errors.Internal(domainName, err).WithOperation("Create")
    }
    return product, nil
}

func (s *service) Update(ctx context.Context, id uint, req *UpdateProductRequest) (*Product, error) {
    product, err := s.repo.FindByID(ctx, id)
    if err != nil {
        if errors.IsNotFound(err) {
            return nil, ErrProductNotFound
        }
        return nil, errors.Internal(domainName, err).WithOperation("Update")
    }

    // Map non-nil fields from request to model
    if err := utils.UpdateModel(product, req); err != nil {
        return nil, errors.Internal(domainName, err).WithOperation("Update")
    }

    if err := s.repo.Update(ctx, product); err != nil {
        return nil, errors.Internal(domainName, err).WithOperation("Update")
    }
    return product, nil
}

func (s *service) Delete(ctx context.Context, id uint) error {
    product, err := s.repo.FindByID(ctx, id)
    if err != nil {
        if errors.IsNotFound(err) {
            return ErrProductNotFound
        }
        return errors.Internal(domainName, err).WithOperation("Delete")
    }

    if err := s.repo.Delete(ctx, product.ID); err != nil {
        return errors.Internal(domainName, err).WithOperation("Delete")
    }
    return nil
}
```

### 6. Create Handler (`handler.go`)

```go
package product

import (
    "github.com/gofiber/fiber/v2"
    "github.com/voidmaindev/go-template/internal/common"
    "github.com/voidmaindev/go-template/internal/common/filter"
)

type Handler struct {
    service Service
}

func NewHandler(service Service) *Handler {
    return &Handler{service: service}
}

func (h *Handler) List(c *fiber.Ctx) error {
    params := filter.ParseFromQuery(c)

    result, err := h.service.ListFiltered(c.Context(), params)
    if err != nil {
        return common.HandleError(c, err)
    }

    return common.SuccessResponse(c, common.MapPaginatedResult(result, func(p Product) ProductResponse {
        return *p.ToResponse()
    }))
}

func (h *Handler) GetByID(c *fiber.Ctx) error {
    id, err := common.ParseID(c, "id", domainName)
    if err != nil {
        return nil // response already sent
    }

    product, err := h.service.GetByID(c.Context(), id)
    if err != nil {
        return common.HandleError(c, err)
    }
    return common.SuccessResponse(c, product.ToResponse())
}

func (h *Handler) Create(c *fiber.Ctx) error {
    req, err := common.ParseAndValidate[CreateProductRequest](c)
    if err != nil {
        return nil // response already sent
    }

    product, err := h.service.Create(c.Context(), req)
    if err != nil {
        return common.HandleError(c, err)
    }
    return common.CreatedResponse(c, product.ToResponse())
}

func (h *Handler) Update(c *fiber.Ctx) error {
    id, err := common.ParseID(c, "id", domainName)
    if err != nil {
        return nil // response already sent
    }

    req, err := common.ParseAndValidate[UpdateProductRequest](c)
    if err != nil {
        return nil // response already sent
    }

    product, err := h.service.Update(c.Context(), id, req)
    if err != nil {
        return common.HandleError(c, err)
    }
    return common.SuccessResponse(c, product.ToResponse())
}

func (h *Handler) Delete(c *fiber.Ctx) error {
    id, err := common.ParseID(c, "id", domainName)
    if err != nil {
        return nil // response already sent
    }

    if err := h.service.Delete(c.Context(), id); err != nil {
        return common.HandleError(c, err)
    }
    return common.DeletedResponse(c)
}
```

### 7. Create Registration (`register.go`)

```go
package product

import (
    "github.com/gofiber/fiber/v2"
    "github.com/voidmaindev/go-template/internal/container"
    "github.com/voidmaindev/go-template/internal/domain/rbac"
    "github.com/voidmaindev/go-template/internal/domain/user"
    "github.com/voidmaindev/go-template/internal/middleware"
)

// Component keys for dependency injection
var (
    RepositoryKey = container.Key[Repository]("product.repository")
    ServiceKey    = container.Key[Service]("product.service")
    HandlerKey    = container.Key[*Handler]("product.handler")
)

type domain struct{}

func NewDomain() container.Domain {
    return &domain{}
}

func (d *domain) Name() string {
    return "product"
}

func (d *domain) Models() []any {
    return []any{&Product{}}
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
    tokenStore := user.TokenStoreKey.MustGet(c)
    enforcer := rbac.EnforcerKey.MustGet(c)
    rateLimiter := middleware.RateLimiterFactoryKey.MustGet(c)
    jwtConfig := &c.Config.JWT

    products := api.Group("/products", middleware.JWTMiddlewareWithInvalidator(jwtConfig, tokenStore, tokenStore))

    // Read endpoints
    products.Get("/", rateLimiter.ForTier(middleware.TierAPIRead),
        middleware.RequirePermission(enforcer, "product", rbac.ActionRead), handler.List)
    products.Get("/:id", rateLimiter.ForTier(middleware.TierAPIRead),
        middleware.RequirePermission(enforcer, "product", rbac.ActionRead), handler.GetByID)

    // Write endpoints
    products.Post("/", rateLimiter.ForTier(middleware.TierAPIWrite),
        middleware.RequirePermission(enforcer, "product", rbac.ActionCreate), handler.Create)
    products.Put("/:id", rateLimiter.ForTier(middleware.TierAPIWrite),
        middleware.RequirePermission(enforcer, "product", rbac.ActionUpdate), handler.Update)
    products.Delete("/:id", rateLimiter.ForTier(middleware.TierAPIWrite),
        middleware.RequirePermission(enforcer, "product", rbac.ActionDelete), handler.Delete)
}
```

### 8. Add Domain to App

Edit `internal/app/main.go` (the only file outside your domain folder):

```go
import (
    // ... existing imports
    "github.com/voidmaindev/go-template/internal/domain/product"
)

func mainDomains() []container.Domain {
    return []container.Domain{
        rbac.NewDomain(),
        user.NewDomain(),
        item.NewDomain(),
        product.NewDomain(),  // Add here
        // ...
    }
}
```

### 9. Run Migrations

```bash
go run ./cmd/api serve main
# Migrations run automatically on startup via GORM AutoMigrate
```

## Optional Files

### Query Specifications (`specs.go`)

Specs must implement both `Apply` and `ApplyGorm`, include constructor functions, and have compile-time interface checks:

```go
package product

import (
    "github.com/voidmaindev/go-template/internal/common"
    "gorm.io/gorm"
)

// BySKU finds product by exact SKU
type BySKUSpec struct {
    SKU string
}

func BySKU(sku string) BySKUSpec {
    return BySKUSpec{SKU: sku}
}

func (s BySKUSpec) Apply(query any) any {
    return s.ApplyGorm(common.AsGormDB(query))
}

func (s BySKUSpec) ApplyGorm(db *gorm.DB) *gorm.DB {
    return db.Where("sku = ?", s.SKU)
}

// ByPriceRange finds products within a price range
type PriceRangeSpec struct {
    Min, Max int64
}

func ByPriceRange(min, max int64) PriceRangeSpec {
    return PriceRangeSpec{Min: min, Max: max}
}

func (s PriceRangeSpec) Apply(query any) any {
    return s.ApplyGorm(common.AsGormDB(query))
}

func (s PriceRangeSpec) ApplyGorm(db *gorm.DB) *gorm.DB {
    return db.Where("price BETWEEN ? AND ?", s.Min, s.Max)
}

// Ensure all specs implement GormSpecification at compile time
var (
    _ common.GormSpecification = BySKUSpec{}
    _ common.GormSpecification = PriceRangeSpec{}
)
```

### Custom Validation (`validation.go`)

```go
package product

import (
    "context"

    "github.com/voidmaindev/go-template/internal/common/validation"
)

type Validator struct {
    validator *validation.Validator
}

func NewValidator() *Validator {
    return &Validator{
        validator: validation.New(),
    }
}

func (v *Validator) ValidateCreate(ctx context.Context, req *CreateProductRequest) *validation.Result {
    return validation.NewBuilder(ctx).
        StructWith(v.validator, req).
        Result()
}

func (v *Validator) ValidateUpdate(ctx context.Context, req *UpdateProductRequest) *validation.Result {
    return validation.NewBuilder(ctx).
        StructWith(v.validator, req).
        Result()
}
```

## Advanced Patterns

### Cross-Domain Dependencies

When your domain depends on another domain, resolve it in `Register()` via typed container keys. Example from `city` depending on `country`:

```go
import "github.com/voidmaindev/go-template/internal/domain/country"

func (d *domain) Register(c *container.Container) {
    repo := NewRepository(c.DB)
    RepositoryKey.Set(c, repo)

    // Get dependency from another domain
    countryRepo := country.RepositoryKey.MustGet(c)

    // Pass cross-domain dependency to service
    service := NewService(repo, countryRepo)
    ServiceKey.Set(c, service)

    handler := NewHandler(service)
    HandlerKey.Set(c, handler)
}
```

**Important:** Registration order in `internal/app/main.go` must respect dependencies — `country` must appear before `city` in the slice.

### Nested Routes

A domain can register routes under another domain's path. Example from `city`:

```go
func (d *domain) Routes(api fiber.Router, c *container.Container) {
    // ... standard city routes at /cities ...

    // Nested route under countries
    countries := api.Group("/countries", middleware.JWTMiddlewareWithInvalidator(jwtConfig, tokenStore, tokenStore))
    countries.Get("/:countryId/cities", rateLimiter.ForTier(middleware.TierAPIRead),
        middleware.RequirePermission(enforcer, "city", rbac.ActionRead), handler.ListByCountry)
}
```

### Optional Domain Interfaces

Domains can implement additional interfaces beyond the required `container.Domain`:

**`DependencyDeclarer`** — declares dependencies for startup validation:

```go
func (d *domain) Dependencies() []string {
    return []string{"country"} // must be registered before this domain
}
```

**`Shutdowner`** — performs cleanup during graceful shutdown:

```go
func (d *domain) Shutdown(ctx context.Context) error {
    // Close connections, flush buffers, etc.
    return nil
}
```

### Multiple Models per Domain

A domain can manage multiple related tables by returning them all from `Models()`:

```go
func (d *domain) Models() []any {
    return []any{&Document{}, &DocumentItem{}}
}
```

### Domains Without Routes or Models

Not all domains need routes or models. For example, the `email` domain provides only a service:

```go
func (d *domain) Models() []any { return nil }
func (d *domain) Routes(api fiber.Router, c *container.Container) {} // no-op
```

## Checklist

- [ ] Created `internal/domain/<name>/` folder
- [ ] Created `model.go` with GORM model, `TableName()`, and `FilterConfig()` (including `updated_at`)
- [ ] Created `dto.go` with request/response types and `ToResponse()` method
- [ ] Created `errors.go` with domain-specific errors
- [ ] Created `repository.go` with `FindOne(ctx, map[string]any{...})` pattern
- [ ] Created `service.go` with `utils.UpdateModel` and `.WithOperation()` on all errors
- [ ] Created `handler.go` with HTTP handlers
- [ ] Created `register.go` with DI and routes
- [ ] Added domain to `internal/app/main.go` (the one external file)
- [ ] Tested CRUD operations
