package email

import (
	"context"
	"crypto/tls"
	"fmt"
	"log/slog"
	"net/smtp"
	"strings"

	"github.com/voidmaindev/go-template/internal/config"
)

// smtpService implements the Service interface using SMTP
type smtpService struct {
	cfg     *config.SMTPConfig
	appName string
	baseURL string
}

// NewSMTPService creates a new SMTP email service
func NewSMTPService(cfg *config.SMTPConfig, appName, baseURL string) Service {
	return &smtpService{
		cfg:     cfg,
		appName: appName,
		baseURL: baseURL,
	}
}

// SendVerificationEmail sends an email verification link
func (s *smtpService) SendVerificationEmail(ctx context.Context, email, name, token string) error {
	verificationLink := fmt.Sprintf("%s/verify-email?token=%s", s.baseURL, token)

	data := &VerificationEmailData{
		Name:             name,
		VerificationLink: verificationLink,
		ExpiresInHours:   24, // matches config default
		AppName:          s.appName,
	}

	htmlBody, err := renderVerificationEmail(data)
	if err != nil {
		slog.Error("failed to render verification email template", "error", err)
		return ErrTemplateRender
	}

	return s.Send(ctx, &SendEmailRequest{
		To:       email,
		Subject:  fmt.Sprintf("Verify your email address - %s", s.appName),
		HTMLBody: htmlBody,
	})
}

// SendPasswordResetEmail sends a password reset link
func (s *smtpService) SendPasswordResetEmail(ctx context.Context, email, name, token string) error {
	resetLink := fmt.Sprintf("%s/reset-password?token=%s", s.baseURL, token)

	data := &PasswordResetEmailData{
		Name:           name,
		ResetLink:      resetLink,
		ExpiresInHours: 1, // matches config default
		AppName:        s.appName,
	}

	htmlBody, err := renderPasswordResetEmail(data)
	if err != nil {
		slog.Error("failed to render password reset email template", "error", err)
		return ErrTemplateRender
	}

	return s.Send(ctx, &SendEmailRequest{
		To:       email,
		Subject:  fmt.Sprintf("Reset your password - %s", s.appName),
		HTMLBody: htmlBody,
	})
}

// SendWelcomeEmail sends a welcome email after verification
func (s *smtpService) SendWelcomeEmail(ctx context.Context, email, name string) error {
	data := &WelcomeEmailData{
		Name:    name,
		AppName: s.appName,
		AppURL:  s.baseURL,
	}

	htmlBody, err := renderWelcomeEmail(data)
	if err != nil {
		slog.Error("failed to render welcome email template", "error", err)
		return ErrTemplateRender
	}

	return s.Send(ctx, &SendEmailRequest{
		To:       email,
		Subject:  fmt.Sprintf("Welcome to %s!", s.appName),
		HTMLBody: htmlBody,
	})
}

// Send sends a custom email via SMTP
func (s *smtpService) Send(ctx context.Context, req *SendEmailRequest) error {
	if req.To == "" {
		return ErrInvalidRecipient
	}

	// Build message
	from := s.cfg.From
	if s.cfg.FromName != "" {
		from = fmt.Sprintf("%s <%s>", s.cfg.FromName, s.cfg.From)
	}

	msg := buildMIMEMessage(from, req.To, req.Subject, req.HTMLBody, req.TextBody)

	addr := fmt.Sprintf("%s:%d", s.cfg.Host, s.cfg.Port)

	// Create auth if credentials provided
	var auth smtp.Auth
	if s.cfg.Username != "" && s.cfg.Password != "" {
		auth = smtp.PlainAuth("", s.cfg.Username, s.cfg.Password, s.cfg.Host)
	}

	// Send with or without TLS
	var err error
	if s.cfg.UseTLS {
		err = sendMailTLS(addr, auth, s.cfg.From, req.To, msg, s.cfg.Host)
	} else {
		err = smtp.SendMail(addr, auth, s.cfg.From, []string{req.To}, msg)
	}

	if err != nil {
		slog.Error("failed to send email", "to", req.To, "error", err)
		return ErrSendFailed
	}

	slog.Info("email sent successfully", "to", req.To, "subject", req.Subject)
	return nil
}

// buildMIMEMessage builds a MIME message with HTML content
func buildMIMEMessage(from, to, subject, htmlBody, textBody string) []byte {
	var msg strings.Builder

	msg.WriteString(fmt.Sprintf("From: %s\r\n", from))
	msg.WriteString(fmt.Sprintf("To: %s\r\n", to))
	msg.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	msg.WriteString("MIME-Version: 1.0\r\n")

	if textBody != "" && htmlBody != "" {
		boundary := "boundary-mixed-12345"
		msg.WriteString(fmt.Sprintf("Content-Type: multipart/alternative; boundary=%s\r\n\r\n", boundary))
		msg.WriteString(fmt.Sprintf("--%s\r\n", boundary))
		msg.WriteString("Content-Type: text/plain; charset=\"UTF-8\"\r\n\r\n")
		msg.WriteString(textBody)
		msg.WriteString(fmt.Sprintf("\r\n--%s\r\n", boundary))
		msg.WriteString("Content-Type: text/html; charset=\"UTF-8\"\r\n\r\n")
		msg.WriteString(htmlBody)
		msg.WriteString(fmt.Sprintf("\r\n--%s--\r\n", boundary))
	} else if htmlBody != "" {
		msg.WriteString("Content-Type: text/html; charset=\"UTF-8\"\r\n\r\n")
		msg.WriteString(htmlBody)
	} else {
		msg.WriteString("Content-Type: text/plain; charset=\"UTF-8\"\r\n\r\n")
		msg.WriteString(textBody)
	}

	return []byte(msg.String())
}

// sendMailTLS sends mail using explicit TLS connection
func sendMailTLS(addr string, auth smtp.Auth, from, to string, msg []byte, host string) error {
	tlsConfig := &tls.Config{
		ServerName: host,
	}

	conn, err := tls.Dial("tcp", addr, tlsConfig)
	if err != nil {
		return fmt.Errorf("TLS dial failed: %w", err)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, host)
	if err != nil {
		return fmt.Errorf("SMTP client creation failed: %w", err)
	}
	defer client.Close()

	if auth != nil {
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("SMTP auth failed: %w", err)
		}
	}

	if err := client.Mail(from); err != nil {
		return fmt.Errorf("SMTP MAIL command failed: %w", err)
	}

	if err := client.Rcpt(to); err != nil {
		return fmt.Errorf("SMTP RCPT command failed: %w", err)
	}

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("SMTP DATA command failed: %w", err)
	}

	if _, err := w.Write(msg); err != nil {
		return fmt.Errorf("SMTP write message failed: %w", err)
	}

	if err := w.Close(); err != nil {
		return fmt.Errorf("SMTP close writer failed: %w", err)
	}

	return client.Quit()
}
