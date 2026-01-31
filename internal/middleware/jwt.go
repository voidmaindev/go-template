package middleware

import (
	"context"
	"strings"
	"time"

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

// TokenInvalidator interface for checking if user's tokens have been invalidated
type TokenInvalidator interface {
	GetTokensInvalidatedAt(ctx context.Context, userID uint) (time.Time, error)
}

// JWTMiddleware creates JWT authentication middleware
func JWTMiddleware(cfg *config.JWTConfig, blacklist TokenBlacklist) fiber.Handler {
	return JWTMiddlewareWithInvalidator(cfg, blacklist, nil)
}

// JWTMiddlewareWithInvalidator creates JWT authentication middleware with token invalidation support
func JWTMiddlewareWithInvalidator(cfg *config.JWTConfig, blacklist TokenBlacklist, invalidator TokenInvalidator) fiber.Handler {
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

			// Validate token claims with safe type assertions
			user, ok := c.Locals("user").(*jwt.Token)
			if !ok || user == nil {
				return common.UnauthorizedResponse(c, "invalid token")
			}
			claims, ok := user.Claims.(jwt.MapClaims)
			if !ok {
				return common.UnauthorizedResponse(c, "invalid token claims")
			}

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

			// Extract user_id from claims
			userIDFloat, ok := claims["user_id"].(float64)
			if !ok {
				return common.UnauthorizedResponse(c, "invalid token claims")
			}
			userID := uint(userIDFloat)

			// Check if token was issued before user's tokens were invalidated (e.g., password change)
			if invalidator != nil {
				invalidatedAt, err := invalidator.GetTokensInvalidatedAt(c.Context(), userID)
				if err != nil {
					return common.InternalServerErrorResponse(c)
				}
				if !invalidatedAt.IsZero() {
					// Get token's issued-at time
					if iatFloat, ok := claims["iat"].(float64); ok {
						issuedAt := time.Unix(int64(iatFloat), 0)
						if issuedAt.Before(invalidatedAt) {
							return common.UnauthorizedResponse(c, "session expired, please login again")
						}
					}
				}
			}

			// Store user_id in context for rate limiting and other middleware
			c.Locals(UserIDKey, userID)

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
		c.Locals(UserIDKey, claims.UserID)
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

// getClaimsFromToken is a helper that extracts MapClaims from the JWT token stored in context.
// This avoids duplicating the same type assertion logic across multiple functions.
func getClaimsFromToken(c *fiber.Ctx) (jwt.MapClaims, bool) {
	user, ok := c.Locals("user").(*jwt.Token)
	if !ok || user == nil {
		return nil, false
	}

	claims, ok := user.Claims.(jwt.MapClaims)
	return claims, ok
}

// GetUserIDFromContext extracts the user ID from the Fiber context
func GetUserIDFromContext(c *fiber.Ctx) (uint, bool) {
	claims, ok := getClaimsFromToken(c)
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
	claims, ok := getClaimsFromToken(c)
	if !ok {
		return "", false
	}

	email, ok := claims["email"].(string)
	return email, ok
}

// GetClaimsFromContext extracts all claims from the Fiber context
func GetClaimsFromContext(c *fiber.Ctx) (jwt.MapClaims, bool) {
	return getClaimsFromToken(c)
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

