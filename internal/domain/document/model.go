package document

import (
	"time"

	"github.com/voidmaindev/go-template/internal/common"
	"github.com/voidmaindev/go-template/internal/common/filter"
	"github.com/voidmaindev/go-template/internal/domain/city"
	"github.com/voidmaindev/go-template/internal/domain/item"
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

// FilterConfig returns the filter configuration for Document
func (Document) FilterConfig() filter.Config {
	return filter.Config{
		TableName: "documents",
		Fields: map[string]filter.FieldConfig{
			"id":            {DBColumn: "id", Type: filter.TypeNumber, Operators: filter.NumberOps, Sortable: true},
			"code":          {DBColumn: "code", Type: filter.TypeString, Operators: filter.StringOps, Sortable: true},
			"city_id":       {DBColumn: "city_id", Type: filter.TypeNumber, Operators: filter.NumberOps, Sortable: true},
			"document_date": {DBColumn: "document_date", Type: filter.TypeDate, Operators: filter.DateOps, Sortable: true},
			"total_amount":  {DBColumn: "total_amount", Type: filter.TypeNumber, Operators: filter.NumberOps, Sortable: true},
			"created_at":    {DBColumn: "created_at", Type: filter.TypeDate, Operators: filter.DateOps, Sortable: true},
			"updated_at":    {DBColumn: "updated_at", Type: filter.TypeDate, Operators: filter.DateOps, Sortable: true},
			"city":          {Relation: "City", RelationFK: "city_id"},
		},
		AllowedRelationFields: map[string][]string{
			"city": {"id", "name", "country_id"}, // Whitelist: only these fields can be filtered/sorted
		},
	}
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

// FilterConfig returns the filter configuration for DocumentItem
func (DocumentItem) FilterConfig() filter.Config {
	return filter.Config{
		TableName: "document_items",
		Fields: map[string]filter.FieldConfig{
			"id":          {DBColumn: "id", Type: filter.TypeNumber, Operators: filter.NumberOps, Sortable: true},
			"document_id": {DBColumn: "document_id", Type: filter.TypeNumber, Operators: filter.NumberOps, Sortable: true},
			"item_id":     {DBColumn: "item_id", Type: filter.TypeNumber, Operators: filter.NumberOps, Sortable: true},
			"quantity":    {DBColumn: "quantity", Type: filter.TypeNumber, Operators: filter.NumberOps, Sortable: true},
			"price":       {DBColumn: "price", Type: filter.TypeNumber, Operators: filter.NumberOps, Sortable: true},
			"created_at":  {DBColumn: "created_at", Type: filter.TypeDate, Operators: filter.DateOps, Sortable: true},
			"updated_at":  {DBColumn: "updated_at", Type: filter.TypeDate, Operators: filter.DateOps, Sortable: true},
			"item":        {Relation: "Item", RelationFK: "item_id"},
		},
		AllowedRelationFields: map[string][]string{
			"item": {"id", "name", "code", "price"}, // Whitelist: only these fields can be filtered/sorted
		},
	}
}

// GetLineTotal returns the total for this line item
func (di *DocumentItem) GetLineTotal() int64 {
	return di.Price * int64(di.Quantity)
}
