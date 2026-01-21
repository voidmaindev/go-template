package common

import (
	"github.com/gofiber/fiber/v2"
	"github.com/voidmaindev/go-template/pkg/validator"
)

// ParseAndValidate parses JSON body into a typed struct and validates it.
// Returns the parsed struct or an error response if parsing/validation fails.
// This is a generic helper to reduce boilerplate in handlers.
//
// Usage:
//
//	func (h *Handler) Create(c *fiber.Ctx) error {
//	    req, err := common.ParseAndValidate[CreateItemRequest](c)
//	    if err != nil {
//	        return err // already returns appropriate response
//	    }
//	    // use req...
//	}
func ParseAndValidate[T any](c *fiber.Ctx) (*T, error) {
	var req T
	if err := c.BodyParser(&req); err != nil {
		return nil, BadRequestResponse(c, "invalid request body")
	}

	if errs := validator.Validate(&req); errs != nil {
		return nil, ValidationErrorResponse(c, errs)
	}

	return &req, nil
}

// ParseID parses an ID parameter from the route and returns it as uint.
// Returns the ID or an error response if parsing fails.
//
// Usage:
//
//	func (h *Handler) GetByID(c *fiber.Ctx) error {
//	    id, err := common.ParseID(c, "id", "item")
//	    if err != nil {
//	        return err // already returns appropriate response
//	    }
//	    // use id...
//	}
func ParseID(c *fiber.Ctx, paramName, resourceName string) (uint, error) {
	id, err := c.ParamsInt(paramName)
	if err != nil {
		return 0, BadRequestResponse(c, "invalid "+resourceName+" ID")
	}
	return uint(id), nil
}
