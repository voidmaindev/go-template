package common

import (
	"github.com/gofiber/fiber/v2"
)

// Response represents a standard API response
type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   interface{} `json:"error,omitempty"`
}

// SuccessResponse sends a success response with data
func SuccessResponse(c *fiber.Ctx, data interface{}) error {
	return c.Status(fiber.StatusOK).JSON(Response{
		Success: true,
		Data:    data,
	})
}

// SuccessResponseWithMessage sends a success response with message and data
func SuccessResponseWithMessage(c *fiber.Ctx, message string, data interface{}) error {
	return c.Status(fiber.StatusOK).JSON(Response{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// CreatedResponse sends a 201 Created response
func CreatedResponse(c *fiber.Ctx, data interface{}) error {
	return c.Status(fiber.StatusCreated).JSON(Response{
		Success: true,
		Message: "created successfully",
		Data:    data,
	})
}

// NoContentResponse sends a 204 No Content response
func NoContentResponse(c *fiber.Ctx) error {
	return c.SendStatus(fiber.StatusNoContent)
}

// DeletedResponse sends a success response for deletion
func DeletedResponse(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON(Response{
		Success: true,
		Message: "deleted successfully",
	})
}

// ErrorResponse sends an error response with a status code
func ErrorResponse(c *fiber.Ctx, statusCode int, message string) error {
	return c.Status(statusCode).JSON(Response{
		Success: false,
		Error:   message,
	})
}

// ErrorResponseWithDetails sends an error response with details
func ErrorResponseWithDetails(c *fiber.Ctx, statusCode int, message string, details interface{}) error {
	return c.Status(statusCode).JSON(Response{
		Success: false,
		Error:   message,
		Data:    details,
	})
}

// ValidationErrorResponse sends a 400 response for validation errors
func ValidationErrorResponse(c *fiber.Ctx, errors interface{}) error {
	return c.Status(fiber.StatusBadRequest).JSON(Response{
		Success: false,
		Error:   "validation failed",
		Data:    errors,
	})
}

// BadRequestResponse sends a 400 Bad Request response
func BadRequestResponse(c *fiber.Ctx, message string) error {
	if message == "" {
		message = "bad request"
	}
	return ErrorResponse(c, fiber.StatusBadRequest, message)
}

// UnauthorizedResponse sends a 401 Unauthorized response
func UnauthorizedResponse(c *fiber.Ctx, message string) error {
	if message == "" {
		message = "unauthorized"
	}
	return ErrorResponse(c, fiber.StatusUnauthorized, message)
}

// ForbiddenResponse sends a 403 Forbidden response
func ForbiddenResponse(c *fiber.Ctx, message string) error {
	if message == "" {
		message = "forbidden"
	}
	return ErrorResponse(c, fiber.StatusForbidden, message)
}

// NotFoundResponse sends a 404 Not Found response
func NotFoundResponse(c *fiber.Ctx, resource string) error {
	message := "resource not found"
	if resource != "" {
		message = resource + " not found"
	}
	return ErrorResponse(c, fiber.StatusNotFound, message)
}

// ConflictResponse sends a 409 Conflict response
func ConflictResponse(c *fiber.Ctx, message string) error {
	if message == "" {
		message = "resource already exists"
	}
	return ErrorResponse(c, fiber.StatusConflict, message)
}

// InternalServerErrorResponse sends a 500 Internal Server Error response
func InternalServerErrorResponse(c *fiber.Ctx) error {
	return ErrorResponse(c, fiber.StatusInternalServerError, "internal server error")
}

// HandleError handles common errors and returns appropriate responses
func HandleError(c *fiber.Ctx, err error) error {
	if err == nil {
		return nil
	}

	switch {
	case IsNotFoundError(err):
		return NotFoundResponse(c, "")
	case IsUnauthorizedError(err):
		return UnauthorizedResponse(c, err.Error())
	case IsForbiddenError(err):
		return ForbiddenResponse(c, err.Error())
	case IsValidationError(err):
		return BadRequestResponse(c, err.Error())
	case IsConflictError(err):
		return ConflictResponse(c, err.Error())
	default:
		return InternalServerErrorResponse(c)
	}
}
