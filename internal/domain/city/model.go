package city

import (
	"github.com/voidmaindev/go-template/internal/common"
	"github.com/voidmaindev/go-template/internal/common/filter"
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

// FilterConfig returns the filter configuration for City
func (City) FilterConfig() filter.Config {
	return filter.Config{
		TableName: "cities",
		Fields: map[string]filter.FieldConfig{
			"id":         {DBColumn: "id", Type: filter.TypeNumber, Operators: filter.NumberOps, Sortable: true},
			"name":       {DBColumn: "name", Type: filter.TypeString, Operators: filter.StringOps, Sortable: true},
			"country_id": {DBColumn: "country_id", Type: filter.TypeNumber, Operators: filter.NumberOps, Sortable: true},
			"created_at": {DBColumn: "created_at", Type: filter.TypeDate, Operators: filter.DateOps, Sortable: true},
			"updated_at": {DBColumn: "updated_at", Type: filter.TypeDate, Operators: filter.DateOps, Sortable: true},
			"country":    {Relation: "Country", RelationFK: "country_id"},
		},
	}
}
