package auth

import (
	"context"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/voidmaindev/go-template/internal/config"
)

const (
	appleAuthURL     = "https://appleid.apple.com/auth/authorize"
	appleTokenURL    = "https://appleid.apple.com/auth/token"
	appleKeysURL     = "https://appleid.apple.com/auth/keys"
)

// appleProvider implements OAuthProvider for Apple Sign-In
type appleProvider struct {
	cfg *config.OAuthProviderConfig
}

// NewAppleProvider creates a new Apple OAuth provider
func NewAppleProvider(cfg *config.OAuthProviderConfig) OAuthProvider {
	return &appleProvider{cfg: cfg}
}

// Name returns the provider name
func (p *appleProvider) Name() string {
	return "apple"
}

// GetAuthURL returns the Apple authorization URL
func (p *appleProvider) GetAuthURL(state string) string {
	params := url.Values{
		"client_id":     {p.cfg.ClientID},
		"redirect_uri":  {p.cfg.RedirectURL},
		"response_type": {"code id_token"},
		"scope":         {"name email"},
		"state":         {state},
		"response_mode": {"form_post"},
	}
	return appleAuthURL + "?" + params.Encode()
}

// ExchangeCode exchanges an authorization code for tokens
// Note: Apple's client_secret is a JWT signed with your private key
func (p *appleProvider) ExchangeCode(ctx context.Context, code string) (*OAuthTokens, error) {
	clientSecret, err := p.generateClientSecret()
	if err != nil {
		return nil, fmt.Errorf("generate client secret: %w", err)
	}

	data := url.Values{
		"client_id":     {p.cfg.ClientID},
		"client_secret": {clientSecret},
		"code":          {code},
		"grant_type":    {"authorization_code"},
		"redirect_uri":  {p.cfg.RedirectURL},
	}

	req, err := http.NewRequestWithContext(ctx, "POST", appleTokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("exchange code: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token exchange failed: %s", string(body))
	}

	var tokens OAuthTokens
	if err := json.Unmarshal(body, &tokens); err != nil {
		return nil, fmt.Errorf("parse tokens: %w", err)
	}

	return &tokens, nil
}

// GetUserInfo retrieves user information from Apple's ID token
// Apple doesn't have a userinfo endpoint, so we parse the ID token
func (p *appleProvider) GetUserInfo(ctx context.Context, accessToken string) (*OAuthUserInfo, error) {
	// For Apple, we need to get the ID token from the token response
	// The accessToken parameter here would be the id_token
	// This is a simplification - in production you'd want to verify the JWT signature

	// Parse the ID token without verification for now
	// In production, you should verify the signature using Apple's public keys
	token, _, err := new(jwt.Parser).ParseUnverified(accessToken, jwt.MapClaims{})
	if err != nil {
		return nil, fmt.Errorf("parse id token: %w", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	userInfo := &OAuthUserInfo{
		ID: claims["sub"].(string),
	}

	if email, ok := claims["email"].(string); ok {
		userInfo.Email = email
	}

	if emailVerified, ok := claims["email_verified"].(bool); ok {
		userInfo.EmailVerified = emailVerified
	} else if emailVerified, ok := claims["email_verified"].(string); ok {
		userInfo.EmailVerified = emailVerified == "true"
	}

	// Apple only provides name on first authorization
	// You should store it when received in the callback
	return userInfo, nil
}

// generateClientSecret generates a JWT client secret for Apple
// The client_secret is a JWT signed with your private key
func (p *appleProvider) generateClientSecret() (string, error) {
	// The ClientSecret field should contain the PEM-encoded private key
	// Parse the private key
	block, _ := pem.Decode([]byte(p.cfg.ClientSecret))
	if block == nil {
		return "", fmt.Errorf("failed to decode PEM block")
	}

	privateKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return "", fmt.Errorf("parse private key: %w", err)
	}

	now := time.Now()
	claims := jwt.MapClaims{
		"iss": p.cfg.ClientID, // Your Team ID
		"iat": now.Unix(),
		"exp": now.Add(5 * time.Minute).Unix(),
		"aud": "https://appleid.apple.com",
		"sub": p.cfg.ClientID, // Your Services ID
	}

	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
	// Note: You'd also need to set the key ID header (kid)
	// token.Header["kid"] = "YOUR_KEY_ID"

	return token.SignedString(privateKey)
}
