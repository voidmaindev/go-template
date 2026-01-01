package document

import (
	"time"

	"github.com/voidmaindev/GoTemplate/internal/domain/city"
	"github.com/voidmaindev/GoTemplate/internal/domain/item"
)

// CreateDocumentRequest represents the create document request
type CreateDocumentRequest struct {
	Code         string                      `json:"code" validate:"required,min=1,max=50"`
	CityID       uint                        `json:"city_id" validate:"required,gt=0"`
	DocumentDate time.Time                   `json:"document_date" validate:"required"`
	Items        []CreateDocumentItemRequest `json:"items" validate:"required,min=1,dive"`
}

// CreateDocumentItemRequest represents a line item in create request
type CreateDocumentItemRequest struct {
	ItemID   uint  `json:"item_id" validate:"required,gt=0"`
	Quantity int   `json:"quantity" validate:"required,gt=0"`
	Price    int64 `json:"price" validate:"required,gte=0"`
}

// UpdateDocumentRequest represents the update document request
type UpdateDocumentRequest struct {
	Code         *string    `json:"code" validate:"omitempty,min=1,max=50"`
	CityID       *uint      `json:"city_id" validate:"omitempty,gt=0"`
	DocumentDate *time.Time `json:"document_date"`
}

// AddDocumentItemRequest represents the add item request
type AddDocumentItemRequest struct {
	ItemID   uint  `json:"item_id" validate:"required,gt=0"`
	Quantity int   `json:"quantity" validate:"required,gt=0"`
	Price    int64 `json:"price" validate:"required,gte=0"`
}

// UpdateDocumentItemRequest represents the update item request
type UpdateDocumentItemRequest struct {
	Quantity *int   `json:"quantity" validate:"omitempty,gt=0"`
	Price    *int64 `json:"price" validate:"omitempty,gte=0"`
}

// DocumentResponse represents the document response
type DocumentResponse struct {
	ID           uint                   `json:"id"`
	Code         string                 `json:"code"`
	CityID       uint                   `json:"city_id"`
	City         *city.CityResponse     `json:"city,omitempty"`
	DocumentDate time.Time              `json:"document_date"`
	TotalAmount  int64                  `json:"total_amount"`
	Items        []DocumentItemResponse `json:"items,omitempty"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
}

// DocumentItemResponse represents a line item response
type DocumentItemResponse struct {
	ID         uint               `json:"id"`
	DocumentID uint               `json:"document_id"`
	ItemID     uint               `json:"item_id"`
	Item       *item.ItemResponse `json:"item,omitempty"`
	Quantity   int                `json:"quantity"`
	Price      int64              `json:"price"`
	LineTotal  int64              `json:"line_total"`
	CreatedAt  time.Time          `json:"created_at"`
	UpdatedAt  time.Time          `json:"updated_at"`
}

// ToResponse converts Document to DocumentResponse
func (d *Document) ToResponse() *DocumentResponse {
	resp := &DocumentResponse{
		ID:           d.ID,
		Code:         d.Code,
		CityID:       d.CityID,
		DocumentDate: d.DocumentDate,
		TotalAmount:  d.TotalAmount,
		CreatedAt:    d.CreatedAt,
		UpdatedAt:    d.UpdatedAt,
	}

	// Include city if loaded
	if d.City.ID != 0 {
		resp.City = d.City.ToResponse()
	}

	// Include items if loaded
	if len(d.Items) > 0 {
		resp.Items = make([]DocumentItemResponse, len(d.Items))
		for i, item := range d.Items {
			resp.Items[i] = *item.ToResponse()
		}
	}

	return resp
}

// ToResponse converts DocumentItem to DocumentItemResponse
func (di *DocumentItem) ToResponse() *DocumentItemResponse {
	resp := &DocumentItemResponse{
		ID:         di.ID,
		DocumentID: di.DocumentID,
		ItemID:     di.ItemID,
		Quantity:   di.Quantity,
		Price:      di.Price,
		LineTotal:  di.GetLineTotal(),
		CreatedAt:  di.CreatedAt,
		UpdatedAt:  di.UpdatedAt,
	}

	// Include item if loaded
	if di.Item.ID != 0 {
		resp.Item = di.Item.ToResponse()
	}

	return resp
}
