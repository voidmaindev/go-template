package sentry

import (
	"strconv"

	sentrygo "github.com/getsentry/sentry-go"
	"github.com/gofiber/fiber/v2"

	"github.com/voidmaindev/go-template/internal/middleware"
)

// Hub returns Fiber middleware that attaches a per-request Sentry hub to the
// request context with request_id, route, method, and (when authenticated)
// user_id. Downstream code reads this via sentrygo.GetHubFromContext.
//
// Routes are tagged using c.Route().Path (template form like /api/v1/users/:id),
// not c.Path() — this keeps Sentry tag cardinality low.
func Hub() fiber.Handler {
	return func(c *fiber.Ctx) error {
		if !enabled {
			return c.Next()
		}

		hub := sentrygo.CurrentHub().Clone()
		hub.Scope().SetTag("request_id", middleware.GetRequestID(c))
		hub.Scope().SetTag("method", c.Method())
		if route := c.Route(); route != nil && route.Path != "" {
			hub.Scope().SetTag("route", route.Path)
		}
		if uid, ok := middleware.GetUserIDFromContext(c); ok {
			hub.Scope().SetUser(sentrygo.User{ID: strconv.FormatUint(uint64(uid), 10)})
		}

		ctx := sentrygo.SetHubOnContext(c.UserContext(), hub)
		c.SetUserContext(ctx)
		return c.Next()
	}
}
