package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/voidmaindev/GoTemplate/internal/config"
)

// CORS configuration constants
const (
	corsMaxAge        = 86400 // 24 hours in seconds
	corsAllowMethods  = "GET,POST,PUT,PATCH,DELETE,OPTIONS"
	corsAllowHeaders  = "Origin,Content-Type,Accept,Authorization,X-Requested-With"
	corsExposeHeaders = "Content-Length,Content-Type"
)

// SetupCORS configures CORS middleware
func SetupCORS(app *fiber.App, cfg *config.Config) {
	app.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.CORS.AllowedOrigins,
		AllowMethods:     corsAllowMethods,
		AllowHeaders:     corsAllowHeaders,
		AllowCredentials: true,
		ExposeHeaders:    corsExposeHeaders,
		MaxAge:           corsMaxAge,
	}))
}

// NewCORSConfig creates a custom CORS configuration
func NewCORSConfig(allowedOrigins []string) cors.Config {
	return cors.Config{
		AllowOriginsFunc: func(origin string) bool {
			// Allow all origins in development
			for _, allowed := range allowedOrigins {
				if allowed == "*" || strings.EqualFold(origin, allowed) {
					return true
				}
			}
			return false
		},
		AllowMethods:     corsAllowMethods,
		AllowHeaders:     corsAllowHeaders,
		AllowCredentials: true,
		ExposeHeaders:    corsExposeHeaders,
		MaxAge:           corsMaxAge,
	}
}
