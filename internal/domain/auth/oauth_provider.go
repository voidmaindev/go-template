package auth

import "context"

// OAuthProvider defines the interface for OAuth providers
type OAuthProvider interface {
	// Name returns the provider name (google, facebook, apple)
	Name() string

	// GetAuthURL returns the authorization URL with the given state and optional PKCE challenge
	GetAuthURL(state string, pkce *PKCEChallenge) string

	// ExchangeCode exchanges an authorization code for tokens
	// If PKCE was used, verifier must be provided
	ExchangeCode(ctx context.Context, code, verifier string) (*OAuthTokens, error)

	// GetUserInfo retrieves user information using the access token
	GetUserInfo(ctx context.Context, accessToken string) (*OAuthUserInfo, error)

	// SupportsPKCE returns whether this provider supports PKCE
	SupportsPKCE() bool
}
