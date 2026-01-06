package errors

// ErrorCode represents a standardized error code for API responses
type ErrorCode string

// Standard error codes
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
