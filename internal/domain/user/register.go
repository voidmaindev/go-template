package user

import (
	"github.com/gofiber/fiber/v2"
	"github.com/voidmaindev/GoTemplate/internal/container"
	"github.com/voidmaindev/GoTemplate/internal/middleware"
)

// Component keys for this domain
const (
	RepositoryKey = "user.repository"
	ServiceKey    = "user.service"
	HandlerKey    = "user.handler"
	TokenStoreKey = "user.tokenStore"
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
	}
}

// Register initializes repositories, services, and handlers
func (d *domain) Register(c *container.Container) {
	// Initialize token store (uses Redis)
	tokenStore := NewTokenStore(c.Redis)
	c.Set(TokenStoreKey, tokenStore)

	// Initialize repository
	repo := NewRepository(c.DB)
	c.Set(RepositoryKey, repo)

	// Initialize service
	service := NewService(repo, tokenStore, &c.Config.JWT)
	c.Set(ServiceKey, service)

	// Initialize handler
	handler := NewHandler(service)
	c.Set(HandlerKey, handler)
}

// Routes registers HTTP routes for this domain
func (d *domain) Routes(api fiber.Router, c *container.Container) {
	handler := c.MustGet(HandlerKey).(*Handler)
	tokenStore := c.MustGet(TokenStoreKey).(*TokenStore)
	jwtConfig := &c.Config.JWT

	// Auth routes (public) - with rate limiting to prevent brute force
	auth := api.Group("/auth", middleware.AuthRateLimiter())
	auth.Post("/register", handler.Register)
	auth.Post("/login", handler.Login)
	auth.Post("/refresh", handler.RefreshToken)

	// Auth routes (protected)
	authProtected := auth.Group("", middleware.JWTMiddleware(jwtConfig, tokenStore))
	authProtected.Post("/logout", handler.Logout)

	// User routes (protected)
	users := api.Group("/users", middleware.JWTMiddleware(jwtConfig, tokenStore))
	users.Get("/me", handler.GetMe)
	users.Put("/me", handler.UpdateMe)
	users.Put("/me/password", handler.ChangePassword)
	users.Get("/", handler.List)
	users.Get("/:id", handler.GetByID)
	users.Delete("/:id", handler.Delete)
}
