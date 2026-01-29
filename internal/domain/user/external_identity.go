package user

import (
	"time"

	"github.com/voidmaindev/go-template/internal/common"
)

// ExternalIdentity represents an OAuth provider identity linked to a user
type ExternalIdentity struct {
	common.BaseModel
	UserID     uint       `gorm:"not null;index" json:"user_id"`
	Provider   string     `gorm:"size:50;not null;index" json:"provider"` // google, facebook, apple
	ProviderID string     `gorm:"size:255;not null" json:"provider_id"`
	Email      string     `gorm:"size:255" json:"email"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty"`
	User       User       `gorm:"foreignKey:UserID" json:"-"`
}

// TableName returns the table name for the ExternalIdentity model
func (ExternalIdentity) TableName() string {
	return "external_identities"
}

// ExternalIdentityResponse represents the response DTO for external identities
type ExternalIdentityResponse struct {
	ID        uint       `json:"id"`
	Provider  string     `json:"provider"`
	Email     string     `json:"email"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}

// ToResponse converts ExternalIdentity to a response DTO
func (e *ExternalIdentity) ToResponse() *ExternalIdentityResponse {
	return &ExternalIdentityResponse{
		ID:        e.ID,
		Provider:  e.Provider,
		Email:     e.Email,
		ExpiresAt: e.ExpiresAt,
		CreatedAt: e.CreatedAt,
	}
}
