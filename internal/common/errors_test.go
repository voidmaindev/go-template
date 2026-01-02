package common

import (
	"errors"
	"testing"
)

func TestAppError_Error(t *testing.T) {
	tests := []struct {
		name     string
		appErr   *AppError
		expected string
	}{
		{
			name: "with message",
			appErr: &AppError{
				Err:     ErrNotFound,
				Message: "user not found",
			},
			expected: "user not found",
		},
		{
			name: "without message, with err",
			appErr: &AppError{
				Err: ErrNotFound,
			},
			expected: "resource not found",
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
	originalErr := ErrNotFound
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
	err := NewAppError(ErrNotFound, "user not found")

	if err.Err != ErrNotFound {
		t.Errorf("Err = %v, want %v", err.Err, ErrNotFound)
	}
	if err.Message != "user not found" {
		t.Errorf("Message = %q, want %q", err.Message, "user not found")
	}
}

func TestNewAppErrorWithCode(t *testing.T) {
	err := NewAppErrorWithCode(ErrNotFound, "user not found", "USER_NOT_FOUND")

	if err.Err != ErrNotFound {
		t.Errorf("Err = %v, want %v", err.Err, ErrNotFound)
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
	err := NewAppErrorWithDetails(ErrValidation, "validation failed", details)

	if err.Err != ErrValidation {
		t.Errorf("Err = %v, want %v", err.Err, ErrValidation)
	}
	if err.Details == nil {
		t.Error("Details should not be nil")
	}
}

func TestWrapError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		message  string
		wantNil  bool
		contains string
	}{
		{
			name:     "wrap non-nil error",
			err:      ErrNotFound,
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
		{"ErrNotFound", ErrNotFound, true},
		{"wrapped ErrNotFound", WrapError(ErrNotFound, "user"), true},
		{"ErrUnauthorized", ErrUnauthorized, false},
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
		{"ErrUnauthorized", ErrUnauthorized, true},
		{"wrapped ErrUnauthorized", WrapError(ErrUnauthorized, "no token"), true},
		{"ErrNotFound", ErrNotFound, false},
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
		{"ErrForbidden", ErrForbidden, true},
		{"wrapped ErrForbidden", WrapError(ErrForbidden, "access denied"), true},
		{"ErrUnauthorized", ErrUnauthorized, false},
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
		{"ErrValidation", ErrValidation, true},
		{"wrapped ErrValidation", WrapError(ErrValidation, "invalid email"), true},
		{"ErrBadRequest", ErrBadRequest, false},
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
		{"ErrConflict", ErrConflict, true},
		{"ErrAlreadyExists", ErrAlreadyExists, true},
		{"wrapped ErrConflict", WrapError(ErrConflict, "duplicate"), true},
		{"wrapped ErrAlreadyExists", WrapError(ErrAlreadyExists, "email exists"), true},
		{"ErrNotFound", ErrNotFound, false},
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
	// Test that standard errors.Is works with our error types
	appErr := NewAppError(ErrNotFound, "user 123 not found")

	if !errors.Is(appErr, ErrNotFound) {
		t.Error("errors.Is should work with AppError")
	}
}

func TestCommonErrors_AreDefined(t *testing.T) {
	// Ensure all common errors are properly defined
	commonErrors := []error{
		ErrNotFound,
		ErrAlreadyExists,
		ErrInvalidInput,
		ErrUnauthorized,
		ErrForbidden,
		ErrInternalServer,
		ErrBadRequest,
		ErrConflict,
		ErrValidation,
		ErrInvalidCredentials,
		ErrTokenExpired,
		ErrTokenInvalid,
		ErrTokenBlacklisted,
	}

	for _, err := range commonErrors {
		if err == nil {
			t.Error("Common error should not be nil")
		}
		if err.Error() == "" {
			t.Error("Common error message should not be empty")
		}
	}
}
