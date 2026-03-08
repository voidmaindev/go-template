package container

import (
	"context"
	"errors"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/voidmaindev/go-template/internal/config"
	"github.com/voidmaindev/go-template/internal/redis"
	"gorm.io/gorm"
)

// Container holds all application dependencies.
// Domains register their repositories, services, and handlers here.
//
// Thread safety: Components are registered sequentially during startup
// (via RegisterAll) and read-only thereafter. Do not call Set() after
// startup or from concurrent goroutines.
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

// MustGet retrieves a component and panics if not found.
// For error handling, use GetRequired instead.
func (c *Container) MustGet(key string) any {
	comp, ok := c.components[key]
	if !ok {
		panic(fmt.Sprintf("component not found: %s (ensure domain is registered before dependent domains)", key))
	}
	return comp
}

// GetRequired retrieves a component and returns an error if not found.
// This is the non-panicking alternative to MustGet.
func (c *Container) GetRequired(key string) (any, error) {
	comp, ok := c.components[key]
	if !ok {
		return nil, fmt.Errorf("component not found: %s (ensure domain is registered before dependent domains)", key)
	}
	return comp, nil
}

// MustGetTyped retrieves a typed component from the container.
// Panics if the component is not found or is not of the expected type.
// Usage: handler := container.MustGetTyped[*Handler](c, "user.handler")
func MustGetTyped[T any](c *Container, key string) T {
	comp, ok := c.components[key]
	if !ok {
		panic(fmt.Sprintf("component not found: %s (ensure domain is registered before dependent domains)", key))
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

// RegisterAll registers all domains (repos, services, handlers).
// Panics if registration fails. For error handling, use TryRegisterAll instead.
func (c *Container) RegisterAll() {
	if err := c.TryRegisterAll(); err != nil {
		panic(err)
	}
}

// DependencyDeclarer is optionally implemented by domains that depend on other
// domains being registered first. The returned names must match the Name() of
// the required domains.
type DependencyDeclarer interface {
	Dependencies() []string
}

// TryRegisterAll registers all domains and returns an error if any registration fails.
// This catches panics from MustGet calls and returns them as errors, providing
// better error handling during application startup.
// If a domain implements DependencyDeclarer, its dependencies are validated
// before registration.
func (c *Container) TryRegisterAll() error {
	registered := make(map[string]bool, len(c.domains))

	for _, d := range c.domains {
		// Validate declared dependencies before registration
		if dd, ok := d.(DependencyDeclarer); ok {
			for _, dep := range dd.Dependencies() {
				if !registered[dep] {
					return fmt.Errorf("domain %q depends on %q which is not yet registered (check registration order)", d.Name(), dep)
				}
			}
		}

		if err := c.tryRegisterDomain(d); err != nil {
			return fmt.Errorf("failed to register domain %q: %w", d.Name(), err)
		}
		registered[d.Name()] = true
	}
	return nil
}

// tryRegisterDomain attempts to register a single domain, recovering from panics.
func (c *Container) tryRegisterDomain(d Domain) (err error) {
	defer func() {
		if r := recover(); r != nil {
			switch v := r.(type) {
			case error:
				err = v
			case string:
				err = fmt.Errorf("%s", v)
			default:
				err = fmt.Errorf("panic during registration: %v", r)
			}
		}
	}()

	d.Register(c)
	return nil
}

// RegisterRoutes registers routes for all domains
func (c *Container) RegisterRoutes(api fiber.Router) {
	for _, d := range c.domains {
		d.Routes(api, c)
	}
}

// GetDomainNames returns the names of all registered domains
// This is used by RBAC to auto-discover available domains
func (c *Container) GetDomainNames() []string {
	names := make([]string, len(c.domains))
	for i, d := range c.domains {
		names[i] = d.Name()
	}
	return names
}

// Shutdowner is an optional interface that domains can implement
// to perform cleanup during graceful shutdown.
type Shutdowner interface {
	Shutdown(ctx context.Context) error
}

// Shutdown gracefully shuts down all domains that implement Shutdowner.
// Domains are shut down in reverse registration order (LIFO).
// Returns an error combining all shutdown errors.
func (c *Container) Shutdown(ctx context.Context) error {
	var errs []error

	// Shutdown in reverse order (last registered, first shutdown)
	for i := len(c.domains) - 1; i >= 0; i-- {
		d := c.domains[i]
		if shutdowner, ok := d.(Shutdowner); ok {
			if err := shutdowner.Shutdown(ctx); err != nil {
				errs = append(errs, fmt.Errorf("domain %q shutdown failed: %w", d.Name(), err))
			}
		}
	}

	return errors.Join(errs...)
}
