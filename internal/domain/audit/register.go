package audit

import (
	"github.com/casbin/casbin/v3"
	"github.com/gofiber/fiber/v2"
	"github.com/voidmaindev/go-template/internal/container"
	"github.com/voidmaindev/go-template/internal/middleware"
)

// External component keys (to avoid import cycles)
const (
	userTokenStoreKey = "user.tokenStore"
	rbacEnforcerKey   = "rbac.enforcer"
)

// Component keys for this domain (typed for compile-time safety)
var (
	RepositoryKey = container.Key[Repository]("audit.repository")
	ServiceKey    = container.Key[Service]("audit.service")
	HandlerKey    = container.Key[*Handler]("audit.handler")
)

// domain implements container.Domain interface
type domain struct{}

// NewDomain creates a new audit domain for registration
func NewDomain() container.Domain {
	return &domain{}
}

// Name returns the domain name
func (d *domain) Name() string {
	return "audit"
}

// Models returns the GORM models for migration
func (d *domain) Models() []any {
	return []any{
		&AuditLog{},
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
	tokenStore := container.MustGetAs[middleware.TokenBlacklist](c, userTokenStoreKey)
	tokenInvalidator := container.MustGetAs[middleware.TokenInvalidator](c, userTokenStoreKey)
	enforcer := container.MustGetAs[*casbin.TransactionalEnforcer](c, rbacEnforcerKey)
	rateLimiter := middleware.RateLimiterFactoryKey.MustGet(c)
	jwtConfig := &c.Config.JWT

	// All audit routes require authentication and rbac:read permission (admin only)
	// Using "read" as the action constant to avoid importing rbac package
	auditGroup := api.Group("/audit",
		middleware.JWTMiddlewareWithInvalidator(jwtConfig, tokenStore, tokenInvalidator),
		rateLimiter.ForTier(middleware.TierAPIRead),
		middleware.RequirePermission(enforcer, "rbac", "read"))

	// Audit log listing
	auditGroup.Get("/logs", handler.List)
	auditGroup.Get("/users/:id/logs", handler.ListByUser)
}
