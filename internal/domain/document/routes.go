package document

import (
	"github.com/gofiber/fiber/v2"
	"github.com/voidmaindev/go-template/internal/config"
	"github.com/voidmaindev/go-template/internal/domain/user"
	"github.com/voidmaindev/go-template/internal/middleware"
)

// RegisterRoutes registers all document routes
func RegisterRoutes(router fiber.Router, handler *Handler, jwtConfig *config.JWTConfig, tokenStore *user.TokenStore) {
	documents := router.Group("/documents", middleware.JWTMiddleware(jwtConfig, tokenStore))

	// Document CRUD
	documents.Post("/", handler.Create)
	documents.Get("/", handler.List)
	documents.Get("/:id", handler.GetByID)
	documents.Put("/:id", handler.Update)
	documents.Delete("/:id", handler.Delete)

	// Document items
	documents.Post("/:id/items", handler.AddItem)
	documents.Put("/:id/items/:itemId", handler.UpdateItem)
	documents.Delete("/:id/items/:itemId", handler.RemoveItem)
}
