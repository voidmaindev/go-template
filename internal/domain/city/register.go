package city

import (
	"github.com/gofiber/fiber/v2"
	"github.com/voidmaindev/go-template/internal/container"
	"github.com/voidmaindev/go-template/internal/domain/country"
	"github.com/voidmaindev/go-template/internal/domain/rbac"
	"github.com/voidmaindev/go-template/internal/domain/user"
	"github.com/voidmaindev/go-template/internal/middleware"
)

// Component keys for this domain (typed for compile-time safety)
var (
	RepositoryKey = container.Key[Repository]("city.repository")
	ServiceKey    = container.Key[Service]("city.service")
	HandlerKey    = container.Key[*Handler]("city.handler")
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
	RepositoryKey.Set(c, repo)

	// Get country repository (cross-domain dependency)
	countryRepo := country.RepositoryKey.MustGet(c)

	// Initialize service with cross-domain dependency
	service := NewService(repo, countryRepo)
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

	cities := api.Group("/cities", middleware.JWTMiddlewareWithInvalidator(jwtConfig, tokenStore, tokenStore))
	// GET endpoints - api_read tier (200 req/min)
	cities.Get("/", rateLimiter.ForTier(middleware.TierAPIRead), middleware.RequirePermission(enforcer, "city", rbac.ActionRead), handler.List)
	cities.Get("/:id", rateLimiter.ForTier(middleware.TierAPIRead), middleware.RequirePermission(enforcer, "city", rbac.ActionRead), handler.GetByID)
	// Write endpoints - api_write tier (60 req/min)
	cities.Post("/", rateLimiter.ForTier(middleware.TierAPIWrite), middleware.RequirePermission(enforcer, "city", rbac.ActionCreate), handler.Create)
	cities.Put("/:id", rateLimiter.ForTier(middleware.TierAPIWrite), middleware.RequirePermission(enforcer, "city", rbac.ActionUpdate), handler.Update)
	cities.Delete("/:id", rateLimiter.ForTier(middleware.TierAPIWrite), middleware.RequirePermission(enforcer, "city", rbac.ActionDelete), handler.Delete)

	// Nested route for cities by country (uses city read permission)
	countries := api.Group("/countries", middleware.JWTMiddlewareWithInvalidator(jwtConfig, tokenStore, tokenStore))
	countries.Get("/:countryId/cities", rateLimiter.ForTier(middleware.TierAPIRead), middleware.RequirePermission(enforcer, "city", rbac.ActionRead), handler.ListByCountry)
}
