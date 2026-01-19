package middleware

import (
	"context"
	"strings"

	"github.com/gofiber/fiber/v2"
	jwtware "github.com/gofiber/contrib/jwt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/voidmaindev/go-template/internal/common"
	"github.com/voidmaindev/go-template/internal/config"
	"github.com/voidmaindev/go-template/pkg/utils"
)

// TokenBlacklist interface for checking blacklisted tokens
type TokenBlacklist interface {
	IsBlacklisted(ctx context.Context, token string) (bool, error)
}

// JWTMiddleware creates JWT authentication middleware
func JWTMiddleware(cfg *config.JWTConfig, blacklist TokenBlacklist) fiber.Handler {
	return jwtware.New(jwtware.Config{
		SigningKey: jwtware.SigningKey{Key: []byte(cfg.SecretKey)},
		ContextKey: "user",
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			if err.Error() == "Missing or malformed JWT" {
				return common.UnauthorizedResponse(c, "missing or malformed token")
			}
			return common.UnauthorizedResponse(c, "invalid or expired token")
		},
		SuccessHandler: func(c *fiber.Ctx) error {
			// Check if token is blacklisted
			if blacklist != nil {
				token := extractToken(c)
				if token != "" {
					isBlacklisted, err := blacklist.IsBlacklisted(c.Context(), token)
					if err != nil {
						return common.InternalServerErrorResponse(c)
					}
					if isBlacklisted {
						return common.UnauthorizedResponse(c, "token has been revoked")
					}
				}
			}

			// Validate token claims
			user := c.Locals("user").(*jwt.Token)
			claims := user.Claims.(jwt.MapClaims)

			// Validate token type is access token
			tokenType, ok := claims["token_type"].(string)
			if !ok || tokenType != string(utils.AccessToken) {
				return common.UnauthorizedResponse(c, "invalid token type")
			}

			// Validate issuer matches expected issuer
			if cfg.Issuer != "" {
				issuer, ok := claims["iss"].(string)
				if !ok || issuer != cfg.Issuer {
					return common.UnauthorizedResponse(c, "invalid token issuer")
				}
			}

			return c.Next()
		},
	})
}

// OptionalJWTMiddleware creates optional JWT middleware (doesn't require auth but parses token if present)
func OptionalJWTMiddleware(cfg *config.JWTConfig) fiber.Handler {
	return func(c *fiber.Ctx) error {
		token := extractToken(c)
		if token == "" {
			return c.Next()
		}

		claims, err := utils.ValidateAccessToken(token, cfg.SecretKey)
		if err != nil {
			// Token is invalid but we don't require auth, so continue
			return c.Next()
		}

		// Store claims in context
		c.Locals("user_id", claims.UserID)
		c.Locals("email", claims.Email)
		c.Locals("claims", claims)

		return c.Next()
	}
}

// extractToken extracts the JWT token from the Authorization header
func extractToken(c *fiber.Ctx) string {
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return ""
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return ""
	}

	return parts[1]
}

// GetUserIDFromContext extracts the user ID from the Fiber context
func GetUserIDFromContext(c *fiber.Ctx) (uint, bool) {
	user, ok := c.Locals("user").(*jwt.Token)
	if !ok {
		return 0, false
	}

	claims, ok := user.Claims.(jwt.MapClaims)
	if !ok {
		return 0, false
	}

	userIDFloat, ok := claims["user_id"].(float64)
	if !ok {
		return 0, false
	}

	return uint(userIDFloat), true
}

// GetEmailFromContext extracts the email from the Fiber context
func GetEmailFromContext(c *fiber.Ctx) (string, bool) {
	user, ok := c.Locals("user").(*jwt.Token)
	if !ok {
		return "", false
	}

	claims, ok := user.Claims.(jwt.MapClaims)
	if !ok {
		return "", false
	}

	email, ok := claims["email"].(string)
	return email, ok
}

// GetClaimsFromContext extracts all claims from the Fiber context
func GetClaimsFromContext(c *fiber.Ctx) (jwt.MapClaims, bool) {
	user, ok := c.Locals("user").(*jwt.Token)
	if !ok {
		return nil, false
	}

	claims, ok := user.Claims.(jwt.MapClaims)
	return claims, ok
}

// GetTokenFromContext extracts the raw token string from the Fiber context
func GetTokenFromContext(c *fiber.Ctx) string {
	return extractToken(c)
}

// RequireAuth is a simple middleware that just checks if user is authenticated
func RequireAuth() fiber.Handler {
	return func(c *fiber.Ctx) error {
		if _, ok := GetUserIDFromContext(c); !ok {
			return common.UnauthorizedResponse(c, "authentication required")
		}
		return c.Next()
	}
}

