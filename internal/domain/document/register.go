package document

import (
	"github.com/casbin/casbin/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/voidmaindev/go-template/internal/container"
	"github.com/voidmaindev/go-template/internal/domain/city"
	"github.com/voidmaindev/go-template/internal/domain/item"
	"github.com/voidmaindev/go-template/internal/domain/rbac"
	"github.com/voidmaindev/go-template/internal/domain/user"
	"github.com/voidmaindev/go-template/internal/middleware"
)

// Component keys for this domain
const (
	RepositoryKey     = "document.repository"
	ItemRepositoryKey = "document.itemRepository"
	ServiceKey        = "document.service"
	HandlerKey        = "document.handler"
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
	c.Set(RepositoryKey, repo)

	itemRepo := NewItemRepository(c.DB)
	c.Set(ItemRepositoryKey, itemRepo)

	// Get cross-domain dependencies
	cityRepo := container.MustGetTyped[city.Repository](c, city.RepositoryKey)
	productRepo := container.MustGetTyped[item.Repository](c, item.RepositoryKey)

	// Initialize service with cross-domain dependencies
	service := NewService(repo, itemRepo, cityRepo, productRepo)
	c.Set(ServiceKey, service)

	// Initialize handler
	handler := NewHandler(service)
	c.Set(HandlerKey, handler)
}

// Routes registers HTTP routes for this domain
func (d *domain) Routes(api fiber.Router, c *container.Container) {
	handler := container.MustGetTyped[*Handler](c, HandlerKey)
	tokenStore := container.MustGetTyped[*user.TokenStore](c, user.TokenStoreKey)
	enforcer := container.MustGetTyped[*casbin.Enforcer](c, rbac.EnforcerKey)
	jwtConfig := &c.Config.JWT

	documents := api.Group("/documents", middleware.JWTMiddleware(jwtConfig, tokenStore))

	// Document operations
	documents.Get("/", middleware.RequirePermission(enforcer, "document", rbac.ActionRead), handler.List)
	documents.Get("/:id", middleware.RequirePermission(enforcer, "document", rbac.ActionRead), handler.GetByID)
	documents.Post("/", middleware.RequirePermission(enforcer, "document", rbac.ActionWrite), handler.Create)
	documents.Put("/:id", middleware.RequirePermission(enforcer, "document", rbac.ActionModify), handler.Update)
	documents.Delete("/:id", middleware.RequirePermission(enforcer, "document", rbac.ActionDelete), handler.Delete)

	// Document items
	documents.Post("/:id/items", middleware.RequirePermission(enforcer, "document", rbac.ActionModify), handler.AddItem)
	documents.Put("/:id/items/:itemId", middleware.RequirePermission(enforcer, "document", rbac.ActionModify), handler.UpdateItem)
	documents.Delete("/:id/items/:itemId", middleware.RequirePermission(enforcer, "document", rbac.ActionModify), handler.RemoveItem)
}
