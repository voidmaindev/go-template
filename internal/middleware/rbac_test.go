package middleware

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/casbin/casbin/v3"
	"github.com/casbin/casbin/v3/model"
	"github.com/gofiber/fiber/v2"
	"github.com/voidmaindev/go-template/internal/config"
	"github.com/voidmaindev/go-template/pkg/utils"
)

// Casbin model for testing
const testCasbinModel = `
[request_definition]
r = sub, dom, act

[policy_definition]
p = sub, dom, act

[role_definition]
g = _, _

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = g(r.sub, p.sub) && (p.dom == "*" || r.dom == p.dom) && (p.act == "*" || r.act == p.act)
`

func createTestEnforcer(t *testing.T) *casbin.Enforcer {
	t.Helper()

	m, err := model.NewModelFromString(testCasbinModel)
	if err != nil {
		t.Fatalf("Failed to create Casbin model: %v", err)
	}

	e, err := casbin.NewEnforcer(m)
	if err != nil {
		t.Fatalf("Failed to create Casbin enforcer: %v", err)
	}

	return e
}

func createTestEnforcerWithPolicies(t *testing.T) *casbin.Enforcer {
	t.Helper()

	e := createTestEnforcer(t)

	// Add role policies
	e.AddPolicy("admin", "*", "*")
	e.AddPolicy("reader", "item", "read")
	e.AddPolicy("writer", "item", "read")
	e.AddPolicy("writer", "item", "write")
	e.AddPolicy("writer", "item", "modify")
	e.AddPolicy("writer", "item", "delete")
	e.AddPolicy("limited", "document", "read")

	// Add user-role mappings
	e.AddRoleForUser("user:1", "admin")
	e.AddRoleForUser("user:2", "reader")
	e.AddRoleForUser("user:3", "writer")
	e.AddRoleForUser("user:4", "limited")
	// user:5 has multiple roles
	e.AddRoleForUser("user:5", "reader")
	e.AddRoleForUser("user:5", "limited")

	return e
}

func getTestRBACJWTConfig() *config.JWTConfig {
	return &config.JWTConfig{
		SecretKey:          "test-secret-key-at-least-32-chars!!",
		AccessTokenExpiry:  15 * time.Minute,
		RefreshTokenExpiry: 7 * 24 * time.Hour,
		Issuer:             "test-issuer",
	}
}

func generateTestRBACToken(userID uint, email string, cfg *config.JWTConfig) string {
	jwtCfg := &utils.JWTConfig{
		SecretKey:          cfg.SecretKey,
		AccessTokenExpiry:  cfg.AccessTokenExpiry,
		RefreshTokenExpiry: cfg.RefreshTokenExpiry,
		Issuer:             cfg.Issuer,
	}
	token, _ := utils.GenerateAccessToken(userID, email, jwtCfg)
	return token
}

func TestRequirePermission_Allowed(t *testing.T) {
	enforcer := createTestEnforcerWithPolicies(t)
	cfg := getTestRBACJWTConfig()
	blacklist := newMockBlacklist()

	app := fiber.New()
	app.Use(JWTMiddleware(cfg, blacklist))
	app.Get("/items", RequirePermission(enforcer, "item", "read"), func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"items": []string{"item1", "item2"}})
	})

	// User 2 has reader role with item:read permission
	token := generateTestRBACToken(2, "reader@example.com", cfg)

	req := httptest.NewRequest(http.MethodGet, "/items", nil)
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

func TestRequirePermission_Denied(t *testing.T) {
	enforcer := createTestEnforcerWithPolicies(t)
	cfg := getTestRBACJWTConfig()
	blacklist := newMockBlacklist()

	app := fiber.New()
	app.Use(JWTMiddleware(cfg, blacklist))
	app.Delete("/items/:id", RequirePermission(enforcer, "item", "delete"), func(c *fiber.Ctx) error {
		return c.SendStatus(http.StatusNoContent)
	})

	// User 2 has reader role - no delete permission
	token := generateTestRBACToken(2, "reader@example.com", cfg)

	req := httptest.NewRequest(http.MethodDelete, "/items/1", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Test request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("Expected status 403, got %d", resp.StatusCode)
	}
}

func TestRequirePermission_NotAuthenticated(t *testing.T) {
	enforcer := createTestEnforcerWithPolicies(t)

	app := fiber.New()
	// No JWT middleware - user is not authenticated
	app.Get("/items", RequirePermission(enforcer, "item", "read"), func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"items": []string{}})
	})

	req := httptest.NewRequest(http.MethodGet, "/items", nil)

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Test request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", resp.StatusCode)
	}
}

func TestRequirePermission_AdminWildcard(t *testing.T) {
	enforcer := createTestEnforcerWithPolicies(t)
	cfg := getTestRBACJWTConfig()
	blacklist := newMockBlacklist()

	app := fiber.New()
	app.Use(JWTMiddleware(cfg, blacklist))
	app.Delete("/users/:id", RequirePermission(enforcer, "user", "delete"), func(c *fiber.Ctx) error {
		return c.SendStatus(http.StatusNoContent)
	})

	// User 1 has admin role with wildcard permission
	token := generateTestRBACToken(1, "admin@example.com", cfg)

	req := httptest.NewRequest(http.MethodDelete, "/users/99", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Test request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		t.Errorf("Expected status 204, got %d: %s", resp.StatusCode, string(body))
	}
}

func TestRequirePermission_UserWithNoRole(t *testing.T) {
	enforcer := createTestEnforcerWithPolicies(t)
	cfg := getTestRBACJWTConfig()
	blacklist := newMockBlacklist()

	app := fiber.New()
	app.Use(JWTMiddleware(cfg, blacklist))
	app.Get("/items", RequirePermission(enforcer, "item", "read"), func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"items": []string{}})
	})

	// User 99 has no roles assigned
	token := generateTestRBACToken(99, "norole@example.com", cfg)

	req := httptest.NewRequest(http.MethodGet, "/items", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Test request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("Expected status 403, got %d", resp.StatusCode)
	}
}

func TestRequireAnyPermission_FirstPermission(t *testing.T) {
	enforcer := createTestEnforcerWithPolicies(t)
	cfg := getTestRBACJWTConfig()
	blacklist := newMockBlacklist()

	app := fiber.New()
	app.Use(JWTMiddleware(cfg, blacklist))
	app.Get("/mixed", RequireAnyPermission(enforcer,
		NewPermission("item", "read"),
		NewPermission("document", "read"),
	), func(c *fiber.Ctx) error {
		return c.SendString("success")
	})

	// User 2 has item:read permission
	token := generateTestRBACToken(2, "reader@example.com", cfg)

	req := httptest.NewRequest(http.MethodGet, "/mixed", nil)
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

func TestRequireAnyPermission_SecondPermission(t *testing.T) {
	enforcer := createTestEnforcerWithPolicies(t)
	cfg := getTestRBACJWTConfig()
	blacklist := newMockBlacklist()

	app := fiber.New()
	app.Use(JWTMiddleware(cfg, blacklist))
	app.Get("/mixed", RequireAnyPermission(enforcer,
		NewPermission("item", "write"),
		NewPermission("document", "read"),
	), func(c *fiber.Ctx) error {
		return c.SendString("success")
	})

	// User 4 has limited role with document:read
	token := generateTestRBACToken(4, "limited@example.com", cfg)

	req := httptest.NewRequest(http.MethodGet, "/mixed", nil)
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

func TestRequireAnyPermission_NoMatchingPermission(t *testing.T) {
	enforcer := createTestEnforcerWithPolicies(t)
	cfg := getTestRBACJWTConfig()
	blacklist := newMockBlacklist()

	app := fiber.New()
	app.Use(JWTMiddleware(cfg, blacklist))
	app.Get("/admin-only", RequireAnyPermission(enforcer,
		NewPermission("user", "delete"),
		NewPermission("rbac", "write"),
	), func(c *fiber.Ctx) error {
		return c.SendString("success")
	})

	// User 2 has reader role - no user:delete or rbac:write
	token := generateTestRBACToken(2, "reader@example.com", cfg)

	req := httptest.NewRequest(http.MethodGet, "/admin-only", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Test request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("Expected status 403, got %d", resp.StatusCode)
	}
}

func TestRequireAnyPermission_NotAuthenticated(t *testing.T) {
	enforcer := createTestEnforcerWithPolicies(t)

	app := fiber.New()
	app.Get("/mixed", RequireAnyPermission(enforcer,
		NewPermission("item", "read"),
	), func(c *fiber.Ctx) error {
		return c.SendString("success")
	})

	req := httptest.NewRequest(http.MethodGet, "/mixed", nil)

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Test request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", resp.StatusCode)
	}
}

func TestRequireAllPermissions_AllGranted(t *testing.T) {
	enforcer := createTestEnforcerWithPolicies(t)
	cfg := getTestRBACJWTConfig()
	blacklist := newMockBlacklist()

	app := fiber.New()
	app.Use(JWTMiddleware(cfg, blacklist))
	app.Post("/items", RequireAllPermissions(enforcer,
		NewPermission("item", "read"),
		NewPermission("item", "write"),
	), func(c *fiber.Ctx) error {
		return c.SendString("created")
	})

	// User 3 has writer role with item:read and item:write
	token := generateTestRBACToken(3, "writer@example.com", cfg)

	req := httptest.NewRequest(http.MethodPost, "/items", nil)
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

func TestRequireAllPermissions_PartiallyDenied(t *testing.T) {
	enforcer := createTestEnforcerWithPolicies(t)
	cfg := getTestRBACJWTConfig()
	blacklist := newMockBlacklist()

	app := fiber.New()
	app.Use(JWTMiddleware(cfg, blacklist))
	app.Post("/items", RequireAllPermissions(enforcer,
		NewPermission("item", "read"),
		NewPermission("item", "write"),
	), func(c *fiber.Ctx) error {
		return c.SendString("created")
	})

	// User 2 has reader role - only has item:read, not item:write
	token := generateTestRBACToken(2, "reader@example.com", cfg)

	req := httptest.NewRequest(http.MethodPost, "/items", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Test request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("Expected status 403, got %d", resp.StatusCode)
	}
}

func TestRequireAllPermissions_NotAuthenticated(t *testing.T) {
	enforcer := createTestEnforcerWithPolicies(t)

	app := fiber.New()
	app.Post("/items", RequireAllPermissions(enforcer,
		NewPermission("item", "read"),
		NewPermission("item", "write"),
	), func(c *fiber.Ctx) error {
		return c.SendString("created")
	})

	req := httptest.NewRequest(http.MethodPost, "/items", nil)

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Test request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", resp.StatusCode)
	}
}

func TestRequireAllPermissions_AdminWildcard(t *testing.T) {
	enforcer := createTestEnforcerWithPolicies(t)
	cfg := getTestRBACJWTConfig()
	blacklist := newMockBlacklist()

	app := fiber.New()
	app.Use(JWTMiddleware(cfg, blacklist))
	app.Delete("/sensitive", RequireAllPermissions(enforcer,
		NewPermission("user", "delete"),
		NewPermission("rbac", "modify"),
		NewPermission("item", "delete"),
	), func(c *fiber.Ctx) error {
		return c.SendString("deleted")
	})

	// User 1 has admin role with wildcard
	token := generateTestRBACToken(1, "admin@example.com", cfg)

	req := httptest.NewRequest(http.MethodDelete, "/sensitive", nil)
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

func TestHasPermission_WithPermission(t *testing.T) {
	enforcer := createTestEnforcerWithPolicies(t)
	cfg := getTestRBACJWTConfig()
	blacklist := newMockBlacklist()

	app := fiber.New()
	app.Use(JWTMiddleware(cfg, blacklist))
	app.Get("/check", func(c *fiber.Ctx) error {
		canRead := HasPermission(c, enforcer, "item", "read")
		canDelete := HasPermission(c, enforcer, "item", "delete")
		return c.JSON(fiber.Map{
			"can_read":   canRead,
			"can_delete": canDelete,
		})
	})

	// User 2 has reader role with item:read only
	token := generateTestRBACToken(2, "reader@example.com", cfg)

	req := httptest.NewRequest(http.MethodGet, "/check", nil)
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

func TestHasPermission_NotAuthenticated(t *testing.T) {
	enforcer := createTestEnforcerWithPolicies(t)

	app := fiber.New()
	app.Get("/check", func(c *fiber.Ctx) error {
		// User is not authenticated
		canRead := HasPermission(c, enforcer, "item", "read")
		return c.JSON(fiber.Map{"can_read": canRead})
	})

	req := httptest.NewRequest(http.MethodGet, "/check", nil)

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Test request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

func TestNewPermission(t *testing.T) {
	perm := NewPermission("item", "read")

	if perm.Domain != "item" {
		t.Errorf("Expected domain 'item', got '%s'", perm.Domain)
	}
	if perm.Action != "read" {
		t.Errorf("Expected action 'read', got '%s'", perm.Action)
	}
}

func TestRequirePermission_MultipleRoles(t *testing.T) {
	enforcer := createTestEnforcerWithPolicies(t)
	cfg := getTestRBACJWTConfig()
	blacklist := newMockBlacklist()

	app := fiber.New()
	app.Use(JWTMiddleware(cfg, blacklist))

	// Route requiring item:read
	app.Get("/items", RequirePermission(enforcer, "item", "read"), func(c *fiber.Ctx) error {
		return c.SendString("items")
	})

	// Route requiring document:read
	app.Get("/documents", RequirePermission(enforcer, "document", "read"), func(c *fiber.Ctx) error {
		return c.SendString("documents")
	})

	// User 5 has both reader (item:read) and limited (document:read) roles
	token := generateTestRBACToken(5, "multirole@example.com", cfg)

	t.Run("can access items via reader role", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/items", nil)
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

	t.Run("can access documents via limited role", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/documents", nil)
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
}

func TestRequirePermission_ChainedMiddleware(t *testing.T) {
	enforcer := createTestEnforcerWithPolicies(t)
	cfg := getTestRBACJWTConfig()
	blacklist := newMockBlacklist()

	app := fiber.New()
	app.Use(JWTMiddleware(cfg, blacklist))

	// Chained permission checks
	items := app.Group("/items")
	items.Use(RequirePermission(enforcer, "item", "read"))
	items.Delete("/:id", RequirePermission(enforcer, "item", "delete"), func(c *fiber.Ctx) error {
		return c.SendStatus(http.StatusNoContent)
	})

	t.Run("writer can delete items", func(t *testing.T) {
		// User 3 has writer role with all item permissions
		token := generateTestRBACToken(3, "writer@example.com", cfg)

		req := httptest.NewRequest(http.MethodDelete, "/items/1", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("Test request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusNoContent {
			t.Errorf("Expected status 204, got %d", resp.StatusCode)
		}
	})

	t.Run("reader cannot delete items", func(t *testing.T) {
		// User 2 has reader role - no delete permission
		token := generateTestRBACToken(2, "reader@example.com", cfg)

		req := httptest.NewRequest(http.MethodDelete, "/items/1", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("Test request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusForbidden {
			t.Errorf("Expected status 403, got %d", resp.StatusCode)
		}
	})
}
