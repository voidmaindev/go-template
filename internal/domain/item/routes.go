package item

import (
	"github.com/gofiber/fiber/v2"
	"github.com/voidmaindev/go-template/internal/config"
	"github.com/voidmaindev/go-template/internal/domain/user"
	"github.com/voidmaindev/go-template/internal/middleware"
)

// RegisterRoutes registers all item routes
func RegisterRoutes(router fiber.Router, handler *Handler, jwtConfig *config.JWTConfig, tokenStore *user.TokenStore) {
	items := router.Group("/items", middleware.JWTMiddleware(jwtConfig, tokenStore))
	items.Post("/", handler.Create)
	items.Get("/", handler.List)
	items.Get("/:id", handler.GetByID)
	items.Put("/:id", handler.Update)
	items.Delete("/:id", handler.Delete)
}
