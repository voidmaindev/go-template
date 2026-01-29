package email

// SendEmailRequest represents a request to send an email
type SendEmailRequest struct {
	To       string
	Subject  string
	HTMLBody string
	TextBody string
}

// VerificationEmailData contains data for verification email template
type VerificationEmailData struct {
	Name             string
	VerificationLink string
	ExpiresInHours   int
	AppName          string
}

// PasswordResetEmailData contains data for password reset email template
type PasswordResetEmailData struct {
	Name           string
	ResetLink      string
	ExpiresInHours int
	AppName        string
}

// WelcomeEmailData contains data for welcome email template
type WelcomeEmailData struct {
	Name    string
	AppName string
	AppURL  string
}
