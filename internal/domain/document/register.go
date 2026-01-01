package document

import (
	"github.com/gofiber/fiber/v2"
	"github.com/voidmaindev/GoTemplate/internal/container"
	"github.com/voidmaindev/GoTemplate/internal/domain/city"
	"github.com/voidmaindev/GoTemplate/internal/domain/item"
	"github.com/voidmaindev/GoTemplate/internal/domain/user"
	"github.com/voidmaindev/GoTemplate/internal/middleware"
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
	c.Set(container.DocumentRepository, repo)

	itemRepo := NewItemRepository(c.DB)
	c.Set(container.DocumentItemRepository, itemRepo)

	// Get cross-domain dependencies
	cityRepo := c.MustGet(container.CityRepository).(city.Repository)
	productRepo := c.MustGet(container.ItemRepository).(item.Repository)

	// Initialize service with cross-domain dependencies
	service := NewService(repo, itemRepo, cityRepo, productRepo)
	c.Set(container.DocumentService, service)

	// Initialize handler
	handler := NewHandler(service)
	c.Set(container.DocumentHandler, handler)
}

// Routes registers HTTP routes for this domain
func (d *domain) Routes(api fiber.Router, c *container.Container) {
	handler := c.MustGet(container.DocumentHandler).(*Handler)
	tokenStore := c.MustGet(container.TokenStore).(*user.TokenStore)
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
