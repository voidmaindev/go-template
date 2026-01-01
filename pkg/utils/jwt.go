package utils

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// TokenType represents the type of JWT token
type TokenType string

const (
	// AccessToken is short-lived token for API access
	AccessToken TokenType = "access"
	// RefreshToken is long-lived token for obtaining new access tokens
	RefreshToken TokenType = "refresh"
)

// Claims represents the JWT claims
type Claims struct {
	UserID    uint      `json:"user_id"`
	Email     string    `json:"email"`
	TokenType TokenType `json:"token_type"`
	jwt.RegisteredClaims
}

// JWTConfig holds JWT configuration
type JWTConfig struct {
	SecretKey          string
	AccessTokenExpiry  time.Duration
	RefreshTokenExpiry time.Duration
	Issuer             string
}

// TokenPair contains both access and refresh tokens
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresAt    int64  `json:"expires_at"`
}

var (
	ErrInvalidToken     = errors.New("invalid token")
	ErrExpiredToken     = errors.New("token has expired")
	ErrInvalidTokenType = errors.New("invalid token type")
)

// GenerateAccessToken generates a new access token
func GenerateAccessToken(userID uint, email string, config *JWTConfig) (string, error) {
	return generateToken(userID, email, AccessToken, config.AccessTokenExpiry, config)
}

// GenerateRefreshToken generates a new refresh token
func GenerateRefreshToken(userID uint, email string, config *JWTConfig) (string, error) {
	return generateToken(userID, email, RefreshToken, config.RefreshTokenExpiry, config)
}

// GenerateTokenPair generates both access and refresh tokens
func GenerateTokenPair(userID uint, email string, config *JWTConfig) (*TokenPair, error) {
	accessToken, err := GenerateAccessToken(userID, email, config)
	if err != nil {
		return nil, err
	}

	refreshToken, err := GenerateRefreshToken(userID, email, config)
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    time.Now().Add(config.AccessTokenExpiry).Unix(),
	}, nil
}

// generateToken creates a new JWT token
func generateToken(userID uint, email string, tokenType TokenType, expiry time.Duration, config *JWTConfig) (string, error) {
	now := time.Now()
	claims := &Claims{
		UserID:    userID,
		Email:     email,
		TokenType: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    config.Issuer,
			Subject:   email,
			ExpiresAt: jwt.NewNumericDate(now.Add(expiry)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(config.SecretKey))
}

// ValidateToken validates a JWT token and returns the claims
func ValidateToken(tokenString string, secretKey string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return []byte(secretKey), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

// ValidateAccessToken validates an access token
func ValidateAccessToken(tokenString string, secretKey string) (*Claims, error) {
	claims, err := ValidateToken(tokenString, secretKey)
	if err != nil {
		return nil, err
	}

	if claims.TokenType != AccessToken {
		return nil, ErrInvalidTokenType
	}

	return claims, nil
}

// ValidateRefreshToken validates a refresh token
func ValidateRefreshToken(tokenString string, secretKey string) (*Claims, error) {
	claims, err := ValidateToken(tokenString, secretKey)
	if err != nil {
		return nil, err
	}

	if claims.TokenType != RefreshToken {
		return nil, ErrInvalidTokenType
	}

	return claims, nil
}

// GetTokenExpiry returns the remaining time until token expiry
func GetTokenExpiry(tokenString string, secretKey string) (time.Duration, error) {
	claims, err := ValidateToken(tokenString, secretKey)
	if err != nil {
		return 0, err
	}

	expiry := claims.ExpiresAt.Time.Sub(time.Now())
	if expiry < 0 {
		return 0, ErrExpiredToken
	}

	return expiry, nil
}
