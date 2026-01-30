# ADR-006: SendGrid Email Provider

## Status

Accepted (Updated)

## Context

The application needs to send transactional emails (verification, password reset, welcome). We need:
- A production-ready email delivery system
- Easy configuration via environment variables
- Reliable delivery with proper error handling

## Decision

Use SendGrid as the sole email provider. The architecture uses a Provider interface pattern for potential future extensibility, but currently only supports SendGrid.

```go
// Provider abstracts email delivery backends
type Provider interface {
    Send(ctx context.Context, msg *Message) error
    Name() string
}

// Message is provider-agnostic
type Message struct {
    From     Address
    To       []Address
    Subject  string
    HTMLBody string
    TextBody string
}
```

Provider implementation:
- `SendGridProvider` - SendGrid API

A factory function creates the provider:

```go
func NewProvider(cfg *config.EmailConfig) (Provider, error) {
    if cfg.SendGrid.APIKey == "" {
        return nil, fmt.Errorf("SENDGRID_API_KEY is required")
    }
    return NewSendGridProvider(cfg.SendGrid.APIKey), nil
}
```

The `Service` interface provides high-level email operations (SendVerificationEmail, etc.) that internally use the Provider.

## Configuration

```yaml
# Required for email functionality
SENDGRID_API_KEY=SG.xxx

# Common settings
EMAIL_FROM=noreply@example.com
EMAIL_FROM_NAME=My App
```

If `SENDGRID_API_KEY` is not configured, the email service will not be initialized and email-dependent features (like email verification) will be unavailable.

## File Structure

```
internal/domain/email/
├── provider.go           # Provider interface + Message types
├── provider_sendgrid.go  # SendGrid implementation
├── provider_factory.go   # Factory function
├── service.go            # Service interface
├── service_impl.go       # Service using Provider
├── register.go           # Domain registration
├── templates.go          # Email templates
├── dto.go               # Data transfer objects
└── errors.go            # Domain errors
```

## Consequences

### Positive
- **Simplicity**: Single provider reduces complexity
- **Production-Ready**: SendGrid is a reliable, well-supported email service
- **Easy Configuration**: Only one API key needed
- **Interface Pattern**: Can add providers later if needed

### Negative
- **No Local Development Email**: Requires SendGrid account even for development
- **Vendor Lock-in**: Tied to SendGrid (but interface allows future changes)

### Mitigations
- For development without email, set self-registration to not require email verification
- The Provider interface allows adding other providers (Mailgun, SES, SMTP) if needed in the future

## Adding New Providers (Future)

1. Create `provider_<name>.go` implementing `Provider`
2. Update `NewProvider()` factory with configuration logic
3. Add config struct to `config.EmailConfig`
4. Bind environment variables in `config.go`
