package country

import (
	"github.com/voidmaindev/GoTemplate/internal/common"
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
