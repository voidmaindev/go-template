package item

import (
	"github.com/gofiber/fiber/v2"
	"github.com/voidmaindev/go-template/internal/container"
	"github.com/voidmaindev/go-template/internal/domain/rbac"
	"github.com/voidmaindev/go-template/internal/domain/user"
	"github.com/voidmaindev/go-template/internal/middleware"
)

// Component keys for this domain (typed for compile-time safety)
var (
	RepositoryKey = container.Key[Repository]("item.repository")
	ServiceKey    = container.Key[Service]("item.service")
	HandlerKey    = container.Key[*Handler]("item.handler")
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

	items := api.Group("/items", middleware.JWTMiddlewareWithInvalidator(jwtConfig, tokenStore, tokenStore))
	// GET endpoints - api_read tier (200 req/min)
	items.Get("/", rateLimiter.ForTier(middleware.TierAPIRead), middleware.RequirePermission(enforcer, "item", rbac.ActionRead), handler.List)
	items.Get("/:id", rateLimiter.ForTier(middleware.TierAPIRead), middleware.RequirePermission(enforcer, "item", rbac.ActionRead), handler.GetByID)
	// Write endpoints - api_write tier (60 req/min)
	items.Post("/", rateLimiter.ForTier(middleware.TierAPIWrite), middleware.RequirePermission(enforcer, "item", rbac.ActionCreate), handler.Create)
	items.Put("/:id", rateLimiter.ForTier(middleware.TierAPIWrite), middleware.RequirePermission(enforcer, "item", rbac.ActionUpdate), handler.Update)
	items.Delete("/:id", rateLimiter.ForTier(middleware.TierAPIWrite), middleware.RequirePermission(enforcer, "item", rbac.ActionDelete), handler.Delete)
}
