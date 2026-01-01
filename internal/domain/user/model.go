package user

import (
	"github.com/voidmaindev/GoTemplate/internal/common"
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
