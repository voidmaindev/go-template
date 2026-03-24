package example_item

import (
	"context"

	"github.com/voidmaindev/go-template/internal/common/validation"
)

// Validator handles item domain validation
type Validator struct {
	validator *validation.Validator
}

// NewValidator creates an item validator
func NewValidator() *Validator {
	return &Validator{
		validator: validation.New(),
	}
}

// ValidateCreate validates create item request
func (v *Validator) ValidateCreate(ctx context.Context, req *CreateItemRequest) *validation.Result {
	return validation.NewBuilder(ctx).
		StructWith(v.validator, req).
		Result()
}

// ValidateUpdate validates update item request
func (v *Validator) ValidateUpdate(ctx context.Context, req *UpdateItemRequest) *validation.Result {
	return validation.NewBuilder(ctx).
		StructWith(v.validator, req).
		Result()
}
