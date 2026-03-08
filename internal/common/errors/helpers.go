package errors

import "errors"

// NotFound creates a not found error for a domain entity
func NotFound(domain, entity string) *DomainError {
	return New(domain, CodeNotFound).
		WithMessagef("%s not found", entity)
}

// AlreadyExists creates an already exists error for a domain entity
func AlreadyExists(domain, entity, field string) *DomainError {
	return New(domain, CodeAlreadyExists).
		WithMessagef("%s with this %s already exists", entity, field)
}

// Validation creates a validation error with a message
func Validation(domain, message string) *DomainError {
	return New(domain, CodeValidation).
		WithMessage(message)
}

// ValidationWithDetails creates a validation error with field details
func ValidationWithDetails(domain string, details map[string]any) *DomainError {
	return New(domain, CodeValidation).
		WithMessage("validation failed").
		WithDetails(details)
}

// Unauthorized creates an unauthorized error
func Unauthorized(domain string) *DomainError {
	return New(domain, CodeUnauthorized).
		WithMessage("unauthorized")
}

// UnauthorizedWithMessage creates an unauthorized error with custom message
func UnauthorizedWithMessage(domain, message string) *DomainError {
	return New(domain, CodeUnauthorized).
		WithMessage(message)
}

// Forbidden creates a forbidden error
func Forbidden(domain string) *DomainError {
	return New(domain, CodeForbidden).
		WithMessage("access denied")
}

// ForbiddenWithMessage creates a forbidden error with custom message
func ForbiddenWithMessage(domain, message string) *DomainError {
	return New(domain, CodeForbidden).
		WithMessage(message)
}

// Internal creates an internal error wrapping another error.
// Stack trace is captured automatically since internal errors are unexpected.
func Internal(domain string, cause error) *DomainError {
	return New(domain, CodeInternal).
		WithStack().
		WithMessage("internal error").
		WithCause(cause)
}

// InternalWithMessage creates an internal error with custom message.
// Stack trace is captured automatically since internal errors are unexpected.
func InternalWithMessage(domain, message string, cause error) *DomainError {
	return New(domain, CodeInternal).
		WithStack().
		WithMessage(message).
		WithCause(cause)
}

// BadRequest creates a bad request error
func BadRequest(domain, message string) *DomainError {
	return New(domain, CodeBadRequest).
		WithMessage(message)
}

// Conflict creates a conflict error
func Conflict(domain, message string) *DomainError {
	return New(domain, CodeConflict).
		WithMessage(message)
}

// IsDomainError checks if an error is a DomainError
func IsDomainError(err error) bool {
	var de *DomainError
	return errors.As(err, &de)
}

// GetDomainError extracts DomainError from an error chain
func GetDomainError(err error) *DomainError {
	var de *DomainError
	if errors.As(err, &de) {
		return de
	}
	return nil
}

// IsCode checks if an error has a specific error code
func IsCode(err error, code ErrorCode) bool {
	if de := GetDomainError(err); de != nil {
		return de.Code == code
	}
	return false
}

// IsNotFound checks if an error is a not found error
func IsNotFound(err error) bool {
	return IsCode(err, CodeNotFound)
}

// IsAlreadyExists checks if an error is an already exists error
func IsAlreadyExists(err error) bool {
	return IsCode(err, CodeAlreadyExists)
}

// IsValidation checks if an error is a validation error
func IsValidation(err error) bool {
	return IsCode(err, CodeValidation)
}

// IsUnauthorized checks if an error is an unauthorized error
func IsUnauthorized(err error) bool {
	return IsCode(err, CodeUnauthorized)
}

// IsForbidden checks if an error is a forbidden error
func IsForbidden(err error) bool {
	return IsCode(err, CodeForbidden)
}

// IsInternal checks if an error is an internal error
func IsInternal(err error) bool {
	return IsCode(err, CodeInternal)
}

// IsBadRequest checks if an error is a bad request error
func IsBadRequest(err error) bool {
	return IsCode(err, CodeBadRequest)
}

// IsConflict checks if an error is a conflict error
func IsConflict(err error) bool {
	return IsCode(err, CodeConflict)
}

// Wrap wraps an error with domain context
func Wrap(domain string, err error, message string) *DomainError {
	if err == nil {
		return nil
	}
	// If it's already a domain error, add context
	if de := GetDomainError(err); de != nil {
		return de.Clone().WithMessage(message)
	}
	// Otherwise create new internal error
	return Internal(domain, err).WithMessage(message)
}
