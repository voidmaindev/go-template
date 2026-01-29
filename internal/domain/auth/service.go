package auth

import (
	"context"

	"github.com/voidmaindev/go-template/internal/domain/user"
)

// Service defines the auth service interface
type Service interface {
	// Email self-registration
	SelfRegister(ctx context.Context, req *SelfRegisterRequest) (*SelfRegisterResponse, error)
	VerifyEmail(ctx context.Context, token string) error
	ResendVerification(ctx context.Context, email string) error

	// Password reset
	ForgotPassword(ctx context.Context, email string) error
	ResetPassword(ctx context.Context, req *ResetPasswordRequest) error

	// OAuth
	GetOAuthURL(ctx context.Context, provider, redirectURL string) (string, string, error) // returns URL and state
	HandleOAuthCallback(ctx context.Context, provider, code, state string) (*OAuthResult, error)

	// Identity management (for authenticated users)
	GetUserIdentities(ctx context.Context, userID uint) ([]*user.ExternalIdentityResponse, error)
	LinkIdentity(ctx context.Context, userID uint, provider, code, state string) error
	UnlinkIdentity(ctx context.Context, userID uint, provider string) error

	// Password for OAuth users
	SetPassword(ctx context.Context, userID uint, password string) error
}
