package rbac

import (
	"context"
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/voidmaindev/go-template/internal/common"
	"github.com/voidmaindev/go-template/internal/common/filter"
	"github.com/voidmaindev/go-template/internal/middleware"
	"github.com/voidmaindev/go-template/pkg/ptr"
)

// AuditEntry mirrors audit.AuditEntry to avoid import cycle
type AuditEntry struct {
	UserID     *uint
	Action     string
	Resource   string
	ResourceID *uint
	IP         string
	UserAgent  string
	Success    bool
	Details    map[string]any
}

// AuditLogger is a minimal interface for audit logging (mirrors audit.Logger)
type AuditLogger interface {
	LogAsync(ctx context.Context, entry *AuditEntry)
}

// Handler handles RBAC HTTP requests
type Handler struct {
	service     Service
	auditLogger AuditLogger
}

// NewHandler creates a new RBAC handler
func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// SetAuditLogger sets the audit logger (optional, for audit logging)
func (h *Handler) SetAuditLogger(logger AuditLogger) {
	h.auditLogger = logger
}

// logAudit safely logs an audit event (no-op if auditLogger is nil)
func (h *Handler) logAudit(c *fiber.Ctx, action string, resourceID *uint, success bool, details map[string]any) {
	if h.auditLogger == nil {
		return
	}
	userID, _ := middleware.GetUserIDFromContext(c)
	h.auditLogger.LogAsync(c.Context(), &AuditEntry{
		UserID:     ptr.To(userID),
		Action:     action,
		Resource:   "rbac",
		ResourceID: resourceID,
		IP:         c.IP(),
		UserAgent:  c.Get("User-Agent"),
		Success:    success,
		Details:    details,
	})
}

// ListRoles lists all roles
// @Summary List all roles
// @Tags RBAC
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} common.PaginatedResult[RoleResponse]
// @Failure 401 {object} common.Response
// @Failure 500 {object} common.Response
// @Router /api/v1/rbac/roles [get]
func (h *Handler) ListRoles(c *fiber.Ctx) error {
	params := filter.ParseFromQuery(c)

	result, err := h.service.ListRoles(c.Context(), params)
	if err != nil {
		return common.InternalServerErrorResponse(c)
	}

	// Convert to response DTOs
	responses := make([]RoleResponse, len(result.Data))
	for i, role := range result.Data {
		responses[i] = *role.ToResponse()
	}

	return common.SuccessResponse(c, common.NewPaginatedResultFromFilter(responses, result.Total, params))
}

// CreateRole creates a new role
// @Summary Create a new role
// @Tags RBAC
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CreateRoleRequest true "Create role request"
// @Success 201 {object} RoleResponse
// @Failure 400 {object} common.Response
// @Failure 401 {object} common.Response
// @Failure 409 {object} common.Response
// @Failure 500 {object} common.Response
// @Router /api/v1/rbac/roles [post]
func (h *Handler) CreateRole(c *fiber.Ctx) error {
	req, err := common.ParseAndValidate[CreateRoleRequest](c)
	if err != nil {
		return nil
	}

	role, err := h.service.CreateRole(c.Context(), req)
	if err != nil {
		if errors.Is(err, ErrRoleCodeExists) {
			return common.ConflictResponse(c, "role code already exists")
		}
		if errors.Is(err, ErrInvalidDomain) {
			return common.BadRequestResponse(c, "invalid domain")
		}
		if errors.Is(err, ErrInvalidAction) {
			return common.BadRequestResponse(c, "invalid action")
		}
		return common.InternalServerErrorResponse(c)
	}

	return common.CreatedResponse(c, role.ToResponse())
}

// GetRole gets a role by code with permissions
// @Summary Get a role by code
// @Tags RBAC
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param code path string true "Role code"
// @Success 200 {object} RoleWithPermissionsResponse
// @Failure 401 {object} common.Response
// @Failure 404 {object} common.Response
// @Failure 500 {object} common.Response
// @Router /api/v1/rbac/roles/{code} [get]
func (h *Handler) GetRole(c *fiber.Ctx) error {
	code := c.Params("code")
	if code == "" {
		return common.BadRequestResponse(c, "role code is required")
	}

	roleWithPerms, err := h.service.GetRoleByCode(c.Context(), code)
	if err != nil {
		if errors.Is(err, ErrRoleNotFound) {
			return common.NotFoundResponse(c, "role")
		}
		return common.InternalServerErrorResponse(c)
	}

	// Convert permissions to response format
	permResponses := make([]PermissionResponse, len(roleWithPerms.Permissions))
	for i, p := range roleWithPerms.Permissions {
		permResponses[i] = PermissionResponse{
			Domain:  p.Domain,
			Actions: p.Actions,
		}
	}

	response := RoleWithPermissionsResponse{
		ID:          roleWithPerms.Role.ID,
		Code:        roleWithPerms.Role.Code,
		Name:        roleWithPerms.Role.Name,
		Description: roleWithPerms.Role.Description,
		IsSystem:    roleWithPerms.Role.IsSystem,
		Permissions: permResponses,
		CreatedAt:   roleWithPerms.Role.CreatedAt,
		UpdatedAt:   roleWithPerms.Role.UpdatedAt,
	}

	return common.SuccessResponse(c, response)
}

// UpdateRolePermissions updates the permissions of a role
// @Summary Update role permissions
// @Tags RBAC
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param code path string true "Role code"
// @Param request body UpdateRolePermissionsRequest true "Update permissions request"
// @Success 200 {object} RoleWithPermissionsResponse
// @Failure 400 {object} common.Response
// @Failure 401 {object} common.Response
// @Failure 403 {object} common.Response
// @Failure 404 {object} common.Response
// @Failure 500 {object} common.Response
// @Router /api/v1/rbac/roles/{code}/permissions [put]
func (h *Handler) UpdateRolePermissions(c *fiber.Ctx) error {
	code := c.Params("code")
	if code == "" {
		return common.BadRequestResponse(c, "role code is required")
	}

	req, err := common.ParseAndValidate[UpdateRolePermissionsRequest](c)
	if err != nil {
		return nil
	}

	roleWithPerms, err := h.service.UpdateRolePermissions(c.Context(), code, req)
	if err != nil {
		if errors.Is(err, ErrRoleNotFound) {
			return common.NotFoundResponse(c, "role")
		}
		if errors.Is(err, ErrSystemRoleCannotBeModified) {
			return common.ForbiddenResponse(c, "system role permissions cannot be modified")
		}
		if errors.Is(err, ErrInvalidDomain) {
			return common.BadRequestResponse(c, "invalid domain")
		}
		if errors.Is(err, ErrInvalidAction) {
			return common.BadRequestResponse(c, "invalid action")
		}
		return common.InternalServerErrorResponse(c)
	}

	// Convert permissions to response format
	permResponses := make([]PermissionResponse, len(roleWithPerms.Permissions))
	for i, p := range roleWithPerms.Permissions {
		permResponses[i] = PermissionResponse{
			Domain:  p.Domain,
			Actions: p.Actions,
		}
	}

	response := RoleWithPermissionsResponse{
		ID:          roleWithPerms.Role.ID,
		Code:        roleWithPerms.Role.Code,
		Name:        roleWithPerms.Role.Name,
		Description: roleWithPerms.Role.Description,
		IsSystem:    roleWithPerms.Role.IsSystem,
		Permissions: permResponses,
		CreatedAt:   roleWithPerms.Role.CreatedAt,
		UpdatedAt:   roleWithPerms.Role.UpdatedAt,
	}

	return common.SuccessResponse(c, response)
}

// DeleteRole deletes a role
// @Summary Delete a role
// @Tags RBAC
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param code path string true "Role code"
// @Success 204
// @Failure 401 {object} common.Response
// @Failure 403 {object} common.Response
// @Failure 404 {object} common.Response
// @Failure 500 {object} common.Response
// @Router /api/v1/rbac/roles/{code} [delete]
func (h *Handler) DeleteRole(c *fiber.Ctx) error {
	code := c.Params("code")
	if code == "" {
		return common.BadRequestResponse(c, "role code is required")
	}

	err := h.service.DeleteRole(c.Context(), code)
	if err != nil {
		if errors.Is(err, ErrRoleNotFound) {
			return common.NotFoundResponse(c, "role")
		}
		if errors.Is(err, ErrSystemRoleCannotBeDeleted) {
			return common.ForbiddenResponse(c, "system roles cannot be deleted")
		}
		return common.InternalServerErrorResponse(c)
	}

	return common.DeletedResponse(c)
}

// GetUserRoles gets the roles assigned to a user
// @Summary Get user roles
// @Tags RBAC
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "User ID"
// @Success 200 {object} UserRolesResponse
// @Failure 401 {object} common.Response
// @Failure 500 {object} common.Response
// @Router /api/v1/rbac/users/{id}/roles [get]
func (h *Handler) GetUserRoles(c *fiber.Ctx) error {
	userID, err := common.ParseID(c, "id", "user")
	if err != nil {
		return nil
	}

	roles, err := h.service.GetUserRoles(c.Context(), userID)
	if err != nil {
		return common.InternalServerErrorResponse(c)
	}

	return common.SuccessResponse(c, UserRolesResponse{
		UserID: userID,
		Roles:  roles,
	})
}

// AssignRole assigns a role to a user
// @Summary Assign role to user
// @Tags RBAC
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "User ID"
// @Param request body AssignRoleRequest true "Assign role request"
// @Success 200 {object} common.Response
// @Failure 400 {object} common.Response
// @Failure 401 {object} common.Response
// @Failure 404 {object} common.Response
// @Failure 409 {object} common.Response
// @Failure 500 {object} common.Response
// @Router /api/v1/rbac/users/{id}/roles [post]
func (h *Handler) AssignRole(c *fiber.Ctx) error {
	userID, err := common.ParseID(c, "id", "user")
	if err != nil {
		return nil
	}

	req, err := common.ParseAndValidate[AssignRoleRequest](c)
	if err != nil {
		return nil
	}

	err = h.service.AssignRole(c.Context(), userID, req.RoleCode)
	if err != nil {
		if errors.Is(err, ErrRoleNotFound) {
			return common.NotFoundResponse(c, "role")
		}
		if errors.Is(err, ErrRoleAlreadyAssigned) {
			return common.ConflictResponse(c, "role is already assigned to this user")
		}
		return common.InternalServerErrorResponse(c)
	}

	// Log role assignment
	h.logAudit(c, "role_assigned", ptr.To(userID), true, map[string]any{
		"role": req.RoleCode,
	})

	return common.SuccessResponseWithMessage(c, "role assigned successfully", nil)
}

// RemoveRole removes a role from a user
// @Summary Remove role from user
// @Tags RBAC
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "User ID"
// @Param code path string true "Role code"
// @Success 204
// @Failure 400 {object} common.Response
// @Failure 401 {object} common.Response
// @Failure 403 {object} common.Response
// @Failure 404 {object} common.Response
// @Failure 500 {object} common.Response
// @Router /api/v1/rbac/users/{id}/roles/{code} [delete]
func (h *Handler) RemoveRole(c *fiber.Ctx) error {
	userID, err := common.ParseID(c, "id", "user")
	if err != nil {
		return nil
	}

	code := c.Params("code")
	if code == "" {
		return common.BadRequestResponse(c, "role code is required")
	}

	err = h.service.RemoveRole(c.Context(), userID, code)
	if err != nil {
		if errors.Is(err, ErrRoleNotAssigned) {
			return common.NotFoundResponse(c, "role assignment")
		}
		if errors.Is(err, ErrCannotRemoveLastAdmin) {
			return common.ForbiddenResponse(c, "cannot remove the last admin")
		}
		return common.InternalServerErrorResponse(c)
	}

	// Log role removal
	h.logAudit(c, "role_removed", ptr.To(userID), true, map[string]any{
		"role": code,
	})

	return common.DeletedResponse(c)
}

// GetDomains returns all available domains
// @Summary Get available domains
// @Tags RBAC
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} DomainsResponse
// @Failure 401 {object} common.Response
// @Router /api/v1/rbac/domains [get]
func (h *Handler) GetDomains(c *fiber.Ctx) error {
	domains := h.service.GetDomains(c.Context())
	return common.SuccessResponse(c, DomainsResponse{Domains: domains})
}

// GetActions returns all available actions
// @Summary Get available actions
// @Tags RBAC
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} ActionsResponse
// @Failure 401 {object} common.Response
// @Router /api/v1/rbac/actions [get]
func (h *Handler) GetActions(c *fiber.Ctx) error {
	actions := h.service.GetActions(c.Context())
	return common.SuccessResponse(c, ActionsResponse{Actions: actions})
}
