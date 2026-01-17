package city

import (
	"github.com/casbin/casbin/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/voidmaindev/go-template/internal/container"
	"github.com/voidmaindev/go-template/internal/domain/country"
	"github.com/voidmaindev/go-template/internal/domain/rbac"
	"github.com/voidmaindev/go-template/internal/domain/user"
	"github.com/voidmaindev/go-template/internal/middleware"
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
	countryRepo := container.MustGetTyped[country.Repository](c, country.RepositoryKey)

	// Initialize service with cross-domain dependency
	service := NewService(repo, countryRepo)
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

	cities := api.Group("/cities", middleware.JWTMiddleware(jwtConfig, tokenStore))
	cities.Get("/", middleware.RequirePermission(enforcer, "city", rbac.ActionRead), handler.List)
	cities.Get("/:id", middleware.RequirePermission(enforcer, "city", rbac.ActionRead), handler.GetByID)
	cities.Post("/", middleware.RequirePermission(enforcer, "city", rbac.ActionWrite), handler.Create)
	cities.Put("/:id", middleware.RequirePermission(enforcer, "city", rbac.ActionModify), handler.Update)
	cities.Delete("/:id", middleware.RequirePermission(enforcer, "city", rbac.ActionDelete), handler.Delete)

	// Nested route for cities by country (uses city read permission)
	countries := api.Group("/countries", middleware.JWTMiddleware(jwtConfig, tokenStore))
	countries.Get("/:countryId/cities", middleware.RequirePermission(enforcer, "city", rbac.ActionRead), handler.ListByCountry)
}
