package email

import (
	"github.com/gofiber/fiber/v2"
	"github.com/voidmaindev/go-template/internal/container"
)

// Component keys for this domain
var (
	ServiceKey = container.Key[Service]("email.service")
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

// Register initializes the email service
func (d *domain) Register(c *container.Container) {
	service := NewSMTPService(
		&c.Config.SMTP,
		c.Config.App.Name,
		c.Config.SelfRegistration.BaseURL,
	)
	ServiceKey.Set(c, service)
}

// Routes registers HTTP routes for this domain (email domain has no routes)
func (d *domain) Routes(api fiber.Router, c *container.Container) {
	// Email domain has no HTTP routes
}
