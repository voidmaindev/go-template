package common

import (
	"fmt"

	commonerrors "github.com/voidmaindev/go-template/internal/common/errors"
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

// IsNotFoundError checks if the error is a not found error
func IsNotFoundError(err error) bool {
	return commonerrors.IsNotFound(err)
}

// IsUnauthorizedError checks if the error is an unauthorized error
func IsUnauthorizedError(err error) bool {
	return commonerrors.IsUnauthorized(err)
}

// IsForbiddenError checks if the error is a forbidden error
func IsForbiddenError(err error) bool {
	return commonerrors.IsForbidden(err)
}

// IsValidationError checks if the error is a validation error
func IsValidationError(err error) bool {
	return commonerrors.IsValidation(err)
}

// IsConflictError checks if the error is a conflict error
func IsConflictError(err error) bool {
	return commonerrors.IsConflict(err) || commonerrors.IsAlreadyExists(err)
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
