package common

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
)

// testValidatable is a test struct with validation tags
type testValidatable struct {
	Name  string `json:"name" validate:"required"`
	Email string `json:"email" validate:"required,email"`
}

// testNoValidation is a test struct without validation tags
type testNoValidation struct {
	Value string `json:"value"`
}

func TestParseAndValidate_ValidJSON(t *testing.T) {
	app := fiber.New()
	app.Post("/test", func(c *fiber.Ctx) error {
		req, err := ParseAndValidate[testValidatable](c)
		if err != nil {
			return nil
		}
		return c.JSON(fiber.Map{"name": req.Name, "email": req.Email})
	})

	body, _ := json.Marshal(map[string]string{
		"name":  "Alice",
		"email": "alice@example.com",
	})
	req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected 200, got %d: %s", resp.StatusCode, string(b))
	}
}

func TestParseAndValidate_InvalidJSON(t *testing.T) {
	app := fiber.New()
	app.Post("/test", func(c *fiber.Ctx) error {
		req, err := ParseAndValidate[testValidatable](c)
		if err != nil {
			return nil // response already sent
		}
		// Should NOT reach here
		return c.JSON(req)
	})

	req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewReader([]byte("not valid json")))
	req.Header.Set("Content-Type", "application/json")

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
	if body.Error == nil || body.Error.Message != "invalid request body" {
		t.Errorf("expected error message 'invalid request body', got %v", body.Error)
	}
}

func TestParseAndValidate_ValidationFailure(t *testing.T) {
	app := fiber.New()
	app.Post("/test", func(c *fiber.Ctx) error {
		req, err := ParseAndValidate[testValidatable](c)
		if err != nil {
			return nil
		}
		return c.JSON(req)
	})

	// Missing required "name", invalid email
	body, _ := json.Marshal(map[string]string{
		"email": "not-an-email",
	})
	req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}

	b, _ := io.ReadAll(resp.Body)
	var body2 Response
	json.Unmarshal(b, &body2)
	if body2.Success {
		t.Error("expected Success=false")
	}
	if body2.Error == nil || body2.Error.Code != "VALIDATION_ERROR" {
		t.Errorf("expected VALIDATION_ERROR code, got %v", body2.Error)
	}
}

func TestParseAndValidate_EmptyBody(t *testing.T) {
	app := fiber.New()
	app.Post("/test", func(c *fiber.Ctx) error {
		_, err := ParseAndValidate[testValidatable](c)
		if err != nil {
			return nil
		}
		// Should not reach here since required fields are missing
		t.Error("ParseAndValidate should have failed on empty body with required fields")
		return nil
	})

	req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewReader([]byte("{}")))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400 for empty body with required fields, got %d", resp.StatusCode)
	}
}

func TestParseAndValidate_NoValidationTags(t *testing.T) {
	app := fiber.New()
	app.Post("/test", func(c *fiber.Ctx) error {
		req, err := ParseAndValidate[testNoValidation](c)
		if err != nil {
			return nil
		}
		return c.JSON(req)
	})

	body, _ := json.Marshal(map[string]string{"value": "hello"})
	req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestParseAndValidate_DoesNotContinueOnFailure(t *testing.T) {
	// Verify the handler does NOT proceed past ParseAndValidate on failure
	continued := false
	app := fiber.New()
	app.Post("/test", func(c *fiber.Ctx) error {
		_, err := ParseAndValidate[testValidatable](c)
		if err != nil {
			return nil
		}
		continued = true
		return c.SendStatus(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewReader([]byte("bad json")))
	req.Header.Set("Content-Type", "application/json")

	resp, _ := app.Test(req, -1)
	defer resp.Body.Close()

	if continued {
		t.Error("handler continued past ParseAndValidate on invalid input")
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}
}

// ================================
// ParseID tests
// ================================

func TestParseID_ValidID(t *testing.T) {
	app := fiber.New()
	app.Get("/items/:id", func(c *fiber.Ctx) error {
		id, err := ParseID(c, "id", "item")
		if err != nil {
			return nil
		}
		return c.JSON(fiber.Map{"id": id})
	})

	req := httptest.NewRequest(http.MethodGet, "/items/42", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestParseID_InvalidID(t *testing.T) {
	app := fiber.New()
	app.Get("/items/:id", func(c *fiber.Ctx) error {
		_, err := ParseID(c, "id", "item")
		if err != nil {
			return nil
		}
		t.Error("should not reach here")
		return nil
	})

	req := httptest.NewRequest(http.MethodGet, "/items/abc", nil)
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
	json.Unmarshal(b, &body)
	if body.Error == nil || body.Error.Message != "invalid item ID" {
		t.Errorf("expected 'invalid item ID', got %v", body.Error)
	}
}

func TestParseID_DoesNotContinueOnFailure(t *testing.T) {
	continued := false
	app := fiber.New()
	app.Get("/items/:id", func(c *fiber.Ctx) error {
		_, err := ParseID(c, "id", "item")
		if err != nil {
			return nil
		}
		continued = true
		return nil
	})

	req := httptest.NewRequest(http.MethodGet, "/items/notanumber", nil)
	resp, _ := app.Test(req, -1)
	defer resp.Body.Close()

	if continued {
		t.Error("handler continued past ParseID on invalid input")
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}
}

func TestParseID_ResourceNameInMessage(t *testing.T) {
	tests := []struct {
		resource    string
		expectedMsg string
	}{
		{"item", "invalid item ID"},
		{"document", "invalid document ID"},
		{"user", "invalid user ID"},
	}

	for _, tt := range tests {
		t.Run(tt.resource, func(t *testing.T) {
			app := fiber.New()
			app.Get("/r/:id", func(c *fiber.Ctx) error {
				_, err := ParseID(c, "id", tt.resource)
				if err != nil {
					return nil
				}
				return nil
			})

			req := httptest.NewRequest(http.MethodGet, "/r/bad", nil)
			resp, _ := app.Test(req, -1)
			defer resp.Body.Close()

			b, _ := io.ReadAll(resp.Body)
			var body Response
			json.Unmarshal(b, &body)

			if body.Error == nil || body.Error.Message != tt.expectedMsg {
				t.Errorf("expected %q, got %v", tt.expectedMsg, body.Error)
			}
		})
	}
}
