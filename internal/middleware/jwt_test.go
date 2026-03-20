package middleware

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/voidmaindev/go-template/internal/config"
	"github.com/voidmaindev/go-template/pkg/utils"
)

// mockBlacklist implements TokenBlacklist interface for testing
type mockBlacklist struct {
	blacklisted map[string]bool
	shouldError bool
}

func newMockBlacklist() *mockBlacklist {
	return &mockBlacklist{
		blacklisted: make(map[string]bool),
	}
}

func (m *mockBlacklist) IsBlacklisted(ctx context.Context, token string) (bool, error) {
	if m.shouldError {
		return false, context.DeadlineExceeded
	}
	return m.blacklisted[token], nil
}

func getTestJWTConfig() *config.JWTConfig {
	return &config.JWTConfig{
		SecretKey:          "test-secret-key-at-least-32-chars!!",
		AccessTokenExpiry:  15 * time.Minute,
		RefreshTokenExpiry: 7 * 24 * time.Hour,
		Issuer:             "test-issuer",
	}
}

func generateTestToken(userID uint, email string, cfg *config.JWTConfig) string {
	jwtCfg := &utils.JWTConfig{
		SecretKey:          cfg.SecretKey,
		AccessTokenExpiry:  cfg.AccessTokenExpiry,
		RefreshTokenExpiry: cfg.RefreshTokenExpiry,
		Issuer:             cfg.Issuer,
	}
	token, _ := utils.GenerateAccessToken(userID, email, jwtCfg)
	return token
}

func generateTestRefreshToken(userID uint, email string, cfg *config.JWTConfig) string {
	jwtCfg := &utils.JWTConfig{
		SecretKey:          cfg.SecretKey,
		AccessTokenExpiry:  cfg.AccessTokenExpiry,
		RefreshTokenExpiry: cfg.RefreshTokenExpiry,
		Issuer:             cfg.Issuer,
	}
	token, _ := utils.GenerateRefreshToken(userID, email, jwtCfg)
	return token
}

func TestJWTMiddleware_ValidToken(t *testing.T) {
	cfg := getTestJWTConfig()
	blacklist := newMockBlacklist()

	app := fiber.New()
	app.Use(JWTMiddleware(cfg, blacklist))
	app.Get("/protected", func(c *fiber.Ctx) error {
		userID, ok := GetUserIDFromContext(c)
		if !ok {
			return c.SendStatus(fiber.StatusUnauthorized)
		}
		return c.JSON(fiber.Map{"user_id": userID})
	})

	token := generateTestToken(123, "test@example.com", cfg)

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Test request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Errorf("Expected status 200, got %d: %s", resp.StatusCode, string(body))
	}
}

func TestJWTMiddleware_MissingToken(t *testing.T) {
	cfg := getTestJWTConfig()
	blacklist := newMockBlacklist()

	app := fiber.New()
	app.Use(JWTMiddleware(cfg, blacklist))
	app.Get("/protected", func(c *fiber.Ctx) error {
		return c.SendString("should not reach here")
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	// No Authorization header

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Test request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", resp.StatusCode)
	}
}

func TestJWTMiddleware_MalformedToken(t *testing.T) {
	cfg := getTestJWTConfig()
	blacklist := newMockBlacklist()

	app := fiber.New()
	app.Use(JWTMiddleware(cfg, blacklist))
	app.Get("/protected", func(c *fiber.Ctx) error {
		return c.SendString("should not reach here")
	})

	tests := []struct {
		name   string
		header string
	}{
		{"no bearer prefix", "notbearer token123"},
		{"empty token", "Bearer "},
		{"garbage token", "Bearer garbage.token.here"},
		{"only bearer", "Bearer"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/protected", nil)
			req.Header.Set("Authorization", tt.header)

			resp, err := app.Test(req)
			if err != nil {
				t.Fatalf("Test request failed: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusUnauthorized {
				t.Errorf("Expected status 401, got %d", resp.StatusCode)
			}
		})
	}
}

func TestJWTMiddleware_WrongSecret(t *testing.T) {
	cfg := getTestJWTConfig()
	blacklist := newMockBlacklist()

	app := fiber.New()
	app.Use(JWTMiddleware(cfg, blacklist))
	app.Get("/protected", func(c *fiber.Ctx) error {
		return c.SendString("should not reach here")
	})

	// Generate token with different secret
	wrongCfg := &config.JWTConfig{
		SecretKey:          "different-secret-key-at-least-32!!",
		AccessTokenExpiry:  15 * time.Minute,
		RefreshTokenExpiry: 7 * 24 * time.Hour,
		Issuer:             "test-issuer",
	}
	token := generateTestToken(123, "test@example.com", wrongCfg)

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Test request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", resp.StatusCode)
	}
}

func TestJWTMiddleware_BlacklistedToken(t *testing.T) {
	cfg := getTestJWTConfig()
	blacklist := newMockBlacklist()

	token := generateTestToken(123, "test@example.com", cfg)
	blacklist.blacklisted[token] = true

	app := fiber.New()
	app.Use(JWTMiddleware(cfg, blacklist))
	app.Get("/protected", func(c *fiber.Ctx) error {
		return c.SendString("should not reach here")
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Test request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", resp.StatusCode)
	}
}

func TestJWTMiddleware_BlacklistError(t *testing.T) {
	cfg := getTestJWTConfig()
	blacklist := newMockBlacklist()
	blacklist.shouldError = true

	token := generateTestToken(123, "test@example.com", cfg)

	app := fiber.New()
	app.Use(JWTMiddleware(cfg, blacklist))
	app.Get("/protected", func(c *fiber.Ctx) error {
		return c.SendString("should not reach here")
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Test request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", resp.StatusCode)
	}
}

func TestJWTMiddleware_RefreshTokenRejected(t *testing.T) {
	cfg := getTestJWTConfig()
	blacklist := newMockBlacklist()

	// Use refresh token instead of access token
	token := generateTestRefreshToken(123, "test@example.com", cfg)

	app := fiber.New()
	app.Use(JWTMiddleware(cfg, blacklist))
	app.Get("/protected", func(c *fiber.Ctx) error {
		return c.SendString("should not reach here")
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Test request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", resp.StatusCode)
	}
}

func TestJWTMiddleware_NilBlacklist(t *testing.T) {
	cfg := getTestJWTConfig()

	app := fiber.New()
	app.Use(JWTMiddleware(cfg, nil)) // No blacklist
	app.Get("/protected", func(c *fiber.Ctx) error {
		userID, ok := GetUserIDFromContext(c)
		if !ok {
			return c.SendStatus(fiber.StatusUnauthorized)
		}
		return c.JSON(fiber.Map{"user_id": userID})
	})

	token := generateTestToken(123, "test@example.com", cfg)

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Test request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

func TestOptionalJWTMiddleware_WithToken(t *testing.T) {
	cfg := getTestJWTConfig()

	app := fiber.New()
	app.Use(OptionalJWTMiddleware(cfg, nil, nil))
	app.Get("/optional", func(c *fiber.Ctx) error {
		userID := c.Locals(UserIDKey)
		if userID == nil {
			return c.JSON(fiber.Map{"authenticated": false})
		}
		return c.JSON(fiber.Map{"authenticated": true, "user_id": userID})
	})

	token := generateTestToken(123, "test@example.com", cfg)

	req := httptest.NewRequest(http.MethodGet, "/optional", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Test request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

func TestOptionalJWTMiddleware_WithoutToken(t *testing.T) {
	cfg := getTestJWTConfig()

	app := fiber.New()
	app.Use(OptionalJWTMiddleware(cfg, nil, nil))
	app.Get("/optional", func(c *fiber.Ctx) error {
		userID := c.Locals(UserIDKey)
		if userID == nil {
			return c.JSON(fiber.Map{"authenticated": false})
		}
		return c.JSON(fiber.Map{"authenticated": true, "user_id": userID})
	})

	req := httptest.NewRequest(http.MethodGet, "/optional", nil)
	// No Authorization header

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Test request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

func TestOptionalJWTMiddleware_InvalidToken(t *testing.T) {
	cfg := getTestJWTConfig()

	app := fiber.New()
	app.Use(OptionalJWTMiddleware(cfg, nil, nil))
	app.Get("/optional", func(c *fiber.Ctx) error {
		userID := c.Locals(UserIDKey)
		if userID == nil {
			return c.JSON(fiber.Map{"authenticated": false})
		}
		return c.JSON(fiber.Map{"authenticated": true, "user_id": userID})
	})

	req := httptest.NewRequest(http.MethodGet, "/optional", nil)
	req.Header.Set("Authorization", "Bearer invalid.token.here")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Test request failed: %v", err)
	}
	defer resp.Body.Close()

	// Should still return 200 but as unauthenticated
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

// mockInvalidator implements TokenInvalidator interface for testing
type mockInvalidator struct {
	invalidatedAt map[uint]time.Time
	shouldError   bool
}

func newMockInvalidator() *mockInvalidator {
	return &mockInvalidator{
		invalidatedAt: make(map[uint]time.Time),
	}
}

func (m *mockInvalidator) GetTokensInvalidatedAt(ctx context.Context, userID uint) (time.Time, error) {
	if m.shouldError {
		return time.Time{}, context.DeadlineExceeded
	}
	t, ok := m.invalidatedAt[userID]
	if !ok {
		return time.Time{}, nil
	}
	return t, nil
}

func TestOptionalJWTMiddleware_BlacklistedToken(t *testing.T) {
	cfg := getTestJWTConfig()
	blacklist := newMockBlacklist()

	token := generateTestToken(123, "test@example.com", cfg)
	blacklist.blacklisted[token] = true

	app := fiber.New()
	app.Use(OptionalJWTMiddleware(cfg, blacklist, nil))
	app.Get("/optional", func(c *fiber.Ctx) error {
		userID := c.Locals(UserIDKey)
		if userID == nil {
			return c.JSON(fiber.Map{"authenticated": false})
		}
		return c.JSON(fiber.Map{"authenticated": true, "user_id": userID})
	})

	req := httptest.NewRequest(http.MethodGet, "/optional", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Test request failed: %v", err)
	}
	defer resp.Body.Close()

	// Should return 200 (optional) but NOT set user context
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	if string(body) != `{"authenticated":false}` {
		t.Errorf("Blacklisted token should NOT authenticate, got: %s", string(body))
	}
}

func TestOptionalJWTMiddleware_InvalidatedToken(t *testing.T) {
	cfg := getTestJWTConfig()
	invalidator := newMockInvalidator()

	// Generate token first, then invalidate
	token := generateTestToken(123, "test@example.com", cfg)

	// Mark tokens as invalidated AFTER the token was issued
	invalidator.invalidatedAt[123] = time.Now().Add(1 * time.Second)

	// Small delay to ensure token's iat is before invalidation time
	time.Sleep(10 * time.Millisecond)

	app := fiber.New()
	app.Use(OptionalJWTMiddleware(cfg, nil, invalidator))
	app.Get("/optional", func(c *fiber.Ctx) error {
		userID := c.Locals(UserIDKey)
		if userID == nil {
			return c.JSON(fiber.Map{"authenticated": false})
		}
		return c.JSON(fiber.Map{"authenticated": true, "user_id": userID})
	})

	req := httptest.NewRequest(http.MethodGet, "/optional", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Test request failed: %v", err)
	}
	defer resp.Body.Close()

	// Should return 200 but NOT set user context (token was invalidated)
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	if string(body) != `{"authenticated":false}` {
		t.Errorf("Invalidated token should NOT authenticate, got: %s", string(body))
	}
}

func TestOptionalJWTMiddleware_ValidTokenWithBlacklistAndInvalidator(t *testing.T) {
	cfg := getTestJWTConfig()
	blacklist := newMockBlacklist()
	invalidator := newMockInvalidator()

	token := generateTestToken(123, "test@example.com", cfg)
	// Token is NOT blacklisted and NOT invalidated

	app := fiber.New()
	app.Use(OptionalJWTMiddleware(cfg, blacklist, invalidator))
	app.Get("/optional", func(c *fiber.Ctx) error {
		userID := c.Locals(UserIDKey)
		if userID == nil {
			return c.JSON(fiber.Map{"authenticated": false})
		}
		return c.JSON(fiber.Map{"authenticated": true, "user_id": userID})
	})

	req := httptest.NewRequest(http.MethodGet, "/optional", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Test request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	if string(body) == `{"authenticated":false}` {
		t.Error("Valid token with passing blacklist+invalidator should authenticate")
	}
}

func TestOptionalJWTMiddleware_BlacklistError_TreatsAsUnauthenticated(t *testing.T) {
	cfg := getTestJWTConfig()
	blacklist := newMockBlacklist()
	blacklist.shouldError = true

	token := generateTestToken(123, "test@example.com", cfg)

	app := fiber.New()
	app.Use(OptionalJWTMiddleware(cfg, blacklist, nil))
	app.Get("/optional", func(c *fiber.Ctx) error {
		userID := c.Locals(UserIDKey)
		if userID == nil {
			return c.JSON(fiber.Map{"authenticated": false})
		}
		return c.JSON(fiber.Map{"authenticated": true, "user_id": userID})
	})

	req := httptest.NewRequest(http.MethodGet, "/optional", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Test request failed: %v", err)
	}
	defer resp.Body.Close()

	// Blacklist error in optional middleware should treat as unauthenticated (not 500)
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	if string(body) != `{"authenticated":false}` {
		t.Errorf("Blacklist error should treat as unauthenticated, got: %s", string(body))
	}
}

func TestExtractToken(t *testing.T) {
	tests := []struct {
		name     string
		header   string
		expected string
	}{
		{"valid bearer", "Bearer mytoken123", "mytoken123"},
		{"lowercase bearer", "bearer mytoken123", "mytoken123"},
		{"no header", "", ""},
		{"only bearer", "Bearer", ""},
		{"wrong format", "Basic credentials", ""},
		{"too many parts", "Bearer token extra", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := fiber.New()
			var extracted string

			app.Get("/test", func(c *fiber.Ctx) error {
				extracted = extractToken(c)
				return c.SendStatus(http.StatusOK)
			})

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			if tt.header != "" {
				req.Header.Set("Authorization", tt.header)
			}

			_, _ = app.Test(req)

			if extracted != tt.expected {
				t.Errorf("extractToken() = %v, want %v", extracted, tt.expected)
			}
		})
	}
}

func TestGetUserIDFromContext(t *testing.T) {
	cfg := getTestJWTConfig()
	blacklist := newMockBlacklist()

	t.Run("with valid token", func(t *testing.T) {
		app := fiber.New()
		app.Use(JWTMiddleware(cfg, blacklist))
		app.Get("/test", func(c *fiber.Ctx) error {
			userID, ok := GetUserIDFromContext(c)
			if !ok {
				return c.SendStatus(http.StatusUnauthorized)
			}
			return c.JSON(fiber.Map{"user_id": userID})
		})

		token := generateTestToken(42, "test@example.com", cfg)

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("Test request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}
	})

	t.Run("without token", func(t *testing.T) {
		app := fiber.New()
		app.Get("/test", func(c *fiber.Ctx) error {
			_, ok := GetUserIDFromContext(c)
			if !ok {
				return c.SendStatus(http.StatusUnauthorized)
			}
			return c.SendStatus(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("Test request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("Expected status 401, got %d", resp.StatusCode)
		}
	})
}

func TestGetUserIDFromContext_ReadsFromLocals(t *testing.T) {
	// Test that GetUserIDFromContext reads from c.Locals(UserIDKey) directly,
	// not from JWT claims. This verifies the M3 fix.
	t.Run("reads from locals when set directly", func(t *testing.T) {
		app := fiber.New()
		app.Get("/test", func(c *fiber.Ctx) error {
			// Set user ID directly in locals (simulating what JWT SuccessHandler does)
			c.Locals(UserIDKey, uint(999))

			userID, ok := GetUserIDFromContext(c)
			if !ok {
				return c.SendStatus(http.StatusUnauthorized)
			}
			if userID != 999 {
				return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
					"error": "wrong user ID",
					"got":   userID,
				})
			}
			return c.JSON(fiber.Map{"user_id": userID})
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("Test request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			t.Errorf("Expected 200, got %d: %s", resp.StatusCode, string(body))
		}
	})

	t.Run("returns false for zero user ID", func(t *testing.T) {
		app := fiber.New()
		app.Get("/test", func(c *fiber.Ctx) error {
			c.Locals(UserIDKey, uint(0))
			_, ok := GetUserIDFromContext(c)
			if ok {
				return c.Status(http.StatusInternalServerError).SendString("should not be ok for zero ID")
			}
			return c.SendStatus(http.StatusUnauthorized)
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		resp, _ := app.Test(req)
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("Expected 401 for zero user ID, got %d", resp.StatusCode)
		}
	})

	t.Run("returns false for wrong type in locals", func(t *testing.T) {
		app := fiber.New()
		app.Get("/test", func(c *fiber.Ctx) error {
			c.Locals(UserIDKey, "not-a-uint")
			_, ok := GetUserIDFromContext(c)
			if ok {
				return c.Status(http.StatusInternalServerError).SendString("should not be ok for string")
			}
			return c.SendStatus(http.StatusUnauthorized)
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		resp, _ := app.Test(req)
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("Expected 401 for wrong type, got %d", resp.StatusCode)
		}
	})
}

func TestGetEmailFromContext(t *testing.T) {
	cfg := getTestJWTConfig()
	blacklist := newMockBlacklist()

	app := fiber.New()
	app.Use(JWTMiddleware(cfg, blacklist))
	app.Get("/test", func(c *fiber.Ctx) error {
		email, ok := GetEmailFromContext(c)
		if !ok {
			return c.SendStatus(http.StatusUnauthorized)
		}
		return c.JSON(fiber.Map{"email": email})
	})

	token := generateTestToken(42, "test@example.com", cfg)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Test request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

func TestRequireAuth(t *testing.T) {
	cfg := getTestJWTConfig()
	blacklist := newMockBlacklist()

	t.Run("with authentication", func(t *testing.T) {
		app := fiber.New()
		app.Use(JWTMiddleware(cfg, blacklist))
		app.Use(RequireAuth())
		app.Get("/test", func(c *fiber.Ctx) error {
			return c.SendString("authenticated")
		})

		token := generateTestToken(42, "test@example.com", cfg)

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("Test request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}
	})

	t.Run("without authentication", func(t *testing.T) {
		app := fiber.New()
		app.Use(RequireAuth())
		app.Get("/test", func(c *fiber.Ctx) error {
			return c.SendString("should not reach here")
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("Test request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("Expected status 401, got %d", resp.StatusCode)
		}
	})
}
