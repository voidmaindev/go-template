package user

import (
	"github.com/voidmaindev/go-template/internal/common"
	"github.com/voidmaindev/go-template/internal/common/filter"
)

// User represents a user entity
type User struct {
	common.BaseModel
	Email    string `gorm:"size:255;not null;uniqueIndex" json:"email"`
	Password string `gorm:"size:255;not null" json:"-"`
	Name     string `gorm:"size:100;not null" json:"name"`
}

// TableName returns the table name for the User model
func (User) TableName() string {
	return "users"
}

// FilterConfig returns the filter configuration for User
func (User) FilterConfig() filter.Config {
	return filter.Config{
		TableName: "users",
		Fields: map[string]filter.FieldConfig{
			"id":         {DBColumn: "id", Type: filter.TypeNumber, Operators: filter.NumberOps, Sortable: true},
			"email":      {DBColumn: "email", Type: filter.TypeString, Operators: filter.StringOps, Sortable: true},
			"name":       {DBColumn: "name", Type: filter.TypeString, Operators: filter.StringOps, Sortable: true},
			"created_at": {DBColumn: "created_at", Type: filter.TypeDate, Operators: filter.DateOps, Sortable: true},
			"updated_at": {DBColumn: "updated_at", Type: filter.TypeDate, Operators: filter.DateOps, Sortable: true},
		},
	}
}

// ToResponse converts User to a response DTO (without sensitive fields)
func (u *User) ToResponse() *UserResponse {
	return &UserResponse{
		ID:        u.ID,
		Email:     u.Email,
		Name:      u.Name,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}
