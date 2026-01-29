package email

import "context"

// Service defines the email service interface
type Service interface {
	// SendVerificationEmail sends an email verification link to the user
	SendVerificationEmail(ctx context.Context, email, name, token string) error

	// SendPasswordResetEmail sends a password reset link to the user
	SendPasswordResetEmail(ctx context.Context, email, name, token string) error

	// SendWelcomeEmail sends a welcome email after verification
	SendWelcomeEmail(ctx context.Context, email, name string) error

	// Send sends a custom email
	Send(ctx context.Context, req *SendEmailRequest) error
}
