package middleware

import (
	"log/slog"
	"runtime/debug"

	sentrygo "github.com/getsentry/sentry-go"
	"github.com/gofiber/fiber/v2"
	fiberrecover "github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/voidmaindev/go-template/internal/common"
)

// SetupRecovery configures panic recovery middleware
func SetupRecovery(app *fiber.App) {
	app.Use(fiberrecover.New(fiberrecover.Config{
		EnableStackTrace: true,
	}))
}

// SetupCustomRecovery sets up a custom recovery middleware
func SetupCustomRecovery(app *fiber.App, isDevelopment bool) {
	app.Use(func(c *fiber.Ctx) error {
		defer func() {
			if r := recover(); r != nil {
				// Log the panic with request context for tracing
				reqID := GetRequestID(c)
				userID, _ := GetUserIDFromContext(c)
				slog.Error("Panic recovered",
					"panic", r,
					"request_id", reqID,
					"user_id", userID,
					"path", c.Path(),
					"method", c.Method(),
				)

				// Capture in Sentry — the per-request hub already carries
				// request_id/route/user_id tags. RecoverWithContext attaches
				// the panic value and a stack trace.
				if hub := sentrygo.GetHubFromContext(c.UserContext()); hub != nil {
					hub.RecoverWithContext(c.UserContext(), r)
				} else {
					sentrygo.CurrentHub().Recover(r)
				}

				if isDevelopment {
					// In development, log the stack trace
					slog.Debug("Stack trace", "trace", string(debug.Stack()))
				}

				// Return a 500 error response
				_ = common.InternalServerErrorResponse(c)
			}
		}()

		return c.Next()
	})
}

// SetupRecoveryWithCallback sets up recovery with a custom callback
func SetupRecoveryWithCallback(app *fiber.App, callback func(c *fiber.Ctx, err any)) {
	app.Use(func(c *fiber.Ctx) error {
		defer func() {
			if r := recover(); r != nil {
				if callback != nil {
					callback(c, r)
				}
				_ = common.InternalServerErrorResponse(c)
			}
		}()

		return c.Next()
	})
}
