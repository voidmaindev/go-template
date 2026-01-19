package rbac

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/voidmaindev/go-template/internal/common"
	commonerrors "github.com/voidmaindev/go-template/internal/common/errors"
	"github.com/voidmaindev/go-template/internal/common/filter"
	"gorm.io/gorm"
)

// mockRepository implements the Repository interface for testing
type mockRepository struct {
	roles      map[uint]*Role
	codeIndex  map[string]*Role
	nextID     uint
	findErr    error
	createErr  error
	updateErr  error
	deleteErr  error
}

func newMockRepository() *mockRepository {
	return &mockRepository{
		roles:     make(map[uint]*Role),
		codeIndex: make(map[string]*Role),
		nextID:    1,
	}
}

func (m *mockRepository) Create(ctx context.Context, entity *Role) error {
	if m.createErr != nil {
		return m.createErr
	}
	entity.ID = m.nextID
	entity.CreatedAt = time.Now()
	entity.UpdatedAt = time.Now()
	m.nextID++
	m.roles[entity.ID] = entity
	m.codeIndex[entity.Code] = entity
	return nil
}

func (m *mockRepository) CreateBatch(ctx context.Context, entities []Role, batchSize int) error {
	for i := range entities {
		if err := m.Create(ctx, &entities[i]); err != nil {
			return err
		}
	}
	return nil
}

func (m *mockRepository) FindByID(ctx context.Context, id uint) (*Role, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	role, ok := m.roles[id]
	if !ok {
		return nil, commonerrors.NotFound("repository", "entity")
	}
	return role, nil
}

func (m *mockRepository) FindAll(ctx context.Context, pagination *common.Pagination) ([]Role, int64, error) {
	var roles []Role
	for _, r := range m.roles {
		roles = append(roles, *r)
	}
	return roles, int64(len(roles)), nil
}

func (m *mockRepository) FindByCondition(ctx context.Context, condition map[string]any, pagination *common.Pagination) ([]Role, int64, error) {
	return m.FindAll(ctx, pagination)
}

func (m *mockRepository) FindAllFiltered(ctx context.Context, params *filter.Params) ([]Role, int64, error) {
	return m.FindAll(ctx, nil)
}

func (m *mockRepository) FindOne(ctx context.Context, condition map[string]any) (*Role, error) {
	if code, ok := condition["code"].(string); ok {
		if role, exists := m.codeIndex[code]; exists {
			return role, nil
		}
	}
	return nil, commonerrors.NotFound("repository", "entity")
}

func (m *mockRepository) Update(ctx context.Context, entity *Role) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	if _, ok := m.roles[entity.ID]; !ok {
		return commonerrors.NotFound("repository", "entity")
	}
	entity.UpdatedAt = time.Now()
	m.roles[entity.ID] = entity
	m.codeIndex[entity.Code] = entity
	return nil
}

func (m *mockRepository) UpdateFields(ctx context.Context, id uint, fields map[string]any) error {
	return m.updateErr
}

func (m *mockRepository) Delete(ctx context.Context, id uint) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	role, ok := m.roles[id]
	if !ok {
		return commonerrors.NotFound("repository", "entity")
	}
	delete(m.roles, id)
	delete(m.codeIndex, role.Code)
	return nil
}

func (m *mockRepository) HardDelete(ctx context.Context, id uint) error {
	return m.Delete(ctx, id)
}

func (m *mockRepository) Exists(ctx context.Context, condition map[string]any) (bool, error) {
	if code, ok := condition["code"].(string); ok {
		_, exists := m.codeIndex[code]
		return exists, nil
	}
	return false, nil
}

func (m *mockRepository) Count(ctx context.Context, condition map[string]any) (int64, error) {
	return int64(len(m.roles)), nil
}

func (m *mockRepository) WithTx(tx *gorm.DB) common.Repository[Role] {
	return m
}

func (m *mockRepository) WithPreload(preloads ...string) common.Repository[Role] {
	return m
}

func (m *mockRepository) Transaction(ctx context.Context, fn func(txRepo common.Repository[Role]) error) error {
	return fn(m)
}

func (m *mockRepository) FindByCode(ctx context.Context, code string) (*Role, error) {
	role, ok := m.codeIndex[code]
	if !ok {
		return nil, commonerrors.NotFound("repository", "entity")
	}
	return role, nil
}

func (m *mockRepository) ExistsByCode(ctx context.Context, code string) (bool, error) {
	_, exists := m.codeIndex[code]
	return exists, nil
}

func (m *mockRepository) FindSystemRoles(ctx context.Context) ([]Role, error) {
	var systemRoles []Role
	for _, r := range m.roles {
		if r.IsSystem {
			systemRoles = append(systemRoles, *r)
		}
	}
	return systemRoles, nil
}

// mockEnforcer implements a simple mock for Casbin enforcer
type mockEnforcer struct {
	policies   [][]string
	groupings  [][]string
	enforceErr error
}

func newMockEnforcer() *mockEnforcer {
	return &mockEnforcer{
		policies:  make([][]string, 0),
		groupings: make([][]string, 0),
	}
}

// mockDomainProvider implements DomainProvider for testing
type mockDomainProvider struct {
	domains []string
}

func newMockDomainProvider() *mockDomainProvider {
	return &mockDomainProvider{
		domains: []string{"user", "rbac", "item", "country", "city", "document"},
	}
}

func (m *mockDomainProvider) GetDomainNames() []string {
	return m.domains
}

func TestPermissions(t *testing.T) {
	t.Run("AllActions returns all actions", func(t *testing.T) {
		actions := AllActions()
		if len(actions) != 4 {
			t.Errorf("AllActions() returned %d actions, want 4", len(actions))
		}

		expected := map[string]bool{
			ActionRead:   true,
			ActionWrite:  true,
			ActionModify: true,
			ActionDelete: true,
		}

		for _, action := range actions {
			if !expected[action] {
				t.Errorf("Unexpected action: %s", action)
			}
		}
	})

	t.Run("IsValidAction", func(t *testing.T) {
		tests := []struct {
			action string
			valid  bool
		}{
			{ActionRead, true},
			{ActionWrite, true},
			{ActionModify, true},
			{ActionDelete, true},
			{"invalid", false},
			{"", false},
		}

		for _, tt := range tests {
			if got := IsValidAction(tt.action); got != tt.valid {
				t.Errorf("IsValidAction(%q) = %v, want %v", tt.action, got, tt.valid)
			}
		}
	})

	t.Run("IsProtectedDomain", func(t *testing.T) {
		tests := []struct {
			domain    string
			protected bool
		}{
			{"user", true},
			{"rbac", true},
			{"item", false},
			{"document", false},
		}

		for _, tt := range tests {
			if got := IsProtectedDomain(tt.domain); got != tt.protected {
				t.Errorf("IsProtectedDomain(%q) = %v, want %v", tt.domain, got, tt.protected)
			}
		}
	})

	t.Run("IsSystemRole", func(t *testing.T) {
		tests := []struct {
			role     string
			isSystem bool
		}{
			{RoleCodeAdmin, true},
			{RoleCodeFullReader, true},
			{RoleCodeFullWriter, true},
			{"custom_role", false},
			{"", false},
		}

		for _, tt := range tests {
			if got := IsSystemRole(tt.role); got != tt.isSystem {
				t.Errorf("IsSystemRole(%q) = %v, want %v", tt.role, got, tt.isSystem)
			}
		}
	})
}

func TestFormatUserSubject(t *testing.T) {
	tests := []struct {
		userID   uint
		expected string
	}{
		{1, "user:1"},
		{100, "user:100"},
		{0, "user:0"},
	}

	for _, tt := range tests {
		got := FormatUserSubject(tt.userID)
		if got != tt.expected {
			t.Errorf("FormatUserSubject(%d) = %s, want %s", tt.userID, got, tt.expected)
		}
	}
}

func TestParseUserSubject(t *testing.T) {
	tests := []struct {
		subject  string
		expected uint
		ok       bool
	}{
		{"user:1", 1, true},
		{"user:100", 100, true},
		{"user:0", 0, true},
		{"invalid", 0, false},
		{"user:", 0, false},
		{"", 0, false},
	}

	for _, tt := range tests {
		got, ok := ParseUserSubject(tt.subject)
		if ok != tt.ok {
			t.Errorf("ParseUserSubject(%q) ok = %v, want %v", tt.subject, ok, tt.ok)
		}
		if ok && got != tt.expected {
			t.Errorf("ParseUserSubject(%q) = %d, want %d", tt.subject, got, tt.expected)
		}
	}
}

func TestModel_ToResponse(t *testing.T) {
	role := &Role{
		Code:        "test_role",
		Name:        "Test Role",
		Description: "A test role",
		IsSystem:    false,
	}
	role.ID = 1
	role.CreatedAt = time.Now()
	role.UpdatedAt = time.Now()

	response := role.ToResponse()

	if response.ID != role.ID {
		t.Errorf("ID = %d, want %d", response.ID, role.ID)
	}
	if response.Code != role.Code {
		t.Errorf("Code = %s, want %s", response.Code, role.Code)
	}
	if response.Name != role.Name {
		t.Errorf("Name = %s, want %s", response.Name, role.Name)
	}
	if response.Description != role.Description {
		t.Errorf("Description = %s, want %s", response.Description, role.Description)
	}
	if response.IsSystem != role.IsSystem {
		t.Errorf("IsSystem = %v, want %v", response.IsSystem, role.IsSystem)
	}
}

func TestErrors(t *testing.T) {
	t.Run("ErrRoleNotFound is not found error", func(t *testing.T) {
		if !commonerrors.IsNotFound(ErrRoleNotFound) {
			t.Error("ErrRoleNotFound should be a not found error")
		}
	})

	t.Run("ErrRoleCodeExists is already exists error", func(t *testing.T) {
		if !commonerrors.IsAlreadyExists(ErrRoleCodeExists) {
			t.Error("ErrRoleCodeExists should be an already exists error")
		}
	})

	t.Run("ErrSystemRoleCannotBeDeleted is forbidden error", func(t *testing.T) {
		if !commonerrors.IsForbidden(ErrSystemRoleCannotBeDeleted) {
			t.Error("ErrSystemRoleCannotBeDeleted should be a forbidden error")
		}
	})

	t.Run("ErrCannotRemoveLastAdmin is forbidden error", func(t *testing.T) {
		if !commonerrors.IsForbidden(ErrCannotRemoveLastAdmin) {
			t.Error("ErrCannotRemoveLastAdmin should be a forbidden error")
		}
	})
}

func TestDomainProvider(t *testing.T) {
	provider := newMockDomainProvider()

	domains := provider.GetDomainNames()
	if len(domains) == 0 {
		t.Error("GetDomainNames() returned empty slice")
	}

	// Check that expected domains are present
	domainMap := make(map[string]bool)
	for _, d := range domains {
		domainMap[d] = true
	}

	expected := []string{"user", "rbac", "item", "country", "city", "document"}
	for _, e := range expected {
		if !domainMap[e] {
			t.Errorf("Expected domain %s not found", e)
		}
	}
}

func TestRepository_FindByCode(t *testing.T) {
	repo := newMockRepository()
	ctx := context.Background()

	// Create a role
	role := &Role{
		Code:        "test_role",
		Name:        "Test Role",
		Description: "Test",
		IsSystem:    false,
	}
	repo.Create(ctx, role)

	t.Run("find existing role", func(t *testing.T) {
		found, err := repo.FindByCode(ctx, "test_role")
		if err != nil {
			t.Fatalf("FindByCode() error = %v", err)
		}
		if found.Code != "test_role" {
			t.Errorf("Code = %s, want test_role", found.Code)
		}
	})

	t.Run("find non-existent role", func(t *testing.T) {
		_, err := repo.FindByCode(ctx, "non_existent")
		if !commonerrors.IsNotFound(err) {
			t.Errorf("FindByCode() error = %v, want not found error", err)
		}
	})
}

func TestRepository_ExistsByCode(t *testing.T) {
	repo := newMockRepository()
	ctx := context.Background()

	repo.Create(ctx, &Role{
		Code:     "existing_role",
		Name:     "Existing",
		IsSystem: false,
	})

	t.Run("existing role", func(t *testing.T) {
		exists, err := repo.ExistsByCode(ctx, "existing_role")
		if err != nil {
			t.Fatalf("ExistsByCode() error = %v", err)
		}
		if !exists {
			t.Error("ExistsByCode() = false, want true")
		}
	})

	t.Run("non-existent role", func(t *testing.T) {
		exists, err := repo.ExistsByCode(ctx, "non_existent")
		if err != nil {
			t.Fatalf("ExistsByCode() error = %v", err)
		}
		if exists {
			t.Error("ExistsByCode() = true, want false")
		}
	})
}

func TestRepository_FindSystemRoles(t *testing.T) {
	repo := newMockRepository()
	ctx := context.Background()

	// Create system and non-system roles
	repo.Create(ctx, &Role{Code: "admin", Name: "Admin", IsSystem: true})
	repo.Create(ctx, &Role{Code: "full_reader", Name: "Full Reader", IsSystem: true})
	repo.Create(ctx, &Role{Code: "custom", Name: "Custom", IsSystem: false})

	systemRoles, err := repo.FindSystemRoles(ctx)
	if err != nil {
		t.Fatalf("FindSystemRoles() error = %v", err)
	}

	if len(systemRoles) != 2 {
		t.Errorf("FindSystemRoles() returned %d roles, want 2", len(systemRoles))
	}

	for _, r := range systemRoles {
		if !r.IsSystem {
			t.Errorf("Found non-system role: %s", r.Code)
		}
	}
}

func TestProtectedDomains(t *testing.T) {
	protected := ProtectedDomains()

	if len(protected) != 2 {
		t.Errorf("ProtectedDomains() returned %d domains, want 2", len(protected))
	}

	// Check that user and rbac are protected
	protectedMap := make(map[string]bool)
	for _, d := range protected {
		protectedMap[d] = true
	}

	if !protectedMap["user"] {
		t.Error("user domain should be protected")
	}
	if !protectedMap["rbac"] {
		t.Error("rbac domain should be protected")
	}
}

func TestSystemRoles(t *testing.T) {
	systemRoles := SystemRoles()

	if len(systemRoles) != 4 {
		t.Errorf("SystemRoles() returned %d roles, want 4", len(systemRoles))
	}

	roleMap := make(map[string]bool)
	for _, r := range systemRoles {
		roleMap[r] = true
	}

	expected := []string{RoleCodeAdmin, RoleCodeFullReader, RoleCodeFullWriter, RoleCodeUser}
	for _, e := range expected {
		if !roleMap[e] {
			t.Errorf("Expected system role %s not found", e)
		}
	}
}

func TestDTO_Validation(t *testing.T) {
	t.Run("CreateRoleRequest validation constraints", func(t *testing.T) {
		// This tests the struct definitions exist
		req := CreateRoleRequest{
			Code:        "test",
			Name:        "Test",
			Description: "Description",
			Permissions: []PermissionInput{
				{Domain: "item", Actions: []string{"read"}},
			},
		}

		if req.Code == "" {
			t.Error("Code should not be empty")
		}
		if len(req.Permissions) == 0 {
			t.Error("Permissions should not be empty")
		}
	})

	t.Run("AssignRoleRequest validation constraints", func(t *testing.T) {
		req := AssignRoleRequest{
			RoleCode: "admin",
		}

		if req.RoleCode == "" {
			t.Error("RoleCode should not be empty")
		}
	})
}

func TestRoleWithPermissions(t *testing.T) {
	role := &Role{
		Code:     "test",
		Name:     "Test",
		IsSystem: false,
	}
	role.ID = 1

	permissions := []Permission{
		{Domain: "item", Actions: []string{"read", "write"}},
		{Domain: "document", Actions: []string{"read"}},
	}

	rwp := &RoleWithPermissions{
		Role:        role,
		Permissions: permissions,
	}

	if rwp.Role.Code != "test" {
		t.Errorf("Role.Code = %s, want test", rwp.Role.Code)
	}
	if len(rwp.Permissions) != 2 {
		t.Errorf("len(Permissions) = %d, want 2", len(rwp.Permissions))
	}
}

func TestUserRole(t *testing.T) {
	ur := UserRole{
		UserID:   1,
		RoleCode: "admin",
		RoleName: "Administrator",
	}

	if ur.UserID != 1 {
		t.Errorf("UserID = %d, want 1", ur.UserID)
	}
	if ur.RoleCode != "admin" {
		t.Errorf("RoleCode = %s, want admin", ur.RoleCode)
	}
}

func TestRoleResponse(t *testing.T) {
	now := time.Now()
	resp := RoleResponse{
		ID:          1,
		Code:        "test",
		Name:        "Test Role",
		Description: "A test role",
		IsSystem:    false,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if resp.ID != 1 {
		t.Errorf("ID = %d, want 1", resp.ID)
	}
	if resp.Code != "test" {
		t.Errorf("Code = %s, want test", resp.Code)
	}
	if resp.IsSystem {
		t.Error("IsSystem should be false")
	}
}

func TestRoleWithPermissionsResponse(t *testing.T) {
	resp := RoleWithPermissionsResponse{
		ID:          1,
		Code:        "test",
		Name:        "Test",
		Description: "Test role",
		IsSystem:    false,
		Permissions: []PermissionResponse{
			{Domain: "item", Actions: []string{"read", "write"}},
		},
	}

	if len(resp.Permissions) != 1 {
		t.Errorf("len(Permissions) = %d, want 1", len(resp.Permissions))
	}
	if resp.Permissions[0].Domain != "item" {
		t.Errorf("Permissions[0].Domain = %s, want item", resp.Permissions[0].Domain)
	}
}

func TestUserRolesResponse(t *testing.T) {
	resp := UserRolesResponse{
		UserID: 1,
		Roles: []UserRoleResponse{
			{Code: "admin", Name: "Administrator"},
			{Code: "full_reader", Name: "Full Reader"},
		},
	}

	if resp.UserID != 1 {
		t.Errorf("UserID = %d, want 1", resp.UserID)
	}
	if len(resp.Roles) != 2 {
		t.Errorf("len(Roles) = %d, want 2", len(resp.Roles))
	}
}

func TestDomainResponse(t *testing.T) {
	resp := DomainResponse{
		Name:        "item",
		IsProtected: false,
	}

	if resp.Name != "item" {
		t.Errorf("Name = %s, want item", resp.Name)
	}
	if resp.IsProtected {
		t.Error("IsProtected should be false")
	}
}

func TestActionsResponse(t *testing.T) {
	resp := ActionsResponse{
		Actions: AllActions(),
	}

	if len(resp.Actions) != 4 {
		t.Errorf("len(Actions) = %d, want 4", len(resp.Actions))
	}
}

func TestFilterConfig(t *testing.T) {
	role := Role{}
	config := role.FilterConfig()

	if config.TableName != "rbac_roles" {
		t.Errorf("TableName = %s, want rbac_roles", config.TableName)
	}

	expectedFields := []string{"id", "code", "name", "description", "is_system", "created_at", "updated_at"}
	for _, field := range expectedFields {
		if _, ok := config.Fields[field]; !ok {
			t.Errorf("Expected field %s not found in FilterConfig", field)
		}
	}
}

func TestTableName(t *testing.T) {
	role := Role{}
	if role.TableName() != "rbac_roles" {
		t.Errorf("TableName() = %s, want rbac_roles", role.TableName())
	}
}

func TestMockRepository_CRUD(t *testing.T) {
	repo := newMockRepository()
	ctx := context.Background()

	// Create
	role := &Role{
		Code:     "test",
		Name:     "Test",
		IsSystem: false,
	}
	err := repo.Create(ctx, role)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if role.ID == 0 {
		t.Error("ID should be set after create")
	}

	// Read
	found, err := repo.FindByID(ctx, role.ID)
	if err != nil {
		t.Fatalf("FindByID() error = %v", err)
	}
	if found.Code != "test" {
		t.Errorf("Code = %s, want test", found.Code)
	}

	// Update
	found.Name = "Updated"
	err = repo.Update(ctx, found)
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	// Verify update
	found, _ = repo.FindByID(ctx, role.ID)
	if found.Name != "Updated" {
		t.Errorf("Name = %s, want Updated", found.Name)
	}

	// Delete
	err = repo.Delete(ctx, role.ID)
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	// Verify delete
	_, err = repo.FindByID(ctx, role.ID)
	if !commonerrors.IsNotFound(err) {
		t.Error("Role should be deleted")
	}
}

func TestMockRepository_Errors(t *testing.T) {
	repo := newMockRepository()
	ctx := context.Background()

	t.Run("create error", func(t *testing.T) {
		repo.createErr = errors.New("create failed")
		err := repo.Create(ctx, &Role{Code: "test"})
		if err == nil {
			t.Error("Expected error")
		}
		repo.createErr = nil
	})

	t.Run("update error", func(t *testing.T) {
		repo.Create(ctx, &Role{Code: "test"})
		repo.updateErr = errors.New("update failed")
		err := repo.Update(ctx, &Role{Code: "test"})
		if err == nil {
			t.Error("Expected error")
		}
		repo.updateErr = nil
	})

	t.Run("delete error", func(t *testing.T) {
		repo.deleteErr = errors.New("delete failed")
		err := repo.Delete(ctx, 1)
		if err == nil {
			t.Error("Expected error")
		}
		repo.deleteErr = nil
	})
}
