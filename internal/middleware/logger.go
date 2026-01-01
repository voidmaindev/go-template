package middleware

import (
	"log/slog"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

// SetupLogger configures request logging middleware using Fiber's built-in logger
func SetupLogger(app *fiber.App) {
	app.Use(logger.New(logger.Config{
		Format:     "${time} | ${status} | ${latency} | ${ip} | ${method} | ${path} | ${error}\n",
		TimeFormat: "2006-01-02 15:04:05",
		TimeZone:   "Local",
	}))
}

// SetupSlogLogger sets up request logging using slog
func SetupSlogLogger(app *fiber.App) {
	app.Use(func(c *fiber.Ctx) error {
		start := time.Now()

		// Process request
		err := c.Next()

		// Calculate latency
		latency := time.Since(start)
		status := c.Response().StatusCode()

		// Log based on status code
		attrs := []any{
			"method", c.Method(),
			"path", c.Path(),
			"status", status,
			"latency", latency,
			"ip", c.IP(),
		}

		if err != nil {
			attrs = append(attrs, "error", err.Error())
		}

		if status >= 500 {
			slog.Error("Request failed", attrs...)
		} else if status >= 400 {
			slog.Warn("Request error", attrs...)
		} else {
			slog.Info("Request completed", attrs...)
		}

		return err
	})
}

// SetupJSONLogger sets up JSON formatted logging (for production)
func SetupJSONLogger(app *fiber.App) {
	app.Use(logger.New(logger.Config{
		Format:     `{"time":"${time}","status":${status},"latency":"${latency}","ip":"${ip}","method":"${method}","path":"${path}","error":"${error}"}` + "\n",
		TimeFormat: time.RFC3339,
		TimeZone:   "UTC",
	}))
}
