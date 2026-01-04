package city

import (
	"github.com/gofiber/fiber/v2"
	"github.com/voidmaindev/go-template/internal/config"
	"github.com/voidmaindev/go-template/internal/domain/user"
	"github.com/voidmaindev/go-template/internal/middleware"
)

// RegisterRoutes registers all city routes
func RegisterRoutes(router fiber.Router, handler *Handler, jwtConfig *config.JWTConfig, tokenStore *user.TokenStore) {
	cities := router.Group("/cities", middleware.JWTMiddleware(jwtConfig, tokenStore))
	cities.Post("/", handler.Create)
	cities.Get("/", handler.List)
	cities.Get("/:id", handler.GetByID)
	cities.Put("/:id", handler.Update)
	cities.Delete("/:id", handler.Delete)

	// Nested route for cities by country
	countries := router.Group("/countries", middleware.JWTMiddleware(jwtConfig, tokenStore))
	countries.Get("/:countryId/cities", handler.ListByCountry)
}
