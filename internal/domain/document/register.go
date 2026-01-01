package document

import (
	"github.com/gofiber/fiber/v2"
	"github.com/voidmaindev/GoTemplate/internal/container"
	"github.com/voidmaindev/GoTemplate/internal/domain/city"
	"github.com/voidmaindev/GoTemplate/internal/domain/item"
	"github.com/voidmaindev/GoTemplate/internal/domain/user"
	"github.com/voidmaindev/GoTemplate/internal/middleware"
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
func (d *domain) Models() []interface{} {
	return []interface{}{
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
	cityRepo := c.MustGet(city.RepositoryKey).(city.Repository)
	productRepo := c.MustGet(item.RepositoryKey).(item.Repository)

	// Initialize service with cross-domain dependencies
	service := NewService(repo, itemRepo, cityRepo, productRepo)
	c.Set(ServiceKey, service)

	// Initialize handler
	handler := NewHandler(service)
	c.Set(HandlerKey, handler)
}

// Routes registers HTTP routes for this domain
func (d *domain) Routes(api fiber.Router, c *container.Container) {
	handler := c.MustGet(HandlerKey).(*Handler)
	tokenStore := c.MustGet(user.TokenStoreKey).(*user.TokenStore)
	jwtConfig := &c.Config.JWT

	documents := api.Group("/documents", middleware.JWTMiddleware(jwtConfig, tokenStore))

	// Document CRUD
	documents.Post("/", handler.Create)
	documents.Get("/", handler.List)
	documents.Get("/:id", handler.GetByID)
	documents.Put("/:id", handler.Update)
	documents.Delete("/:id", handler.Delete)

	// Document items
	documents.Post("/:id/items", handler.AddItem)
	documents.Put("/:id/items/:itemId", handler.UpdateItem)
	documents.Delete("/:id/items/:itemId", handler.RemoveItem)
}
