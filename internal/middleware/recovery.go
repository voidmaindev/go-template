package middleware

import (
	"log"
	"runtime/debug"

	"github.com/gofiber/fiber/v2"
	fiberrecover "github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/voidmaindev/GoTemplate/internal/common"
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
				// Log the panic
				log.Printf("Panic recovered: %v\n", r)

				if isDevelopment {
					// In development, log the stack trace
					log.Printf("Stack trace:\n%s", debug.Stack())
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
