package user

import (
	"github.com/voidmaindev/go-template/internal/common"
)

// Role represents user roles
type Role string

const (
	RoleUser  Role = "user"
	RoleAdmin Role = "admin"
)

// User represents a user entity
type User struct {
	common.BaseModel
	Email    string `gorm:"size:255;not null;uniqueIndex" json:"email"`
	Password string `gorm:"size:255;not null" json:"-"`
	Name     string `gorm:"size:100;not null" json:"name"`
	Role     Role   `gorm:"size:20;not null;default:'user'" json:"role"`
}

// TableName returns the table name for the User model
func (User) TableName() string {
	return "users"
}

// ToResponse converts User to a response DTO (without sensitive fields)
func (u *User) ToResponse() *UserResponse {
	return &UserResponse{
		ID:        u.ID,
		Email:     u.Email,
		Name:      u.Name,
		Role:      string(u.Role),
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}

// IsAdmin checks if the user has admin role
func (u *User) IsAdmin() bool {
	return u.Role == RoleAdmin
}
