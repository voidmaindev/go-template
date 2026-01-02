package user

import (
	"context"
	"errors"
	"log"

	"github.com/voidmaindev/GoTemplate/internal/common"
	"github.com/voidmaindev/GoTemplate/internal/config"
	"github.com/voidmaindev/GoTemplate/pkg/utils"
)

// Service errors
var (
	ErrEmailAlreadyExists = errors.New("email already exists")
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidPassword    = errors.New("invalid password")
	ErrSamePassword       = errors.New("new password must be different from current password")
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
}

// service implements the Service interface
type service struct {
	repo       Repository
	tokenStore *TokenStore
	jwtConfig  *config.JWTConfig
}

// NewService creates a new user service
func NewService(repo Repository, tokenStore *TokenStore, jwtConfig *config.JWTConfig) Service {
	return &service{
		repo:       repo,
		tokenStore: tokenStore,
		jwtConfig:  jwtConfig,
	}
}

// Register creates a new user and returns tokens
func (s *service) Register(ctx context.Context, req *RegisterRequest) (*TokenResponse, error) {
	// Check if email already exists
	exists, err := s.repo.ExistsByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrEmailAlreadyExists
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	// Create user
	user := &User{
		Email:    req.Email,
		Password: hashedPassword,
		Name:     req.Name,
	}

	if err := s.repo.Create(ctx, user); err != nil {
		return nil, err
	}

	// Generate tokens
	return s.generateTokenResponse(user)
}

// Login authenticates a user and returns tokens
func (s *service) Login(ctx context.Context, req *LoginRequest) (*TokenResponse, error) {
	// Find user by email
	user, err := s.repo.FindByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, common.ErrNotFound) {
			return nil, common.ErrInvalidCredentials
		}
		return nil, err
	}

	// Verify password
	if !utils.CheckPassword(req.Password, user.Password) {
		return nil, common.ErrInvalidCredentials
	}

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
		return nil, common.ErrTokenInvalid
	}

	// Check if token is blacklisted
	isBlacklisted, err := s.tokenStore.IsBlacklisted(ctx, refreshToken)
	if err != nil {
		return nil, err
	}
	if isBlacklisted {
		return nil, common.ErrTokenBlacklisted
	}

	// Get user
	user, err := s.repo.FindByID(ctx, claims.UserID)
	if err != nil {
		if errors.Is(err, common.ErrNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	// Blacklist old refresh token
	// We log errors but don't fail - the user should still get new tokens
	expiry, _ := utils.GetTokenExpiry(refreshToken, s.jwtConfig.SecretKey)
	if expiry > 0 {
		if err := s.tokenStore.Blacklist(ctx, refreshToken, expiry); err != nil {
			log.Printf("warning: failed to blacklist refresh token: %v", err)
		}
	}

	// Generate new tokens
	return s.generateTokenResponse(user)
}

// GetByID retrieves a user by ID
func (s *service) GetByID(ctx context.Context, id uint) (*User, error) {
	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, common.ErrNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return user, nil
}

// GetByEmail retrieves a user by email
func (s *service) GetByEmail(ctx context.Context, email string) (*User, error) {
	user, err := s.repo.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, common.ErrNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return user, nil
}

// Update updates a user
func (s *service) Update(ctx context.Context, id uint, req *UpdateUserRequest) (*User, error) {
	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, common.ErrNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	// Map non-nil/non-empty fields from request to model
	if err := utils.UpdateModel(user, req); err != nil {
		return nil, err
	}

	if err := s.repo.Update(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

// ChangePassword changes a user's password
func (s *service) ChangePassword(ctx context.Context, id uint, req *ChangePasswordRequest) error {
	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, common.ErrNotFound) {
			return ErrUserNotFound
		}
		return err
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
		return err
	}

	// Update password
	return s.repo.UpdateFields(ctx, id, map[string]any{
		"password": hashedPassword,
	})
}

// Delete soft-deletes a user
func (s *service) Delete(ctx context.Context, id uint) error {
	_, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, common.ErrNotFound) {
			return ErrUserNotFound
		}
		return err
	}

	return s.repo.Delete(ctx, id)
}

// List retrieves all users with pagination
func (s *service) List(ctx context.Context, pagination *common.Pagination) (*common.PaginatedResult[User], error) {
	users, total, err := s.repo.FindAll(ctx, pagination)
	if err != nil {
		return nil, err
	}

	return common.NewPaginatedResult(users, total, pagination), nil
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
