package document

import (
	"github.com/gofiber/fiber/v2"
	"github.com/voidmaindev/go-template/internal/container"
	"github.com/voidmaindev/go-template/internal/domain/city"
	"github.com/voidmaindev/go-template/internal/domain/item"
	"github.com/voidmaindev/go-template/internal/domain/rbac"
	"github.com/voidmaindev/go-template/internal/domain/user"
	"github.com/voidmaindev/go-template/internal/middleware"
)

// Component keys for this domain (typed for compile-time safety)
var (
	RepositoryKey     = container.Key[Repository]("document.repository")
	ItemRepositoryKey = container.Key[ItemRepository]("document.itemRepository")
	ServiceKey        = container.Key[Service]("document.service")
	HandlerKey        = container.Key[*Handler]("document.handler")
)

// domain implements container.Domain interface
type domain struct{}

// NewDomain creates a new document domain for registration
func NewDomain() container.Domain {
	return &domain{}
}

// Name returns the domain name
func (d *domain) Name() string {
	return "document"
}

// Models returns the GORM models for migration
func (d *domain) Models() []any {
	return []any{
		&Document{},
		&DocumentItem{},
	}
}

// Register initializes repositories, services, and handlers
func (d *domain) Register(c *container.Container) {
	// Initialize repositories
	repo := NewRepository(c.DB)
	RepositoryKey.Set(c, repo)

	itemRepo := NewItemRepository(c.DB)
	ItemRepositoryKey.Set(c, itemRepo)

	// Get cross-domain dependencies
	cityRepo := city.RepositoryKey.MustGet(c)
	productRepo := item.RepositoryKey.MustGet(c)

	// Initialize service with cross-domain dependencies
	service := NewService(repo, itemRepo, cityRepo, productRepo)
	ServiceKey.Set(c, service)

	// Initialize handler
	handler := NewHandler(service)
	HandlerKey.Set(c, handler)
}

// Routes registers HTTP routes for this domain
func (d *domain) Routes(api fiber.Router, c *container.Container) {
	handler := HandlerKey.MustGet(c)
	tokenStore := user.TokenStoreKey.MustGet(c)
	enforcer := rbac.EnforcerKey.MustGet(c)
	rateLimiter := middleware.RateLimiterFactoryKey.MustGet(c)
	jwtConfig := &c.Config.JWT

	documents := api.Group("/documents", middleware.JWTMiddleware(jwtConfig, tokenStore))

	// Document operations - GET endpoints (api_read tier)
	documents.Get("/", rateLimiter.ForTier(middleware.TierAPIRead), middleware.RequirePermission(enforcer, "document", rbac.ActionRead), handler.List)
	documents.Get("/:id", rateLimiter.ForTier(middleware.TierAPIRead), middleware.RequirePermission(enforcer, "document", rbac.ActionRead), handler.GetByID)
	// Document operations - Write endpoints (api_write tier)
	documents.Post("/", rateLimiter.ForTier(middleware.TierAPIWrite), middleware.RequirePermission(enforcer, "document", rbac.ActionCreate), handler.Create)
	documents.Put("/:id", rateLimiter.ForTier(middleware.TierAPIWrite), middleware.RequirePermission(enforcer, "document", rbac.ActionUpdate), handler.Update)
	documents.Delete("/:id", rateLimiter.ForTier(middleware.TierAPIWrite), middleware.RequirePermission(enforcer, "document", rbac.ActionDelete), handler.Delete)

	// Document items (api_write tier - all are mutations)
	documents.Post("/:id/items", rateLimiter.ForTier(middleware.TierAPIWrite), middleware.RequirePermission(enforcer, "document", rbac.ActionUpdate), handler.AddItem)
	documents.Put("/:id/items/:itemId", rateLimiter.ForTier(middleware.TierAPIWrite), middleware.RequirePermission(enforcer, "document", rbac.ActionUpdate), handler.UpdateItem)
	documents.Delete("/:id/items/:itemId", rateLimiter.ForTier(middleware.TierAPIWrite), middleware.RequirePermission(enforcer, "document", rbac.ActionUpdate), handler.RemoveItem)
}
