package email

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

// SendGridProvider implements Provider using SendGrid API
type SendGridProvider struct {
	client *sendgrid.Client
}

// NewSendGridProvider creates a new SendGrid email provider
func NewSendGridProvider(apiKey string) *SendGridProvider {
	return &SendGridProvider{
		client: sendgrid.NewSendClient(apiKey),
	}
}

// Name returns the provider name
func (p *SendGridProvider) Name() string {
	return "sendgrid"
}

// Send delivers an email via SendGrid API
func (p *SendGridProvider) Send(ctx context.Context, msg *Message) error {
	if len(msg.To) == 0 {
		return ErrInvalidRecipient
	}

	from := mail.NewEmail(msg.From.Name, msg.From.Email)

	// Build personalization for all recipients
	m := mail.NewV3Mail()
	m.SetFrom(from)
	m.Subject = msg.Subject

	personalization := mail.NewPersonalization()
	for _, addr := range msg.To {
		personalization.AddTos(mail.NewEmail(addr.Name, addr.Email))
	}
	m.AddPersonalizations(personalization)

	// Add content
	if msg.TextBody != "" {
		m.AddContent(mail.NewContent("text/plain", msg.TextBody))
	}
	if msg.HTMLBody != "" {
		m.AddContent(mail.NewContent("text/html", msg.HTMLBody))
	}

	resp, err := p.client.SendWithContext(ctx, m)
	if err != nil {
		slog.Error("SendGrid send failed", "provider", p.Name(), "error", err)
		return ErrSendFailed
	}

	// SendGrid returns 2xx for success
	if resp.StatusCode >= 300 {
		slog.Error("SendGrid API error",
			"provider", p.Name(),
			"status", resp.StatusCode,
			"body", resp.Body,
		)
		return fmt.Errorf("%w: status %d", ErrSendFailed, resp.StatusCode)
	}

	recipients := make([]string, len(msg.To))
	for i, addr := range msg.To {
		recipients[i] = addr.Email
	}
	slog.Info("email sent", "provider", p.Name(), "to", recipients, "subject", msg.Subject)
	return nil
}
