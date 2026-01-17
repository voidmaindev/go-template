package rbac

import (
	"time"

	"github.com/voidmaindev/go-template/internal/common"
	"github.com/voidmaindev/go-template/internal/common/filter"
)

// Role represents a role entity for metadata (Casbin handles actual permissions)
type Role struct {
	common.BaseModel
	Code        string `gorm:"size:50;not null;uniqueIndex" json:"code"`
	Name        string `gorm:"size:100;not null" json:"name"`
	Description string `gorm:"size:255" json:"description"`
	IsSystem    bool   `gorm:"not null;default:false" json:"is_system"`
}

// TableName returns the table name for the Role model
func (Role) TableName() string {
	return "rbac_roles"
}

// FilterConfig returns the filter configuration for Role
func (Role) FilterConfig() filter.Config {
	return filter.Config{
		TableName: "rbac_roles",
		Fields: map[string]filter.FieldConfig{
			"id":          {DBColumn: "id", Type: filter.TypeNumber, Operators: filter.NumberOps, Sortable: true},
			"code":        {DBColumn: "code", Type: filter.TypeString, Operators: filter.StringOps, Sortable: true},
			"name":        {DBColumn: "name", Type: filter.TypeString, Operators: filter.StringOps, Sortable: true},
			"description": {DBColumn: "description", Type: filter.TypeString, Operators: filter.StringOps, Sortable: true},
			"is_system":   {DBColumn: "is_system", Type: filter.TypeBool, Operators: filter.BoolOps, Sortable: true},
			"created_at":  {DBColumn: "created_at", Type: filter.TypeDate, Operators: filter.DateOps, Sortable: true},
			"updated_at":  {DBColumn: "updated_at", Type: filter.TypeDate, Operators: filter.DateOps, Sortable: true},
		},
	}
}

// ToResponse converts Role to a response DTO
func (r *Role) ToResponse() *RoleResponse {
	return &RoleResponse{
		ID:          r.ID,
		Code:        r.Code,
		Name:        r.Name,
		Description: r.Description,
		IsSystem:    r.IsSystem,
		CreatedAt:   r.CreatedAt,
		UpdatedAt:   r.UpdatedAt,
	}
}

// Permission represents a permission (domain + action pair)
type Permission struct {
	Domain  string   `json:"domain"`
	Actions []string `json:"actions"`
}

// RoleWithPermissions combines role metadata with its Casbin policies
type RoleWithPermissions struct {
	Role        *Role        `json:"role"`
	Permissions []Permission `json:"permissions"`
}

// UserRole represents the assignment of a role to a user (stored in Casbin)
type UserRole struct {
	UserID   uint      `json:"user_id"`
	RoleCode string    `json:"role_code"`
	RoleName string    `json:"role_name"`
	AssignedAt time.Time `json:"assigned_at,omitempty"`
}
