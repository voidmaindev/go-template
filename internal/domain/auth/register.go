package auth

import (
	"github.com/gofiber/fiber/v2"
	"github.com/voidmaindev/go-template/internal/container"
	"github.com/voidmaindev/go-template/internal/domain/audit"
	"github.com/voidmaindev/go-template/internal/domain/email"
	"github.com/voidmaindev/go-template/internal/domain/rbac"
	"github.com/voidmaindev/go-template/internal/domain/user"
	"github.com/voidmaindev/go-template/internal/middleware"
)

// Component keys for this domain
var (
	ServiceKey    = container.Key[Service]("auth.service")
	HandlerKey    = container.Key[*Handler]("auth.handler")
	TokenStoreKey = container.Key[*TokenStore]("auth.tokenStore")
)

// domain implements container.Domain interface
type domain struct{}

// NewDomain creates a new auth domain for registration
func NewDomain() container.Domain {
	return &domain{}
}

// Name returns the domain name
func (d *domain) Name() string {
	return "auth"
}

// Models returns the GORM models for migration (auth domain uses user's models)
func (d *domain) Models() []any {
	return nil
}

// Register initializes the auth service
func (d *domain) Register(c *container.Container) {
	// Initialize token store
	tokenStore := NewTokenStore(c.Redis)
	TokenStoreKey.Set(c, tokenStore)

	// Get dependencies
	userRepo := user.RepositoryKey.MustGet(c)
	userSvc := user.ServiceKey.MustGet(c)
	emailSvc := email.ServiceKey.MustGet(c)
	rbacSvc := rbac.ServiceKey.MustGet(c)
	auditSvc := audit.ServiceKey.MustGet(c)
	enforcer := rbac.EnforcerKey.MustGet(c)

	// Get user token store for session invalidation on password reset
	userTokenStore := user.TokenStoreKey.MustGet(c)

	// Initialize service
	service := NewService(
		c.DB,
		userRepo,
		tokenStore,
		userTokenStore,
		emailSvc,
		rbacSvc,
		auditSvc,
		enforcer,
		&c.Config.SelfRegistration,
		&c.Config.OAuth,
	)
	ServiceKey.Set(c, service)

	// Initialize handler
	handler := NewHandler(service, userSvc, &c.Config.JWT, &c.Config.SelfRegistration)
	HandlerKey.Set(c, handler)
}

// Routes registers HTTP routes for this domain
func (d *domain) Routes(api fiber.Router, c *container.Container) {
	handler := HandlerKey.MustGet(c)
	userTokenStore := user.TokenStoreKey.MustGet(c)
	rateLimiter := middleware.RateLimiterFactoryKey.MustGet(c)
	jwtConfig := &c.Config.JWT

	// Self-registration routes (public, gated by config in handler)
	selfAuth := api.Group("/auth/self", rateLimiter.ForTier(middleware.TierAuth))
	selfAuth.Post("/register", handler.SelfRegister)
	selfAuth.Post("/verify-email", handler.VerifyEmail)
	selfAuth.Post("/resend-verification", handler.ResendVerification)
	selfAuth.Post("/forgot-password", handler.ForgotPassword)
	selfAuth.Post("/reset-password", handler.ResetPassword)

	// OAuth routes (public, gated by config in handler)
	oauth := api.Group("/auth/oauth", rateLimiter.ForTier(middleware.TierAuth))
	oauth.Get("/:provider", handler.GetOAuthURL)
	oauth.Get("/:provider/callback", handler.OAuthCallback)
	oauth.Post("/:provider/token", handler.OAuthToken)

	// Identity management routes (authenticated)
	userRoutes := api.Group("/users", middleware.JWTMiddlewareWithInvalidator(jwtConfig, userTokenStore, userTokenStore))
	userRoutes.Get("/me/identities", rateLimiter.ForTier(middleware.TierAPIRead), handler.GetUserIdentities)
	userRoutes.Post("/me/identities/:provider", rateLimiter.ForTier(middleware.TierAPIWrite), handler.LinkIdentity)
	userRoutes.Delete("/me/identities/:provider", rateLimiter.ForTier(middleware.TierAPIWrite), handler.UnlinkIdentity)
	userRoutes.Post("/me/set-password", rateLimiter.ForTier(middleware.TierAuthUser), handler.SetPassword)
}
