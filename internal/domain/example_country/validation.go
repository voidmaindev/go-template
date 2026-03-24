package example_country

import (
	"context"
	"strings"

	"github.com/voidmaindev/go-template/internal/common/validation"
)

// Validator handles country domain validation
type Validator struct {
	repo      Repository
	validator *validation.Validator
}

// NewValidator creates a country validator
func NewValidator(repo Repository) *Validator {
	return &Validator{
		repo:      repo,
		validator: validation.New(),
	}
}

// ValidateCreate validates create country request
func (v *Validator) ValidateCreate(ctx context.Context, req *CreateCountryRequest) *validation.Result {
	return validation.NewBuilder(ctx).
		StructWith(v.validator, req).
		CustomWithCode("code", "ALREADY_EXISTS", func(ctx context.Context) error {
			code := strings.ToUpper(req.Code)
			exists, err := v.repo.ExistsByCode(ctx, code)
			if err != nil {
				return err
			}
			if exists {
				return ErrCountryCodeExists
			}
			return nil
		}).
		Result()
}

// ValidateUpdate validates update country request
func (v *Validator) ValidateUpdate(ctx context.Context, id uint, req *UpdateCountryRequest) *validation.Result {
	builder := validation.NewBuilder(ctx).
		StructWith(v.validator, req)

	// If code is being updated, check for uniqueness
	if req.Code != nil {
		builder.CustomWithCode("code", "ALREADY_EXISTS", func(ctx context.Context) error {
			code := strings.ToUpper(*req.Code)
			existing, err := v.repo.FindByCode(ctx, code)
			if err != nil {
				// Not found is OK - code doesn't exist
				return nil
			}
			// If found and it's not the same country, it's a conflict
			if existing.ID != id {
				return ErrCountryCodeExists
			}
			return nil
		})
	}

	return builder.Result()
}
