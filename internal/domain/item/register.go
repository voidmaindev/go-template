package item

import (
	"github.com/gofiber/fiber/v2"
	"github.com/voidmaindev/GoTemplate/internal/container"
	"github.com/voidmaindev/GoTemplate/internal/domain/user"
	"github.com/voidmaindev/GoTemplate/internal/middleware"
)

// domain implements container.Domain interface
type domain struct{}

// NewDomain creates a new item domain for registration
func NewDomain() container.Domain {
	return &domain{}
}

// Name returns the domain name
func (d *domain) Name() string {
	return "item"
}

// Models returns the GORM models for migration
func (d *domain) Models() []interface{} {
	return []interface{}{
		&Item{},
	}
}

// Register initializes repositories, services, and handlers
func (d *domain) Register(c *container.Container) {
	// Initialize repository
	repo := NewRepository(c.DB)
	c.Set(container.ItemRepository, repo)

	// Initialize service
	service := NewService(repo)
	c.Set(container.ItemService, service)

	// Initialize handler
	handler := NewHandler(service)
	c.Set(container.ItemHandler, handler)
}

// Routes registers HTTP routes for this domain
func (d *domain) Routes(api fiber.Router, c *container.Container) {
	handler := c.MustGet(container.ItemHandler).(*Handler)
	tokenStore := c.MustGet(container.TokenStore).(*user.TokenStore)
	jwtConfig := &c.Config.JWT

	items := api.Group("/items", middleware.JWTMiddleware(jwtConfig, tokenStore))
	items.Post("/", handler.Create)
	items.Get("/", handler.List)
	items.Get("/:id", handler.GetByID)
	items.Put("/:id", handler.Update)
	items.Delete("/:id", handler.Delete)
}
