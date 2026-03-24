package example_item

import (
	"time"
)

// CreateItemRequest represents the create item request
type CreateItemRequest struct {
	Name        string `json:"name" validate:"required,min=1,max=200"`
	Description string `json:"description" validate:"omitempty,max=5000"`
	Price       int64  `json:"price" validate:"gte=0,lte=100000000000"`
}

// UpdateItemRequest represents the update item request
type UpdateItemRequest struct {
	Name        *string `json:"name" validate:"omitempty,min=1,max=200"`
	Description *string `json:"description" validate:"omitempty,max=5000"`
	Price       *int64  `json:"price" validate:"omitempty,gte=0,lte=100000000000"`
}

// ItemResponse represents the item response
type ItemResponse struct {
	ID          uint      `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Price       int64     `json:"price"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ToResponse converts Item to ItemResponse
func (i *Item) ToResponse() *ItemResponse {
	return &ItemResponse{
		ID:          i.ID,
		Name:        i.Name,
		Description: i.Description,
		Price:       i.Price,
		CreatedAt:   i.CreatedAt,
		UpdatedAt:   i.UpdatedAt,
	}
}
