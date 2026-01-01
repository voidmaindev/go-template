package item

import (
	"github.com/gofiber/fiber/v2"
	"github.com/voidmaindev/GoTemplate/internal/container"
	"github.com/voidmaindev/GoTemplate/internal/domain/user"
	"github.com/voidmaindev/GoTemplate/internal/middleware"
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
	handler := c.MustGet(HandlerKey).(*Handler)
	tokenStore := c.MustGet(user.TokenStoreKey).(*user.TokenStore)
	jwtConfig := &c.Config.JWT

	items := api.Group("/items", middleware.JWTMiddleware(jwtConfig, tokenStore))
	items.Post("/", handler.Create)
	items.Get("/", handler.List)
	items.Get("/:id", handler.GetByID)
	items.Put("/:id", handler.Update)
	items.Delete("/:id", handler.Delete)
}
