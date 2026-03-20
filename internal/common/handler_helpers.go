package common

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/voidmaindev/go-template/pkg/validator"
)

// errResponseSent is a sentinel error indicating that an HTTP response has
// already been written. Handlers receiving this error should return nil to
// Fiber (the response is already buffered).
var errResponseSent = errors.New("response already sent")

// ParseAndValidate parses JSON body into a typed struct and validates it.
// On failure, it writes the appropriate 400 response and returns a non-nil error.
// Handlers should return nil when this error is received.
//
// Usage:
//
//	func (h *Handler) Create(c *fiber.Ctx) error {
//	    req, err := common.ParseAndValidate[CreateItemRequest](c)
//	    if err != nil {
//	        return nil // response already sent
//	    }
//	    item, err := h.service.Create(c.Context(), req)
//	    if err != nil {
//	        return common.HandleError(c, err)
//	    }
//	    return common.CreatedResponse(c, item.ToResponse())
//	}
func ParseAndValidate[T any](c *fiber.Ctx) (*T, error) {
	var req T
	if err := c.BodyParser(&req); err != nil {
		_ = BadRequestResponse(c, "invalid request body")
		return nil, errResponseSent
	}

	if errs := validator.Validate(&req); errs != nil {
		_ = ValidationErrorResponse(c, errs)
		return nil, errResponseSent
	}

	return &req, nil
}

// ParseID parses an ID parameter from the route and returns it as uint.
// On failure, it writes a 400 response and returns a non-nil error.
// Handlers should return nil when this error is received.
//
// Usage:
//
//	func (h *Handler) GetByID(c *fiber.Ctx) error {
//	    id, err := common.ParseID(c, "id", "item")
//	    if err != nil {
//	        return nil // response already sent
//	    }
//	    item, err := h.service.GetByID(c.Context(), id)
//	    if err != nil {
//	        return common.HandleError(c, err)
//	    }
//	    return common.SuccessResponse(c, item.ToResponse())
//	}
func ParseID(c *fiber.Ctx, paramName, resourceName string) (uint, error) {
	id, err := c.ParamsInt(paramName)
	if err != nil || id < 1 {
		_ = BadRequestResponse(c, "invalid "+resourceName+" ID")
		return 0, errResponseSent
	}
	return uint(id), nil
}
