package auth

import "context"

// OAuthProvider defines the interface for OAuth providers
type OAuthProvider interface {
	// Name returns the provider name (google, facebook, apple)
	Name() string

	// GetAuthURL returns the authorization URL with the given state
	GetAuthURL(state string) string

	// ExchangeCode exchanges an authorization code for tokens
	ExchangeCode(ctx context.Context, code string) (*OAuthTokens, error)

	// GetUserInfo retrieves user information using the access token
	GetUserInfo(ctx context.Context, accessToken string) (*OAuthUserInfo, error)
}
