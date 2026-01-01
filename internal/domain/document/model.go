package document

import (
	"time"

	"github.com/voidmaindev/GoTemplate/internal/common"
	"github.com/voidmaindev/GoTemplate/internal/domain/city"
	"github.com/voidmaindev/GoTemplate/internal/domain/item"
)

// Document represents a document entity (like an order/invoice)
type Document struct {
	common.BaseModel
	Code         string         `gorm:"size:50;not null;uniqueIndex" json:"code"`
	CityID       uint           `gorm:"not null;index" json:"city_id"`
	City         city.City      `gorm:"foreignKey:CityID" json:"city,omitempty"`
	DocumentDate time.Time      `gorm:"not null" json:"document_date"`
	TotalAmount  int64          `gorm:"not null;default:0" json:"total_amount"` // Total in cents
	Items        []DocumentItem `gorm:"foreignKey:DocumentID" json:"items,omitempty"`
}

// TableName returns the table name for the Document model
func (Document) TableName() string {
	return "documents"
}

// CalculateTotal calculates the total amount from items
func (d *Document) CalculateTotal() int64 {
	var total int64
	for _, item := range d.Items {
		total += item.Price * int64(item.Quantity)
	}
	return total
}

// DocumentItem represents a line item in a document
type DocumentItem struct {
	common.BaseModel
	DocumentID uint      `gorm:"not null;index" json:"document_id"`
	ItemID     uint      `gorm:"not null;index" json:"item_id"`
	Item       item.Item `gorm:"foreignKey:ItemID" json:"item,omitempty"`
	Quantity   int       `gorm:"not null" json:"quantity"`
	Price      int64     `gorm:"not null" json:"price"` // Price per unit in cents
}

// TableName returns the table name for the DocumentItem model
func (DocumentItem) TableName() string {
	return "document_items"
}

// GetLineTotal returns the total for this line item
func (di *DocumentItem) GetLineTotal() int64 {
	return di.Price * int64(di.Quantity)
}
