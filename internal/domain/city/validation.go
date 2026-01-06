package city

import (
	"context"

	"github.com/voidmaindev/go-template/internal/common/errors"
	"github.com/voidmaindev/go-template/internal/common/validation"
	"github.com/voidmaindev/go-template/internal/domain/country"
)

// Validator handles city domain validation
type Validator struct {
	countryRepo country.Repository
	validator   *validation.Validator
}

// NewValidator creates a city validator
func NewValidator(countryRepo country.Repository) *Validator {
	return &Validator{
		countryRepo: countryRepo,
		validator:   validation.New(),
	}
}

// ValidateCreate validates create city request
func (v *Validator) ValidateCreate(ctx context.Context, req *CreateCityRequest) *validation.Result {
	return validation.NewBuilder(ctx).
		StructWith(v.validator, req).
		CustomWithCode("country_id", "NOT_FOUND", func(ctx context.Context) error {
			_, err := v.countryRepo.FindByID(ctx, req.CountryID)
			if err != nil {
				if errors.IsNotFound(err) {
					return ErrCountryNotFound
				}
				return err
			}
			return nil
		}).
		Result()
}

// ValidateUpdate validates update city request
func (v *Validator) ValidateUpdate(ctx context.Context, req *UpdateCityRequest) *validation.Result {
	builder := validation.NewBuilder(ctx).
		StructWith(v.validator, req)

	// If country_id is being updated, verify it exists
	if req.CountryID != nil {
		builder.CustomWithCode("country_id", "NOT_FOUND", func(ctx context.Context) error {
			_, err := v.countryRepo.FindByID(ctx, *req.CountryID)
			if err != nil {
				if errors.IsNotFound(err) {
					return ErrCountryNotFound
				}
				return err
			}
			return nil
		})
	}

	return builder.Result()
}
