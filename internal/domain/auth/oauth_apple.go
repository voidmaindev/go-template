package auth

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/voidmaindev/go-template/internal/config"
)

const (
	appleAuthURL  = "https://appleid.apple.com/auth/authorize"
	appleTokenURL = "https://appleid.apple.com/auth/token"
	appleKeysURL  = "https://appleid.apple.com/auth/keys"
	appleIssuer   = "https://appleid.apple.com"
	jwksCacheTTL  = 24 * time.Hour
	// jwksMinRefetchGap rate-limits forced JWKS refreshes triggered by a
	// token carrying an unknown `kid`. Without it, an attacker could spam
	// unknown-kid tokens to hammer Apple's keys endpoint.
	jwksMinRefetchGap = 5 * time.Minute
)

// appleJWKS represents Apple's JWKS response
type appleJWKS struct {
	Keys []appleJWK `json:"keys"`
}

// appleJWK represents a single JSON Web Key
type appleJWK struct {
	Kty string `json:"kty"` // Key type (RSA, EC)
	Kid string `json:"kid"` // Key ID
	Use string `json:"use"` // Key use (sig)
	Alg string `json:"alg"` // Algorithm (RS256, ES256)
	N   string `json:"n"`   // RSA modulus (for RSA keys)
	E   string `json:"e"`   // RSA exponent (for RSA keys)
	Crv string `json:"crv"` // Curve (for EC keys)
	X   string `json:"x"`   // X coordinate (for EC keys)
	Y   string `json:"y"`   // Y coordinate (for EC keys)
}

// jwksCache caches Apple's public keys
type jwksCache struct {
	mu        sync.RWMutex
	keys      map[string]*ecdsa.PublicKey
	fetchedAt time.Time
}

// appleProvider implements OAuthProvider for Apple Sign-In
type appleProvider struct {
	cfg   *config.OAuthProviderConfig
	cache *jwksCache
}

// Global JWKS cache shared across Apple provider instances
var globalAppleJWKSCache = &jwksCache{
	keys: make(map[string]*ecdsa.PublicKey),
}

// NewAppleProvider creates a new Apple OAuth provider
func NewAppleProvider(cfg *config.OAuthProviderConfig) OAuthProvider {
	return &appleProvider{
		cfg:   cfg,
		cache: globalAppleJWKSCache,
	}
}

// Name returns the provider name
func (p *appleProvider) Name() string {
	return "apple"
}

// SupportsPKCE returns whether this provider supports PKCE
// Note: Apple has limited PKCE support with form_post response mode
func (p *appleProvider) SupportsPKCE() bool {
	return false
}

// GetAuthURL returns the Apple authorization URL
func (p *appleProvider) GetAuthURL(state string, pkce *PKCEChallenge) string {
	params := url.Values{
		"client_id":     {p.cfg.ClientID},
		"redirect_uri":  {p.cfg.RedirectURL},
		"response_type": {"code id_token"},
		"scope":         {"name email"},
		"state":         {state},
		"response_mode": {"form_post"},
	}
	// Apple doesn't support PKCE with form_post response mode
	return appleAuthURL + "?" + params.Encode()
}

// ExchangeCode exchanges an authorization code for tokens
// Note: Apple's client_secret is a JWT signed with your private key
func (p *appleProvider) ExchangeCode(ctx context.Context, code, verifier string) (*OAuthTokens, error) {
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

	// Note: verifier is ignored as Apple doesn't support PKCE with form_post
	_ = verifier

	req, err := http.NewRequestWithContext(ctx, "POST", appleTokenURL, strings.NewReader(data.Encode()))
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

// GetUserInfo retrieves user information from Apple's ID token
// Apple doesn't have a userinfo endpoint, so we parse and verify the ID token
func (p *appleProvider) GetUserInfo(ctx context.Context, idToken string) (*OAuthUserInfo, error) {
	// Parse and verify the ID token using Apple's public keys
	token, err := jwt.Parse(idToken, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodECDSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		// Get key ID from token header
		kid, ok := token.Header["kid"].(string)
		if !ok {
			return nil, fmt.Errorf("missing kid in token header")
		}

		// Get public key for this key ID
		publicKey, err := p.getPublicKey(ctx, kid)
		if err != nil {
			return nil, fmt.Errorf("get public key: %w", err)
		}

		return publicKey, nil
	}, jwt.WithValidMethods([]string{"ES256"}),
		jwt.WithIssuer(appleIssuer),
		jwt.WithAudience(p.cfg.ClientID),
		jwt.WithExpirationRequired(),
	)

	if err != nil {
		return nil, fmt.Errorf("verify id token: %w", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	sub, ok := claims["sub"].(string)
	if !ok {
		return nil, fmt.Errorf("missing sub claim")
	}

	userInfo := &OAuthUserInfo{
		ID: sub,
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

// getPublicKey retrieves the public key for a given key ID.
//
// Three paths:
//  1. Hit: cached kid, TTL fresh → return.
//  2. Cache stale or empty → regular refresh.
//  3. Cache fresh but kid missing (likely Apple rotated keys) → force-refresh,
//     rate-limited so unknown-kid floods can't DoS Apple's JWKS endpoint.
func (p *appleProvider) getPublicKey(ctx context.Context, kid string) (*ecdsa.PublicKey, error) {
	p.cache.mu.RLock()
	fresh := len(p.cache.keys) > 0 && time.Since(p.cache.fetchedAt) < jwksCacheTTL
	if key, ok := p.cache.keys[kid]; ok && fresh {
		p.cache.mu.RUnlock()
		return key, nil
	}
	sinceFetch := time.Since(p.cache.fetchedAt)
	p.cache.mu.RUnlock()

	// Cache is fresh but this kid isn't in it → Apple likely rotated.
	// Force a refresh, but not more often than jwksMinRefetchGap.
	force := fresh && sinceFetch >= jwksMinRefetchGap
	if err := p.fetchJWKS(ctx, force); err != nil {
		return nil, err
	}

	p.cache.mu.RLock()
	key, ok := p.cache.keys[kid]
	p.cache.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("key not found: %s", kid)
	}

	return key, nil
}

// fetchJWKS fetches Apple's public keys from the JWKS endpoint.
// When force is true, the TTL guard is bypassed — used for cache-miss refreshes
// triggered by an unknown `kid` in a token header.
func (p *appleProvider) fetchJWKS(ctx context.Context, force bool) error {
	p.cache.mu.Lock()
	defer p.cache.mu.Unlock()

	// Double-check after acquiring write lock
	if !force && time.Since(p.cache.fetchedAt) < jwksCacheTTL && len(p.cache.keys) > 0 {
		return nil
	}

	req, err := http.NewRequestWithContext(ctx, "GET", appleKeysURL, nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	resp, err := oauthHTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("fetch JWKS: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("JWKS request failed with status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read JWKS response: %w", err)
	}

	var jwks appleJWKS
	if err := json.Unmarshal(body, &jwks); err != nil {
		return fmt.Errorf("parse JWKS: %w", err)
	}

	// Parse and cache keys
	newKeys := make(map[string]*ecdsa.PublicKey)
	for _, jwk := range jwks.Keys {
		if jwk.Kty != "EC" || jwk.Crv != "P-256" {
			continue // Apple uses ES256 (P-256 curve)
		}

		key, err := parseECPublicKey(jwk)
		if err != nil {
			continue // Skip invalid keys
		}
		newKeys[jwk.Kid] = key
	}

	p.cache.keys = newKeys
	p.cache.fetchedAt = time.Now()

	return nil
}

// parseECPublicKey parses an EC public key from JWK format
func parseECPublicKey(jwk appleJWK) (*ecdsa.PublicKey, error) {
	xBytes, err := base64.RawURLEncoding.DecodeString(jwk.X)
	if err != nil {
		return nil, fmt.Errorf("decode x: %w", err)
	}

	yBytes, err := base64.RawURLEncoding.DecodeString(jwk.Y)
	if err != nil {
		return nil, fmt.Errorf("decode y: %w", err)
	}

	return &ecdsa.PublicKey{
		Curve: elliptic.P256(),
		X:     new(big.Int).SetBytes(xBytes),
		Y:     new(big.Int).SetBytes(yBytes),
	}, nil
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
