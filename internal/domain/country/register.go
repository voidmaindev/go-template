package country

import (
	"github.com/casbin/casbin/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/voidmaindev/go-template/internal/container"
	"github.com/voidmaindev/go-template/internal/domain/rbac"
	"github.com/voidmaindev/go-template/internal/domain/user"
	"github.com/voidmaindev/go-template/internal/middleware"
)

// Component keys for this domain
const (
	RepositoryKey = "country.repository"
	ServiceKey    = "country.service"
	HandlerKey    = "country.handler"
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
	c.Set(RepositoryKey, repo)

	// Initialize service
	service := NewService(repo)
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

	countries := api.Group("/countries", middleware.JWTMiddleware(jwtConfig, tokenStore))
	countries.Get("/", middleware.RequirePermission(enforcer, "country", rbac.ActionRead), handler.List)
	countries.Get("/:id", middleware.RequirePermission(enforcer, "country", rbac.ActionRead), handler.GetByID)
	countries.Post("/", middleware.RequirePermission(enforcer, "country", rbac.ActionWrite), handler.Create)
	countries.Put("/:id", middleware.RequirePermission(enforcer, "country", rbac.ActionModify), handler.Update)
	countries.Delete("/:id", middleware.RequirePermission(enforcer, "country", rbac.ActionDelete), handler.Delete)
}
