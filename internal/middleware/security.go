package middleware

import (
	"github.com/gofiber/fiber/v2"
)

// SecurityHeaders adds security-related HTTP headers to responses
func SecurityHeaders() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Prevent MIME type sniffing
		c.Set("X-Content-Type-Options", "nosniff")

		// Prevent clickjacking attacks
		c.Set("X-Frame-Options", "DENY")

		// Enable browser XSS filtering
		c.Set("X-XSS-Protection", "1; mode=block")

		// Control referrer information
		c.Set("Referrer-Policy", "strict-origin-when-cross-origin")

		// Disable caching for API responses by default
		c.Set("Cache-Control", "no-store, no-cache, must-revalidate, proxy-revalidate")
		c.Set("Pragma", "no-cache")
		c.Set("Expires", "0")

		return c.Next()
	}
}

// StrictTransportSecurity adds HSTS header for HTTPS connections
// Should only be used when the server is behind HTTPS
func StrictTransportSecurity(maxAge int) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Only add HSTS header when connection is secure
		if c.Secure() || c.Get("X-Forwarded-Proto") == "https" {
			c.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}
		return c.Next()
	}
}

// ContentSecurityPolicy adds a basic CSP header for API responses
func ContentSecurityPolicy() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Restrictive CSP for API responses (no scripts, styles, etc.)
		c.Set("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'")
		return c.Next()
	}
}
