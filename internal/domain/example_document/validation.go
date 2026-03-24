package example_document

import (
	"context"

	"github.com/voidmaindev/go-template/internal/common/errors"
	"github.com/voidmaindev/go-template/internal/common/validation"
	"github.com/voidmaindev/go-template/internal/domain/example_city"
	"github.com/voidmaindev/go-template/internal/domain/example_item"
)

// Validator handles document domain validation
type Validator struct {
	repo        Repository
	cityRepo    example_city.Repository
	productRepo example_item.Repository
	validator   *validation.Validator
}

// NewValidator creates a document validator
func NewValidator(repo Repository, cityRepo example_city.Repository, productRepo example_item.Repository) *Validator {
	return &Validator{
		repo:        repo,
		cityRepo:    cityRepo,
		productRepo: productRepo,
		validator:   validation.New(),
	}
}

// ValidateCreate validates create document request
func (v *Validator) ValidateCreate(ctx context.Context, req *CreateDocumentRequest) *validation.Result {
	builder := validation.NewBuilder(ctx).
		StructWith(v.validator, req)

	// Validate city exists
	builder.CustomWithCode("city_id", "NOT_FOUND", func(ctx context.Context) error {
		_, err := v.cityRepo.FindByID(ctx, req.CityID)
		if err != nil {
			if errors.IsNotFound(err) {
				return ErrCityNotFound
			}
			return err
		}
		return nil
	})

	// Validate document code uniqueness
	builder.CustomWithCode("code", "ALREADY_EXISTS", func(ctx context.Context) error {
		exists, err := v.repo.ExistsByCode(ctx, req.Code)
		if err != nil {
			return err
		}
		if exists {
			return ErrDocumentCodeExists
		}
		return nil
	})

	// Validate all items exist
	for i, reqItem := range req.Items {
		itemIndex := i
		itemID := reqItem.ItemID
		builder.CustomWithCode("items", "ITEM_NOT_FOUND", func(ctx context.Context) error {
			_, err := v.productRepo.FindByID(ctx, itemID)
			if err != nil {
				if errors.IsNotFound(err) {
					return &validationItemError{index: itemIndex, itemID: itemID}
				}
				return err
			}
			return nil
		})
	}

	return builder.Result()
}

// ValidateUpdate validates update document request
func (v *Validator) ValidateUpdate(ctx context.Context, id uint, req *UpdateDocumentRequest) *validation.Result {
	builder := validation.NewBuilder(ctx).
		StructWith(v.validator, req)

	// Validate city if being updated
	if req.CityID != nil {
		builder.CustomWithCode("city_id", "NOT_FOUND", func(ctx context.Context) error {
			_, err := v.cityRepo.FindByID(ctx, *req.CityID)
			if err != nil {
				if errors.IsNotFound(err) {
					return ErrCityNotFound
				}
				return err
			}
			return nil
		})
	}

	// Validate code uniqueness if being updated
	if req.Code != nil {
		builder.CustomWithCode("code", "ALREADY_EXISTS", func(ctx context.Context) error {
			existing, err := v.repo.FindByCode(ctx, *req.Code)
			if err != nil {
				// Not found is OK
				return nil
			}
			if existing.ID != id {
				return ErrDocumentCodeExists
			}
			return nil
		})
	}

	return builder.Result()
}

// ValidateAddItem validates add item request
func (v *Validator) ValidateAddItem(ctx context.Context, req *AddDocumentItemRequest) *validation.Result {
	builder := validation.NewBuilder(ctx).
		StructWith(v.validator, req)

	// Validate item exists
	builder.CustomWithCode("item_id", "NOT_FOUND", func(ctx context.Context) error {
		_, err := v.productRepo.FindByID(ctx, req.ItemID)
		if err != nil {
			if errors.IsNotFound(err) {
				return ErrItemNotFound
			}
			return err
		}
		return nil
	})

	return builder.Result()
}

// ValidateUpdateItem validates update item request
func (v *Validator) ValidateUpdateItem(ctx context.Context, req *UpdateDocumentItemRequest) *validation.Result {
	return validation.NewBuilder(ctx).
		StructWith(v.validator, req).
		Result()
}

// validationItemError is a helper for item validation errors
type validationItemError struct {
	index  int
	itemID uint
}

func (e *validationItemError) Error() string {
	return "item not found"
}
