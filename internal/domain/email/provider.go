package email

import (
	"context"
	"fmt"
)

// Provider abstracts email delivery backends
type Provider interface {
	// Send delivers an email message
	Send(ctx context.Context, msg *Message) error
	// Name returns the provider name for logging
	Name() string
}

// Message represents a provider-agnostic email to be sent
type Message struct {
	From     Address
	To       []Address
	Subject  string
	HTMLBody string
	TextBody string
}

// Address represents an email address with optional display name
type Address struct {
	Email string
	Name  string
}

// String returns formatted "Name <email>" or just email
func (a Address) String() string {
	if a.Name == "" {
		return a.Email
	}
	return fmt.Sprintf("%s <%s>", a.Name, a.Email)
}
