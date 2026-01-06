package middleware

import (
	"log/slog"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/voidmaindev/go-template/internal/telemetry"
)

// RequestIDKey is the key used to store the request ID in fiber context
const RequestIDKey = "request_id"

// RequestIDMiddleware adds a unique request ID to each request for tracing
func RequestIDMiddleware() fiber.Handler {
	return requestid.New(requestid.Config{
		Header: "X-Request-ID",
		ContextKey: RequestIDKey,
	})
}

// GetRequestID retrieves the request ID from the context
func GetRequestID(c *fiber.Ctx) string {
	if id := c.Locals(RequestIDKey); id != nil {
		if strID, ok := id.(string); ok {
			return strID
		}
	}
	return ""
}

// SetupLogger configures request logging middleware using Fiber's built-in logger
func SetupLogger(app *fiber.App) {
	app.Use(logger.New(logger.Config{
		Format:     "${time} | ${status} | ${latency} | ${ip} | ${method} | ${path} | ${error}\n",
		TimeFormat: "2006-01-02 15:04:05",
		TimeZone:   "Local",
	}))
}

// SetupSlogLogger sets up request logging using slog with request ID support
func SetupSlogLogger(app *fiber.App) {
	app.Use(func(c *fiber.Ctx) error {
		start := time.Now()

		// Copy method and path before c.Next() as Fiber reuses buffers
		method := c.Method()
		path := c.Path()
		// Make copies of the strings to avoid buffer reuse issues
		method = string([]byte(method))
		path = string([]byte(path))

		// Process request
		err := c.Next()

		// Calculate latency
		latency := time.Since(start)
		status := c.Response().StatusCode()

		// Record Prometheus metrics (skip /metrics endpoint to avoid collection conflicts)
		if path != "/metrics" {
			telemetry.RecordHTTPRequest(
				method,
				path,
				strconv.Itoa(status),
				latency.Seconds(),
			)
		}

		// Log based on status code, include request ID for tracing
		attrs := []any{
			"request_id", GetRequestID(c),
			"method", method,
			"path", path,
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
