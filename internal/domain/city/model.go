package city

import (
	"github.com/voidmaindev/go-template/internal/common"
	"github.com/voidmaindev/go-template/internal/domain/country"
)

// City represents a city entity
type City struct {
	common.BaseModel
	Name      string          `gorm:"size:100;not null;index" json:"name"`
	CountryID uint            `gorm:"not null;index" json:"country_id"`
	Country   country.Country `gorm:"foreignKey:CountryID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT" json:"country,omitempty"`
}

// TableName returns the table name for the City model
func (City) TableName() string {
	return "cities"
}
