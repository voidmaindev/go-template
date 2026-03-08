package email

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"github.com/voidmaindev/go-template/internal/common/logging"
	"github.com/voidmaindev/go-template/internal/container"
)

// Component keys for this domain
var (
	ServiceKey  = container.Key[Service]("email.service")
	ProviderKey = container.Key[Provider]("email.provider")
)

// domain implements container.Domain interface
type domain struct{}

// NewDomain creates a new email domain for registration
func NewDomain() container.Domain {
	return &domain{}
}

// Name returns the domain name
func (d *domain) Name() string {
	return "email"
}

// Models returns the GORM models for migration (email domain has no models)
func (d *domain) Models() []any {
	return nil
}

// Register initializes the email service with SendGrid provider
func (d *domain) Register(c *container.Container) {
	logger := logging.New(domainName)
	cfg := &c.Config.Email

	provider, err := NewProvider(cfg)
	if err != nil {
		logger.Warn(context.Background(), "email provider not configured", "error", err)
		// Don't set provider or service - email functionality will be disabled
		return
	}
	ProviderKey.Set(c, provider)

	from := Address{
		Email: cfg.From,
		Name:  cfg.FromName,
	}

	service := NewService(
		provider,
		from,
		c.Config.App.Name,
		c.Config.SelfRegistration.BaseURL,
	)
	ServiceKey.Set(c, service)

	logger.Info(context.Background(), "email service initialized", "provider", provider.Name())
}

// Routes registers HTTP routes for this domain (email domain has no routes)
func (d *domain) Routes(api fiber.Router, c *container.Container) {
	// Email domain has no HTTP routes
}
