package container

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/voidmaindev/go-template/internal/config"
	"github.com/voidmaindev/go-template/internal/redis"
	"gorm.io/gorm"
)

// Container holds all application dependencies.
// Domains register their repositories, services, and handlers here.
type Container struct {
	// Core dependencies
	DB     *gorm.DB
	Redis  *redis.Client
	Config *config.Config

	// Registered domains
	domains []Domain

	// Domain components (populated by Register calls)
	components map[string]any
}

// Domain interface that each domain package implements
type Domain interface {
	// Name returns the domain name (e.g., "user", "item")
	Name() string

	// Models returns the GORM models for migration
	Models() []any

	// Register initializes repositories, services, and handlers
	// Domains can access other domain's components via container.Get()
	Register(c *Container)

	// Routes registers HTTP routes for this domain
	Routes(api fiber.Router, c *Container)
}

// New creates a new container with core dependencies
func New(db *gorm.DB, redisClient *redis.Client, cfg *config.Config) *Container {
	return &Container{
		DB:         db,
		Redis:      redisClient,
		Config:     cfg,
		domains:    make([]Domain, 0),
		components: make(map[string]any),
	}
}

// AddDomain adds a domain to the container
func (c *Container) AddDomain(d Domain) {
	c.domains = append(c.domains, d)
}

// Set stores a component in the container
func (c *Container) Set(key string, component any) {
	c.components[key] = component
}

// Get retrieves a component from the container
func (c *Container) Get(key string) any {
	return c.components[key]
}

// MustGet retrieves a component and panics if not found
func (c *Container) MustGet(key string) any {
	comp, ok := c.components[key]
	if !ok {
		panic("component not found: " + key)
	}
	return comp
}

// MustGetTyped retrieves a typed component from the container.
// Panics if the component is not found or is not of the expected type.
// Usage: handler := container.MustGetTyped[*Handler](c, "user.handler")
func MustGetTyped[T any](c *Container, key string) T {
	comp, ok := c.components[key]
	if !ok {
		panic("component not found: " + key)
	}
	typed, ok := comp.(T)
	if !ok {
		panic(fmt.Sprintf("component %s is not of expected type %T, got %T", key, *new(T), comp))
	}
	return typed
}

// GetTyped retrieves a typed component from the container.
// Returns the zero value and false if not found or wrong type.
// Usage: handler, ok := container.GetTyped[*Handler](c, "user.handler")
func GetTyped[T any](c *Container, key string) (T, bool) {
	comp, ok := c.components[key]
	if !ok {
		var zero T
		return zero, false
	}
	typed, ok := comp.(T)
	if !ok {
		var zero T
		return zero, false
	}
	return typed, true
}

// GetAllModels returns all models from all registered domains
func (c *Container) GetAllModels() []any {
	var models []any
	for _, d := range c.domains {
		models = append(models, d.Models()...)
	}
	return models
}

// RegisterAll registers all domains (repos, services, handlers)
func (c *Container) RegisterAll() {
	for _, d := range c.domains {
		d.Register(c)
	}
}

// RegisterRoutes registers routes for all domains
func (c *Container) RegisterRoutes(api fiber.Router) {
	for _, d := range c.domains {
		d.Routes(api, c)
	}
}
