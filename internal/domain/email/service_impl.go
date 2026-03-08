package email

import (
	"context"
	"fmt"

	"github.com/voidmaindev/go-template/internal/common/logging"
)

// serviceImpl implements the Service interface using a Provider
type serviceImpl struct {
	provider Provider
	from     Address
	appName  string
	baseURL  string
	logger   *logging.Logger
}

// NewService creates a new email service with the given provider
func NewService(provider Provider, from Address, appName, baseURL string) Service {
	return &serviceImpl{
		provider: provider,
		from:     from,
		appName:  appName,
		baseURL:  baseURL,
		logger:   logging.New(domainName),
	}
}

// SendVerificationEmail sends an email verification link
func (s *serviceImpl) SendVerificationEmail(ctx context.Context, email, name, token string) error {
	verificationLink := fmt.Sprintf("%s/verify-email?token=%s", s.baseURL, token)

	data := &VerificationEmailData{
		Name:             name,
		VerificationLink: verificationLink,
		ExpiresInHours:   24,
		AppName:          s.appName,
	}

	htmlBody, err := renderVerificationEmail(data)
	if err != nil {
		s.logger.Error(ctx, "failed to render verification email template", err)
		return ErrTemplateRender
	}

	return s.Send(ctx, &SendEmailRequest{
		To:       email,
		Subject:  fmt.Sprintf("Verify your email address - %s", s.appName),
		HTMLBody: htmlBody,
	})
}

// SendPasswordResetEmail sends a password reset link
func (s *serviceImpl) SendPasswordResetEmail(ctx context.Context, email, name, token string) error {
	resetLink := fmt.Sprintf("%s/reset-password?token=%s", s.baseURL, token)

	data := &PasswordResetEmailData{
		Name:           name,
		ResetLink:      resetLink,
		ExpiresInHours: 1,
		AppName:        s.appName,
	}

	htmlBody, err := renderPasswordResetEmail(data)
	if err != nil {
		s.logger.Error(ctx, "failed to render password reset email template", err)
		return ErrTemplateRender
	}

	return s.Send(ctx, &SendEmailRequest{
		To:       email,
		Subject:  fmt.Sprintf("Reset your password - %s", s.appName),
		HTMLBody: htmlBody,
	})
}

// SendWelcomeEmail sends a welcome email after verification
func (s *serviceImpl) SendWelcomeEmail(ctx context.Context, email, name string) error {
	data := &WelcomeEmailData{
		Name:    name,
		AppName: s.appName,
		AppURL:  s.baseURL,
	}

	htmlBody, err := renderWelcomeEmail(data)
	if err != nil {
		s.logger.Error(ctx, "failed to render welcome email template", err)
		return ErrTemplateRender
	}

	return s.Send(ctx, &SendEmailRequest{
		To:       email,
		Subject:  fmt.Sprintf("Welcome to %s!", s.appName),
		HTMLBody: htmlBody,
	})
}

// Send sends a custom email via the configured provider
func (s *serviceImpl) Send(ctx context.Context, req *SendEmailRequest) error {
	if req.To == "" {
		return ErrInvalidRecipient
	}

	msg := &Message{
		From:     s.from,
		To:       []Address{{Email: req.To}},
		Subject:  req.Subject,
		HTMLBody: req.HTMLBody,
		TextBody: req.TextBody,
	}

	return s.provider.Send(ctx, msg)
}
