package middleware

import (
	"log/slog"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/voidmaindev/go-template/internal/config"
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
	allowedOrigins := parseOrigins(cfg.CORS.AllowedOrigins)
	allowCredentials := true

	// Security: Wildcard origin with credentials violates CORS specification
	// In production, we must use specific origins when credentials are enabled
	if containsWildcard(allowedOrigins) {
		if cfg.App.IsProduction() {
			slog.Warn("CORS: Wildcard origin (*) with credentials is insecure, disabling credentials")
			allowCredentials = false
		} else {
			slog.Warn("CORS: Wildcard origin (*) detected in development mode")
		}
	}

	app.Use(cors.New(cors.Config{
		AllowOriginsFunc: func(origin string) bool {
			return isOriginAllowed(origin, allowedOrigins, allowCredentials)
		},
		AllowMethods:     corsAllowMethods,
		AllowHeaders:     corsAllowHeaders,
		AllowCredentials: allowCredentials,
		ExposeHeaders:    corsExposeHeaders,
		MaxAge:           corsMaxAge,
	}))
}

// parseOrigins splits the comma-separated origins string into a slice
func parseOrigins(origins string) []string {
	if origins == "" {
		return nil
	}
	parts := strings.Split(origins, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		if trimmed := strings.TrimSpace(p); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

// containsWildcard checks if the allowed origins list contains a wildcard
func containsWildcard(origins []string) bool {
	for _, o := range origins {
		if o == "*" {
			return true
		}
	}
	return false
}

// isOriginAllowed checks if the origin is in the allowed list
func isOriginAllowed(origin string, allowedOrigins []string, allowCredentials bool) bool {
	for _, allowed := range allowedOrigins {
		if allowed == "*" {
			// Wildcard with credentials is forbidden by CORS spec
			// If credentials are enabled, deny wildcard matches
			if allowCredentials {
				continue
			}
			return true
		}
		if strings.EqualFold(origin, allowed) {
			return true
		}
	}
	return false
}

// NewCORSConfig creates a custom CORS configuration
func NewCORSConfig(allowedOrigins []string, allowCredentials bool) cors.Config {
	return cors.Config{
		AllowOriginsFunc: func(origin string) bool {
			return isOriginAllowed(origin, allowedOrigins, allowCredentials)
		},
		AllowMethods:     corsAllowMethods,
		AllowHeaders:     corsAllowHeaders,
		AllowCredentials: allowCredentials,
		ExposeHeaders:    corsExposeHeaders,
		MaxAge:           corsMaxAge,
	}
}
