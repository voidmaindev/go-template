package email

import "github.com/voidmaindev/go-template/internal/common/errors"

const domainName = "email"

// These errors are package-level singletons. NEVER chain builder methods
// (WithOperation, WithContext, etc.) on them at runtime — doing so would
// mutate the shared instance. Return them directly or create new errors
// with errors.New()/errors.Internal() for context-enriched variants.
//
// Domain-specific errors for email operations
var (
	// ErrSMTPConnection is returned when SMTP connection fails
	ErrSMTPConnection = errors.Internal(domainName, nil).
				WithMessage("failed to connect to SMTP server")

	// ErrSendFailed is returned when sending email fails
	ErrSendFailed = errors.Internal(domainName, nil).
			WithMessage("failed to send email")

	// ErrTemplateRender is returned when email template rendering fails
	ErrTemplateRender = errors.Internal(domainName, nil).
				WithMessage("failed to render email template")

	// ErrInvalidRecipient is returned when email address is invalid
	ErrInvalidRecipient = errors.Validation(domainName, "invalid email address")
)
