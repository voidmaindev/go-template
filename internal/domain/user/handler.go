package user

import (
	"errors"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/voidmaindev/go-template/internal/common"
	"github.com/voidmaindev/go-template/internal/common/filter"
	"github.com/voidmaindev/go-template/internal/middleware"
	"github.com/voidmaindev/go-template/pkg/validator"
)

// Minimum response time for sensitive operations to prevent timing attacks
const minPasswordChangeResponseTime = 200 * time.Millisecond

// Handler handles HTTP requests for users
type Handler struct {
	service Service
}

// NewHandler creates a new user handler
func NewHandler(service Service) *Handler {
	return &Handler{
		service: service,
	}
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
		if errors.Is(err, ErrEmailAlreadyExists) {
			return common.ConflictResponse(c, "email already exists")
		}
		return common.InternalServerErrorResponse(c)
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

	// Login user
	response, err := h.service.Login(c.Context(), &req)
	if err != nil {
		if errors.Is(err, common.ErrInvalidCredentials) {
			return common.UnauthorizedResponse(c, "invalid email or password")
		}
		return common.InternalServerErrorResponse(c)
	}

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
	token := middleware.GetTokenFromContext(c)
	if token == "" {
		return common.UnauthorizedResponse(c, "no token provided")
	}

	if err := h.service.Logout(c.Context(), token); err != nil {
		return common.InternalServerErrorResponse(c)
	}

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
		if errors.Is(err, common.ErrTokenInvalid) || errors.Is(err, common.ErrTokenBlacklisted) {
			return common.UnauthorizedResponse(c, "invalid or expired refresh token")
		}
		if errors.Is(err, ErrTokenRefreshUnavailable) {
			return common.ServiceUnavailableResponse(c, "token refresh temporarily unavailable")
		}
		return common.InternalServerErrorResponse(c)
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
		return common.InternalServerErrorResponse(c)
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
		return common.InternalServerErrorResponse(c)
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
		if elapsed < minPasswordChangeResponseTime {
			time.Sleep(minPasswordChangeResponseTime - elapsed)
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
			return common.InternalServerErrorResponse(c)
		}
	}

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
	currentUserID, ok := middleware.GetUserIDFromContext(c)
	if !ok {
		return common.UnauthorizedResponse(c, "")
	}

	targetID, err := c.ParamsInt("id")
	if err != nil {
		return common.BadRequestResponse(c, "invalid user ID")
	}

	// Authorization: only allow self-view or admin
	if uint(targetID) != currentUserID && !middleware.IsAdmin(c) {
		return common.ForbiddenResponse(c, "cannot view other users")
	}

	user, err := h.service.GetByID(c.Context(), uint(targetID))
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return common.NotFoundResponse(c, "user")
		}
		return common.InternalServerErrorResponse(c)
	}

	return common.SuccessResponse(c, user.ToResponse())
}

// List handles listing all users with filtering and sorting
// @Summary List users (admin only)
// @Tags Users
// @Security BearerAuth
// @Produce json
// @Param page query int false "Page number"
// @Param page_size query int false "Page size"
// @Param name__contains query string false "Filter by name (contains)"
// @Param email__contains query string false "Filter by email (contains)"
// @Param role query string false "Filter by role (eq)"
// @Param sort query string false "Sort field"
// @Param order query string false "Sort order (asc/desc)"
// @Success 200 {object} common.Response{data=common.FilteredResult[UserResponse]}
// @Failure 401 {object} common.Response
// @Failure 403 {object} common.Response
// @Router /users [get]
func (h *Handler) List(c *fiber.Ctx) error {
	// Authorization: only admin can list all users
	if !middleware.IsAdmin(c) {
		return common.ForbiddenResponse(c, "admin access required")
	}

	params := filter.ParseFromQuery(c)

	result, err := h.service.ListFiltered(c.Context(), params)
	if err != nil {
		return common.InternalServerErrorResponse(c)
	}

	// Convert to response DTOs
	responses := make([]UserResponse, len(result.Data))
	for i, user := range result.Data {
		responses[i] = *user.ToResponse()
	}

	return common.SuccessResponse(c, common.NewFilteredResult(responses, result.Total, params))
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
	// Get the current authenticated user
	currentUserID, ok := middleware.GetUserIDFromContext(c)
	if !ok {
		return common.UnauthorizedResponse(c, "")
	}

	targetID, err := c.ParamsInt("id")
	if err != nil {
		return common.BadRequestResponse(c, "invalid user ID")
	}

	// Authorization: only allow self-delete or admin
	if uint(targetID) != currentUserID && !middleware.IsAdmin(c) {
		return common.ForbiddenResponse(c, "cannot delete other users")
	}

	if err := h.service.Delete(c.Context(), uint(targetID)); err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return common.NotFoundResponse(c, "user")
		}
		return common.InternalServerErrorResponse(c)
	}

	return common.DeletedResponse(c)
}
