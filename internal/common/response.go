package common

import (
	"github.com/gofiber/fiber/v2"
	"github.com/voidmaindev/go-template/internal/common/errors"
)

// Response represents a standard API response
type Response struct {
	Success   bool       `json:"success"`
	Message   string     `json:"message,omitempty"`
	Data      any        `json:"data,omitempty"`
	Error     *ErrorInfo `json:"error,omitempty"`
	RequestID string     `json:"request_id,omitempty"`
}

// ErrorInfo provides a consistent error structure for all API error responses.
// Both simple errors (BadRequest, NotFound) and domain errors (HandleDomainError)
// use this same shape, so API consumers always see the same structure.
type ErrorInfo struct {
	Code    string         `json:"code"`
	Message string         `json:"message"`
	Domain  string         `json:"domain,omitempty"`
	Details map[string]any `json:"details,omitempty"`
}

// SuccessResponse sends a success response with data
func SuccessResponse(c *fiber.Ctx, data any) error {
	return c.Status(fiber.StatusOK).JSON(Response{
		Success: true,
		Data:    data,
	})
}

// SuccessResponseWithMessage sends a success response with message and data
func SuccessResponseWithMessage(c *fiber.Ctx, message string, data any) error {
	return c.Status(fiber.StatusOK).JSON(Response{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// CreatedResponse sends a 201 Created response
func CreatedResponse(c *fiber.Ctx, data any) error {
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

// DeletedResponse sends a 204 No Content response for deletion
func DeletedResponse(c *fiber.Ctx) error {
	return c.SendStatus(fiber.StatusNoContent)
}

// httpStatusToCode maps HTTP status codes to error codes for simple error responses.
func httpStatusToCode(statusCode int) string {
	switch statusCode {
	case 400:
		return string(errors.CodeBadRequest)
	case 401:
		return string(errors.CodeUnauthorized)
	case 403:
		return string(errors.CodeForbidden)
	case 404:
		return string(errors.CodeNotFound)
	case 409:
		return string(errors.CodeConflict)
	case 429:
		return string(errors.CodeTooManyRequests)
	case 503:
		return "SERVICE_UNAVAILABLE"
	default:
		return string(errors.CodeInternal)
	}
}

// ErrorResponse sends an error response with a status code
func ErrorResponse(c *fiber.Ctx, statusCode int, message string) error {
	requestID := getRequestID(c)
	return c.Status(statusCode).JSON(Response{
		Success: false,
		Error: &ErrorInfo{
			Code:    httpStatusToCode(statusCode),
			Message: message,
		},
		RequestID: requestID,
	})
}

// getRequestID extracts the request ID from the Fiber context.
func getRequestID(c *fiber.Ctx) string {
	if id := c.Locals("request_id"); id != nil {
		if str, ok := id.(string); ok {
			return str
		}
	}
	return ""
}

// ErrorResponseWithDetails sends an error response with details
func ErrorResponseWithDetails(c *fiber.Ctx, statusCode int, message string, details any) error {
	requestID := getRequestID(c)
	return c.Status(statusCode).JSON(Response{
		Success: false,
		Error: &ErrorInfo{
			Code:    httpStatusToCode(statusCode),
			Message: message,
		},
		Data:      details,
		RequestID: requestID,
	})
}

// ValidationErrorResponse sends a 400 response for validation errors
func ValidationErrorResponse(c *fiber.Ctx, validationErrors any) error {
	requestID := getRequestID(c)
	return c.Status(fiber.StatusBadRequest).JSON(Response{
		Success: false,
		Error: &ErrorInfo{
			Code:    string(errors.CodeValidation),
			Message: "validation failed",
		},
		Data:      validationErrors,
		RequestID: requestID,
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

// TooManyRequestsResponse sends a 429 Too Many Requests response
func TooManyRequestsResponse(c *fiber.Ctx, message string) error {
	if message == "" {
		message = "too many requests"
	}
	return ErrorResponse(c, fiber.StatusTooManyRequests, message)
}

// InternalServerErrorResponse sends a 500 Internal Server Error response
func InternalServerErrorResponse(c *fiber.Ctx) error {
	return ErrorResponse(c, fiber.StatusInternalServerError, "internal server error")
}

// ServiceUnavailableResponse sends a 503 Service Unavailable response
func ServiceUnavailableResponse(c *fiber.Ctx, message string) error {
	if message == "" {
		message = "service temporarily unavailable"
	}
	return ErrorResponse(c, fiber.StatusServiceUnavailable, message)
}

// HandleError handles errors and returns appropriate HTTP responses.
// All domain errors are typed DomainError instances with HTTP status codes.
func HandleError(c *fiber.Ctx, err error) error {
	if err == nil {
		return nil
	}

	// Check for typed domain errors
	if de := errors.GetDomainError(err); de != nil {
		return HandleDomainError(c, de)
	}

	// Unknown error — return generic 500
	return InternalServerErrorResponse(c)
}

// HandleDomainError handles typed domain errors with structured response
func HandleDomainError(c *fiber.Ctx, de *errors.DomainError) error {
	requestID := getRequestID(c)

	errInfo := &ErrorInfo{
		Code:    string(de.Code),
		Message: de.Message,
		Domain:  de.Domain,
	}

	if len(de.Details) > 0 {
		errInfo.Details = de.Details
	}

	return c.Status(de.HTTPStatus()).JSON(Response{
		Success:   false,
		Error:     errInfo,
		RequestID: requestID,
	})
}

// HandleErrorWithDomain is a convenience function that wraps errors in domain context
func HandleErrorWithDomain(c *fiber.Ctx, domain string, err error) error {
	if err == nil {
		return nil
	}

	// If it's already a domain error, handle it directly
	if de := errors.GetDomainError(err); de != nil {
		return HandleDomainError(c, de)
	}

	// Wrap unknown errors as internal errors
	de := errors.Internal(domain, err)
	return HandleDomainError(c, de)
}
