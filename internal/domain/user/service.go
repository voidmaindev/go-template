package user

import (
	"context"
	"log/slog"

	"github.com/voidmaindev/go-template/internal/common"
	"github.com/voidmaindev/go-template/internal/common/errors"
	"github.com/voidmaindev/go-template/internal/common/filter"
	"github.com/voidmaindev/go-template/internal/config"
	"github.com/voidmaindev/go-template/internal/domain/rbac"
	"github.com/voidmaindev/go-template/internal/telemetry"
	"github.com/voidmaindev/go-template/pkg/utils"
)

// Token blacklist retry settings
const (
	blacklistMaxRetries = 3
)

// Service defines the user service interface
type Service interface {
	// Authentication
	Register(ctx context.Context, req *RegisterRequest) (*TokenResponse, error)
	Login(ctx context.Context, req *LoginRequest) (*TokenResponse, error)
	Logout(ctx context.Context, token string) error
	RefreshToken(ctx context.Context, refreshToken string) (*TokenResponse, error)

	// User management
	GetByID(ctx context.Context, id uint) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	Update(ctx context.Context, id uint, req *UpdateUserRequest) (*User, error)
	ChangePassword(ctx context.Context, id uint, req *ChangePasswordRequest) error
	Delete(ctx context.Context, id uint) error

	// Listing
	List(ctx context.Context, pagination *common.Pagination) (*common.PaginatedResult[User], error)
	ListFiltered(ctx context.Context, params *filter.Params) (*common.FilteredResult[User], error)
}

// service implements the Service interface
type service struct {
	repo         Repository
	tokenStore   *TokenStore
	jwtConfig    *config.JWTConfig
	isProduction bool
	rbacSvc      rbac.Service
}

// NewService creates a new user service
func NewService(repo Repository, tokenStore *TokenStore, jwtConfig *config.JWTConfig, isProduction bool, rbacSvc rbac.Service) Service {
	return &service{
		repo:         repo,
		tokenStore:   tokenStore,
		jwtConfig:    jwtConfig,
		isProduction: isProduction,
		rbacSvc:      rbacSvc,
	}
}

// Register creates a new user and returns tokens
func (s *service) Register(ctx context.Context, req *RegisterRequest) (*TokenResponse, error) {
	// Check if email already exists
	exists, err := s.repo.ExistsByEmail(ctx, req.Email)
	if err != nil {
		return nil, errors.Internal(domainName, err).WithOperation("Register")
	}
	if exists {
		return nil, ErrEmailExists
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, errors.Internal(domainName, err).WithOperation("Register")
	}

	// Create user
	user := &User{
		Email:    req.Email,
		Password: hashedPassword,
		Name:     req.Name,
	}

	if err := s.repo.Create(ctx, user); err != nil {
		return nil, errors.Internal(domainName, err).WithOperation("Register")
	}

	// Assign RBAC roles (default to user if none specified)
	roleCodes := req.RoleCodes
	if len(roleCodes) == 0 {
		roleCodes = []string{rbac.RoleCodeUser}
	}

	for _, roleCode := range roleCodes {
		if err := s.rbacSvc.AssignRole(ctx, user.ID, roleCode); err != nil {
			// Log warning but don't fail registration
			slog.Warn("failed to assign role during registration", "role", roleCode, "userID", user.ID, "error", err)
		}
	}

	// Record metric
	telemetry.IncrementUsersRegistered()

	// Generate tokens
	return s.generateTokenResponse(user)
}

// Login authenticates a user and returns tokens
func (s *service) Login(ctx context.Context, req *LoginRequest) (*TokenResponse, error) {
	// Find user by email
	user, err := s.repo.FindByEmail(ctx, req.Email)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, ErrInvalidCredentials
		}
		return nil, errors.Internal(domainName, err).WithOperation("Login")
	}

	// Verify password
	if !utils.CheckPassword(req.Password, user.Password) {
		return nil, ErrInvalidCredentials
	}

	// Record metric
	telemetry.IncrementUsersLoggedIn()

	// Generate tokens
	return s.generateTokenResponse(user)
}

// Logout invalidates a token
func (s *service) Logout(ctx context.Context, token string) error {
	// Get token expiry
	expiry, err := utils.GetTokenExpiry(token, s.jwtConfig.SecretKey)
	if err != nil {
		// Token might already be expired, but we still try to blacklist it
		expiry = s.jwtConfig.AccessTokenExpiry
	}

	// Add token to blacklist
	return s.tokenStore.Blacklist(ctx, token, expiry)
}

// RefreshToken generates new tokens from a refresh token
func (s *service) RefreshToken(ctx context.Context, refreshToken string) (*TokenResponse, error) {
	// Validate refresh token
	claims, err := utils.ValidateRefreshToken(refreshToken, s.jwtConfig.SecretKey)
	if err != nil {
		return nil, ErrTokenInvalid
	}

	// Check if token is blacklisted
	isBlacklisted, err := s.tokenStore.IsBlacklisted(ctx, refreshToken)
	if err != nil {
		return nil, errors.Internal(domainName, err).WithOperation("RefreshToken")
	}
	if isBlacklisted {
		return nil, ErrTokenBlacklisted
	}

	// Get user
	user, err := s.repo.FindByID(ctx, claims.UserID)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, ErrUserNotFound
		}
		return nil, errors.Internal(domainName, err).WithOperation("RefreshToken")
	}

	// Blacklist old refresh token with retry logic
	// In production, fail the operation if blacklisting fails to prevent token reuse
	expiry, _ := utils.GetTokenExpiry(refreshToken, s.jwtConfig.SecretKey)
	if expiry > 0 {
		if err := s.tokenStore.BlacklistWithRetry(ctx, refreshToken, expiry, blacklistMaxRetries); err != nil {
			slog.Error("failed to blacklist refresh token after retries",
				"error", err,
				"retries", blacklistMaxRetries,
				"userID", claims.UserID,
			)
			// In production, fail the request to prevent stolen token reuse
			if s.isProduction {
				return nil, ErrTokenRefreshUnavailable
			}
		}
	}

	// Generate new tokens
	return s.generateTokenResponse(user)
}

// GetByID retrieves a user by ID
func (s *service) GetByID(ctx context.Context, id uint) (*User, error) {
	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, ErrUserNotFound
		}
		return nil, errors.Internal(domainName, err).WithOperation("GetByID")
	}
	return user, nil
}

// GetByEmail retrieves a user by email
func (s *service) GetByEmail(ctx context.Context, email string) (*User, error) {
	user, err := s.repo.FindByEmail(ctx, email)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, ErrUserNotFound
		}
		return nil, errors.Internal(domainName, err).WithOperation("GetByEmail")
	}
	return user, nil
}

// Update updates a user
func (s *service) Update(ctx context.Context, id uint, req *UpdateUserRequest) (*User, error) {
	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, ErrUserNotFound
		}
		return nil, errors.Internal(domainName, err).WithOperation("Update")
	}

	// Map non-nil/non-empty fields from request to model
	if err := utils.UpdateModel(user, req); err != nil {
		return nil, errors.Internal(domainName, err).WithOperation("Update")
	}

	if err := s.repo.Update(ctx, user); err != nil {
		return nil, errors.Internal(domainName, err).WithOperation("Update")
	}

	return user, nil
}

// ChangePassword changes a user's password
func (s *service) ChangePassword(ctx context.Context, id uint, req *ChangePasswordRequest) error {
	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.IsNotFound(err) {
			return ErrUserNotFound
		}
		return errors.Internal(domainName, err).WithOperation("ChangePassword")
	}

	// Verify current password
	if !utils.CheckPassword(req.CurrentPassword, user.Password) {
		return ErrInvalidPassword
	}

	// Check if new password is different
	if utils.CheckPassword(req.NewPassword, user.Password) {
		return ErrSamePassword
	}

	// Hash new password
	hashedPassword, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		return errors.Internal(domainName, err).WithOperation("ChangePassword")
	}

	// Update password
	if err := s.repo.UpdateFields(ctx, id, map[string]any{
		"password": hashedPassword,
	}); err != nil {
		return errors.Internal(domainName, err).WithOperation("ChangePassword")
	}
	return nil
}

// Delete soft-deletes a user
func (s *service) Delete(ctx context.Context, id uint) error {
	_, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.IsNotFound(err) {
			return ErrUserNotFound
		}
		return errors.Internal(domainName, err).WithOperation("Delete")
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return errors.Internal(domainName, err).WithOperation("Delete")
	}
	return nil
}

// List retrieves all users with pagination
func (s *service) List(ctx context.Context, pagination *common.Pagination) (*common.PaginatedResult[User], error) {
	users, total, err := s.repo.FindAll(ctx, pagination)
	if err != nil {
		return nil, errors.Internal(domainName, err).WithOperation("List")
	}

	return common.NewPaginatedResult(users, total, pagination), nil
}

// ListFiltered retrieves users with dynamic filtering, sorting, and pagination
func (s *service) ListFiltered(ctx context.Context, params *filter.Params) (*common.FilteredResult[User], error) {
	users, total, err := s.repo.FindAllFiltered(ctx, params)
	if err != nil {
		return nil, errors.Internal(domainName, err).WithOperation("ListFiltered")
	}

	return common.NewFilteredResult(users, total, params), nil
}

// generateTokenResponse generates a token response for a user
func (s *service) generateTokenResponse(user *User) (*TokenResponse, error) {
	jwtConfig := &utils.JWTConfig{
		SecretKey:          s.jwtConfig.SecretKey,
		AccessTokenExpiry:  s.jwtConfig.AccessTokenExpiry,
		RefreshTokenExpiry: s.jwtConfig.RefreshTokenExpiry,
		Issuer:             s.jwtConfig.Issuer,
	}

	tokenPair, err := utils.GenerateTokenPair(user.ID, user.Email, jwtConfig)
	if err != nil {
		return nil, err
	}

	return &TokenResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresAt:    tokenPair.ExpiresAt,
		User:         user.ToResponse(),
	}, nil
}
