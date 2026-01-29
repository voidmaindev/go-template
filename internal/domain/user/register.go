package user

import (
	"github.com/gofiber/fiber/v2"
	"github.com/voidmaindev/go-template/internal/container"
	"github.com/voidmaindev/go-template/internal/domain/rbac"
	"github.com/voidmaindev/go-template/internal/middleware"
)

// Component keys for this domain (typed for compile-time safety)
var (
	RepositoryKey = container.Key[Repository]("user.repository")
	ServiceKey    = container.Key[Service]("user.service")
	HandlerKey    = container.Key[*Handler]("user.handler")
	TokenStoreKey = container.Key[*TokenStore]("user.tokenStore")
)

// domain implements container.Domain interface
type domain struct{}

// NewDomain creates a new user domain for registration
func NewDomain() container.Domain {
	return &domain{}
}

// Name returns the domain name
func (d *domain) Name() string {
	return "user"
}

// Models returns the GORM models for migration
func (d *domain) Models() []any {
	return []any{
		&User{},
		&ExternalIdentity{},
	}
}

// Register initializes repositories, services, and handlers
func (d *domain) Register(c *container.Container) {
	// Initialize token store (uses Redis)
	tokenStore := NewTokenStore(c.Redis)
	TokenStoreKey.Set(c, tokenStore)

	// Initialize repository
	repo := NewRepository(c.DB)
	RepositoryKey.Set(c, repo)

	// Get RBAC service and enforcer (must be registered before user domain)
	rbacSvc := rbac.ServiceKey.MustGet(c)
	enforcer := rbac.EnforcerKey.MustGet(c)

	// Initialize service with enforcer for transactional role assignment
	service := NewService(repo, tokenStore, &c.Config.JWT, c.Config.App.IsProduction(), rbacSvc, enforcer)
	ServiceKey.Set(c, service)

	// Initialize handler
	handler := NewHandler(service, &c.Config.JWT)
	HandlerKey.Set(c, handler)
}

// Routes registers HTTP routes for this domain
func (d *domain) Routes(api fiber.Router, c *container.Container) {
	handler := HandlerKey.MustGet(c)
	tokenStore := TokenStoreKey.MustGet(c)
	enforcer := rbac.EnforcerKey.MustGet(c)
	rateLimiter := middleware.RateLimiterFactoryKey.MustGet(c)
	jwtConfig := &c.Config.JWT

	// Auth routes (public) - with strict rate limiting to prevent brute force
	auth := api.Group("/auth", rateLimiter.ForTier(middleware.TierAuth))
	auth.Post("/login", handler.Login)
	auth.Post("/refresh", handler.RefreshToken)

	// Register requires authentication and user:write permission (admin-only)
	authRegister := auth.Group("", middleware.JWTMiddleware(jwtConfig, tokenStore))
	authRegister.Post("/register", middleware.RequirePermission(enforcer, "user", rbac.ActionCreate), handler.Register)

	// Auth routes (protected) - with auth_user tier rate limiting
	authProtected := auth.Group("", middleware.JWTMiddleware(jwtConfig, tokenStore), rateLimiter.ForTier(middleware.TierAuthUser))
	authProtected.Post("/logout", handler.Logout)

	// User routes (protected)
	users := api.Group("/users", middleware.JWTMiddleware(jwtConfig, tokenStore))
	// Self-management routes (any authenticated user)
	users.Get("/me", rateLimiter.ForTier(middleware.TierAPIRead), handler.GetMe)
	users.Put("/me", rateLimiter.ForTier(middleware.TierAPIWrite), handler.UpdateMe)
	users.Put("/me/password", rateLimiter.ForTier(middleware.TierAuthUser), handler.ChangePassword)
	// Admin routes (require RBAC permission)
	users.Get("/", rateLimiter.ForTier(middleware.TierAPIRead), middleware.RequirePermission(enforcer, "user", rbac.ActionRead), handler.List)
	users.Get("/:id", rateLimiter.ForTier(middleware.TierAPIRead), middleware.RequirePermission(enforcer, "user", rbac.ActionRead), handler.GetByID)
	users.Delete("/:id", rateLimiter.ForTier(middleware.TierAPIWrite), middleware.RequirePermission(enforcer, "user", rbac.ActionDelete), handler.Delete)
}
