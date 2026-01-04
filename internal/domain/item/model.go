package item

import (
	"github.com/voidmaindev/go-template/internal/common"
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
