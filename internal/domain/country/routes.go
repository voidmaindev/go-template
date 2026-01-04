package country

import (
	"github.com/gofiber/fiber/v2"
	"github.com/voidmaindev/go-template/internal/config"
	"github.com/voidmaindev/go-template/internal/domain/user"
	"github.com/voidmaindev/go-template/internal/middleware"
)

// RegisterRoutes registers all country routes
func RegisterRoutes(router fiber.Router, handler *Handler, jwtConfig *config.JWTConfig, tokenStore *user.TokenStore) {
	countries := router.Group("/countries", middleware.JWTMiddleware(jwtConfig, tokenStore))
	countries.Post("/", handler.Create)
	countries.Get("/", handler.List)
	countries.Get("/:id", handler.GetByID)
	countries.Put("/:id", handler.Update)
	countries.Delete("/:id", handler.Delete)
}
