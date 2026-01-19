package rbac

import (
	"context"
	"log/slog"

	"github.com/casbin/casbin/v3"
	"github.com/gofiber/fiber/v2"
	"github.com/voidmaindev/go-template/internal/container"
	"github.com/voidmaindev/go-template/internal/middleware"
)

// External component keys (to avoid import cycles with user domain)
const (
	userTokenStoreKey = "user.tokenStore"
)

// Component keys for this domain
const (
	RepositoryKey = "rbac.repository"
	ServiceKey    = "rbac.service"
	HandlerKey    = "rbac.handler"
)

// domain implements container.Domain interface
type domain struct{}

// NewDomain creates a new RBAC domain for registration
func NewDomain() container.Domain {
	return &domain{}
}

// Name returns the domain name
func (d *domain) Name() string {
	return "rbac"
}

// Models returns the GORM models for migration
func (d *domain) Models() []any {
	return []any{
		&Role{},
	}
}

// Register initializes repositories, services, and handlers
func (d *domain) Register(c *container.Container) {
	// Initialize repository
	repo := NewRepository(c.DB)
	c.Set(RepositoryKey, repo)

	// Initialize Casbin enforcer
	enforcer, err := NewEnforcer(c.DB, &c.Config.RBAC)
	if err != nil {
		slog.Error("failed to create RBAC enforcer", "error", err)
		panic(err)
	}
	c.Set(EnforcerKey, enforcer)

	// Initialize service with container as domain provider
	service := NewService(repo, enforcer, c)
	c.Set(ServiceKey, service)

	// Initialize handler
	handler := NewHandler(service)
	c.Set(HandlerKey, handler)

	// Sync global roles on startup
	if err := service.SyncGlobalRoles(context.Background()); err != nil {
		slog.Error("failed to sync global RBAC roles", "error", err)
	}
}

// Routes registers HTTP routes for this domain
func (d *domain) Routes(api fiber.Router, c *container.Container) {
	handler := container.MustGetTyped[*Handler](c, HandlerKey)
	tokenStore := container.MustGetTyped[middleware.TokenBlacklist](c, userTokenStoreKey)
	enforcer := container.MustGetTyped[*casbin.Enforcer](c, EnforcerKey)
	rateLimiter := container.MustGetTyped[*middleware.RateLimiterFactory](c, middleware.RateLimiterFactoryKey)
	jwtConfig := &c.Config.JWT

	// All RBAC routes require authentication and admin-tier rate limiting
	rbacGroup := api.Group("/rbac", middleware.JWTMiddleware(jwtConfig, tokenStore), rateLimiter.ForTier(middleware.TierRBACAdmin))

	// Role management (admin only via RBAC)
	roles := rbacGroup.Group("/roles", middleware.RequirePermission(enforcer, "rbac", ActionRead))
	roles.Get("/", handler.ListRoles)
	roles.Get("/:code", handler.GetRole)
	roles.Post("/", middleware.RequirePermission(enforcer, "rbac", ActionWrite), handler.CreateRole)
	roles.Put("/:code/permissions", middleware.RequirePermission(enforcer, "rbac", ActionModify), handler.UpdateRolePermissions)
	roles.Delete("/:code", middleware.RequirePermission(enforcer, "rbac", ActionDelete), handler.DeleteRole)

	// User-role management (admin only via RBAC)
	users := rbacGroup.Group("/users", middleware.RequirePermission(enforcer, "rbac", ActionRead))
	users.Get("/:id/roles", handler.GetUserRoles)
	users.Post("/:id/roles", middleware.RequirePermission(enforcer, "rbac", ActionModify), handler.AssignRole)
	users.Delete("/:id/roles/:code", middleware.RequirePermission(enforcer, "rbac", ActionModify), handler.RemoveRole)

	// Discovery endpoints (any authenticated user)
	rbacGroup.Get("/domains", handler.GetDomains)
	rbacGroup.Get("/actions", handler.GetActions)
}
