package email

import (
	"fmt"

	"github.com/voidmaindev/go-template/internal/config"
)

// NewProvider creates a SendGrid email provider from configuration
func NewProvider(cfg *config.EmailConfig) (Provider, error) {
	if cfg.SendGrid.APIKey == "" {
		return nil, fmt.Errorf("SENDGRID_API_KEY is required for email delivery")
	}
	return NewSendGridProvider(cfg.SendGrid.APIKey), nil
}
