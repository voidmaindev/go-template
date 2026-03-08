package rbac

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/casbin/casbin/v3"
	"github.com/gofiber/fiber/v2"
	"github.com/voidmaindev/go-template/internal/common"
	"github.com/voidmaindev/go-template/internal/common/filter"
)

// mockService implements the Service interface for testing
type mockService struct {
	roles           map[string]*RoleWithPermissions
	userRoles       map[uint][]UserRoleResponse
	createRoleErr   error
	getRoleErr      error
	listRolesErr    error
	updateRoleErr   error
	deleteRoleErr   error
	getUserRolesErr error
	assignRoleErr   error
	removeRoleErr   error
	domains         []DomainResponse
	actions         []string
	adminCount      int
}

func newMockService() *mockService {
	return &mockService{
		roles:     make(map[string]*RoleWithPermissions),
		userRoles: make(map[uint][]UserRoleResponse),
		domains: []DomainResponse{
			{Name: "user", IsProtected: true},
			{Name: "rbac", IsProtected: true},
			{Name: "item", IsProtected: false},
		},
		actions:    AllActions(),
		adminCount: 1,
	}
}

func (m *mockService) CreateRole(ctx context.Context, req *CreateRoleRequest) (*Role, error) {
	if m.createRoleErr != nil {
		return nil, m.createRoleErr
	}
	role := &Role{
		Code:        req.Code,
		Name:        req.Name,
		Description: req.Description,
		IsSystem:    false,
	}
	role.ID = 1
	m.roles[req.Code] = &RoleWithPermissions{Role: role}
	return role, nil
}

func (m *mockService) GetRoleByCode(ctx context.Context, code string) (*RoleWithPermissions, error) {
	if m.getRoleErr != nil {
		return nil, m.getRoleErr
	}
	rwp, ok := m.roles[code]
	if !ok {
		return nil, ErrRoleNotFound
	}
	return rwp, nil
}

func (m *mockService) ListRoles(ctx context.Context, params *filter.Params) (*common.PaginatedResult[Role], error) {
	if m.listRolesErr != nil {
		return nil, m.listRolesErr
	}
	var roles []Role
	for _, rwp := range m.roles {
		roles = append(roles, *rwp.Role)
	}
	return common.NewPaginatedResultFromFilter(roles, int64(len(roles)), params), nil
}

func (m *mockService) UpdateRolePermissions(ctx context.Context, code string, req *UpdateRolePermissionsRequest) (*RoleWithPermissions, error) {
	if m.updateRoleErr != nil {
		return nil, m.updateRoleErr
	}
	rwp, ok := m.roles[code]
	if !ok {
		return nil, ErrRoleNotFound
	}
	// Convert permissions
	var perms []Permission
	for _, p := range req.Permissions {
		perms = append(perms, Permission{Domain: p.Domain, Actions: p.Actions})
	}
	rwp.Permissions = perms
	return rwp, nil
}

func (m *mockService) DeleteRole(ctx context.Context, code string) error {
	if m.deleteRoleErr != nil {
		return m.deleteRoleErr
	}
	if _, ok := m.roles[code]; !ok {
		return ErrRoleNotFound
	}
	delete(m.roles, code)
	return nil
}

func (m *mockService) GetUserRoles(ctx context.Context, userID uint) ([]UserRoleResponse, error) {
	if m.getUserRolesErr != nil {
		return nil, m.getUserRolesErr
	}
	return m.userRoles[userID], nil
}

func (m *mockService) AssignRole(ctx context.Context, userID uint, roleCode string) error {
	if m.assignRoleErr != nil {
		return m.assignRoleErr
	}
	m.userRoles[userID] = append(m.userRoles[userID], UserRoleResponse{Code: roleCode, Name: roleCode})
	return nil
}

func (m *mockService) RemoveRole(ctx context.Context, userID uint, roleCode string) error {
	if m.removeRoleErr != nil {
		return m.removeRoleErr
	}
	roles := m.userRoles[userID]
	for i, r := range roles {
		if r.Code == roleCode {
			m.userRoles[userID] = append(roles[:i], roles[i+1:]...)
			return nil
		}
	}
	return ErrRoleNotAssigned
}

func (m *mockService) CheckPermission(ctx context.Context, userID uint, domain, action string) (bool, error) {
	return true, nil
}

func (m *mockService) GetDomains(ctx context.Context) []DomainResponse {
	return m.domains
}

func (m *mockService) GetActions(ctx context.Context) []string {
	return m.actions
}

func (m *mockService) SyncGlobalRoles(ctx context.Context) error {
	return nil
}

func (m *mockService) CountAdminUsers(ctx context.Context) (int, error) {
	return m.adminCount, nil
}

func (m *mockService) AssignRoleInTx(tx *casbin.Transaction, ctx context.Context, userID uint, roleCode string) error {
	return m.AssignRole(ctx, userID, roleCode)
}

func (m *mockService) GetTransactionalEnforcer() *casbin.TransactionalEnforcer {
	return nil
}

func setupTestApp(svc Service) *fiber.App {
	app := fiber.New()
	handler := NewHandler(svc)

	// Register routes
	rbac := app.Group("/api/v1/rbac")

	// Roles
	rbac.Get("/roles", handler.ListRoles)
	rbac.Post("/roles", handler.CreateRole)
	rbac.Get("/roles/:code", handler.GetRole)
	rbac.Put("/roles/:code/permissions", handler.UpdateRolePermissions)
	rbac.Delete("/roles/:code", handler.DeleteRole)

	// User roles
	rbac.Get("/users/:id/roles", handler.GetUserRoles)
	rbac.Post("/users/:id/roles", handler.AssignRole)
	rbac.Delete("/users/:id/roles/:code", handler.RemoveRole)

	// Discovery
	rbac.Get("/domains", handler.GetDomains)
	rbac.Get("/actions", handler.GetActions)

	return app
}

func TestHandler_ListRoles(t *testing.T) {
	t.Run("successful list", func(t *testing.T) {
		svc := newMockService()
		svc.roles["admin"] = &RoleWithPermissions{
			Role: &Role{Code: "admin", Name: "Administrator", IsSystem: true},
		}

		app := setupTestApp(svc)
		req := httptest.NewRequest(http.MethodGet, "/api/v1/rbac/roles", nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("Test request failed: %v", err)
		}

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Status = %d, want %d", resp.StatusCode, http.StatusOK)
		}
	})

	t.Run("list error", func(t *testing.T) {
		svc := newMockService()
		svc.listRolesErr = errors.New("database error")

		app := setupTestApp(svc)
		req := httptest.NewRequest(http.MethodGet, "/api/v1/rbac/roles", nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("Test request failed: %v", err)
		}

		if resp.StatusCode != http.StatusInternalServerError {
			t.Errorf("Status = %d, want %d", resp.StatusCode, http.StatusInternalServerError)
		}
	})
}

func TestHandler_CreateRole(t *testing.T) {
	t.Run("successful creation", func(t *testing.T) {
		svc := newMockService()
		app := setupTestApp(svc)

		body := CreateRoleRequest{
			Code:        "editor",
			Name:        "Editor",
			Description: "Can edit content",
			Permissions: []PermissionInput{
				{Domain: "item", Actions: []string{"read", "create", "update"}},
			},
		}
		bodyBytes, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/rbac/roles", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("Test request failed: %v", err)
		}

		if resp.StatusCode != http.StatusCreated {
			body, _ := io.ReadAll(resp.Body)
			t.Errorf("Status = %d, want %d. Body: %s", resp.StatusCode, http.StatusCreated, string(body))
		}
	})

	t.Run("invalid request body", func(t *testing.T) {
		svc := newMockService()
		app := setupTestApp(svc)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/rbac/roles", bytes.NewReader([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("Test request failed: %v", err)
		}

		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Status = %d, want %d", resp.StatusCode, http.StatusBadRequest)
		}
	})

	t.Run("validation error - missing fields", func(t *testing.T) {
		svc := newMockService()
		app := setupTestApp(svc)

		body := CreateRoleRequest{
			Code: "", // Missing required field
		}
		bodyBytes, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/rbac/roles", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("Test request failed: %v", err)
		}

		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Status = %d, want %d", resp.StatusCode, http.StatusBadRequest)
		}
	})

	t.Run("role code already exists", func(t *testing.T) {
		svc := newMockService()
		svc.createRoleErr = ErrRoleCodeExists
		app := setupTestApp(svc)

		body := CreateRoleRequest{
			Code:        "existing",
			Name:        "Existing Role",
			Permissions: []PermissionInput{{Domain: "item", Actions: []string{"read"}}},
		}
		bodyBytes, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/rbac/roles", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("Test request failed: %v", err)
		}

		if resp.StatusCode != http.StatusConflict {
			t.Errorf("Status = %d, want %d", resp.StatusCode, http.StatusConflict)
		}
	})
}

func TestHandler_GetRole(t *testing.T) {
	t.Run("successful get", func(t *testing.T) {
		svc := newMockService()
		svc.roles["admin"] = &RoleWithPermissions{
			Role:        &Role{Code: "admin", Name: "Administrator"},
			Permissions: []Permission{{Domain: "*", Actions: []string{"*"}}},
		}

		app := setupTestApp(svc)
		req := httptest.NewRequest(http.MethodGet, "/api/v1/rbac/roles/admin", nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("Test request failed: %v", err)
		}

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Status = %d, want %d", resp.StatusCode, http.StatusOK)
		}
	})

	t.Run("role not found", func(t *testing.T) {
		svc := newMockService()
		app := setupTestApp(svc)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/rbac/roles/nonexistent", nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("Test request failed: %v", err)
		}

		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("Status = %d, want %d", resp.StatusCode, http.StatusNotFound)
		}
	})
}

func TestHandler_UpdateRolePermissions(t *testing.T) {
	t.Run("successful update", func(t *testing.T) {
		svc := newMockService()
		svc.roles["editor"] = &RoleWithPermissions{
			Role: &Role{Code: "editor", Name: "Editor", IsSystem: false},
		}

		app := setupTestApp(svc)

		body := UpdateRolePermissionsRequest{
			Permissions: []PermissionInput{
				{Domain: "item", Actions: []string{"read", "create"}},
			},
		}
		bodyBytes, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPut, "/api/v1/rbac/roles/editor/permissions", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("Test request failed: %v", err)
		}

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Status = %d, want %d", resp.StatusCode, http.StatusOK)
		}
	})

	t.Run("role not found", func(t *testing.T) {
		svc := newMockService()
		app := setupTestApp(svc)

		body := UpdateRolePermissionsRequest{
			Permissions: []PermissionInput{
				{Domain: "item", Actions: []string{"read"}},
			},
		}
		bodyBytes, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPut, "/api/v1/rbac/roles/nonexistent/permissions", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("Test request failed: %v", err)
		}

		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("Status = %d, want %d", resp.StatusCode, http.StatusNotFound)
		}
	})

	t.Run("system role cannot be modified", func(t *testing.T) {
		svc := newMockService()
		svc.updateRoleErr = ErrSystemRoleCannotBeModified
		svc.roles["admin"] = &RoleWithPermissions{
			Role: &Role{Code: "admin", Name: "Admin", IsSystem: true},
		}

		app := setupTestApp(svc)

		body := UpdateRolePermissionsRequest{
			Permissions: []PermissionInput{
				{Domain: "item", Actions: []string{"read"}},
			},
		}
		bodyBytes, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPut, "/api/v1/rbac/roles/admin/permissions", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("Test request failed: %v", err)
		}

		if resp.StatusCode != http.StatusForbidden {
			t.Errorf("Status = %d, want %d", resp.StatusCode, http.StatusForbidden)
		}
	})
}

func TestHandler_DeleteRole(t *testing.T) {
	t.Run("successful delete", func(t *testing.T) {
		svc := newMockService()
		svc.roles["custom"] = &RoleWithPermissions{
			Role: &Role{Code: "custom", Name: "Custom", IsSystem: false},
		}

		app := setupTestApp(svc)
		req := httptest.NewRequest(http.MethodDelete, "/api/v1/rbac/roles/custom", nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("Test request failed: %v", err)
		}

		if resp.StatusCode != http.StatusNoContent {
			t.Errorf("Status = %d, want %d", resp.StatusCode, http.StatusNoContent)
		}
	})

	t.Run("role not found", func(t *testing.T) {
		svc := newMockService()
		app := setupTestApp(svc)

		req := httptest.NewRequest(http.MethodDelete, "/api/v1/rbac/roles/nonexistent", nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("Test request failed: %v", err)
		}

		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("Status = %d, want %d", resp.StatusCode, http.StatusNotFound)
		}
	})

	t.Run("system role cannot be deleted", func(t *testing.T) {
		svc := newMockService()
		svc.deleteRoleErr = ErrSystemRoleCannotBeDeleted
		svc.roles["admin"] = &RoleWithPermissions{
			Role: &Role{Code: "admin", Name: "Admin", IsSystem: true},
		}

		app := setupTestApp(svc)
		req := httptest.NewRequest(http.MethodDelete, "/api/v1/rbac/roles/admin", nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("Test request failed: %v", err)
		}

		if resp.StatusCode != http.StatusForbidden {
			t.Errorf("Status = %d, want %d", resp.StatusCode, http.StatusForbidden)
		}
	})
}

func TestHandler_GetUserRoles(t *testing.T) {
	t.Run("successful get", func(t *testing.T) {
		svc := newMockService()
		svc.userRoles[1] = []UserRoleResponse{
			{Code: "admin", Name: "Administrator"},
			{Code: "editor", Name: "Editor"},
		}

		app := setupTestApp(svc)
		req := httptest.NewRequest(http.MethodGet, "/api/v1/rbac/users/1/roles", nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("Test request failed: %v", err)
		}

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Status = %d, want %d", resp.StatusCode, http.StatusOK)
		}
	})

	t.Run("invalid user ID", func(t *testing.T) {
		svc := newMockService()
		app := setupTestApp(svc)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/rbac/users/invalid/roles", nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("Test request failed: %v", err)
		}

		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Status = %d, want %d", resp.StatusCode, http.StatusBadRequest)
		}
	})
}

func TestHandler_AssignRole(t *testing.T) {
	t.Run("successful assignment", func(t *testing.T) {
		svc := newMockService()
		svc.roles["editor"] = &RoleWithPermissions{
			Role: &Role{Code: "editor", Name: "Editor"},
		}

		app := setupTestApp(svc)

		body := AssignRoleRequest{RoleCode: "editor"}
		bodyBytes, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/rbac/users/1/roles", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("Test request failed: %v", err)
		}

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Status = %d, want %d", resp.StatusCode, http.StatusOK)
		}
	})

	t.Run("role not found", func(t *testing.T) {
		svc := newMockService()
		svc.assignRoleErr = ErrRoleNotFound

		app := setupTestApp(svc)

		body := AssignRoleRequest{RoleCode: "nonexistent"}
		bodyBytes, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/rbac/users/1/roles", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("Test request failed: %v", err)
		}

		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("Status = %d, want %d", resp.StatusCode, http.StatusNotFound)
		}
	})

	t.Run("role already assigned", func(t *testing.T) {
		svc := newMockService()
		svc.assignRoleErr = ErrRoleAlreadyAssigned

		app := setupTestApp(svc)

		body := AssignRoleRequest{RoleCode: "editor"}
		bodyBytes, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/rbac/users/1/roles", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("Test request failed: %v", err)
		}

		if resp.StatusCode != http.StatusConflict {
			t.Errorf("Status = %d, want %d", resp.StatusCode, http.StatusConflict)
		}
	})
}

func TestHandler_RemoveRole(t *testing.T) {
	t.Run("successful removal", func(t *testing.T) {
		svc := newMockService()
		svc.userRoles[1] = []UserRoleResponse{{Code: "editor", Name: "Editor"}}

		app := setupTestApp(svc)
		req := httptest.NewRequest(http.MethodDelete, "/api/v1/rbac/users/1/roles/editor", nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("Test request failed: %v", err)
		}

		if resp.StatusCode != http.StatusNoContent {
			t.Errorf("Status = %d, want %d", resp.StatusCode, http.StatusNoContent)
		}
	})

	t.Run("role not assigned", func(t *testing.T) {
		svc := newMockService()
		app := setupTestApp(svc)

		req := httptest.NewRequest(http.MethodDelete, "/api/v1/rbac/users/1/roles/nonexistent", nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("Test request failed: %v", err)
		}

		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("Status = %d, want %d", resp.StatusCode, http.StatusNotFound)
		}
	})

	t.Run("cannot remove last admin", func(t *testing.T) {
		svc := newMockService()
		svc.removeRoleErr = ErrCannotRemoveLastAdmin

		app := setupTestApp(svc)
		req := httptest.NewRequest(http.MethodDelete, "/api/v1/rbac/users/1/roles/admin", nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("Test request failed: %v", err)
		}

		if resp.StatusCode != http.StatusForbidden {
			t.Errorf("Status = %d, want %d", resp.StatusCode, http.StatusForbidden)
		}
	})
}

func TestHandler_GetDomains(t *testing.T) {
	svc := newMockService()
	app := setupTestApp(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/rbac/domains", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Test request failed: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	// Parse response
	body, _ := io.ReadAll(resp.Body)
	var result common.Response
	json.Unmarshal(body, &result)

	if !result.Success {
		t.Error("Response should be successful")
	}
}

func TestHandler_GetActions(t *testing.T) {
	svc := newMockService()
	app := setupTestApp(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/rbac/actions", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Test request failed: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	// Parse response
	body, _ := io.ReadAll(resp.Body)
	var result common.Response
	json.Unmarshal(body, &result)

	if !result.Success {
		t.Error("Response should be successful")
	}
}
