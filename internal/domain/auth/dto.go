package auth

// SelfRegisterRequest represents a self-registration request
type SelfRegisterRequest struct {
	Email    string `json:"email" validate:"required,email,max=255"`
	Password string `json:"password" validate:"required,password,max=128"`
	Name     string `json:"name" validate:"required,min=2,max=100"`
}

// SelfRegisterResponse represents the response after self-registration
type SelfRegisterResponse struct {
	Message string `json:"message"`
	UserID  uint   `json:"user_id,omitempty"`
}

// VerifyEmailRequest represents an email verification request
type VerifyEmailRequest struct {
	Token string `json:"token" validate:"required"`
}

// ResendVerificationRequest represents a request to resend verification email
type ResendVerificationRequest struct {
	Email string `json:"email" validate:"required,email,max=255"`
}

// ForgotPasswordRequest represents a forgot password request
type ForgotPasswordRequest struct {
	Email string `json:"email" validate:"required,email,max=255"`
}

// ResetPasswordRequest represents a password reset request
type ResetPasswordRequest struct {
	Token       string `json:"token" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,password,max=128"`
}

// SetPasswordRequest represents a request to set password for OAuth users
type SetPasswordRequest struct {
	NewPassword string `json:"new_password" validate:"required,password,max=128"`
}

// OAuthCallbackRequest represents the OAuth callback parameters
type OAuthCallbackRequest struct {
	Code  string `query:"code" validate:"required"`
	State string `query:"state" validate:"required"`
}

// OAuthTokenRequest represents a request to exchange OAuth code for tokens (SPA flow)
type OAuthTokenRequest struct {
	Code     string `json:"code" validate:"required"`
	State    string `json:"state" validate:"required"`
	Provider string `json:"provider" validate:"required"`
}

// OAuthUserInfo represents user information from OAuth provider
type OAuthUserInfo struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	Name          string `json:"name"`
	EmailVerified bool   `json:"email_verified"`
	Picture       string `json:"picture,omitempty"`
}

// OAuthTokens represents tokens returned from OAuth provider
type OAuthTokens struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token,omitempty"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	IDToken      string `json:"id_token,omitempty"`
}

// OAuthResult represents the result of OAuth authentication
type OAuthResult struct {
	IsNewUser   bool   `json:"is_new_user"`
	UserID      uint   `json:"user_id"`
	Email       string `json:"email"`
	RedirectURL string `json:"-"` // Frontend redirect URL (not serialized)
}

// LinkIdentityRequest represents a request to link an OAuth identity
type LinkIdentityRequest struct {
	Code  string `json:"code" validate:"required"`
	State string `json:"state" validate:"required"`
}

// MessageResponse represents a generic message response
type MessageResponse struct {
	Message string `json:"message"`
}
