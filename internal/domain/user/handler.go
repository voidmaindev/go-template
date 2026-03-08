package user

import (
	"errors"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/voidmaindev/go-template/internal/common"
	"github.com/voidmaindev/go-template/internal/common/filter"
	"github.com/voidmaindev/go-template/internal/config"
	"github.com/voidmaindev/go-template/internal/domain/audit"
	"github.com/voidmaindev/go-template/internal/middleware"
	"github.com/voidmaindev/go-template/pkg/ptr"
	"github.com/voidmaindev/go-template/pkg/validator"
)

// Handler handles HTTP requests for users
type Handler struct {
	service     Service
	jwtConfig   *config.JWTConfig
	auditLogger audit.Logger
}

// NewHandler creates a new user handler
func NewHandler(service Service, jwtConfig *config.JWTConfig) *Handler {
	return &Handler{
		service:   service,
		jwtConfig: jwtConfig,
	}
}

// SetAuditLogger sets the audit logger (optional, for audit logging)
func (h *Handler) SetAuditLogger(logger audit.Logger) {
	h.auditLogger = logger
}

// logAudit safely logs an audit event (no-op if auditLogger is nil)
func (h *Handler) logAudit(c *fiber.Ctx, action string, userID *uint, success bool, details map[string]any) {
	if h.auditLogger == nil {
		return
	}
	h.auditLogger.LogAsync(c.Context(), &audit.AuditEntry{
		UserID:    userID,
		Action:    action,
		Resource:  "user",
		IP:        c.IP(),
		UserAgent: c.Get("User-Agent"),
		Success:   success,
		Details:   details,
	})
}

// Register handles user registration
// @Summary Register a new user
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body RegisterRequest true "Registration request"
// @Success 201 {object} common.Response{data=TokenResponse}
// @Failure 400 {object} common.Response
// @Failure 409 {object} common.Response
// @Router /auth/register [post]
func (h *Handler) Register(c *fiber.Ctx) error {
	var req RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return common.BadRequestResponse(c, "invalid request body")
	}

	// Validate request
	if errs := validator.Validate(&req); errs != nil {
		return common.ValidationErrorResponse(c, errs)
	}

	// Register user
	response, err := h.service.Register(c.Context(), &req)
	if err != nil {
		if errors.Is(err, ErrEmailExists) {
			return common.ConflictResponse(c, "email already exists")
		}
		return common.HandleError(c, err)
	}

	return common.CreatedResponse(c, response)
}

// Login handles user login
// @Summary Login user
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Login request"
// @Success 200 {object} common.Response{data=TokenResponse}
// @Failure 400 {object} common.Response
// @Failure 401 {object} common.Response
// @Failure 403 {object} common.Response "Email not verified (self-registered users only)"
// @Failure 429 {object} common.Response "Too many login attempts"
// @Router /auth/login [post]
func (h *Handler) Login(c *fiber.Ctx) error {
	var req LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return common.BadRequestResponse(c, "invalid request body")
	}

	// Validate request
	if errs := validator.Validate(&req); errs != nil {
		return common.ValidationErrorResponse(c, errs)
	}

	// Build login context with client IP and user agent
	loginCtx := &LoginContext{
		IP:        c.IP(),
		UserAgent: c.Get("User-Agent"),
	}

	// Login user
	response, err := h.service.Login(c.Context(), &req, loginCtx)
	if err != nil {
		// Log failed login attempt
		reason := "unknown"
		if errors.Is(err, ErrInvalidCredentials) {
			reason = "invalid_credentials"
		} else if errors.Is(err, ErrEmailNotVerified) {
			reason = "email_not_verified"
		} else if errors.Is(err, ErrTooManyLoginAttempts) {
			reason = "rate_limited"
		}
		h.logAudit(c, audit.ActionLoginFailed, nil, false, map[string]any{
			"email":  req.Email,
			"reason": reason,
		})

		if errors.Is(err, ErrInvalidCredentials) {
			return common.UnauthorizedResponse(c, "invalid email or password")
		}
		if errors.Is(err, ErrEmailNotVerified) {
			return common.ForbiddenResponse(c, "email address not verified")
		}
		if errors.Is(err, ErrTooManyLoginAttempts) {
			return common.TooManyRequestsResponse(c, "too many login attempts, please try again later")
		}
		return common.HandleError(c, err)
	}

	// Log successful login
	h.logAudit(c, audit.ActionLoginSuccess, ptr.To(response.User.ID), true, map[string]any{
		"email": req.Email,
	})

	return common.SuccessResponse(c, response)
}

// Logout handles user logout
// @Summary Logout user
// @Tags Auth
// @Security BearerAuth
// @Produce json
// @Success 200 {object} common.Response
// @Failure 401 {object} common.Response
// @Router /auth/logout [post]
func (h *Handler) Logout(c *fiber.Ctx) error {
	accessToken := middleware.GetTokenFromContext(c)
	if accessToken == "" {
		return common.UnauthorizedResponse(c, "no token provided")
	}

	userID, _ := middleware.GetUserIDFromContext(c)

	// Try to extract refresh token from body (optional)
	var req LogoutRequest
	_ = c.BodyParser(&req) // Ignore error - refresh token is optional

	if err := h.service.Logout(c.Context(), accessToken, req.RefreshToken); err != nil {
		return common.HandleError(c, err)
	}

	// Log logout
	h.logAudit(c, audit.ActionLogout, ptr.To(userID), true, nil)

	return common.SuccessResponseWithMessage(c, "logged out successfully", nil)
}

// RefreshToken handles token refresh
// @Summary Refresh access token
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body RefreshTokenRequest true "Refresh token request"
// @Success 200 {object} common.Response{data=TokenResponse}
// @Failure 400 {object} common.Response
// @Failure 401 {object} common.Response
// @Router /auth/refresh [post]
func (h *Handler) RefreshToken(c *fiber.Ctx) error {
	var req RefreshTokenRequest
	if err := c.BodyParser(&req); err != nil {
		return common.BadRequestResponse(c, "invalid request body")
	}

	// Validate request
	if errs := validator.Validate(&req); errs != nil {
		return common.ValidationErrorResponse(c, errs)
	}

	// Refresh token
	response, err := h.service.RefreshToken(c.Context(), req.RefreshToken)
	if err != nil {
		if errors.Is(err, ErrTokenInvalid) || errors.Is(err, ErrTokenBlacklisted) {
			return common.UnauthorizedResponse(c, "invalid or expired refresh token")
		}
		if errors.Is(err, ErrTokenRefreshUnavailable) {
			return common.ServiceUnavailableResponse(c, "token refresh temporarily unavailable")
		}
		return common.HandleError(c, err)
	}

	return common.SuccessResponse(c, response)
}

// GetMe handles getting current user
// @Summary Get current user
// @Tags Users
// @Security BearerAuth
// @Produce json
// @Success 200 {object} common.Response{data=UserResponse}
// @Failure 401 {object} common.Response
// @Router /users/me [get]
func (h *Handler) GetMe(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserIDFromContext(c)
	if !ok {
		return common.UnauthorizedResponse(c, "")
	}

	user, err := h.service.GetByID(c.Context(), userID)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return common.NotFoundResponse(c, "user")
		}
		return common.HandleError(c, err)
	}

	return common.SuccessResponse(c, user.ToResponse())
}

// UpdateMe handles updating current user
// @Summary Update current user
// @Tags Users
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body UpdateUserRequest true "Update request"
// @Success 200 {object} common.Response{data=UserResponse}
// @Failure 400 {object} common.Response
// @Failure 401 {object} common.Response
// @Router /users/me [put]
func (h *Handler) UpdateMe(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserIDFromContext(c)
	if !ok {
		return common.UnauthorizedResponse(c, "")
	}

	var req UpdateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return common.BadRequestResponse(c, "invalid request body")
	}

	// Validate request
	if errs := validator.Validate(&req); errs != nil {
		return common.ValidationErrorResponse(c, errs)
	}

	user, err := h.service.Update(c.Context(), userID, &req)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return common.NotFoundResponse(c, "user")
		}
		return common.HandleError(c, err)
	}

	return common.SuccessResponse(c, user.ToResponse())
}

// ChangePassword handles password change
// @Summary Change password
// @Tags Users
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body ChangePasswordRequest true "Change password request"
// @Success 200 {object} common.Response
// @Failure 400 {object} common.Response
// @Failure 401 {object} common.Response
// @Router /users/me/password [put]
func (h *Handler) ChangePassword(c *fiber.Ctx) error {
	// Add constant-time delay to prevent timing attacks
	start := time.Now()
	defer func() {
		elapsed := time.Since(start)
		if elapsed < h.jwtConfig.MinPasswordResponseTime {
			time.Sleep(h.jwtConfig.MinPasswordResponseTime - elapsed)
		}
	}()

	userID, ok := middleware.GetUserIDFromContext(c)
	if !ok {
		return common.UnauthorizedResponse(c, "")
	}

	var req ChangePasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return common.BadRequestResponse(c, "invalid request body")
	}

	// Validate request
	if errs := validator.Validate(&req); errs != nil {
		return common.ValidationErrorResponse(c, errs)
	}

	if err := h.service.ChangePassword(c.Context(), userID, &req); err != nil {
		switch {
		case errors.Is(err, ErrUserNotFound), errors.Is(err, ErrInvalidPassword):
			// Return generic error to prevent user enumeration
			return common.BadRequestResponse(c, "current password is incorrect")
		case errors.Is(err, ErrSamePassword):
			return common.BadRequestResponse(c, "new password must be different")
		default:
			return common.HandleError(c, err)
		}
	}

	// Log password change
	h.logAudit(c, audit.ActionPasswordChange, ptr.To(userID), true, nil)

	return common.SuccessResponseWithMessage(c, "password changed successfully", nil)
}

// GetByID handles getting user by ID
// @Summary Get user by ID
// @Tags Users
// @Security BearerAuth
// @Produce json
// @Param id path int true "User ID"
// @Success 200 {object} common.Response{data=UserResponse}
// @Failure 401 {object} common.Response
// @Failure 403 {object} common.Response
// @Failure 404 {object} common.Response
// @Router /users/{id} [get]
func (h *Handler) GetByID(c *fiber.Ctx) error {
	_, ok := middleware.GetUserIDFromContext(c)
	if !ok {
		return common.UnauthorizedResponse(c, "")
	}

	targetID, err := c.ParamsInt("id")
	if err != nil {
		return common.BadRequestResponse(c, "invalid user ID")
	}

	// Authorization is handled by RBAC middleware at route level
	// This endpoint requires user:read permission

	user, err := h.service.GetByID(c.Context(), uint(targetID))
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return common.NotFoundResponse(c, "user")
		}
		return common.HandleError(c, err)
	}

	return common.SuccessResponse(c, user.ToResponse())
}

// List handles listing all users with filtering and sorting
// @Summary List users (requires user:read permission)
// @Tags Users
// @Security BearerAuth
// @Produce json
// @Param page query int false "Page number"
// @Param page_size query int false "Page size"
// @Param name__contains query string false "Filter by name (contains)"
// @Param email__contains query string false "Filter by email (contains)"
// @Param sort query string false "Sort field"
// @Param order query string false "Sort order (asc/desc)"
// @Success 200 {object} common.Response{data=common.PaginatedResult[UserResponse]}
// @Failure 401 {object} common.Response
// @Failure 403 {object} common.Response
// @Router /users [get]
func (h *Handler) List(c *fiber.Ctx) error {
	// Authorization is handled by RBAC middleware at route level
	// This endpoint requires user:read permission

	params := filter.ParseFromQuery(c)

	result, err := h.service.ListFiltered(c.Context(), params)
	if err != nil {
		return common.HandleError(c, err)
	}

	// Convert to response DTOs
	responses := make([]UserResponse, len(result.Data))
	for i, user := range result.Data {
		responses[i] = *user.ToResponse()
	}

	return common.SuccessResponse(c, common.NewPaginatedResultFromFilter(responses, result.Total, params))
}

// Delete handles deleting a user
// @Summary Delete user
// @Tags Users
// @Security BearerAuth
// @Produce json
// @Param id path int true "User ID"
// @Success 200 {object} common.Response
// @Failure 401 {object} common.Response
// @Failure 403 {object} common.Response
// @Failure 404 {object} common.Response
// @Router /users/{id} [delete]
func (h *Handler) Delete(c *fiber.Ctx) error {
	// Authorization is handled by RequirePermission middleware at route level
	// which checks user:delete permission for all delete operations

	targetID, err := c.ParamsInt("id")
	if err != nil {
		return common.BadRequestResponse(c, "invalid user ID")
	}

	if err := h.service.Delete(c.Context(), uint(targetID)); err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return common.NotFoundResponse(c, "user")
		}
		return common.HandleError(c, err)
	}

	return common.DeletedResponse(c)
}
