package item

import (
	"github.com/gofiber/fiber/v2"
	"github.com/voidmaindev/go-template/internal/container"
	"github.com/voidmaindev/go-template/internal/domain/user"
	"github.com/voidmaindev/go-template/internal/middleware"
)

// Component keys for this domain
const (
	RepositoryKey = "item.repository"
	ServiceKey    = "item.service"
	HandlerKey    = "item.handler"
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
func (d *domain) Models() []any {
	return []any{
		&Item{},
	}
}

// Register initializes repositories, services, and handlers
func (d *domain) Register(c *container.Container) {
	// Initialize repository
	repo := NewRepository(c.DB)
	c.Set(RepositoryKey, repo)

	// Initialize service
	service := NewService(repo)
	c.Set(ServiceKey, service)

	// Initialize handler
	handler := NewHandler(service)
	c.Set(HandlerKey, handler)
}

// Routes registers HTTP routes for this domain
func (d *domain) Routes(api fiber.Router, c *container.Container) {
	handler := container.MustGetTyped[*Handler](c, HandlerKey)
	tokenStore := container.MustGetTyped[*user.TokenStore](c, user.TokenStoreKey)
	jwtConfig := &c.Config.JWT

	items := api.Group("/items", middleware.JWTMiddleware(jwtConfig, tokenStore))
	items.Get("/", handler.List)
	items.Get("/:id", handler.GetByID)
	// Write operations require admin role
	items.Post("/", middleware.RequireAdmin(), handler.Create)
	items.Put("/:id", middleware.RequireAdmin(), handler.Update)
	items.Delete("/:id", middleware.RequireAdmin(), handler.Delete)
}
