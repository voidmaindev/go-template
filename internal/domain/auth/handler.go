package auth

import (
	"encoding/json"
	"net/url"

	"github.com/gofiber/fiber/v2"
	"github.com/voidmaindev/go-template/internal/common"
	"github.com/voidmaindev/go-template/internal/config"
	"github.com/voidmaindev/go-template/internal/domain/user"
	"github.com/voidmaindev/go-template/internal/middleware"
	"github.com/voidmaindev/go-template/pkg/utils"
)

// Handler handles auth HTTP requests
type Handler struct {
	service       Service
	userSvc       user.Service
	jwtConfig     *config.JWTConfig
	selfRegConfig *config.SelfRegistrationConfig
}

// NewHandler creates a new auth handler
func NewHandler(service Service, userSvc user.Service, jwtConfig *config.JWTConfig, selfRegConfig *config.SelfRegistrationConfig) *Handler {
	return &Handler{
		service:       service,
		userSvc:       userSvc,
		jwtConfig:     jwtConfig,
		selfRegConfig: selfRegConfig,
	}
}

// SelfRegister handles self-registration requests
// POST /auth/self/register
func (h *Handler) SelfRegister(c *fiber.Ctx) error {
	req, err := common.ParseAndValidate[SelfRegisterRequest](c)
	if err != nil {
		return nil
	}

	resp, err := h.service.SelfRegister(c.Context(), req)
	if err != nil {
		return common.HandleError(c, err)
	}

	return common.CreatedResponse(c, resp)
}

// VerifyEmail handles email verification
// POST /auth/self/verify-email
func (h *Handler) VerifyEmail(c *fiber.Ctx) error {
	req, err := common.ParseAndValidate[VerifyEmailRequest](c)
	if err != nil {
		return nil
	}

	if err := h.service.VerifyEmail(c.Context(), req.Token); err != nil {
		return common.HandleError(c, err)
	}

	return common.SuccessResponseWithMessage(c, "Email verified successfully", nil)
}

// ResendVerification handles resending verification email
// POST /auth/self/resend-verification
func (h *Handler) ResendVerification(c *fiber.Ctx) error {
	req, err := common.ParseAndValidate[ResendVerificationRequest](c)
	if err != nil {
		return nil
	}

	if err := h.service.ResendVerification(c.Context(), req.Email); err != nil {
		return common.HandleError(c, err)
	}

	return common.SuccessResponseWithMessage(c, "If your email is registered and unverified, you will receive a verification email", nil)
}

// ForgotPassword handles forgot password requests
// POST /auth/self/forgot-password
func (h *Handler) ForgotPassword(c *fiber.Ctx) error {
	req, err := common.ParseAndValidate[ForgotPasswordRequest](c)
	if err != nil {
		return nil
	}

	if err := h.service.ForgotPassword(c.Context(), req.Email); err != nil {
		return common.HandleError(c, err)
	}

	return common.SuccessResponseWithMessage(c, "If an account exists with this email, you will receive a password reset email", nil)
}

// ResetPassword handles password reset
// POST /auth/self/reset-password
func (h *Handler) ResetPassword(c *fiber.Ctx) error {
	req, err := common.ParseAndValidate[ResetPasswordRequest](c)
	if err != nil {
		return nil
	}

	if err := h.service.ResetPassword(c.Context(), req); err != nil {
		return common.HandleError(c, err)
	}

	return common.SuccessResponseWithMessage(c, "Password reset successfully", nil)
}

// GetOAuthURL redirects to OAuth provider authorization page
// GET /auth/oauth/:provider
func (h *Handler) GetOAuthURL(c *fiber.Ctx) error {
	provider := c.Params("provider")
	redirectURL := c.Query("redirect_url", "")

	authURL, _, err := h.service.GetOAuthURL(c.Context(), provider, redirectURL)
	if err != nil {
		return common.HandleError(c, err)
	}

	return c.Redirect(authURL, fiber.StatusTemporaryRedirect)
}

// OAuthCallback handles OAuth provider callback
// GET /auth/oauth/:provider/callback
func (h *Handler) OAuthCallback(c *fiber.Ctx) error {
	provider := c.Params("provider")
	code := c.Query("code")
	state := c.Query("state")

	if code == "" || state == "" {
		return common.BadRequestResponse(c, "missing code or state parameter")
	}

	result, err := h.service.HandleOAuthCallback(c.Context(), provider, code, state)
	if err != nil {
		// If redirect URL was provided, redirect with error
		if result != nil && result.RedirectURL != "" {
			redirectURL := result.RedirectURL + "?error=" + url.QueryEscape(err.Error())
			return c.Redirect(redirectURL, fiber.StatusTemporaryRedirect)
		}
		return common.HandleError(c, err)
	}

	// Get user and generate tokens
	u, err := h.userSvc.GetByID(c.Context(), result.UserID)
	if err != nil {
		return common.HandleError(c, err)
	}

	tokenPair, err := utils.GenerateTokenPair(u.ID, u.Email, toUtilsJWTConfig(h.jwtConfig))
	if err != nil {
		return common.InternalServerErrorResponse(c)
	}

	// If redirect URL was provided (SPA popup flow), redirect to frontend with tokens in URL hash
	if result.RedirectURL != "" {
		// Prepare auth data as JSON
		authData := map[string]any{
			"access_token":  tokenPair.AccessToken,
			"refresh_token": tokenPair.RefreshToken,
			"expires_at":    tokenPair.ExpiresAt,
			"user":          u.ToResponse(),
		}
		authJSON, err := json.Marshal(authData)
		if err != nil {
			return common.InternalServerErrorResponse(c)
		}
		// Redirect with tokens in URL fragment (not sent to server, safe from logs)
		redirectURL := result.RedirectURL + "#auth=" + url.QueryEscape(string(authJSON))
		return c.Redirect(redirectURL, fiber.StatusTemporaryRedirect)
	}

	// No redirect URL - return JSON response (direct API flow)
	return common.SuccessResponse(c, user.TokenResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresAt:    tokenPair.ExpiresAt,
		User:         u.ToResponse(),
	})
}

// OAuthToken handles OAuth token exchange for SPAs
// POST /auth/oauth/:provider/token
func (h *Handler) OAuthToken(c *fiber.Ctx) error {
	provider := c.Params("provider")

	var req OAuthTokenRequest
	if err := c.BodyParser(&req); err != nil {
		return common.BadRequestResponse(c, "invalid request body")
	}

	result, err := h.service.HandleOAuthCallback(c.Context(), provider, req.Code, req.State)
	if err != nil {
		return common.HandleError(c, err)
	}

	// Get user and generate tokens
	u, err := h.userSvc.GetByID(c.Context(), result.UserID)
	if err != nil {
		return common.HandleError(c, err)
	}

	// Generate JWT tokens
	tokenPair, err := utils.GenerateTokenPair(u.ID, u.Email, toUtilsJWTConfig(h.jwtConfig))
	if err != nil {
		return common.InternalServerErrorResponse(c)
	}

	return common.SuccessResponse(c, user.TokenResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresAt:    tokenPair.ExpiresAt,
		User:         u.ToResponse(),
	})
}

// GetUserIdentities returns the user's linked OAuth identities
// GET /users/me/identities
func (h *Handler) GetUserIdentities(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserIDFromContext(c)
	if !ok {
		return common.UnauthorizedResponse(c, "")
	}

	identities, err := h.service.GetUserIdentities(c.Context(), userID)
	if err != nil {
		return common.HandleError(c, err)
	}

	return common.SuccessResponse(c, fiber.Map{"identities": identities})
}

// LinkIdentity links a new OAuth identity to the user's account
// POST /users/me/identities/:provider
func (h *Handler) LinkIdentity(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserIDFromContext(c)
	if !ok {
		return common.UnauthorizedResponse(c, "")
	}
	provider := c.Params("provider")

	var req LinkIdentityRequest
	if err := c.BodyParser(&req); err != nil {
		return common.BadRequestResponse(c, "invalid request body")
	}

	if err := h.service.LinkIdentity(c.Context(), userID, provider, req.Code, req.State); err != nil {
		return common.HandleError(c, err)
	}

	return common.SuccessResponseWithMessage(c, "Identity linked successfully", nil)
}

// UnlinkIdentity unlinks an OAuth identity from the user's account
// DELETE /users/me/identities/:provider
func (h *Handler) UnlinkIdentity(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserIDFromContext(c)
	if !ok {
		return common.UnauthorizedResponse(c, "")
	}
	provider := c.Params("provider")

	if err := h.service.UnlinkIdentity(c.Context(), userID, provider); err != nil {
		return common.HandleError(c, err)
	}

	return common.SuccessResponseWithMessage(c, "Identity unlinked successfully", nil)
}

// SetPassword sets a password for OAuth-only users
// POST /users/me/set-password
func (h *Handler) SetPassword(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserIDFromContext(c)
	if !ok {
		return common.UnauthorizedResponse(c, "")
	}

	req, err := common.ParseAndValidate[SetPasswordRequest](c)
	if err != nil {
		return nil
	}

	if err := h.service.SetPassword(c.Context(), userID, req.NewPassword); err != nil {
		return common.HandleError(c, err)
	}

	return common.SuccessResponseWithMessage(c, "Password set successfully", nil)
}

// toUtilsJWTConfig converts config.JWTConfig to utils.JWTConfig
func toUtilsJWTConfig(cfg *config.JWTConfig) *utils.JWTConfig {
	return &utils.JWTConfig{
		SecretKey:          cfg.SecretKey,
		AccessTokenExpiry:  cfg.AccessTokenExpiry,
		RefreshTokenExpiry: cfg.RefreshTokenExpiry,
		Issuer:             cfg.Issuer,
	}
}

