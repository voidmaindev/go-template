package city

import (
	"time"

	"github.com/voidmaindev/GoTemplate/internal/domain/country"
)

// CreateCityRequest represents the create city request
type CreateCityRequest struct {
	Name      string `json:"name" validate:"required,min=1,max=100"`
	CountryID uint   `json:"country_id" validate:"required,gt=0"`
}

// UpdateCityRequest represents the update city request
type UpdateCityRequest struct {
	Name      *string `json:"name" validate:"omitempty,min=1,max=100"`
	CountryID *uint   `json:"country_id" validate:"omitempty,gt=0"`
}

// CityResponse represents the city response
type CityResponse struct {
	ID        uint                     `json:"id"`
	Name      string                   `json:"name"`
	CountryID uint                     `json:"country_id"`
	Country   *country.CountryResponse `json:"country,omitempty"`
	CreatedAt time.Time                `json:"created_at"`
	UpdatedAt time.Time                `json:"updated_at"`
}

// ToResponse converts City to CityResponse
func (c *City) ToResponse() *CityResponse {
	resp := &CityResponse{
		ID:        c.ID,
		Name:      c.Name,
		CountryID: c.CountryID,
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
	}

	// Include country if loaded
	if c.Country.ID != 0 {
		resp.Country = c.Country.ToResponse()
	}

	return resp
}
