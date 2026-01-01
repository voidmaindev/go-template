package country

import (
	"time"
)

// CreateCountryRequest represents the create country request
type CreateCountryRequest struct {
	Name string `json:"name" validate:"required,min=1,max=100"`
	Code string `json:"code" validate:"required,len=3,alpha"`
}

// UpdateCountryRequest represents the update country request
type UpdateCountryRequest struct {
	Name *string `json:"name" validate:"omitempty,min=1,max=100"`
	Code *string `json:"code" validate:"omitempty,len=3,alpha"`
}

// CountryResponse represents the country response
type CountryResponse struct {
	ID        uint      `json:"id"`
	Name      string    `json:"name"`
	Code      string    `json:"code"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ToResponse converts Country to CountryResponse
func (c *Country) ToResponse() *CountryResponse {
	return &CountryResponse{
		ID:        c.ID,
		Name:      c.Name,
		Code:      c.Code,
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
	}
}
