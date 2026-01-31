package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/voidmaindev/go-template/internal/config"
)

const (
	facebookAuthURL     = "https://www.facebook.com/v18.0/dialog/oauth"
	facebookTokenURL    = "https://graph.facebook.com/v18.0/oauth/access_token"
	facebookUserInfoURL = "https://graph.facebook.com/me"
)

// facebookProvider implements OAuthProvider for Facebook
type facebookProvider struct {
	cfg *config.OAuthProviderConfig
}

// NewFacebookProvider creates a new Facebook OAuth provider
func NewFacebookProvider(cfg *config.OAuthProviderConfig) OAuthProvider {
	return &facebookProvider{cfg: cfg}
}

// Name returns the provider name
func (p *facebookProvider) Name() string {
	return "facebook"
}

// SupportsPKCE returns whether this provider supports PKCE
func (p *facebookProvider) SupportsPKCE() bool {
	return true
}

// GetAuthURL returns the Facebook authorization URL
func (p *facebookProvider) GetAuthURL(state string, pkce *PKCEChallenge) string {
	params := url.Values{
		"client_id":     {p.cfg.ClientID},
		"redirect_uri":  {p.cfg.RedirectURL},
		"response_type": {"code"},
		"scope":         {"email,public_profile"},
		"state":         {state},
	}

	// Add PKCE parameters if provided
	if pkce != nil {
		params.Set("code_challenge", pkce.Challenge)
		params.Set("code_challenge_method", pkce.Method)
	}

	return facebookAuthURL + "?" + params.Encode()
}

// ExchangeCode exchanges an authorization code for tokens
func (p *facebookProvider) ExchangeCode(ctx context.Context, code, verifier string) (*OAuthTokens, error) {
	data := url.Values{
		"client_id":     {p.cfg.ClientID},
		"client_secret": {p.cfg.ClientSecret},
		"code":          {code},
		"redirect_uri":  {p.cfg.RedirectURL},
	}

	// Add PKCE verifier if provided
	if verifier != "" {
		data.Set("code_verifier", verifier)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", facebookTokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := oauthHTTPClient.Do(req)
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

// GetUserInfo retrieves user information from Facebook
func (p *facebookProvider) GetUserInfo(ctx context.Context, accessToken string) (*OAuthUserInfo, error) {
	// Facebook requires specifying the fields to retrieve
	// Use Authorization header instead of query parameter for security
	reqURL := facebookUserInfoURL + "?fields=id,name,email,picture"

	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := oauthHTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("get user info: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("user info request failed: %s", string(body))
	}

	var fbUser struct {
		ID      string `json:"id"`
		Name    string `json:"name"`
		Email   string `json:"email"`
		Picture struct {
			Data struct {
				URL string `json:"url"`
			} `json:"data"`
		} `json:"picture"`
	}

	if err := json.Unmarshal(body, &fbUser); err != nil {
		return nil, fmt.Errorf("parse user info: %w", err)
	}

	return &OAuthUserInfo{
		ID:            fbUser.ID,
		Email:         fbUser.Email,
		Name:          fbUser.Name,
		EmailVerified: fbUser.Email != "", // Facebook only returns email if verified
		Picture:       fbUser.Picture.Data.URL,
	}, nil
}
