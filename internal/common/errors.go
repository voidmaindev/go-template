package common

import (
	"errors"
	"fmt"
)

// Common application errors
var (
	ErrNotFound          = errors.New("resource not found")
	ErrAlreadyExists     = errors.New("resource already exists")
	ErrInvalidInput      = errors.New("invalid input")
	ErrUnauthorized      = errors.New("unauthorized")
	ErrForbidden         = errors.New("forbidden")
	ErrInternalServer    = errors.New("internal server error")
	ErrBadRequest        = errors.New("bad request")
	ErrConflict          = errors.New("conflict")
	ErrValidation        = errors.New("validation error")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrTokenExpired      = errors.New("token expired")
	ErrTokenInvalid      = errors.New("invalid token")
	ErrTokenBlacklisted  = errors.New("token has been revoked")
)

// AppError represents an application error with additional context
type AppError struct {
	Err     error
	Message string
	Code    string
	Details any
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
	return errors.Is(err, ErrNotFound)
}

// IsUnauthorizedError checks if the error is an unauthorized error
func IsUnauthorizedError(err error) bool {
	return errors.Is(err, ErrUnauthorized)
}

// IsForbiddenError checks if the error is a forbidden error
func IsForbiddenError(err error) bool {
	return errors.Is(err, ErrForbidden)
}

// IsValidationError checks if the error is a validation error
func IsValidationError(err error) bool {
	return errors.Is(err, ErrValidation)
}

// IsConflictError checks if the error is a conflict error
func IsConflictError(err error) bool {
	return errors.Is(err, ErrConflict) || errors.Is(err, ErrAlreadyExists)
}
