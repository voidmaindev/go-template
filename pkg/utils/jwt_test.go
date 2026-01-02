package utils

import (
	"testing"
	"time"
)

func getTestJWTConfig() *JWTConfig {
	return &JWTConfig{
		SecretKey:          "test-secret-key-at-least-32-chars!!",
		AccessTokenExpiry:  15 * time.Minute,
		RefreshTokenExpiry: 7 * 24 * time.Hour,
		Issuer:             "test-issuer",
	}
}

func TestGenerateAccessToken(t *testing.T) {
	config := getTestJWTConfig()

	token, err := GenerateAccessToken(1, "test@example.com", config)
	if err != nil {
		t.Fatalf("GenerateAccessToken() error = %v", err)
	}

	if token == "" {
		t.Error("GenerateAccessToken() returned empty token")
	}

	// Validate the token
	claims, err := ValidateAccessToken(token, config.SecretKey)
	if err != nil {
		t.Fatalf("ValidateAccessToken() error = %v", err)
	}

	if claims.UserID != 1 {
		t.Errorf("claims.UserID = %v, want 1", claims.UserID)
	}
	if claims.Email != "test@example.com" {
		t.Errorf("claims.Email = %v, want test@example.com", claims.Email)
	}
	if claims.TokenType != AccessToken {
		t.Errorf("claims.TokenType = %v, want %v", claims.TokenType, AccessToken)
	}
}

func TestGenerateRefreshToken(t *testing.T) {
	config := getTestJWTConfig()

	token, err := GenerateRefreshToken(1, "test@example.com", config)
	if err != nil {
		t.Fatalf("GenerateRefreshToken() error = %v", err)
	}

	if token == "" {
		t.Error("GenerateRefreshToken() returned empty token")
	}

	// Validate the token
	claims, err := ValidateRefreshToken(token, config.SecretKey)
	if err != nil {
		t.Fatalf("ValidateRefreshToken() error = %v", err)
	}

	if claims.UserID != 1 {
		t.Errorf("claims.UserID = %v, want 1", claims.UserID)
	}
	if claims.TokenType != RefreshToken {
		t.Errorf("claims.TokenType = %v, want %v", claims.TokenType, RefreshToken)
	}
}

func TestGenerateTokenPair(t *testing.T) {
	config := getTestJWTConfig()

	pair, err := GenerateTokenPair(1, "test@example.com", config)
	if err != nil {
		t.Fatalf("GenerateTokenPair() error = %v", err)
	}

	if pair.AccessToken == "" {
		t.Error("AccessToken is empty")
	}
	if pair.RefreshToken == "" {
		t.Error("RefreshToken is empty")
	}
	if pair.ExpiresAt == 0 {
		t.Error("ExpiresAt is 0")
	}

	// Tokens should be different
	if pair.AccessToken == pair.RefreshToken {
		t.Error("AccessToken and RefreshToken should be different")
	}
}

func TestValidateToken_InvalidToken(t *testing.T) {
	config := getTestJWTConfig()

	tests := []struct {
		name  string
		token string
	}{
		{"empty token", ""},
		{"garbage token", "not.a.valid.token"},
		{"malformed jwt", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.invalid.signature"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ValidateToken(tt.token, config.SecretKey)
			if err == nil {
				t.Error("ValidateToken() should return error for invalid token")
			}
		})
	}
}

func TestValidateToken_WrongSecret(t *testing.T) {
	config := getTestJWTConfig()

	token, err := GenerateAccessToken(1, "test@example.com", config)
	if err != nil {
		t.Fatalf("GenerateAccessToken() error = %v", err)
	}

	// Try to validate with wrong secret
	_, err = ValidateToken(token, "wrong-secret-key-at-least-32-chars!!")
	if err == nil {
		t.Error("ValidateToken() should return error for wrong secret")
	}
}

func TestValidateAccessToken_WithRefreshToken(t *testing.T) {
	config := getTestJWTConfig()

	// Generate a refresh token
	token, err := GenerateRefreshToken(1, "test@example.com", config)
	if err != nil {
		t.Fatalf("GenerateRefreshToken() error = %v", err)
	}

	// Try to validate as access token - should fail
	_, err = ValidateAccessToken(token, config.SecretKey)
	if err != ErrInvalidTokenType {
		t.Errorf("ValidateAccessToken() error = %v, want %v", err, ErrInvalidTokenType)
	}
}

func TestValidateRefreshToken_WithAccessToken(t *testing.T) {
	config := getTestJWTConfig()

	// Generate an access token
	token, err := GenerateAccessToken(1, "test@example.com", config)
	if err != nil {
		t.Fatalf("GenerateAccessToken() error = %v", err)
	}

	// Try to validate as refresh token - should fail
	_, err = ValidateRefreshToken(token, config.SecretKey)
	if err != ErrInvalidTokenType {
		t.Errorf("ValidateRefreshToken() error = %v, want %v", err, ErrInvalidTokenType)
	}
}

func TestValidateToken_ExpiredToken(t *testing.T) {
	config := &JWTConfig{
		SecretKey:          "test-secret-key-at-least-32-chars!!",
		AccessTokenExpiry:  -1 * time.Hour, // Already expired
		RefreshTokenExpiry: 7 * 24 * time.Hour,
		Issuer:             "test-issuer",
	}

	token, err := GenerateAccessToken(1, "test@example.com", config)
	if err != nil {
		t.Fatalf("GenerateAccessToken() error = %v", err)
	}

	_, err = ValidateToken(token, config.SecretKey)
	if err != ErrExpiredToken {
		t.Errorf("ValidateToken() error = %v, want %v", err, ErrExpiredToken)
	}
}

func TestGetTokenExpiry(t *testing.T) {
	config := getTestJWTConfig()

	token, err := GenerateAccessToken(1, "test@example.com", config)
	if err != nil {
		t.Fatalf("GenerateAccessToken() error = %v", err)
	}

	expiry, err := GetTokenExpiry(token, config.SecretKey)
	if err != nil {
		t.Fatalf("GetTokenExpiry() error = %v", err)
	}

	// Expiry should be close to 15 minutes (with some tolerance)
	expectedExpiry := config.AccessTokenExpiry
	tolerance := 5 * time.Second

	if expiry < expectedExpiry-tolerance || expiry > expectedExpiry+tolerance {
		t.Errorf("GetTokenExpiry() = %v, want ~%v", expiry, expectedExpiry)
	}
}

func TestGetTokenExpiry_ExpiredToken(t *testing.T) {
	config := &JWTConfig{
		SecretKey:          "test-secret-key-at-least-32-chars!!",
		AccessTokenExpiry:  -1 * time.Hour,
		RefreshTokenExpiry: 7 * 24 * time.Hour,
		Issuer:             "test-issuer",
	}

	token, err := GenerateAccessToken(1, "test@example.com", config)
	if err != nil {
		t.Fatalf("GenerateAccessToken() error = %v", err)
	}

	_, err = GetTokenExpiry(token, config.SecretKey)
	if err != ErrExpiredToken {
		t.Errorf("GetTokenExpiry() error = %v, want %v", err, ErrExpiredToken)
	}
}
