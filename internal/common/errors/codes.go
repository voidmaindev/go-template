// Package errors provides domain-specific error types with HTTP mapping.
// DomainError supports error codes, stack traces, and error chaining
// via the standard errors.Is() and errors.As() interfaces.
package errors

// ErrorCode represents a standardized error code for API responses.
// Each code maps to a specific HTTP status code via HTTPStatus().
type ErrorCode string

// Standard error codes for domain errors.
// These codes are used in API responses and can be mapped to HTTP status codes.
const (
	CodeNotFound      ErrorCode = "NOT_FOUND"
	CodeAlreadyExists ErrorCode = "ALREADY_EXISTS"
	CodeValidation    ErrorCode = "VALIDATION_ERROR"
	CodeUnauthorized  ErrorCode = "UNAUTHORIZED"
	CodeForbidden     ErrorCode = "FORBIDDEN"
	CodeConflict      ErrorCode = "CONFLICT"
	CodeInternal      ErrorCode = "INTERNAL_ERROR"
	CodeBadRequest    ErrorCode = "BAD_REQUEST"
)

// HTTPStatus returns the HTTP status code for this error code
func (c ErrorCode) HTTPStatus() int {
	switch c {
	case CodeNotFound:
		return 404
	case CodeAlreadyExists, CodeConflict:
		return 409
	case CodeValidation, CodeBadRequest:
		return 400
	case CodeUnauthorized:
		return 401
	case CodeForbidden:
		return 403
	default:
		return 500
	}
}

// String returns the string representation of the error code
func (c ErrorCode) String() string {
	return string(c)
}
