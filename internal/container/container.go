package container

import (
	"github.com/gofiber/fiber/v2"
	"github.com/voidmaindev/GoTemplate/internal/config"
	"github.com/voidmaindev/GoTemplate/internal/redis"
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
	components map[string]interface{}
}

// Domain interface that each domain package implements
type Domain interface {
	// Name returns the domain name (e.g., "user", "item")
	Name() string

	// Models returns the GORM models for migration
	Models() []interface{}

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
		components: make(map[string]interface{}),
	}
}

// AddDomain adds a domain to the container
func (c *Container) AddDomain(d Domain) {
	c.domains = append(c.domains, d)
}

// Set stores a component in the container
func (c *Container) Set(key string, component interface{}) {
	c.components[key] = component
}

// Get retrieves a component from the container
func (c *Container) Get(key string) interface{} {
	return c.components[key]
}

// MustGet retrieves a component and panics if not found
func (c *Container) MustGet(key string) interface{} {
	comp, ok := c.components[key]
	if !ok {
		panic("component not found: " + key)
	}
	return comp
}

// GetAllModels returns all models from all registered domains
func (c *Container) GetAllModels() []interface{} {
	var models []interface{}
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

// Component keys for type-safe access
const (
	// User domain
	UserRepository = "user.repository"
	UserService    = "user.service"
	UserHandler    = "user.handler"
	TokenStore     = "user.tokenStore"

	// Item domain
	ItemRepository = "item.repository"
	ItemService    = "item.service"
	ItemHandler    = "item.handler"

	// Country domain
	CountryRepository = "country.repository"
	CountryService    = "country.service"
	CountryHandler    = "country.handler"

	// City domain
	CityRepository = "city.repository"
	CityService    = "city.service"
	CityHandler    = "city.handler"

	// Document domain
	DocumentRepository     = "document.repository"
	DocumentItemRepository = "document.itemRepository"
	DocumentService        = "document.service"
	DocumentHandler        = "document.handler"
)
