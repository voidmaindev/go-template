package docs

import (
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
)

func TestScalarHandler(t *testing.T) {
	app := fiber.New()
	app.Get("/docs", ScalarHandler("/openapi.json"))

	tests := []struct {
		name          string
		specURL       string
		wantStatus    int
		wantContains  []string
	}{
		{
			name:       "returns HTML with spec URL",
			specURL:    "/openapi.json",
			wantStatus: 200,
			wantContains: []string{
				"<!DOCTYPE html>",
				"<title>Go Template API - Documentation</title>",
				`data-url="/openapi.json"`,
				"@scalar/api-reference",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a new app for each test to set the correct spec URL
			testApp := fiber.New()
			testApp.Get("/docs", ScalarHandler(tt.specURL))

			req := httptest.NewRequest("GET", "/docs", nil)
			resp, err := testApp.Test(req)
			if err != nil {
				t.Fatalf("Test request failed: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != tt.wantStatus {
				t.Errorf("Status = %d, want %d", resp.StatusCode, tt.wantStatus)
			}

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("Failed to read body: %v", err)
			}
			bodyStr := string(body)

			for _, want := range tt.wantContains {
				if !strings.Contains(bodyStr, want) {
					t.Errorf("Body does not contain %q", want)
				}
			}

			// Check Content-Type header
			contentType := resp.Header.Get("Content-Type")
			if !strings.Contains(contentType, "text/html") {
				t.Errorf("Content-Type = %s, want text/html", contentType)
			}
		})
	}
}

func TestScalarHandlerDifferentSpecURLs(t *testing.T) {
	specURLs := []string{
		"/openapi.json",
		"/api/spec.json",
		"/v1/openapi.yaml",
		"https://example.com/api.json",
	}

	for _, specURL := range specURLs {
		t.Run(specURL, func(t *testing.T) {
			app := fiber.New()
			app.Get("/docs", ScalarHandler(specURL))

			req := httptest.NewRequest("GET", "/docs", nil)
			resp, err := app.Test(req)
			if err != nil {
				t.Fatalf("Test request failed: %v", err)
			}
			defer resp.Body.Close()

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("Failed to read body: %v", err)
			}

			expected := `data-url="` + specURL + `"`
			if !strings.Contains(string(body), expected) {
				t.Errorf("Body does not contain expected spec URL reference: %s", expected)
			}
		})
	}
}
