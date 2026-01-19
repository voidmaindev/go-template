package user

import (
	"time"
)

// RegisterRequest represents the registration request
type RegisterRequest struct {
	Email     string   `json:"email" validate:"required,email,max=255"`
	Password  string   `json:"password" validate:"required,password,max=128"`
	Name      string   `json:"name" validate:"required,min=2,max=100"`
	RoleCodes []string `json:"role_codes" validate:"omitempty,dive,min=1"`
}

// LoginRequest represents the login request
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email,max=255"`
	Password string `json:"password" validate:"required,max=128"`
}

// RefreshTokenRequest represents the refresh token request
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// LogoutRequest represents the logout request (optional body)
type LogoutRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// ChangePasswordRequest represents the change password request
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" validate:"required,max=128"`
	NewPassword     string `json:"new_password" validate:"required,password,max=128"`
}

// UpdateUserRequest represents the update user request
type UpdateUserRequest struct {
	Name *string `json:"name" validate:"omitempty,min=2,max=100"`
}

// UserResponse represents the user response (without sensitive data)
type UserResponse struct {
	ID        uint      `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TokenResponse represents the token response
type TokenResponse struct {
	AccessToken  string        `json:"access_token"`
	RefreshToken string        `json:"refresh_token"`
	ExpiresAt    int64         `json:"expires_at"`
	User         *UserResponse `json:"user"`
}

// AuthResponse represents the authentication response
type AuthResponse struct {
	User   *UserResponse `json:"user"`
	Tokens *TokenResponse `json:"tokens,omitempty"`
}
