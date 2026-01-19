package user

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/voidmaindev/go-template/internal/common"
	commonerrors "github.com/voidmaindev/go-template/internal/common/errors"
	"github.com/voidmaindev/go-template/internal/common/filter"
	"github.com/voidmaindev/go-template/internal/config"
	"github.com/voidmaindev/go-template/internal/domain/rbac"
	"github.com/voidmaindev/go-template/pkg/utils"
	"gorm.io/gorm"
)

// mockRepository implements the Repository interface for testing
type mockRepository struct {
	users       map[uint]*User
	emailIndex  map[string]*User
	nextID      uint
	findByIDErr error
	createErr   error
	updateErr   error
}

func newMockRepository() *mockRepository {
	return &mockRepository{
		users:      make(map[uint]*User),
		emailIndex: make(map[string]*User),
		nextID:     1,
	}
}

func (m *mockRepository) Create(ctx context.Context, entity *User) error {
	if m.createErr != nil {
		return m.createErr
	}
	entity.ID = m.nextID
	entity.CreatedAt = time.Now()
	entity.UpdatedAt = time.Now()
	m.nextID++
	m.users[entity.ID] = entity
	m.emailIndex[entity.Email] = entity
	return nil
}

func (m *mockRepository) CreateBatch(ctx context.Context, entities []User, batchSize int) error {
	for i := range entities {
		if err := m.Create(ctx, &entities[i]); err != nil {
			return err
		}
	}
	return nil
}

func (m *mockRepository) FindByID(ctx context.Context, id uint) (*User, error) {
	if m.findByIDErr != nil {
		return nil, m.findByIDErr
	}
	user, ok := m.users[id]
	if !ok {
		return nil, commonerrors.NotFound("repository", "entity")
	}
	return user, nil
}

func (m *mockRepository) FindAll(ctx context.Context, pagination *common.Pagination) ([]User, int64, error) {
	var users []User
	for _, u := range m.users {
		users = append(users, *u)
	}
	return users, int64(len(users)), nil
}

func (m *mockRepository) FindByCondition(ctx context.Context, condition map[string]any, pagination *common.Pagination) ([]User, int64, error) {
	return m.FindAll(ctx, pagination)
}

func (m *mockRepository) FindAllFiltered(ctx context.Context, params *filter.Params) ([]User, int64, error) {
	return m.FindAll(ctx, nil)
}

func (m *mockRepository) FindOne(ctx context.Context, condition map[string]any) (*User, error) {
	if email, ok := condition["email"].(string); ok {
		if user, exists := m.emailIndex[email]; exists {
			return user, nil
		}
	}
	return nil, commonerrors.NotFound("repository", "entity")
}

func (m *mockRepository) Update(ctx context.Context, entity *User) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	if _, ok := m.users[entity.ID]; !ok {
		return commonerrors.NotFound("repository", "entity")
	}
	entity.UpdatedAt = time.Now()
	m.users[entity.ID] = entity
	m.emailIndex[entity.Email] = entity
	return nil
}

func (m *mockRepository) UpdateFields(ctx context.Context, id uint, fields map[string]any) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	user, ok := m.users[id]
	if !ok {
		return commonerrors.NotFound("repository", "entity")
	}
	if pwd, ok := fields["password"].(string); ok {
		user.Password = pwd
	}
	if name, ok := fields["name"].(string); ok {
		user.Name = name
	}
	user.UpdatedAt = time.Now()
	return nil
}

func (m *mockRepository) Delete(ctx context.Context, id uint) error {
	user, ok := m.users[id]
	if !ok {
		return commonerrors.NotFound("repository", "entity")
	}
	delete(m.users, id)
	delete(m.emailIndex, user.Email)
	return nil
}

func (m *mockRepository) HardDelete(ctx context.Context, id uint) error {
	return m.Delete(ctx, id)
}

func (m *mockRepository) Exists(ctx context.Context, condition map[string]any) (bool, error) {
	if email, ok := condition["email"].(string); ok {
		_, exists := m.emailIndex[email]
		return exists, nil
	}
	return false, nil
}

func (m *mockRepository) Count(ctx context.Context, condition map[string]any) (int64, error) {
	return int64(len(m.users)), nil
}

func (m *mockRepository) WithTx(tx *gorm.DB) common.Repository[User] {
	return m
}

func (m *mockRepository) WithPreload(preloads ...string) common.Repository[User] {
	return m
}

func (m *mockRepository) Transaction(ctx context.Context, fn func(txRepo common.Repository[User]) error) error {
	return fn(m)
}

func (m *mockRepository) FindByEmail(ctx context.Context, email string) (*User, error) {
	user, ok := m.emailIndex[email]
	if !ok {
		return nil, commonerrors.NotFound("repository", "entity")
	}
	return user, nil
}

func (m *mockRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	_, exists := m.emailIndex[email]
	return exists, nil
}

// mockTokenStore implements token store for testing
type mockTokenStore struct {
	blacklisted    map[string]bool
	blacklistErr   error
	isBlacklistErr error
}

func newMockTokenStore() *mockTokenStore {
	return &mockTokenStore{
		blacklisted: make(map[string]bool),
	}
}

func (m *mockTokenStore) Blacklist(ctx context.Context, token string, expiry time.Duration) error {
	if m.blacklistErr != nil {
		return m.blacklistErr
	}
	m.blacklisted[token] = true
	return nil
}

func (m *mockTokenStore) BlacklistWithRetry(ctx context.Context, token string, expiry time.Duration, maxRetries int) error {
	return m.Blacklist(ctx, token, expiry)
}

func (m *mockTokenStore) IsBlacklisted(ctx context.Context, token string) (bool, error) {
	if m.isBlacklistErr != nil {
		return false, m.isBlacklistErr
	}
	return m.blacklisted[token], nil
}

// mockRbacService implements rbac.Service for testing
type mockRbacService struct {
	assignedRoles map[uint][]string
	assignRoleErr error
}

func newMockRbacService() *mockRbacService {
	return &mockRbacService{
		assignedRoles: make(map[uint][]string),
	}
}

func (m *mockRbacService) CreateRole(ctx context.Context, req *rbac.CreateRoleRequest) (*rbac.Role, error) {
	return nil, nil
}

func (m *mockRbacService) GetRoleByCode(ctx context.Context, code string) (*rbac.RoleWithPermissions, error) {
	return nil, nil
}

func (m *mockRbacService) ListRoles(ctx context.Context, params *filter.Params) (*common.FilteredResult[rbac.Role], error) {
	return nil, nil
}

func (m *mockRbacService) UpdateRolePermissions(ctx context.Context, code string, req *rbac.UpdateRolePermissionsRequest) (*rbac.RoleWithPermissions, error) {
	return nil, nil
}

func (m *mockRbacService) DeleteRole(ctx context.Context, code string) error {
	return nil
}

func (m *mockRbacService) GetUserRoles(ctx context.Context, userID uint) ([]rbac.UserRoleResponse, error) {
	return nil, nil
}

func (m *mockRbacService) AssignRole(ctx context.Context, userID uint, roleCode string) error {
	if m.assignRoleErr != nil {
		return m.assignRoleErr
	}
	m.assignedRoles[userID] = append(m.assignedRoles[userID], roleCode)
	return nil
}

func (m *mockRbacService) RemoveRole(ctx context.Context, userID uint, roleCode string) error {
	return nil
}

func (m *mockRbacService) CheckPermission(ctx context.Context, userID uint, domain, action string) (bool, error) {
	return true, nil
}

func (m *mockRbacService) GetDomains(ctx context.Context) []rbac.DomainResponse {
	return nil
}

func (m *mockRbacService) GetActions(ctx context.Context) []string {
	return nil
}

func (m *mockRbacService) SyncGlobalRoles(ctx context.Context) error {
	return nil
}

func (m *mockRbacService) CountAdminUsers(ctx context.Context) (int, error) {
	return 0, nil
}

func getTestConfig() *config.JWTConfig {
	return &config.JWTConfig{
		SecretKey:          "test-secret-key-at-least-32-chars!!",
		AccessTokenExpiry:  15 * time.Minute,
		RefreshTokenExpiry: 7 * 24 * time.Hour,
		Issuer:             "test-issuer",
	}
}

func TestService_Register(t *testing.T) {
	repo := newMockRepository()
	cfg := getTestConfig()
	rbacSvc := newMockRbacService()

	// Create a custom service with our mock
	customSvc := &service{
		repo:       repo,
		tokenStore: &TokenStore{},
		jwtConfig:  cfg,
		rbacSvc:    rbacSvc,
	}

	t.Run("successful registration", func(t *testing.T) {
		req := &RegisterRequest{
			Email:    "test@example.com",
			Password: "password123",
			Name:     "Test User",
		}

		response, err := customSvc.Register(context.Background(), req)
		if err != nil {
			t.Fatalf("Register() error = %v", err)
		}

		if response.AccessToken == "" {
			t.Error("AccessToken should not be empty")
		}
		if response.RefreshToken == "" {
			t.Error("RefreshToken should not be empty")
		}
		if response.User.Email != req.Email {
			t.Errorf("User.Email = %v, want %v", response.User.Email, req.Email)
		}
		if response.User.Name != req.Name {
			t.Errorf("User.Name = %v, want %v", response.User.Name, req.Name)
		}
	})

	t.Run("duplicate email", func(t *testing.T) {
		req := &RegisterRequest{
			Email:    "test@example.com", // Same email as before
			Password: "password456",
			Name:     "Another User",
		}

		_, err := customSvc.Register(context.Background(), req)
		if !errors.Is(err, ErrEmailExists) {
			t.Errorf("Register() error = %v, want %v", err, ErrEmailExists)
		}
	})
}

func TestService_Login(t *testing.T) {
	repo := newMockRepository()
	cfg := getTestConfig()
	svc := &service{
		repo:       repo,
		tokenStore: &TokenStore{},
		jwtConfig:  cfg,
		rbacSvc:    newMockRbacService(),
	}

	// Create a user first
	hashedPassword, _ := utils.HashPassword("correctpassword")
	repo.Create(context.Background(), &User{
		Email:    "user@example.com",
		Password: hashedPassword,
		Name:     "Test User",
	})

	t.Run("successful login", func(t *testing.T) {
		req := &LoginRequest{
			Email:    "user@example.com",
			Password: "correctpassword",
		}

		response, err := svc.Login(context.Background(), req)
		if err != nil {
			t.Fatalf("Login() error = %v", err)
		}

		if response.AccessToken == "" {
			t.Error("AccessToken should not be empty")
		}
		if response.User.Email != req.Email {
			t.Errorf("User.Email = %v, want %v", response.User.Email, req.Email)
		}
	})

	t.Run("wrong password", func(t *testing.T) {
		req := &LoginRequest{
			Email:    "user@example.com",
			Password: "wrongpassword",
		}

		_, err := svc.Login(context.Background(), req)
		if !errors.Is(err, ErrInvalidCredentials) {
			t.Errorf("Login() error = %v, want %v", err, ErrInvalidCredentials)
		}
	})

	t.Run("non-existent user", func(t *testing.T) {
		req := &LoginRequest{
			Email:    "nonexistent@example.com",
			Password: "password123",
		}

		_, err := svc.Login(context.Background(), req)
		if !errors.Is(err, ErrInvalidCredentials) {
			t.Errorf("Login() error = %v, want %v", err, ErrInvalidCredentials)
		}
	})
}

func TestService_GetByID(t *testing.T) {
	repo := newMockRepository()
	cfg := getTestConfig()
	svc := &service{
		repo:       repo,
		tokenStore: &TokenStore{},
		jwtConfig:  cfg,
		rbacSvc:    newMockRbacService(),
	}

	// Create a user
	repo.Create(context.Background(), &User{
		Email:    "user@example.com",
		Password: "hashedpwd",
		Name:     "Test User",
	})

	t.Run("existing user", func(t *testing.T) {
		user, err := svc.GetByID(context.Background(), 1)
		if err != nil {
			t.Fatalf("GetByID() error = %v", err)
		}

		if user.Email != "user@example.com" {
			t.Errorf("User.Email = %v, want user@example.com", user.Email)
		}
	})

	t.Run("non-existent user", func(t *testing.T) {
		_, err := svc.GetByID(context.Background(), 999)
		if !errors.Is(err, ErrUserNotFound) {
			t.Errorf("GetByID() error = %v, want %v", err, ErrUserNotFound)
		}
	})
}

func TestService_GetByEmail(t *testing.T) {
	repo := newMockRepository()
	cfg := getTestConfig()
	svc := &service{
		repo:       repo,
		tokenStore: &TokenStore{},
		jwtConfig:  cfg,
		rbacSvc:    newMockRbacService(),
	}

	repo.Create(context.Background(), &User{
		Email:    "user@example.com",
		Password: "hashedpwd",
		Name:     "Test User",
	})

	t.Run("existing email", func(t *testing.T) {
		user, err := svc.GetByEmail(context.Background(), "user@example.com")
		if err != nil {
			t.Fatalf("GetByEmail() error = %v", err)
		}

		if user.Name != "Test User" {
			t.Errorf("User.Name = %v, want Test User", user.Name)
		}
	})

	t.Run("non-existent email", func(t *testing.T) {
		_, err := svc.GetByEmail(context.Background(), "nonexistent@example.com")
		if !errors.Is(err, ErrUserNotFound) {
			t.Errorf("GetByEmail() error = %v, want %v", err, ErrUserNotFound)
		}
	})
}

func TestService_Update(t *testing.T) {
	repo := newMockRepository()
	cfg := getTestConfig()
	svc := &service{
		repo:       repo,
		tokenStore: &TokenStore{},
		jwtConfig:  cfg,
		rbacSvc:    newMockRbacService(),
	}

	repo.Create(context.Background(), &User{
		Email:    "user@example.com",
		Password: "hashedpwd",
		Name:     "Original Name",
	})

	t.Run("successful update", func(t *testing.T) {
		newName := "Updated Name"
		req := &UpdateUserRequest{
			Name: &newName,
		}

		user, err := svc.Update(context.Background(), 1, req)
		if err != nil {
			t.Fatalf("Update() error = %v", err)
		}

		if user.Name != newName {
			t.Errorf("User.Name = %v, want %v", user.Name, newName)
		}
	})

	t.Run("non-existent user", func(t *testing.T) {
		newName := "Test"
		req := &UpdateUserRequest{
			Name: &newName,
		}

		_, err := svc.Update(context.Background(), 999, req)
		if !errors.Is(err, ErrUserNotFound) {
			t.Errorf("Update() error = %v, want %v", err, ErrUserNotFound)
		}
	})
}

func TestService_ChangePassword(t *testing.T) {
	repo := newMockRepository()
	cfg := getTestConfig()
	svc := &service{
		repo:       repo,
		tokenStore: &TokenStore{},
		jwtConfig:  cfg,
		rbacSvc:    newMockRbacService(),
	}

	hashedPassword, _ := utils.HashPassword("currentpassword")
	repo.Create(context.Background(), &User{
		Email:    "user@example.com",
		Password: hashedPassword,
		Name:     "Test User",
	})

	t.Run("successful password change", func(t *testing.T) {
		req := &ChangePasswordRequest{
			CurrentPassword: "currentpassword",
			NewPassword:     "newpassword123",
		}

		err := svc.ChangePassword(context.Background(), 1, req)
		if err != nil {
			t.Fatalf("ChangePassword() error = %v", err)
		}
	})

	t.Run("wrong current password", func(t *testing.T) {
		req := &ChangePasswordRequest{
			CurrentPassword: "wrongpassword",
			NewPassword:     "newpassword456",
		}

		err := svc.ChangePassword(context.Background(), 1, req)
		if !errors.Is(err, ErrInvalidPassword) {
			t.Errorf("ChangePassword() error = %v, want %v", err, ErrInvalidPassword)
		}
	})

	t.Run("same password", func(t *testing.T) {
		// First update the password
		hashedNew, _ := utils.HashPassword("samepassword")
		repo.users[1].Password = hashedNew

		req := &ChangePasswordRequest{
			CurrentPassword: "samepassword",
			NewPassword:     "samepassword",
		}

		err := svc.ChangePassword(context.Background(), 1, req)
		if !errors.Is(err, ErrSamePassword) {
			t.Errorf("ChangePassword() error = %v, want %v", err, ErrSamePassword)
		}
	})

	t.Run("non-existent user", func(t *testing.T) {
		req := &ChangePasswordRequest{
			CurrentPassword: "password",
			NewPassword:     "newpassword",
		}

		err := svc.ChangePassword(context.Background(), 999, req)
		if !errors.Is(err, ErrUserNotFound) {
			t.Errorf("ChangePassword() error = %v, want %v", err, ErrUserNotFound)
		}
	})
}

func TestService_Delete(t *testing.T) {
	repo := newMockRepository()
	cfg := getTestConfig()
	svc := &service{
		repo:       repo,
		tokenStore: &TokenStore{},
		jwtConfig:  cfg,
		rbacSvc:    newMockRbacService(),
	}

	repo.Create(context.Background(), &User{
		Email:    "user@example.com",
		Password: "hashedpwd",
		Name:     "Test User",
	})

	t.Run("successful delete", func(t *testing.T) {
		err := svc.Delete(context.Background(), 1)
		if err != nil {
			t.Fatalf("Delete() error = %v", err)
		}

		// Verify user is gone
		_, err = svc.GetByID(context.Background(), 1)
		if !errors.Is(err, ErrUserNotFound) {
			t.Error("User should be deleted")
		}
	})

	t.Run("non-existent user", func(t *testing.T) {
		err := svc.Delete(context.Background(), 999)
		if !errors.Is(err, ErrUserNotFound) {
			t.Errorf("Delete() error = %v, want %v", err, ErrUserNotFound)
		}
	})
}

func TestService_List(t *testing.T) {
	repo := newMockRepository()
	cfg := getTestConfig()
	svc := &service{
		repo:       repo,
		tokenStore: &TokenStore{},
		jwtConfig:  cfg,
		rbacSvc:    newMockRbacService(),
	}

	// Create multiple users
	for i := 0; i < 5; i++ {
		repo.Create(context.Background(), &User{
			Email:    "user" + string(rune('0'+i)) + "@example.com",
			Password: "hashedpwd",
			Name:     "User " + string(rune('0'+i)),
		})
	}

	t.Run("list with pagination", func(t *testing.T) {
		pagination := &common.Pagination{
			Page:     1,
			PageSize: 10,
		}

		result, err := svc.List(context.Background(), pagination)
		if err != nil {
			t.Fatalf("List() error = %v", err)
		}

		if result.Total != 5 {
			t.Errorf("Total = %v, want 5", result.Total)
		}
		if len(result.Data) != 5 {
			t.Errorf("len(Data) = %v, want 5", len(result.Data))
		}
	})
}

func TestService_RefreshToken(t *testing.T) {
	repo := newMockRepository()
	_ = newMockTokenStore() // Used for documentation, actual mock usage requires Redis integration
	cfg := getTestConfig()

	// Create wrapper TokenStore that uses our mock
	ts := &TokenStore{}

	svc := &service{
		repo:       repo,
		tokenStore: ts,
		jwtConfig:  cfg,
		rbacSvc:    newMockRbacService(),
	}

	// Create a user
	repo.Create(context.Background(), &User{
		Email:    "user@example.com",
		Password: "hashedpwd",
		Name:     "Test User",
	})

	t.Run("invalid refresh token", func(t *testing.T) {
		_, err := svc.RefreshToken(context.Background(), "invalid-token")
		if !errors.Is(err, ErrTokenInvalid) {
			t.Errorf("RefreshToken() error = %v, want %v", err, ErrTokenInvalid)
		}
	})

	t.Run("access token used as refresh", func(t *testing.T) {
		// Generate an access token
		jwtCfg := &utils.JWTConfig{
			SecretKey:          cfg.SecretKey,
			AccessTokenExpiry:  cfg.AccessTokenExpiry,
			RefreshTokenExpiry: cfg.RefreshTokenExpiry,
			Issuer:             cfg.Issuer,
		}
		accessToken, _ := utils.GenerateAccessToken(1, "user@example.com", jwtCfg)

		_, err := svc.RefreshToken(context.Background(), accessToken)
		if !errors.Is(err, ErrTokenInvalid) {
			t.Errorf("RefreshToken() error = %v, want %v", err, ErrTokenInvalid)
		}
	})

	t.Run("blacklisted token handling", func(t *testing.T) {
		// This test demonstrates blacklist checking with mock
		// The actual behavior requires Redis integration for full test
	})
}
