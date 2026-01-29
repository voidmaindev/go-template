package email

import (
	"bytes"
	"embed"
	"html/template"
)

//go:embed templates/*.html
var templatesFS embed.FS

var templates *template.Template

func init() {
	var err error
	templates, err = template.ParseFS(templatesFS, "templates/*.html")
	if err != nil {
		panic("failed to parse email templates: " + err.Error())
	}
}

// renderTemplate renders an email template with the given data
func renderTemplate(name string, data any) (string, error) {
	var buf bytes.Buffer
	if err := templates.ExecuteTemplate(&buf, name, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// renderVerificationEmail renders the email verification template
func renderVerificationEmail(data *VerificationEmailData) (string, error) {
	return renderTemplate("verification.html", data)
}

// renderPasswordResetEmail renders the password reset email template
func renderPasswordResetEmail(data *PasswordResetEmailData) (string, error) {
	return renderTemplate("password_reset.html", data)
}

// renderWelcomeEmail renders the welcome email template
func renderWelcomeEmail(data *WelcomeEmailData) (string, error) {
	return renderTemplate("welcome.html", data)
}
