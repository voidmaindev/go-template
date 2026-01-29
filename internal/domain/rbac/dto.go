package rbac

import "time"

// Request DTOs

// CreateRoleRequest is the request body for creating a new role
type CreateRoleRequest struct {
	Code        string             `json:"code" validate:"required,min=2,max=50,alphanum"`
	Name        string             `json:"name" validate:"required,min=2,max=100"`
	Description string             `json:"description" validate:"max=255"`
	Permissions []PermissionInput  `json:"permissions" validate:"required,min=1,dive"`
}

// PermissionInput represents a permission input in API requests
type PermissionInput struct {
	Domain  string   `json:"domain" validate:"required,min=1,max=50"`
	Actions []string `json:"actions" validate:"required,min=1,dive,oneof=read create update delete"`
}

// UpdateRolePermissionsRequest is the request body for updating role permissions
type UpdateRolePermissionsRequest struct {
	Permissions []PermissionInput `json:"permissions" validate:"required,min=1,dive"`
}

// AssignRoleRequest is the request body for assigning a role to a user
type AssignRoleRequest struct {
	RoleCode string `json:"role_code" validate:"required,min=2,max=50"`
}

// Response DTOs

// RoleResponse is the response DTO for a role
type RoleResponse struct {
	ID          uint      `json:"id"`
	Code        string    `json:"code"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	IsSystem    bool      `json:"is_system"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// RoleWithPermissionsResponse is the response DTO for a role with permissions
type RoleWithPermissionsResponse struct {
	ID          uint                 `json:"id"`
	Code        string               `json:"code"`
	Name        string               `json:"name"`
	Description string               `json:"description"`
	IsSystem    bool                 `json:"is_system"`
	Permissions []PermissionResponse `json:"permissions"`
	CreatedAt   time.Time            `json:"created_at"`
	UpdatedAt   time.Time            `json:"updated_at"`
}

// PermissionResponse represents a permission in API responses
type PermissionResponse struct {
	Domain  string   `json:"domain"`
	Actions []string `json:"actions"`
}

// UserRolesResponse is the response DTO for a user's roles
type UserRolesResponse struct {
	UserID uint               `json:"user_id"`
	Roles  []UserRoleResponse `json:"roles"`
}

// UserRoleResponse represents a role assigned to a user
type UserRoleResponse struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

// DomainResponse is the response DTO for a domain
type DomainResponse struct {
	Name        string `json:"name"`
	IsProtected bool   `json:"is_protected"`
}

// DomainsResponse is the response DTO for listing domains
type DomainsResponse struct {
	Domains []DomainResponse `json:"domains"`
}

// ActionsResponse is the response DTO for listing actions
type ActionsResponse struct {
	Actions []string `json:"actions"`
}
