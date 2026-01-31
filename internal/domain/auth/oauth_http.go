package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"net/http"
	"time"
)

// oauthHTTPClient is a shared HTTP client with timeout for all OAuth operations
var oauthHTTPClient = &http.Client{
	Timeout: 30 * time.Second,
}

// PKCEChallenge holds PKCE code challenge data
type PKCEChallenge struct {
	Verifier  string // code_verifier (stored server-side)
	Challenge string // code_challenge (sent to auth server)
	Method    string // code_challenge_method (always "S256")
}

// GeneratePKCE generates a new PKCE code verifier and challenge
// Uses S256 method as per RFC 7636 recommendation
func GeneratePKCE() (*PKCEChallenge, error) {
	// Generate 32 bytes of randomness for the verifier
	verifierBytes := make([]byte, 32)
	if _, err := rand.Read(verifierBytes); err != nil {
		return nil, err
	}

	// Base64 URL encode the verifier
	verifier := base64.RawURLEncoding.EncodeToString(verifierBytes)

	// Generate challenge using SHA256
	hash := sha256.Sum256([]byte(verifier))
	challenge := base64.RawURLEncoding.EncodeToString(hash[:])

	return &PKCEChallenge{
		Verifier:  verifier,
		Challenge: challenge,
		Method:    "S256",
	}, nil
}
