package user

import (
	"github.com/gofiber/fiber/v2"
	"github.com/voidmaindev/GoTemplate/internal/config"
	"github.com/voidmaindev/GoTemplate/internal/middleware"
)

// RegisterRoutes registers all user routes
func RegisterRoutes(router fiber.Router, handler *Handler, jwtConfig *config.JWTConfig, tokenStore *TokenStore) {
	// Auth routes (public)
	auth := router.Group("/auth")
	auth.Post("/register", handler.Register)
	auth.Post("/login", handler.Login)
	auth.Post("/refresh", handler.RefreshToken)

	// Auth routes (protected)
	authProtected := auth.Group("", middleware.JWTMiddleware(jwtConfig, tokenStore))
	authProtected.Post("/logout", handler.Logout)

	// User routes (protected)
	users := router.Group("/users", middleware.JWTMiddleware(jwtConfig, tokenStore))
	users.Get("/me", handler.GetMe)
	users.Put("/me", handler.UpdateMe)
	users.Put("/me/password", handler.ChangePassword)
	users.Get("/", handler.List)
	users.Get("/:id", handler.GetByID)
	users.Delete("/:id", handler.Delete)
}
