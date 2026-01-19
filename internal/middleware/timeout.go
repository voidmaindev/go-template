package middleware

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/timeout"
)

// TimeoutConfig holds timeout middleware configuration.
type TimeoutConfig struct {
	Default time.Duration `mapstructure:"default"`
	Auth    time.Duration `mapstructure:"auth"`
	Read    time.Duration `mapstructure:"read"`
	Write   time.Duration `mapstructure:"write"`
	Export  time.Duration `mapstructure:"export"`
}

// DefaultTimeoutConfig returns the default timeout configuration.
func DefaultTimeoutConfig() TimeoutConfig {
	return TimeoutConfig{
		Default: 30 * time.Second,
		Auth:    10 * time.Second,
		Read:    15 * time.Second,
		Write:   30 * time.Second,
		Export:  120 * time.Second,
	}
}

// TimeoutMiddleware creates a middleware that enforces request timeouts
// using Fiber's official timeout middleware pattern.
// Returns 408 Request Timeout when the handler exceeds the timeout.
func TimeoutMiddleware(t time.Duration) fiber.Handler {
	return timeout.NewWithContext(func(c *fiber.Ctx) error {
		return c.Next()
	}, t)
}

// TimeoutMiddlewareWithConfig creates a timeout middleware using the provided configuration.
// It selects the appropriate timeout based on the request path and method.
func TimeoutMiddlewareWithConfig(cfg TimeoutConfig) fiber.Handler {
	return func(c *fiber.Ctx) error {
		t := selectTimeout(c, cfg)
		handler := timeout.NewWithContext(func(c *fiber.Ctx) error {
			return c.Next()
		}, t)
		return handler(c)
	}
}

// selectTimeout selects the appropriate timeout based on the request.
func selectTimeout(c *fiber.Ctx, cfg TimeoutConfig) time.Duration {
	path := c.Path()
	method := c.Method()

	// Auth endpoints get shorter timeout
	if isAuthPath(path) {
		return cfg.Auth
	}

	// Export endpoints get longer timeout
	if isExportPath(path) {
		return cfg.Export
	}

	// Read vs Write operations
	switch method {
	case fiber.MethodGet, fiber.MethodHead, fiber.MethodOptions:
		return cfg.Read
	case fiber.MethodPost, fiber.MethodPut, fiber.MethodPatch, fiber.MethodDelete:
		return cfg.Write
	default:
		return cfg.Default
	}
}

// isAuthPath checks if the path is an authentication endpoint.
func isAuthPath(path string) bool {
	authPaths := []string{
		"/api/v1/auth/login",
		"/api/v1/auth/register",
		"/api/v1/auth/logout",
		"/api/v1/auth/refresh",
	}
	for _, p := range authPaths {
		if path == p {
			return true
		}
	}
	return false
}

// isExportPath checks if the path is an export endpoint.
func isExportPath(path string) bool {
	// Add export paths as needed
	// Example: /api/v1/documents/export
	return false
}

// ContextTimeout returns the remaining timeout from the context.
// Returns 0 if no deadline is set.
func ContextTimeout(ctx context.Context) time.Duration {
	deadline, ok := ctx.Deadline()
	if !ok {
		return 0
	}
	return time.Until(deadline)
}

// HasContextDeadline checks if the context has a deadline set.
func HasContextDeadline(ctx context.Context) bool {
	_, ok := ctx.Deadline()
	return ok
}
