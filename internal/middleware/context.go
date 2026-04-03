package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/voidmaindev/go-template/internal/common/ctxutil"
)

// ContextEnrichment adds request context to Go context
// This should be used after RequestIDMiddleware
func ContextEnrichment() fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := c.UserContext()

		// Add request ID to context
		if reqID := GetRequestID(c); reqID != "" {
			ctx = ctxutil.WithRequestID(ctx, reqID)
		}

		// Add trace ID if available from headers
		if traceID := c.Get("X-Trace-ID"); traceID != "" {
			ctx = ctxutil.WithTraceID(ctx, traceID)
		}

		// Add span ID if available from headers
		if spanID := c.Get("X-Span-ID"); spanID != "" {
			ctx = ctxutil.WithSpanID(ctx, spanID)
		}

		c.SetUserContext(ctx)
		return c.Next()
	}
}

// UserContextEnrichment adds authenticated user info to Go context
// This should be used after JWTMiddleware
func UserContextEnrichment() fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := c.UserContext()

		// Add user ID to context
		if userID, ok := GetUserIDFromContext(c); ok {
			ctx = ctxutil.WithUserID(ctx, userID)
		}

		// Add user email to context
		if email := GetUserEmailFromContext(c); email != "" {
			ctx = ctxutil.WithUserEmail(ctx, email)
		}

		// Add user role to context
		if role := GetUserRoleFromContext(c); role != "" {
			ctx = ctxutil.WithUserRole(ctx, role)
		}

		c.SetUserContext(ctx)
		return c.Next()
	}
}

// Note: GetRequestID is defined in logger.go

// GetUserEmailFromContext retrieves user email from fiber context
func GetUserEmailFromContext(c *fiber.Ctx) string {
	if email := c.Locals(UserEmailKey); email != nil {
		if str, ok := email.(string); ok {
			return str
		}
	}
	return ""
}

// GetUserRoleFromContext retrieves user role from fiber context
func GetUserRoleFromContext(c *fiber.Ctx) string {
	if role := c.Locals(UserRoleKey); role != nil {
		if str, ok := role.(string); ok {
			return str
		}
	}
	return ""
}
