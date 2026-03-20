package common

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	domainerrors "github.com/voidmaindev/go-template/internal/common/errors"
)

// helper: perform a Fiber request against a handler and return status + decoded body
func doRequest(t *testing.T, handler fiber.Handler) (int, Response) {
	t.Helper()
	app := fiber.New()
	app.Get("/test", handler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test failed: %v", err)
	}
	defer resp.Body.Close()

	var body Response
	if resp.StatusCode != http.StatusNoContent {
		b, _ := io.ReadAll(resp.Body)
		if len(b) > 0 {
			if err := json.Unmarshal(b, &body); err != nil {
				t.Fatalf("failed to decode response body: %v", err)
			}
		}
	}
	return resp.StatusCode, body
}

// ================================
// HandleError tests
// ================================

func TestHandleError_NilError(t *testing.T) {
	app := fiber.New()
	app.Get("/test", func(c *fiber.Ctx) error {
		result := HandleError(c, nil)
		if result != nil {
			return result
		}
		// HandleError returned nil, meaning no error was sent — send success
		return c.SendStatus(fiber.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200 for nil error, got %d", resp.StatusCode)
	}
}

func TestHandleError_DomainErrors(t *testing.T) {
	tests := []struct {
		name         string
		err          error
		expectedCode int
		expectedMsg  string
	}{
		{
			name:         "NotFound",
			err:          domainerrors.NotFound("user", "user"),
			expectedCode: http.StatusNotFound,
			expectedMsg:  "user not found",
		},
		{
			name:         "AlreadyExists",
			err:          domainerrors.AlreadyExists("user", "user", "email"),
			expectedCode: http.StatusConflict,
			expectedMsg:  "user with this email already exists",
		},
		{
			name:         "Validation",
			err:          domainerrors.Validation("user", "invalid email format"),
			expectedCode: http.StatusBadRequest,
			expectedMsg:  "invalid email format",
		},
		{
			name:         "Unauthorized",
			err:          domainerrors.Unauthorized("auth"),
			expectedCode: http.StatusUnauthorized,
			expectedMsg:  "unauthorized",
		},
		{
			name:         "Forbidden",
			err:          domainerrors.Forbidden("rbac"),
			expectedCode: http.StatusForbidden,
			expectedMsg:  "access denied",
		},
		{
			name:         "Conflict",
			err:          domainerrors.Conflict("user", "duplicate entry"),
			expectedCode: http.StatusConflict,
			expectedMsg:  "duplicate entry",
		},
		{
			name:         "Internal",
			err:          domainerrors.Internal("db", errors.New("connection failed")),
			expectedCode: http.StatusInternalServerError,
			expectedMsg:  "internal error",
		},
		{
			name:         "BadRequest",
			err:          domainerrors.BadRequest("api", "malformed JSON"),
			expectedCode: http.StatusBadRequest,
			expectedMsg:  "malformed JSON",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, body := doRequest(t, func(c *fiber.Ctx) error {
				return HandleError(c, tt.err)
			})

			if status != tt.expectedCode {
				t.Errorf("status = %d, want %d", status, tt.expectedCode)
			}
			if body.Success {
				t.Error("expected Success=false")
			}

			if body.Error == nil {
				t.Fatal("expected error to be non-nil")
			}
			if body.Error.Message != tt.expectedMsg {
				t.Errorf("message = %q, want %q", body.Error.Message, tt.expectedMsg)
			}
		})
	}
}

func TestHandleError_WrappedDomainError(t *testing.T) {
	// DomainError wrapped with fmt.Errorf should still be detected
	inner := domainerrors.NotFound("user", "item")
	wrapped := fmt.Errorf("service layer: %w", inner)

	status, body := doRequest(t, func(c *fiber.Ctx) error {
		return HandleError(c, wrapped)
	})

	if status != http.StatusNotFound {
		t.Errorf("status = %d, want 404 for wrapped DomainError", status)
	}
	if body.Success {
		t.Error("expected Success=false")
	}
}

func TestHandleError_PlainError_Returns500(t *testing.T) {
	// A plain error (not DomainError) should return 500
	plainErr := errors.New("something unexpected")

	status, body := doRequest(t, func(c *fiber.Ctx) error {
		return HandleError(c, plainErr)
	})

	if status != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500 for plain error", status)
	}
	if body.Success {
		t.Error("expected Success=false")
	}
	if body.Error == nil {
		t.Fatal("expected error to be non-nil")
	}
	if body.Error.Message != "internal server error" {
		t.Errorf("error = %q, want %q", body.Error.Message, "internal server error")
	}
}

func TestHandleError_WrappedPlainError_Returns500(t *testing.T) {
	// A wrapped plain error (no DomainError in chain) should return 500
	wrapped := fmt.Errorf("layer 1: %w", fmt.Errorf("layer 2: %w", errors.New("root cause")))

	status, _ := doRequest(t, func(c *fiber.Ctx) error {
		return HandleError(c, wrapped)
	})

	if status != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500 for wrapped plain error", status)
	}
}

// ================================
// HandleDomainError tests
// ================================

func TestHandleDomainError_IncludesDomainAndCode(t *testing.T) {
	de := domainerrors.NotFound("user", "account").WithOperation("GetByID")

	status, body := doRequest(t, func(c *fiber.Ctx) error {
		return HandleDomainError(c, de)
	})

	if status != http.StatusNotFound {
		t.Errorf("status = %d, want 404", status)
	}

	if body.Error == nil {
		t.Fatal("expected error to be non-nil")
	}
	if body.Error.Code != "NOT_FOUND" {
		t.Errorf("code = %v, want NOT_FOUND", body.Error.Code)
	}
	if body.Error.Domain != "user" {
		t.Errorf("domain = %v, want user", body.Error.Domain)
	}
}

func TestHandleDomainError_IncludesDetails(t *testing.T) {
	details := map[string]any{
		"field": "email",
		"rule":  "required",
	}
	de := domainerrors.Validation("user", "validation failed").WithDetails(details)

	_, body := doRequest(t, func(c *fiber.Ctx) error {
		return HandleDomainError(c, de)
	})

	if body.Error == nil {
		t.Fatal("expected error to be non-nil")
	}
	if body.Error.Details == nil {
		t.Fatal("expected details to be non-nil")
	}
	if body.Error.Details["field"] != "email" {
		t.Errorf("details.field = %v, want email", body.Error.Details["field"])
	}
}

func TestHandleDomainError_OmitsEmptyDomainAndDetails(t *testing.T) {
	// Error with no domain and no details — those fields should be omitted
	de := domainerrors.New("", domainerrors.CodeInternal).WithMessage("oops")

	_, body := doRequest(t, func(c *fiber.Ctx) error {
		return HandleDomainError(c, de)
	})

	if body.Error == nil {
		t.Fatal("expected error to be non-nil")
	}
	if body.Error.Domain != "" {
		t.Error("domain should be empty when not set")
	}
	if body.Error.Details != nil {
		t.Error("details should be nil when empty")
	}
}

func TestHandleDomainError_RequestIDPropagated(t *testing.T) {
	de := domainerrors.NotFound("user", "item")

	app := fiber.New()
	app.Get("/test", func(c *fiber.Ctx) error {
		c.Locals("request_id", "req-abc-123")
		return HandleDomainError(c, de)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	resp, _ := app.Test(req, -1)
	defer resp.Body.Close()

	b, _ := io.ReadAll(resp.Body)
	var body Response
	json.Unmarshal(b, &body)

	if body.RequestID != "req-abc-123" {
		t.Errorf("RequestID = %q, want %q", body.RequestID, "req-abc-123")
	}
}

// ================================
// HandleErrorWithDomain tests
// ================================

func TestHandleErrorWithDomain_NilError(t *testing.T) {
	app := fiber.New()
	app.Get("/test", func(c *fiber.Ctx) error {
		result := HandleErrorWithDomain(c, "user", nil)
		if result != nil {
			return result
		}
		return c.SendStatus(fiber.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	resp, _ := app.Test(req, -1)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200 for nil error, got %d", resp.StatusCode)
	}
}

func TestHandleErrorWithDomain_DomainError(t *testing.T) {
	// Already a DomainError — should be handled directly
	de := domainerrors.NotFound("user", "account")

	status, body := doRequest(t, func(c *fiber.Ctx) error {
		return HandleErrorWithDomain(c, "user", de)
	})

	if status != http.StatusNotFound {
		t.Errorf("status = %d, want 404", status)
	}

	if body.Error == nil {
		t.Fatal("expected error to be non-nil")
	}
	if body.Error.Code != "NOT_FOUND" {
		t.Errorf("code = %v, want NOT_FOUND", body.Error.Code)
	}
}

func TestHandleErrorWithDomain_PlainError_WrappedAsInternal(t *testing.T) {
	// Plain error should be wrapped as internal with the specified domain
	plainErr := errors.New("db connection timeout")

	status, body := doRequest(t, func(c *fiber.Ctx) error {
		return HandleErrorWithDomain(c, "user", plainErr)
	})

	if status != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", status)
	}

	if body.Error == nil {
		t.Fatal("expected error to be non-nil")
	}
	if body.Error.Code != "INTERNAL_ERROR" {
		t.Errorf("code = %v, want INTERNAL_ERROR", body.Error.Code)
	}
	if body.Error.Domain != "user" {
		t.Errorf("domain = %v, want user", body.Error.Domain)
	}
}

// ================================
// ValidationErrorResponse tests
// ================================

func TestValidationErrorResponse_DetailsInErrorField(t *testing.T) {
	app := fiber.New()
	app.Get("/test", func(c *fiber.Ctx) error {
		validationErrors := []map[string]any{
			{"field": "name", "message": "name is required"},
			{"field": "email", "message": "email must be valid"},
		}
		return ValidationErrorResponse(c, validationErrors)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}

	b, _ := io.ReadAll(resp.Body)
	var body Response
	if err := json.Unmarshal(b, &body); err != nil {
		t.Fatalf("failed to decode body: %v", err)
	}

	if body.Success {
		t.Error("expected Success=false")
	}
	if body.Error == nil {
		t.Fatal("expected error to be non-nil")
	}
	if body.Error.Code != "VALIDATION_ERROR" {
		t.Errorf("code = %q, want VALIDATION_ERROR", body.Error.Code)
	}
	// Validation errors should be in error.details.fields, not in data
	if body.Error.Details == nil {
		t.Fatal("expected error.details to be non-nil")
	}
	fields, ok := body.Error.Details["fields"]
	if !ok || fields == nil {
		t.Fatal("expected error.details.fields to contain validation errors")
	}
	// Verify data field is NOT populated with validation errors
	if body.Data != nil {
		t.Error("validation errors should not be in data field")
	}
}

// ================================
// ErrorResponseWithDetails tests
// ================================

func TestErrorResponseWithDetails_DetailsInErrorField(t *testing.T) {
	app := fiber.New()
	app.Get("/test", func(c *fiber.Ctx) error {
		details := map[string]any{
			"field":   "email",
			"allowed": []string{"user@example.com"},
		}
		return ErrorResponseWithDetails(c, fiber.StatusBadRequest, "invalid input", details)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}

	b, _ := io.ReadAll(resp.Body)
	var body Response
	if err := json.Unmarshal(b, &body); err != nil {
		t.Fatalf("failed to decode body: %v", err)
	}

	if body.Error == nil {
		t.Fatal("expected error to be non-nil")
	}
	if body.Error.Message != "invalid input" {
		t.Errorf("message = %q, want %q", body.Error.Message, "invalid input")
	}
	// Details should be in error.details, not in data
	if body.Error.Details == nil {
		t.Fatal("expected error.details to be non-nil")
	}
	if body.Error.Details["field"] != "email" {
		t.Errorf("details.field = %v, want email", body.Error.Details["field"])
	}
	if body.Data != nil {
		t.Error("details should not be in data field")
	}
}
