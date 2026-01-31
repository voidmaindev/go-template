package user

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/voidmaindev/go-template/internal/common"
	"github.com/voidmaindev/go-template/internal/common/filter"
	"github.com/voidmaindev/go-template/internal/config"
	"github.com/voidmaindev/go-template/internal/middleware"
	"github.com/voidmaindev/go-template/pkg/utils"
)

// mockService implements Service interface for testing
type mockService struct {
	registerResponse       *TokenResponse
	registerErr            error
	loginResponse          *TokenResponse
	loginErr               error
	logoutErr              error
	refreshResponse        *TokenResponse
	refreshErr             error
	getByIDResponse        *User
	getByIDErr             error
	getByEmailResponse     *User
	getByEmailErr          error
	updateResponse         *User
	updateErr              error
	changePasswordErr      error
	deleteErr              error
	listResponse           *common.PaginatedResult[User]
	listErr                error
	listFilteredResponse   *common.FilteredResult[User]
	listFilteredErr        error
}

func (m *mockService) Register(ctx context.Context, req *RegisterRequest) (*TokenResponse, error) {
	if m.registerErr != nil {
		return nil, m.registerErr
	}
	return m.registerResponse, nil
}

func (m *mockService) Login(ctx context.Context, req *LoginRequest, loginCtx *LoginContext) (*TokenResponse, error) {
	if m.loginErr != nil {
		return nil, m.loginErr
	}
	return m.loginResponse, nil
}

func (m *mockService) Logout(ctx context.Context, accessToken, refreshToken string) error {
	return m.logoutErr
}

func (m *mockService) RefreshToken(ctx context.Context, refreshToken string) (*TokenResponse, error) {
	if m.refreshErr != nil {
		return nil, m.refreshErr
	}
	return m.refreshResponse, nil
}

func (m *mockService) GetByID(ctx context.Context, id uint) (*User, error) {
	if m.getByIDErr != nil {
		return nil, m.getByIDErr
	}
	return m.getByIDResponse, nil
}

func (m *mockService) GetByEmail(ctx context.Context, email string) (*User, error) {
	if m.getByEmailErr != nil {
		return nil, m.getByEmailErr
	}
	return m.getByEmailResponse, nil
}

func (m *mockService) Update(ctx context.Context, id uint, req *UpdateUserRequest) (*User, error) {
	if m.updateErr != nil {
		return nil, m.updateErr
	}
	return m.updateResponse, nil
}

func (m *mockService) ChangePassword(ctx context.Context, id uint, req *ChangePasswordRequest) error {
	return m.changePasswordErr
}

func (m *mockService) Delete(ctx context.Context, id uint) error {
	return m.deleteErr
}

func (m *mockService) List(ctx context.Context, pagination *common.Pagination) (*common.PaginatedResult[User], error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	return m.listResponse, nil
}

func (m *mockService) ListFiltered(ctx context.Context, params *filter.Params) (*common.FilteredResult[User], error) {
	if m.listFilteredErr != nil {
		return nil, m.listFilteredErr
	}
	return m.listFilteredResponse, nil
}

// Helper functions for tests
func createTestApp(handler *Handler) *fiber.App {
	app := fiber.New()
	return app
}

func getHandlerTestConfig() *config.JWTConfig {
	return &config.JWTConfig{
		SecretKey:               "test-secret-key-at-least-32-chars!!",
		AccessTokenExpiry:       15 * time.Minute,
		RefreshTokenExpiry:      7 * 24 * time.Hour,
		Issuer:                  "test-issuer",
		MinPasswordResponseTime: 200 * time.Millisecond,
	}
}

func generateHandlerTestToken(userID uint, email string) string {
	cfg := getHandlerTestConfig()
	jwtCfg := &utils.JWTConfig{
		SecretKey:          cfg.SecretKey,
		AccessTokenExpiry:  cfg.AccessTokenExpiry,
		RefreshTokenExpiry: cfg.RefreshTokenExpiry,
		Issuer:             cfg.Issuer,
	}
	token, _ := utils.GenerateAccessToken(userID, email, jwtCfg)
	return token
}


func createTestUser() *User {
	return &User{
		BaseModel: common.BaseModel{
			ID:        1,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		Email: "test@example.com",
		Name:  "Test User",
	}
}

func createTestTokenResponse() *TokenResponse {
	user := createTestUser()
	return &TokenResponse{
		AccessToken:  "access-token",
		RefreshToken: "refresh-token",
		ExpiresAt:    time.Now().Add(15 * time.Minute).Unix(),
		User:         user.ToResponse(),
	}
}

func TestHandler_Register(t *testing.T) {
	t.Run("successful registration", func(t *testing.T) {
		svc := &mockService{
			registerResponse: createTestTokenResponse(),
		}
		handler := NewHandler(svc, getHandlerTestConfig())

		app := fiber.New()
		app.Post("/register", handler.Register)

		reqBody := RegisterRequest{
			Email:    "new@example.com",
			Password: "Password123!",
			Name:     "New User",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("Test request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusCreated {
			bodyBytes, _ := io.ReadAll(resp.Body)
			t.Errorf("Expected status 201, got %d: %s", resp.StatusCode, string(bodyBytes))
		}
	})

	t.Run("invalid request body", func(t *testing.T) {
		svc := &mockService{}
		handler := NewHandler(svc, getHandlerTestConfig())

		app := fiber.New()
		app.Post("/register", handler.Register)

		req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("Test request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", resp.StatusCode)
		}
	})

	t.Run("validation error - missing fields", func(t *testing.T) {
		svc := &mockService{}
		handler := NewHandler(svc, getHandlerTestConfig())

		app := fiber.New()
		app.Post("/register", handler.Register)

		reqBody := map[string]string{
			"email": "invalid-email",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("Test request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", resp.StatusCode)
		}
	})

	t.Run("email already exists", func(t *testing.T) {
		svc := &mockService{
			registerErr: ErrEmailExists,
		}
		handler := NewHandler(svc, getHandlerTestConfig())

		app := fiber.New()
		app.Post("/register", handler.Register)

		reqBody := RegisterRequest{
			Email:    "existing@example.com",
			Password: "Password123!",
			Name:     "User",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("Test request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusConflict {
			t.Errorf("Expected status 409, got %d", resp.StatusCode)
		}
	})

	t.Run("internal server error", func(t *testing.T) {
		svc := &mockService{
			registerErr: errors.New("database error"),
		}
		handler := NewHandler(svc, getHandlerTestConfig())

		app := fiber.New()
		app.Post("/register", handler.Register)

		reqBody := RegisterRequest{
			Email:    "test@example.com",
			Password: "Password123!",
			Name:     "User",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("Test request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusInternalServerError {
			t.Errorf("Expected status 500, got %d", resp.StatusCode)
		}
	})
}

func TestHandler_Login(t *testing.T) {
	t.Run("successful login", func(t *testing.T) {
		svc := &mockService{
			loginResponse: createTestTokenResponse(),
		}
		handler := NewHandler(svc, getHandlerTestConfig())

		app := fiber.New()
		app.Post("/login", handler.Login)

		reqBody := LoginRequest{
			Email:    "test@example.com",
			Password: "password123",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("Test request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			bodyBytes, _ := io.ReadAll(resp.Body)
			t.Errorf("Expected status 200, got %d: %s", resp.StatusCode, string(bodyBytes))
		}
	})

	t.Run("invalid credentials", func(t *testing.T) {
		svc := &mockService{
			loginErr: ErrInvalidCredentials,
		}
		handler := NewHandler(svc, getHandlerTestConfig())

		app := fiber.New()
		app.Post("/login", handler.Login)

		reqBody := LoginRequest{
			Email:    "test@example.com",
			Password: "wrongpassword",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("Test request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("Expected status 401, got %d", resp.StatusCode)
		}
	})

	t.Run("invalid request body", func(t *testing.T) {
		svc := &mockService{}
		handler := NewHandler(svc, getHandlerTestConfig())

		app := fiber.New()
		app.Post("/login", handler.Login)

		req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader([]byte("invalid")))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("Test request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", resp.StatusCode)
		}
	})
}

func TestHandler_Logout(t *testing.T) {
	cfg := getHandlerTestConfig()

	t.Run("successful logout", func(t *testing.T) {
		svc := &mockService{}
		handler := NewHandler(svc, getHandlerTestConfig())

		app := fiber.New()
		app.Use(middleware.JWTMiddleware(cfg, nil))
		app.Post("/logout", handler.Logout)

		token := generateHandlerTestToken(1, "test@example.com")

		req := httptest.NewRequest(http.MethodPost, "/logout", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("Test request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			bodyBytes, _ := io.ReadAll(resp.Body)
			t.Errorf("Expected status 200, got %d: %s", resp.StatusCode, string(bodyBytes))
		}
	})

	t.Run("no token provided", func(t *testing.T) {
		svc := &mockService{}
		handler := NewHandler(svc, getHandlerTestConfig())

		app := fiber.New()
		// No JWT middleware - simulating missing token
		app.Post("/logout", handler.Logout)

		req := httptest.NewRequest(http.MethodPost, "/logout", nil)

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("Test request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("Expected status 401, got %d", resp.StatusCode)
		}
	})

	t.Run("logout error", func(t *testing.T) {
		svc := &mockService{
			logoutErr: errors.New("redis error"),
		}
		handler := NewHandler(svc, getHandlerTestConfig())

		app := fiber.New()
		app.Use(middleware.JWTMiddleware(cfg, nil))
		app.Post("/logout", handler.Logout)

		token := generateHandlerTestToken(1, "test@example.com")

		req := httptest.NewRequest(http.MethodPost, "/logout", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("Test request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusInternalServerError {
			t.Errorf("Expected status 500, got %d", resp.StatusCode)
		}
	})
}

func TestHandler_RefreshToken(t *testing.T) {
	t.Run("successful refresh", func(t *testing.T) {
		svc := &mockService{
			refreshResponse: createTestTokenResponse(),
		}
		handler := NewHandler(svc, getHandlerTestConfig())

		app := fiber.New()
		app.Post("/refresh", handler.RefreshToken)

		reqBody := RefreshTokenRequest{
			RefreshToken: "valid-refresh-token",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/refresh", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("Test request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			bodyBytes, _ := io.ReadAll(resp.Body)
			t.Errorf("Expected status 200, got %d: %s", resp.StatusCode, string(bodyBytes))
		}
	})

	t.Run("invalid refresh token", func(t *testing.T) {
		svc := &mockService{
			refreshErr: ErrTokenInvalid,
		}
		handler := NewHandler(svc, getHandlerTestConfig())

		app := fiber.New()
		app.Post("/refresh", handler.RefreshToken)

		reqBody := RefreshTokenRequest{
			RefreshToken: "invalid-token",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/refresh", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("Test request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("Expected status 401, got %d", resp.StatusCode)
		}
	})

	t.Run("blacklisted token", func(t *testing.T) {
		svc := &mockService{
			refreshErr: ErrTokenBlacklisted,
		}
		handler := NewHandler(svc, getHandlerTestConfig())

		app := fiber.New()
		app.Post("/refresh", handler.RefreshToken)

		reqBody := RefreshTokenRequest{
			RefreshToken: "blacklisted-token",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/refresh", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

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

func TestHandler_GetMe(t *testing.T) {
	cfg := getHandlerTestConfig()

	t.Run("successful get", func(t *testing.T) {
		svc := &mockService{
			getByIDResponse: createTestUser(),
		}
		handler := NewHandler(svc, getHandlerTestConfig())

		app := fiber.New()
		app.Use(middleware.JWTMiddleware(cfg, nil))
		app.Get("/me", handler.GetMe)

		token := generateHandlerTestToken(1, "test@example.com")

		req := httptest.NewRequest(http.MethodGet, "/me", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("Test request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			bodyBytes, _ := io.ReadAll(resp.Body)
			t.Errorf("Expected status 200, got %d: %s", resp.StatusCode, string(bodyBytes))
		}
	})

	t.Run("user not found", func(t *testing.T) {
		svc := &mockService{
			getByIDErr: ErrUserNotFound,
		}
		handler := NewHandler(svc, getHandlerTestConfig())

		app := fiber.New()
		app.Use(middleware.JWTMiddleware(cfg, nil))
		app.Get("/me", handler.GetMe)

		token := generateHandlerTestToken(999, "deleted@example.com")

		req := httptest.NewRequest(http.MethodGet, "/me", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("Test request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("Expected status 404, got %d", resp.StatusCode)
		}
	})

	t.Run("unauthorized - no token", func(t *testing.T) {
		svc := &mockService{}
		handler := NewHandler(svc, getHandlerTestConfig())

		app := fiber.New()
		app.Use(middleware.JWTMiddleware(cfg, nil))
		app.Get("/me", handler.GetMe)

		req := httptest.NewRequest(http.MethodGet, "/me", nil)

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

func TestHandler_ChangePassword(t *testing.T) {
	cfg := getHandlerTestConfig()

	t.Run("successful change", func(t *testing.T) {
		svc := &mockService{}
		handler := NewHandler(svc, getHandlerTestConfig())

		app := fiber.New()
		app.Use(middleware.JWTMiddleware(cfg, nil))
		app.Put("/password", handler.ChangePassword)

		token := generateHandlerTestToken(1, "test@example.com")

		reqBody := ChangePasswordRequest{
			CurrentPassword: "oldpassword",
			NewPassword:     "NewPassword123!",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPut, "/password", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("Test request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			bodyBytes, _ := io.ReadAll(resp.Body)
			t.Errorf("Expected status 200, got %d: %s", resp.StatusCode, string(bodyBytes))
		}
	})

	t.Run("invalid password", func(t *testing.T) {
		svc := &mockService{
			changePasswordErr: ErrInvalidPassword,
		}
		handler := NewHandler(svc, getHandlerTestConfig())

		app := fiber.New()
		app.Use(middleware.JWTMiddleware(cfg, nil))
		app.Put("/password", handler.ChangePassword)

		token := generateHandlerTestToken(1, "test@example.com")

		reqBody := ChangePasswordRequest{
			CurrentPassword: "wrongpassword",
			NewPassword:     "NewPassword123!",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPut, "/password", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("Test request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", resp.StatusCode)
		}
	})

	t.Run("same password", func(t *testing.T) {
		svc := &mockService{
			changePasswordErr: ErrSamePassword,
		}
		handler := NewHandler(svc, getHandlerTestConfig())

		app := fiber.New()
		app.Use(middleware.JWTMiddleware(cfg, nil))
		app.Put("/password", handler.ChangePassword)

		token := generateHandlerTestToken(1, "test@example.com")

		reqBody := ChangePasswordRequest{
			CurrentPassword: "samepassword",
			NewPassword:     "SamePassword123!",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPut, "/password", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("Test request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", resp.StatusCode)
		}
	})
}

func TestHandler_GetByID(t *testing.T) {
	cfg := getHandlerTestConfig()

	t.Run("successful get", func(t *testing.T) {
		svc := &mockService{
			getByIDResponse: createTestUser(),
		}
		handler := NewHandler(svc, getHandlerTestConfig())

		app := fiber.New()
		app.Use(middleware.JWTMiddleware(cfg, nil))
		app.Get("/users/:id", handler.GetByID)

		token := generateHandlerTestToken(1, "test@example.com")

		req := httptest.NewRequest(http.MethodGet, "/users/1", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("Test request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			bodyBytes, _ := io.ReadAll(resp.Body)
			t.Errorf("Expected status 200, got %d: %s", resp.StatusCode, string(bodyBytes))
		}
	})

	t.Run("user not found", func(t *testing.T) {
		svc := &mockService{
			getByIDErr: ErrUserNotFound,
		}
		handler := NewHandler(svc, getHandlerTestConfig())

		app := fiber.New()
		app.Use(middleware.JWTMiddleware(cfg, nil))
		app.Get("/users/:id", handler.GetByID)

		// Authorization is handled by RBAC middleware at route level
		token := generateHandlerTestToken(1, "test@example.com")

		req := httptest.NewRequest(http.MethodGet, "/users/999", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("Test request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("Expected status 404, got %d", resp.StatusCode)
		}
	})

	t.Run("invalid user ID", func(t *testing.T) {
		svc := &mockService{}
		handler := NewHandler(svc, getHandlerTestConfig())

		app := fiber.New()
		app.Use(middleware.JWTMiddleware(cfg, nil))
		app.Get("/users/:id", handler.GetByID)

		token := generateHandlerTestToken(1, "test@example.com")

		req := httptest.NewRequest(http.MethodGet, "/users/invalid", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("Test request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", resp.StatusCode)
		}
	})
}

func TestHandler_List(t *testing.T) {
	cfg := getHandlerTestConfig()

	t.Run("successful list", func(t *testing.T) {
		users := []User{*createTestUser()}
		svc := &mockService{
			listFilteredResponse: &common.FilteredResult[User]{
				Data:       users,
				Total:      1,
				Page:       1,
				PageSize:   10,
				TotalPages: 1,
			},
		}
		handler := NewHandler(svc, getHandlerTestConfig())

		app := fiber.New()
		app.Use(middleware.JWTMiddleware(cfg, nil))
		app.Get("/users", handler.List)

		// Authorization is handled by RBAC middleware at route level
		token := generateHandlerTestToken(1, "test@example.com")

		req := httptest.NewRequest(http.MethodGet, "/users?page=1&page_size=10", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("Test request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			bodyBytes, _ := io.ReadAll(resp.Body)
			t.Errorf("Expected status 200, got %d: %s", resp.StatusCode, string(bodyBytes))
		}
	})

	t.Run("list with default pagination", func(t *testing.T) {
		svc := &mockService{
			listFilteredResponse: &common.FilteredResult[User]{
				Data:       []User{},
				Total:      0,
				Page:       1,
				PageSize:   10,
				TotalPages: 0,
			},
		}
		handler := NewHandler(svc, getHandlerTestConfig())

		app := fiber.New()
		app.Use(middleware.JWTMiddleware(cfg, nil))
		app.Get("/users", handler.List)

		// Authorization is handled by RBAC middleware at route level
		token := generateHandlerTestToken(1, "test@example.com")

		req := httptest.NewRequest(http.MethodGet, "/users", nil)
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

func TestHandler_Delete(t *testing.T) {
	cfg := getHandlerTestConfig()

	t.Run("successful delete", func(t *testing.T) {
		// Note: Authorization (user:delete permission) is handled by RequirePermission
		// middleware at route level. This test only verifies handler behavior.
		svc := &mockService{}
		handler := NewHandler(svc, getHandlerTestConfig())

		app := fiber.New()
		app.Use(middleware.JWTMiddleware(cfg, nil))
		app.Delete("/users/:id", handler.Delete)

		token := generateHandlerTestToken(1, "test@example.com")

		req := httptest.NewRequest(http.MethodDelete, "/users/1", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("Test request failed: %v", err)
		}
		defer resp.Body.Close()

		// Delete now returns 204 No Content
		if resp.StatusCode != http.StatusNoContent {
			bodyBytes, _ := io.ReadAll(resp.Body)
			t.Errorf("Expected status 204, got %d: %s", resp.StatusCode, string(bodyBytes))
		}
	})

	t.Run("user not found", func(t *testing.T) {
		svc := &mockService{
			deleteErr: ErrUserNotFound,
		}
		handler := NewHandler(svc, getHandlerTestConfig())

		app := fiber.New()
		app.Use(middleware.JWTMiddleware(cfg, nil))
		app.Delete("/users/:id", handler.Delete)

		token := generateHandlerTestToken(1, "test@example.com")

		req := httptest.NewRequest(http.MethodDelete, "/users/1", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("Test request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("Expected status 404, got %d", resp.StatusCode)
		}
	})

	t.Run("invalid user ID", func(t *testing.T) {
		svc := &mockService{}
		handler := NewHandler(svc, getHandlerTestConfig())

		app := fiber.New()
		app.Use(middleware.JWTMiddleware(cfg, nil))
		app.Delete("/users/:id", handler.Delete)

		token := generateHandlerTestToken(1, "test@example.com")

		req := httptest.NewRequest(http.MethodDelete, "/users/invalid", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("Test request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", resp.StatusCode)
		}
	})
}
