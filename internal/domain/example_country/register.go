package example_country

import (
	"github.com/gofiber/fiber/v2"
	"github.com/voidmaindev/go-template/internal/container"
	"github.com/voidmaindev/go-template/internal/domain/rbac"
	"github.com/voidmaindev/go-template/internal/domain/user"
	"github.com/voidmaindev/go-template/internal/middleware"
)

// Component keys for this domain (typed for compile-time safety)
var (
	RepositoryKey = container.Key[Repository]("country.repository")
	ServiceKey    = container.Key[Service]("country.service")
	HandlerKey    = container.Key[*Handler]("country.handler")
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
	RepositoryKey.Set(c, repo)

	// Initialize service
	service := NewService(repo)
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

	countries := api.Group("/countries", middleware.JWTMiddlewareWithInvalidator(jwtConfig, tokenStore, tokenStore))
	// GET endpoints - api_read tier (200 req/min)
	countries.Get("/", rateLimiter.ForTier(middleware.TierAPIRead), middleware.RequirePermission(enforcer, "country", rbac.ActionRead), handler.List)
	countries.Get("/:id", rateLimiter.ForTier(middleware.TierAPIRead), middleware.RequirePermission(enforcer, "country", rbac.ActionRead), handler.GetByID)
	// Write endpoints - api_write tier (60 req/min)
	countries.Post("/", rateLimiter.ForTier(middleware.TierAPIWrite), middleware.RequirePermission(enforcer, "country", rbac.ActionCreate), handler.Create)
	countries.Put("/:id", rateLimiter.ForTier(middleware.TierAPIWrite), middleware.RequirePermission(enforcer, "country", rbac.ActionUpdate), handler.Update)
	countries.Delete("/:id", rateLimiter.ForTier(middleware.TierAPIWrite), middleware.RequirePermission(enforcer, "country", rbac.ActionDelete), handler.Delete)
}
