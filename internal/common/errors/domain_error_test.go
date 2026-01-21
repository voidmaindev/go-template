package errors

import (
	"errors"
	"fmt"
	"testing"
)

func TestNew(t *testing.T) {
	err := New("user", CodeNotFound)

	if err.Domain != "user" {
		t.Errorf("expected domain 'user', got '%s'", err.Domain)
	}
	if err.Code != CodeNotFound {
		t.Errorf("expected code CodeNotFound, got '%s'", err.Code)
	}
	if len(err.stack) == 0 {
		t.Error("expected stack trace to be captured")
	}
}

func TestDomainError_FluentAPI(t *testing.T) {
	cause := errors.New("underlying error")
	err := New("user", CodeNotFound).
		WithMessage("user not found").
		WithOperation("GetByID").
		WithCause(cause).
		WithDetail("user_id", 123).
		WithContext("req-123", "trace-456")

	if err.Message != "user not found" {
		t.Errorf("expected message 'user not found', got '%s'", err.Message)
	}
	if err.Operation != "GetByID" {
		t.Errorf("expected operation 'GetByID', got '%s'", err.Operation)
	}
	if err.Cause != cause {
		t.Error("expected cause to be set")
	}
	if err.Details["user_id"] != 123 {
		t.Errorf("expected detail user_id=123, got %v", err.Details["user_id"])
	}
	if err.RequestID != "req-123" {
		t.Errorf("expected request_id 'req-123', got '%s'", err.RequestID)
	}
	if err.TraceID != "trace-456" {
		t.Errorf("expected trace_id 'trace-456', got '%s'", err.TraceID)
	}
}

func TestDomainError_WithMessagef(t *testing.T) {
	err := New("item", CodeNotFound).WithMessagef("item %d not found", 42)

	if err.Message != "item 42 not found" {
		t.Errorf("expected 'item 42 not found', got '%s'", err.Message)
	}
}

func TestDomainError_WithDetails(t *testing.T) {
	details := map[string]any{"field": "email", "reason": "invalid format"}
	err := New("user", CodeValidation).WithDetails(details)

	if err.Details["field"] != "email" {
		t.Errorf("expected field 'email', got %v", err.Details["field"])
	}
	if err.Details["reason"] != "invalid format" {
		t.Errorf("expected reason 'invalid format', got %v", err.Details["reason"])
	}
}

func TestDomainError_WithDetail_InitializesMap(t *testing.T) {
	err := New("user", CodeNotFound)
	if err.Details != nil {
		t.Error("expected Details to be nil initially")
	}

	err.WithDetail("key", "value")
	if err.Details == nil {
		t.Error("expected Details to be initialized")
	}
	if err.Details["key"] != "value" {
		t.Errorf("expected 'value', got %v", err.Details["key"])
	}
}

func TestDomainError_Error(t *testing.T) {
	t.Run("with message", func(t *testing.T) {
		err := New("user", CodeNotFound).WithMessage("custom message")
		if err.Error() != "custom message" {
			t.Errorf("expected 'custom message', got '%s'", err.Error())
		}
	})

	t.Run("without message", func(t *testing.T) {
		err := New("user", CodeNotFound)
		if err.Error() != "NOT_FOUND" {
			t.Errorf("expected 'NOT_FOUND', got '%s'", err.Error())
		}
	})
}

func TestDomainError_Unwrap(t *testing.T) {
	cause := errors.New("underlying error")
	err := New("user", CodeInternal).WithCause(cause)

	unwrapped := err.Unwrap()
	if unwrapped != cause {
		t.Error("Unwrap should return the cause")
	}
}

func TestDomainError_HTTPStatus(t *testing.T) {
	tests := []struct {
		code     ErrorCode
		expected int
	}{
		{CodeNotFound, 404},
		{CodeAlreadyExists, 409},
		{CodeConflict, 409},
		{CodeValidation, 400},
		{CodeBadRequest, 400},
		{CodeUnauthorized, 401},
		{CodeForbidden, 403},
		{CodeInternal, 500},
	}

	for _, tt := range tests {
		t.Run(string(tt.code), func(t *testing.T) {
			err := New("test", tt.code)
			if err.HTTPStatus() != tt.expected {
				t.Errorf("expected HTTP status %d for %s, got %d", tt.expected, tt.code, err.HTTPStatus())
			}
		})
	}
}

func TestDomainError_Is(t *testing.T) {
	t.Run("same domain and code", func(t *testing.T) {
		err1 := New("user", CodeNotFound)
		err2 := New("user", CodeNotFound)

		if !errors.Is(err1, err2) {
			t.Error("errors with same domain and code should match")
		}
	})

	t.Run("different domain", func(t *testing.T) {
		err1 := New("user", CodeNotFound)
		err2 := New("item", CodeNotFound)

		if errors.Is(err1, err2) {
			t.Error("errors with different domain should not match")
		}
	})

	t.Run("different code", func(t *testing.T) {
		err1 := New("user", CodeNotFound)
		err2 := New("user", CodeInternal)

		if errors.Is(err1, err2) {
			t.Error("errors with different code should not match")
		}
	})

	t.Run("non-DomainError target", func(t *testing.T) {
		err1 := New("user", CodeNotFound)
		err2 := errors.New("regular error")

		if errors.Is(err1, err2) {
			t.Error("DomainError should not match non-DomainError")
		}
	})
}

func TestDomainError_Clone(t *testing.T) {
	original := New("user", CodeNotFound).
		WithMessage("original").
		WithOperation("GetByID").
		WithDetail("id", 123).
		WithContext("req-1", "trace-1")

	clone := original.Clone()

	// Verify all fields are copied
	if clone.Code != original.Code {
		t.Error("clone Code mismatch")
	}
	if clone.Domain != original.Domain {
		t.Error("clone Domain mismatch")
	}
	if clone.Message != original.Message {
		t.Error("clone Message mismatch")
	}
	if clone.Operation != original.Operation {
		t.Error("clone Operation mismatch")
	}
	if clone.RequestID != original.RequestID {
		t.Error("clone RequestID mismatch")
	}
	if clone.TraceID != original.TraceID {
		t.Error("clone TraceID mismatch")
	}

	// Verify details are deep copied
	if clone.Details["id"] != original.Details["id"] {
		t.Error("clone Details mismatch")
	}

	// Modify clone details and verify original unchanged
	clone.Details["id"] = 999
	if original.Details["id"] == 999 {
		t.Error("original Details should not be affected by clone modification")
	}

	// Verify new stack trace
	if &clone.stack[0] == &original.stack[0] {
		t.Error("clone should have its own stack trace")
	}
}

func TestDomainError_StackTrace(t *testing.T) {
	err := New("user", CodeNotFound)
	trace := err.StackTrace()

	if trace == "" {
		t.Error("expected non-empty stack trace")
	}

	// Should contain function name
	if !containsString(trace, "TestDomainError_StackTrace") {
		t.Error("stack trace should contain the test function name")
	}
}

func TestDomainError_Stack_Empty(t *testing.T) {
	err := &DomainError{Code: CodeNotFound}
	trace := err.StackTrace()

	if trace != "" {
		t.Error("expected empty stack trace for error without stack")
	}
}

// Helper tests
func TestHelpers_NotFound(t *testing.T) {
	err := NotFound("user", "user")

	if err.Code != CodeNotFound {
		t.Errorf("expected CodeNotFound, got %s", err.Code)
	}
	if err.Domain != "user" {
		t.Errorf("expected domain 'user', got '%s'", err.Domain)
	}
	if err.Message != "user not found" {
		t.Errorf("expected 'user not found', got '%s'", err.Message)
	}
}

func TestHelpers_AlreadyExists(t *testing.T) {
	err := AlreadyExists("user", "user", "email")

	if err.Code != CodeAlreadyExists {
		t.Errorf("expected CodeAlreadyExists, got %s", err.Code)
	}
	if err.Message != "user with this email already exists" {
		t.Errorf("expected 'user with this email already exists', got '%s'", err.Message)
	}
}

func TestHelpers_Validation(t *testing.T) {
	err := Validation("user", "invalid email format")

	if err.Code != CodeValidation {
		t.Errorf("expected CodeValidation, got %s", err.Code)
	}
	if err.Message != "invalid email format" {
		t.Errorf("expected 'invalid email format', got '%s'", err.Message)
	}
}

func TestHelpers_ValidationWithDetails(t *testing.T) {
	details := map[string]any{"email": "invalid format"}
	err := ValidationWithDetails("user", details)

	if err.Code != CodeValidation {
		t.Errorf("expected CodeValidation, got %s", err.Code)
	}
	if err.Details["email"] != "invalid format" {
		t.Errorf("expected detail 'invalid format', got %v", err.Details["email"])
	}
}

func TestHelpers_Unauthorized(t *testing.T) {
	err := Unauthorized("auth")

	if err.Code != CodeUnauthorized {
		t.Errorf("expected CodeUnauthorized, got %s", err.Code)
	}
}

func TestHelpers_UnauthorizedWithMessage(t *testing.T) {
	err := UnauthorizedWithMessage("auth", "token expired")

	if err.Code != CodeUnauthorized {
		t.Errorf("expected CodeUnauthorized, got %s", err.Code)
	}
	if err.Message != "token expired" {
		t.Errorf("expected 'token expired', got '%s'", err.Message)
	}
}

func TestHelpers_Forbidden(t *testing.T) {
	err := Forbidden("rbac")

	if err.Code != CodeForbidden {
		t.Errorf("expected CodeForbidden, got %s", err.Code)
	}
}

func TestHelpers_ForbiddenWithMessage(t *testing.T) {
	err := ForbiddenWithMessage("rbac", "admin only")

	if err.Code != CodeForbidden {
		t.Errorf("expected CodeForbidden, got %s", err.Code)
	}
	if err.Message != "admin only" {
		t.Errorf("expected 'admin only', got '%s'", err.Message)
	}
}

func TestHelpers_Internal(t *testing.T) {
	cause := errors.New("db connection failed")
	err := Internal("user", cause)

	if err.Code != CodeInternal {
		t.Errorf("expected CodeInternal, got %s", err.Code)
	}
	if err.Cause != cause {
		t.Error("expected cause to be set")
	}
}

func TestHelpers_InternalWithMessage(t *testing.T) {
	cause := errors.New("db error")
	err := InternalWithMessage("user", "failed to create user", cause)

	if err.Code != CodeInternal {
		t.Errorf("expected CodeInternal, got %s", err.Code)
	}
	if err.Message != "failed to create user" {
		t.Errorf("expected 'failed to create user', got '%s'", err.Message)
	}
}

func TestHelpers_BadRequest(t *testing.T) {
	err := BadRequest("api", "missing required field")

	if err.Code != CodeBadRequest {
		t.Errorf("expected CodeBadRequest, got %s", err.Code)
	}
}

func TestHelpers_Conflict(t *testing.T) {
	err := Conflict("user", "user already has this role")

	if err.Code != CodeConflict {
		t.Errorf("expected CodeConflict, got %s", err.Code)
	}
}

func TestHelpers_IsDomainError(t *testing.T) {
	t.Run("is DomainError", func(t *testing.T) {
		err := New("user", CodeNotFound)
		if !IsDomainError(err) {
			t.Error("expected IsDomainError to return true")
		}
	})

	t.Run("wrapped DomainError", func(t *testing.T) {
		err := fmt.Errorf("wrapped: %w", New("user", CodeNotFound))
		if !IsDomainError(err) {
			t.Error("expected IsDomainError to return true for wrapped error")
		}
	})

	t.Run("not DomainError", func(t *testing.T) {
		err := errors.New("regular error")
		if IsDomainError(err) {
			t.Error("expected IsDomainError to return false")
		}
	})
}

func TestHelpers_GetDomainError(t *testing.T) {
	t.Run("direct DomainError", func(t *testing.T) {
		original := New("user", CodeNotFound)
		result := GetDomainError(original)

		if result != original {
			t.Error("expected to get the original DomainError")
		}
	})

	t.Run("wrapped DomainError", func(t *testing.T) {
		original := New("user", CodeNotFound)
		wrapped := fmt.Errorf("context: %w", original)
		result := GetDomainError(wrapped)

		if result != original {
			t.Error("expected to extract the wrapped DomainError")
		}
	})

	t.Run("not DomainError", func(t *testing.T) {
		err := errors.New("regular error")
		result := GetDomainError(err)

		if result != nil {
			t.Error("expected nil for non-DomainError")
		}
	})
}

func TestHelpers_IsCode(t *testing.T) {
	err := New("user", CodeNotFound)

	if !IsCode(err, CodeNotFound) {
		t.Error("expected IsCode to return true for matching code")
	}
	if IsCode(err, CodeInternal) {
		t.Error("expected IsCode to return false for non-matching code")
	}
}

func TestHelpers_IsNotFound(t *testing.T) {
	if !IsNotFound(New("user", CodeNotFound)) {
		t.Error("expected IsNotFound to return true")
	}
	if IsNotFound(New("user", CodeInternal)) {
		t.Error("expected IsNotFound to return false")
	}
}

func TestHelpers_IsAlreadyExists(t *testing.T) {
	if !IsAlreadyExists(New("user", CodeAlreadyExists)) {
		t.Error("expected IsAlreadyExists to return true")
	}
}

func TestHelpers_IsValidation(t *testing.T) {
	if !IsValidation(New("user", CodeValidation)) {
		t.Error("expected IsValidation to return true")
	}
}

func TestHelpers_IsUnauthorized(t *testing.T) {
	if !IsUnauthorized(New("auth", CodeUnauthorized)) {
		t.Error("expected IsUnauthorized to return true")
	}
}

func TestHelpers_IsForbidden(t *testing.T) {
	if !IsForbidden(New("rbac", CodeForbidden)) {
		t.Error("expected IsForbidden to return true")
	}
}

func TestHelpers_IsInternal(t *testing.T) {
	if !IsInternal(New("db", CodeInternal)) {
		t.Error("expected IsInternal to return true")
	}
}

func TestHelpers_IsBadRequest(t *testing.T) {
	if !IsBadRequest(New("api", CodeBadRequest)) {
		t.Error("expected IsBadRequest to return true")
	}
}

func TestHelpers_IsConflict(t *testing.T) {
	if !IsConflict(New("user", CodeConflict)) {
		t.Error("expected IsConflict to return true")
	}
}

func TestHelpers_Wrap(t *testing.T) {
	t.Run("wrap nil error", func(t *testing.T) {
		result := Wrap("user", nil, "some message")
		if result != nil {
			t.Error("expected nil when wrapping nil error")
		}
	})

	t.Run("wrap DomainError", func(t *testing.T) {
		original := New("user", CodeNotFound).WithMessage("original message")
		result := Wrap("user", original, "new message")

		if result.Message != "new message" {
			t.Errorf("expected message 'new message', got '%s'", result.Message)
		}
		if result.Code != CodeNotFound {
			t.Error("code should be preserved")
		}
	})

	t.Run("wrap regular error", func(t *testing.T) {
		err := errors.New("regular error")
		result := Wrap("user", err, "wrapped message")

		if result.Code != CodeInternal {
			t.Errorf("expected CodeInternal, got %s", result.Code)
		}
		if result.Message != "wrapped message" {
			t.Errorf("expected 'wrapped message', got '%s'", result.Message)
		}
		if result.Cause != err {
			t.Error("cause should be set")
		}
	})
}

func TestErrorCode_String(t *testing.T) {
	if CodeNotFound.String() != "NOT_FOUND" {
		t.Errorf("expected 'NOT_FOUND', got '%s'", CodeNotFound.String())
	}
}

// Helper function for string contains check
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsStringHelper(s, substr))
}

func containsStringHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
