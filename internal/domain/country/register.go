package country

import (
	"github.com/gofiber/fiber/v2"
	"github.com/voidmaindev/GoTemplate/internal/container"
	"github.com/voidmaindev/GoTemplate/internal/domain/user"
	"github.com/voidmaindev/GoTemplate/internal/middleware"
)

// Component keys for this domain
const (
	RepositoryKey = "country.repository"
	ServiceKey    = "country.service"
	HandlerKey    = "country.handler"
)

// domain implements container.Domain interface
type domain struct{}

// NewDomain creates a new country domain for registration
func NewDomain() container.Domain {
	return &domain{}
}

// Name returns the domain name
func (d *domain) Name() string {
	return "country"
}

// Models returns the GORM models for migration
func (d *domain) Models() []any {
	return []any{
		&Country{},
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

	countries := api.Group("/countries", middleware.JWTMiddleware(jwtConfig, tokenStore))
	countries.Post("/", handler.Create)
	countries.Get("/", handler.List)
	countries.Get("/:id", handler.GetByID)
	countries.Put("/:id", handler.Update)
	countries.Delete("/:id", handler.Delete)
}
