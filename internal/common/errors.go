package common

import (
	"errors"
	"fmt"

	commonerrors "github.com/voidmaindev/go-template/internal/common/errors"
)

// Legacy application errors.
// Deprecated: Use the typed DomainError system from internal/common/errors package instead.
// These are kept for backward compatibility. New code should use:
//   - errors.NotFound(domain, resource) instead of ErrNotFound
//   - errors.AlreadyExists(domain, resource) instead of ErrAlreadyExists
//   - errors.Validation(domain, message) instead of ErrValidation
//   - errors.Unauthorized(domain, message) instead of ErrUnauthorized
//   - errors.Forbidden(domain, message) instead of ErrForbidden
//   - errors.Internal(domain, err) instead of ErrInternalServer
var (
	// Deprecated: Use errors.NotFound() instead
	ErrNotFound = errors.New("resource not found")
	// Deprecated: Use errors.AlreadyExists() instead
	ErrAlreadyExists = errors.New("resource already exists")
	// Deprecated: Use errors.Validation() instead
	ErrInvalidInput = errors.New("invalid input")
	// Deprecated: Use errors.Unauthorized() instead
	ErrUnauthorized = errors.New("unauthorized")
	// Deprecated: Use errors.Forbidden() instead
	ErrForbidden = errors.New("forbidden")
	// Deprecated: Use errors.Internal() instead
	ErrInternalServer = errors.New("internal server error")
	// Deprecated: Use errors.Validation() instead
	ErrBadRequest = errors.New("bad request")
	// Deprecated: Use errors.Conflict() instead
	ErrConflict = errors.New("conflict")
	// Deprecated: Use errors.Validation() instead
	ErrValidation = errors.New("validation error")
	// Deprecated: Use errors.Unauthorized() with appropriate message instead
	ErrInvalidCredentials = errors.New("invalid credentials")
	// Deprecated: Use errors.Unauthorized() with appropriate message instead
	ErrTokenExpired = errors.New("token expired")
	// Deprecated: Use errors.Unauthorized() with appropriate message instead
	ErrTokenInvalid = errors.New("invalid token")
	// Deprecated: Use errors.Unauthorized() with appropriate message instead
	ErrTokenBlacklisted = errors.New("token has been revoked")
)

// AppError represents an application error with additional context
type AppError struct {
	Err       error
	Message   string
	Code      string
	Details   any
	RequestID string // Correlation ID from X-Request-ID header
	TraceID   string // OpenTelemetry trace ID
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	if e.Err != nil {
		return e.Err.Error()
	}
	return "unknown error"
}

// Unwrap returns the wrapped error
func (e *AppError) Unwrap() error {
	return e.Err
}

// NewAppError creates a new AppError
func NewAppError(err error, message string) *AppError {
	return &AppError{
		Err:     err,
		Message: message,
	}
}

// NewAppErrorWithCode creates a new AppError with a code
func NewAppErrorWithCode(err error, message, code string) *AppError {
	return &AppError{
		Err:     err,
		Message: message,
		Code:    code,
	}
}

// NewAppErrorWithDetails creates a new AppError with details
func NewAppErrorWithDetails(err error, message string, details any) *AppError {
	return &AppError{
		Err:     err,
		Message: message,
		Details: details,
	}
}

// WrapError wraps an error with a message
func WrapError(err error, message string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", message, err)
}

// IsNotFoundError checks if the error is a not found error (supports both legacy and domain errors)
func IsNotFoundError(err error) bool {
	return errors.Is(err, ErrNotFound) || commonerrors.IsNotFound(err)
}

// IsUnauthorizedError checks if the error is an unauthorized error (supports both legacy and domain errors)
func IsUnauthorizedError(err error) bool {
	return errors.Is(err, ErrUnauthorized) || commonerrors.IsUnauthorized(err)
}

// IsForbiddenError checks if the error is a forbidden error (supports both legacy and domain errors)
func IsForbiddenError(err error) bool {
	return errors.Is(err, ErrForbidden) || commonerrors.IsForbidden(err)
}

// IsValidationError checks if the error is a validation error (supports both legacy and domain errors)
func IsValidationError(err error) bool {
	return errors.Is(err, ErrValidation) || commonerrors.IsValidation(err)
}

// IsConflictError checks if the error is a conflict error (supports both legacy and domain errors)
func IsConflictError(err error) bool {
	return errors.Is(err, ErrConflict) || errors.Is(err, ErrAlreadyExists) || commonerrors.IsConflict(err) || commonerrors.IsAlreadyExists(err)
}

// NewAppErrorWithContext creates an AppError with request and trace context.
func NewAppErrorWithContext(err error, message, requestID, traceID string) *AppError {
	return &AppError{
		Err:       err,
		Message:   message,
		RequestID: requestID,
		TraceID:   traceID,
	}
}

// WithContext adds request and trace context to an existing AppError.
func (e *AppError) WithContext(requestID, traceID string) *AppError {
	e.RequestID = requestID
	e.TraceID = traceID
	return e
}
