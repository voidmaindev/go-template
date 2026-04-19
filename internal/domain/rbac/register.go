package rbac

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"github.com/voidmaindev/go-template/internal/common/logging"
	"github.com/voidmaindev/go-template/internal/container"
	"github.com/voidmaindev/go-template/internal/domain/audit"
	"github.com/voidmaindev/go-template/internal/middleware"
)

// auditAdapter adapts audit.Service to rbac.AuditLogger interface
type auditAdapter struct {
	svc audit.Service
}

func (a *auditAdapter) LogAsync(ctx context.Context, entry *AuditEntry) {
	a.svc.LogAsync(ctx, &audit.AuditEntry{
		UserID:     entry.UserID,
		Action:     entry.Action,
		Resource:   entry.Resource,
		ResourceID: entry.ResourceID,
		IP:         entry.IP,
		UserAgent:  entry.UserAgent,
		Success:    entry.Success,
		Details:    entry.Details,
	})
}

// Component keys for this domain (typed for compile-time safety)
var (
	RepositoryKey = container.Key[Repository]("rbac.repository")
	ServiceKey    = container.Key[Service]("rbac.service")
	HandlerKey    = container.Key[*Handler]("rbac.handler")
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
	logger := logging.New(domainName)

	// Initialize repository
	repo := NewRepository(c.DB)
	RepositoryKey.Set(c, repo)

	// Initialize Casbin enforcer
	enforcer, err := NewEnforcer(c.DB, &c.Config.RBAC)
	if err != nil {
		logger.Error(context.Background(), "failed to create RBAC enforcer", err)
		panic(err)
	}
	EnforcerKey.Set(c, enforcer)

	// Initialize service with container as domain provider
	service := NewService(repo, enforcer, c)
	ServiceKey.Set(c, service)

	// Initialize handler
	handler := NewHandler(service)
	HandlerKey.Set(c, handler)

	// Sync global roles on startup
	if err := service.SyncGlobalRoles(context.Background()); err != nil {
		logger.Error(context.Background(), "failed to sync global RBAC roles", err)
	}
}

// Routes registers HTTP routes for this domain
func (d *domain) Routes(api fiber.Router, c *container.Container) {
	handler := HandlerKey.MustGet(c)
	tokenStore := middleware.TokenBlacklistKey.MustGet(c)
	tokenInvalidator := middleware.TokenInvalidatorKey.MustGet(c)
	enforcer := EnforcerKey.MustGet(c)
	rateLimiter := middleware.RateLimiterFactoryKey.MustGet(c)
	jwtConfig := &c.Config.JWT

	// Wire up audit logger if available (audit domain may not be registered)
	if auditSvc, ok := audit.ServiceKey.Get(c); ok {
		handler.SetAuditLogger(&auditAdapter{svc: auditSvc})
	}

	// All RBAC routes require authentication and admin-tier rate limiting
	rbacGroup := api.Group("/rbac", middleware.JWTMiddlewareWithInvalidator(jwtConfig, tokenStore, tokenInvalidator), rateLimiter.ForTier(middleware.TierRBACAdmin))

	// Role management (admin only via RBAC)
	roles := rbacGroup.Group("/roles", middleware.RequirePermission(enforcer, "rbac", ActionRead))
	roles.Get("/", handler.ListRoles)
	roles.Get("/:code", handler.GetRole)
	roles.Post("/", middleware.RequirePermission(enforcer, "rbac", ActionCreate), handler.CreateRole)
	roles.Put("/:code/permissions", middleware.RequirePermission(enforcer, "rbac", ActionUpdate), handler.UpdateRolePermissions)
	roles.Delete("/:code", middleware.RequirePermission(enforcer, "rbac", ActionDelete), handler.DeleteRole)

	// User-role management (admin only via RBAC)
	users := rbacGroup.Group("/users", middleware.RequirePermission(enforcer, "rbac", ActionRead))
	users.Get("/:id/roles", handler.GetUserRoles)
	users.Post("/:id/roles", middleware.RequirePermission(enforcer, "rbac", ActionUpdate), handler.AssignRole)
	users.Delete("/:id/roles/:code", middleware.RequirePermission(enforcer, "rbac", ActionUpdate), handler.RemoveRole)

	// Discovery endpoints (require rbac:read permission to prevent structure exposure)
	discovery := rbacGroup.Group("", middleware.RequirePermission(enforcer, "rbac", ActionRead))
	discovery.Get("/domains", handler.GetDomains)
	discovery.Get("/actions", handler.GetActions)
}
