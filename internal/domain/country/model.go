package country

import (
	"github.com/voidmaindev/go-template/internal/common"
	"github.com/voidmaindev/go-template/internal/common/filter"
)

// Country represents a country entity
type Country struct {
	common.BaseModel
	Name string `gorm:"size:100;not null;uniqueIndex" json:"name"`
	Code string `gorm:"size:3;not null;uniqueIndex" json:"code"` // ISO 3166-1 alpha-3
}

// TableName returns the table name for the Country model
func (Country) TableName() string {
	return "countries"
}

// FilterConfig returns the filter configuration for Country
func (Country) FilterConfig() filter.Config {
	return filter.Config{
		TableName: "countries",
		Fields: map[string]filter.FieldConfig{
			"id":         {DBColumn: "id", Type: filter.TypeNumber, Operators: filter.NumberOps, Sortable: true},
			"name":       {DBColumn: "name", Type: filter.TypeString, Operators: filter.StringOps, Sortable: true},
			"code":       {DBColumn: "code", Type: filter.TypeString, Operators: filter.StringOps, Sortable: true},
			"created_at": {DBColumn: "created_at", Type: filter.TypeDate, Operators: filter.DateOps, Sortable: true},
			"updated_at": {DBColumn: "updated_at", Type: filter.TypeDate, Operators: filter.DateOps, Sortable: true},
		},
	}
}
