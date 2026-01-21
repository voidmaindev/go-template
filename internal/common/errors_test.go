package common

import (
	"errors"
	"testing"

	commonerrors "github.com/voidmaindev/go-template/internal/common/errors"
)

func TestAppError_Error(t *testing.T) {
	baseErr := errors.New("base error")
	tests := []struct {
		name     string
		appErr   *AppError
		expected string
	}{
		{
			name: "with message",
			appErr: &AppError{
				Err:     baseErr,
				Message: "user not found",
			},
			expected: "user not found",
		},
		{
			name: "without message, with err",
			appErr: &AppError{
				Err: baseErr,
			},
			expected: "base error",
		},
		{
			name:     "without message and err",
			appErr:   &AppError{},
			expected: "unknown error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.appErr.Error(); got != tt.expected {
				t.Errorf("Error() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestAppError_Unwrap(t *testing.T) {
	originalErr := errors.New("original error")
	appErr := &AppError{
		Err:     originalErr,
		Message: "custom message",
	}

	unwrapped := appErr.Unwrap()
	if unwrapped != originalErr {
		t.Errorf("Unwrap() = %v, want %v", unwrapped, originalErr)
	}
}

func TestNewAppError(t *testing.T) {
	baseErr := errors.New("test error")
	err := NewAppError(baseErr, "user not found")

	if err.Err != baseErr {
		t.Errorf("Err = %v, want %v", err.Err, baseErr)
	}
	if err.Message != "user not found" {
		t.Errorf("Message = %q, want %q", err.Message, "user not found")
	}
}

func TestNewAppErrorWithCode(t *testing.T) {
	baseErr := errors.New("test error")
	err := NewAppErrorWithCode(baseErr, "user not found", "USER_NOT_FOUND")

	if err.Err != baseErr {
		t.Errorf("Err = %v, want %v", err.Err, baseErr)
	}
	if err.Message != "user not found" {
		t.Errorf("Message = %q, want %q", err.Message, "user not found")
	}
	if err.Code != "USER_NOT_FOUND" {
		t.Errorf("Code = %q, want %q", err.Code, "USER_NOT_FOUND")
	}
}

func TestNewAppErrorWithDetails(t *testing.T) {
	details := map[string]string{"field": "email"}
	baseErr := errors.New("validation error")
	err := NewAppErrorWithDetails(baseErr, "validation failed", details)

	if err.Err != baseErr {
		t.Errorf("Err = %v, want %v", err.Err, baseErr)
	}
	if err.Details == nil {
		t.Error("Details should not be nil")
	}
}

func TestWrapError(t *testing.T) {
	baseErr := errors.New("base error")
	tests := []struct {
		name     string
		err      error
		message  string
		wantNil  bool
		contains string
	}{
		{
			name:     "wrap non-nil error",
			err:      baseErr,
			message:  "failed to get user",
			wantNil:  false,
			contains: "failed to get user",
		},
		{
			name:    "wrap nil error",
			err:     nil,
			message: "some message",
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wrapped := WrapError(tt.err, tt.message)

			if tt.wantNil {
				if wrapped != nil {
					t.Error("WrapError(nil) should return nil")
				}
			} else {
				if wrapped == nil {
					t.Error("WrapError() should not return nil for non-nil error")
				}
				if !errors.Is(wrapped, tt.err) {
					t.Error("Wrapped error should match original with errors.Is")
				}
			}
		})
	}
}

func TestIsNotFoundError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{"DomainError NotFound", commonerrors.NotFound("test", "resource"), true},
		{"wrapped DomainError NotFound", WrapError(commonerrors.NotFound("test", "resource"), "context"), true},
		{"DomainError Unauthorized", commonerrors.Unauthorized("test"), false},
		{"nil error", nil, false},
		{"random error", errors.New("random"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsNotFoundError(tt.err); got != tt.expected {
				t.Errorf("IsNotFoundError() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestIsUnauthorizedError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{"DomainError Unauthorized", commonerrors.Unauthorized("test"), true},
		{"wrapped DomainError Unauthorized", WrapError(commonerrors.Unauthorized("test"), "context"), true},
		{"DomainError NotFound", commonerrors.NotFound("test", "resource"), false},
		{"nil error", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsUnauthorizedError(tt.err); got != tt.expected {
				t.Errorf("IsUnauthorizedError() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestIsForbiddenError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{"DomainError Forbidden", commonerrors.Forbidden("test"), true},
		{"wrapped DomainError Forbidden", WrapError(commonerrors.Forbidden("test"), "context"), true},
		{"DomainError Unauthorized", commonerrors.Unauthorized("test"), false},
		{"nil error", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsForbiddenError(tt.err); got != tt.expected {
				t.Errorf("IsForbiddenError() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestIsValidationError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{"DomainError Validation", commonerrors.Validation("test", "message"), true},
		{"wrapped DomainError Validation", WrapError(commonerrors.Validation("test", "invalid email"), "context"), true},
		{"DomainError NotFound", commonerrors.NotFound("test", "resource"), false},
		{"nil error", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsValidationError(tt.err); got != tt.expected {
				t.Errorf("IsValidationError() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestIsConflictError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{"DomainError Conflict", commonerrors.Conflict("test", "resource"), true},
		{"DomainError AlreadyExists", commonerrors.AlreadyExists("test", "resource", "email"), true},
		{"wrapped DomainError Conflict", WrapError(commonerrors.Conflict("test", "duplicate"), "context"), true},
		{"wrapped DomainError AlreadyExists", WrapError(commonerrors.AlreadyExists("test", "user", "email"), "context"), true},
		{"DomainError NotFound", commonerrors.NotFound("test", "resource"), false},
		{"nil error", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsConflictError(tt.err); got != tt.expected {
				t.Errorf("IsConflictError() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestErrorsIs_Integration(t *testing.T) {
	// Test that standard errors.Is works with DomainError
	domainErr := commonerrors.NotFound("test", "user 123")

	// DomainErrors should be identifiable by type checking
	if !commonerrors.IsNotFound(domainErr) {
		t.Error("IsNotFound should identify domain error")
	}

	// Test wrapping
	wrapped := WrapError(domainErr, "context")
	if !commonerrors.IsNotFound(wrapped) {
		t.Error("IsNotFound should work with wrapped domain error")
	}
}
