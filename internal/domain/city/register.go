package city

import (
	"github.com/gofiber/fiber/v2"
	"github.com/voidmaindev/GoTemplate/internal/container"
	"github.com/voidmaindev/GoTemplate/internal/domain/country"
	"github.com/voidmaindev/GoTemplate/internal/domain/user"
	"github.com/voidmaindev/GoTemplate/internal/middleware"
)

// Component keys for this domain
const (
	RepositoryKey = "city.repository"
	ServiceKey    = "city.service"
	HandlerKey    = "city.handler"
)

// domain implements container.Domain interface
type domain struct{}

// NewDomain creates a new city domain for registration
func NewDomain() container.Domain {
	return &domain{}
}

// Name returns the domain name
func (d *domain) Name() string {
	return "city"
}

// Models returns the GORM models for migration
func (d *domain) Models() []any {
	return []any{
		&City{},
	}
}

// Register initializes repositories, services, and handlers
func (d *domain) Register(c *container.Container) {
	// Initialize repository
	repo := NewRepository(c.DB)
	c.Set(RepositoryKey, repo)

	// Get country repository (cross-domain dependency)
	countryRepo := c.MustGet(country.RepositoryKey).(country.Repository)

	// Initialize service with cross-domain dependency
	service := NewService(repo, countryRepo)
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

	cities := api.Group("/cities", middleware.JWTMiddleware(jwtConfig, tokenStore))
	cities.Post("/", handler.Create)
	cities.Get("/", handler.List)
	cities.Get("/:id", handler.GetByID)
	cities.Put("/:id", handler.Update)
	cities.Delete("/:id", handler.Delete)

	// Nested route for cities by country
	countries := api.Group("/countries", middleware.JWTMiddleware(jwtConfig, tokenStore))
	countries.Get("/:countryId/cities", handler.ListByCountry)
}
