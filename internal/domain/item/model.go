package item

import (
	"github.com/voidmaindev/go-template/internal/common"
	"github.com/voidmaindev/go-template/internal/common/filter"
)

// Item represents an item entity
type Item struct {
	common.BaseModel
	Name        string `gorm:"size:200;not null" json:"name"`
	Description string `gorm:"type:text" json:"description"`
	Price       int64  `gorm:"not null;default:0" json:"price"` // Price in cents (1999 = $19.99)
}

// TableName returns the table name for the Item model
func (Item) TableName() string {
	return "items"
}

// FilterConfig returns the filter configuration for Item
func (Item) FilterConfig() filter.Config {
	return filter.Config{
		TableName: "items",
		Fields: map[string]filter.FieldConfig{
			"id":          {DBColumn: "id", Type: filter.TypeNumber, Operators: filter.NumberOps, Sortable: true},
			"name":        {DBColumn: "name", Type: filter.TypeString, Operators: filter.StringOps, Sortable: true},
			"description": {DBColumn: "description", Type: filter.TypeString, Operators: filter.StringOps, Sortable: false},
			"price":       {DBColumn: "price", Type: filter.TypeNumber, Operators: filter.NumberOps, Sortable: true},
			"created_at":  {DBColumn: "created_at", Type: filter.TypeDate, Operators: filter.DateOps, Sortable: true},
			"updated_at":  {DBColumn: "updated_at", Type: filter.TypeDate, Operators: filter.DateOps, Sortable: true},
		},
	}
}
