package middleware

import (
	"fmt"

	"github.com/casbin/casbin/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/voidmaindev/go-template/internal/common"
)

// RequirePermission creates middleware that checks RBAC permission
// using Casbin. The user must be authenticated (via JWTMiddleware)
// before this middleware runs.
func RequirePermission(enforcer *casbin.Enforcer, domain, action string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID, ok := GetUserIDFromContext(c)
		if !ok {
			return common.UnauthorizedResponse(c, "authentication required")
		}

		subject := fmt.Sprintf("user:%d", userID)
		allowed, err := enforcer.Enforce(subject, domain, action)
		if err != nil {
			return common.InternalServerErrorResponse(c)
		}

		if !allowed {
			return common.ForbiddenResponse(c, "insufficient permissions")
		}

		return c.Next()
	}
}

// RequireAnyPermission creates middleware that passes if the user has
// ANY of the specified permissions
func RequireAnyPermission(enforcer *casbin.Enforcer, permissions ...Permission) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID, ok := GetUserIDFromContext(c)
		if !ok {
			return common.UnauthorizedResponse(c, "authentication required")
		}

		subject := fmt.Sprintf("user:%d", userID)

		for _, perm := range permissions {
			allowed, err := enforcer.Enforce(subject, perm.Domain, perm.Action)
			if err != nil {
				continue
			}
			if allowed {
				return c.Next()
			}
		}

		return common.ForbiddenResponse(c, "insufficient permissions")
	}
}

// RequireAllPermissions creates middleware that requires ALL specified permissions
func RequireAllPermissions(enforcer *casbin.Enforcer, permissions ...Permission) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID, ok := GetUserIDFromContext(c)
		if !ok {
			return common.UnauthorizedResponse(c, "authentication required")
		}

		subject := fmt.Sprintf("user:%d", userID)

		for _, perm := range permissions {
			allowed, err := enforcer.Enforce(subject, perm.Domain, perm.Action)
			if err != nil {
				return common.InternalServerErrorResponse(c)
			}
			if !allowed {
				return common.ForbiddenResponse(c, "insufficient permissions")
			}
		}

		return c.Next()
	}
}

// Permission represents a domain-action pair for permission checking
type Permission struct {
	Domain string
	Action string
}

// NewPermission creates a new Permission
func NewPermission(domain, action string) Permission {
	return Permission{Domain: domain, Action: action}
}

// HasPermission checks if the current user has a specific permission.
// Returns false if the user is not authenticated or doesn't have the permission.
func HasPermission(c *fiber.Ctx, enforcer *casbin.Enforcer, domain, action string) bool {
	userID, ok := GetUserIDFromContext(c)
	if !ok {
		return false
	}

	subject := fmt.Sprintf("user:%d", userID)
	allowed, err := enforcer.Enforce(subject, domain, action)
	if err != nil {
		return false
	}

	return allowed
}
