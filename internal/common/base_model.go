package common

import (
	"time"

	"gorm.io/gorm"
)

// BaseModel contains common fields for all entities
type BaseModel struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

// IsNew returns true if the model has not been saved to the database
func (m *BaseModel) IsNew() bool {
	return m.ID == 0
}

// IsDeleted returns true if the model has been soft-deleted
func (m *BaseModel) IsDeleted() bool {
	return m.DeletedAt.Valid
}
