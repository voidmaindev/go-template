package example_item

import (
	"context"

	"github.com/voidmaindev/go-template/internal/common"
	"github.com/voidmaindev/go-template/internal/common/errors"
	"github.com/voidmaindev/go-template/internal/common/filter"
	"github.com/voidmaindev/go-template/pkg/utils"
)

// Service defines the item service interface
type Service interface {
	Create(ctx context.Context, req *CreateItemRequest) (*Item, error)
	GetByID(ctx context.Context, id uint) (*Item, error)
	Update(ctx context.Context, id uint, req *UpdateItemRequest) (*Item, error)
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context, pagination *common.Pagination) (*common.PaginatedResult[Item], error)
	ListFiltered(ctx context.Context, params *filter.Params) (*common.PaginatedResult[Item], error)
}

// service implements the Service interface
type service struct {
	repo Repository
}

// NewService creates a new item service
func NewService(repo Repository) Service {
	return &service{
		repo: repo,
	}
}

// Create creates a new item
func (s *service) Create(ctx context.Context, req *CreateItemRequest) (*Item, error) {
	item := &Item{
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
	}

	if err := s.repo.Create(ctx, item); err != nil {
		return nil, errors.Internal(domainName, err).WithOperation("Create")
	}

	return item, nil
}

// GetByID retrieves an item by ID
func (s *service) GetByID(ctx context.Context, id uint) (*Item, error) {
	item, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, ErrItemNotFound
		}
		return nil, errors.Internal(domainName, err).WithOperation("GetByID")
	}
	return item, nil
}

// Update updates an item
func (s *service) Update(ctx context.Context, id uint, req *UpdateItemRequest) (*Item, error) {
	item, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, ErrItemNotFound
		}
		return nil, errors.Internal(domainName, err).WithOperation("Update")
	}

	// Map non-nil fields from request to model
	if err := utils.UpdateModel(item, req); err != nil {
		return nil, errors.Internal(domainName, err).WithOperation("Update")
	}

	if err := s.repo.Update(ctx, item); err != nil {
		return nil, errors.Internal(domainName, err).WithOperation("Update")
	}

	return item, nil
}

// Delete soft-deletes an item
func (s *service) Delete(ctx context.Context, id uint) error {
	item, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.IsNotFound(err) {
			return ErrItemNotFound
		}
		return errors.Internal(domainName, err).WithOperation("Delete")
	}

	if err := s.repo.Delete(ctx, item.ID); err != nil {
		return errors.Internal(domainName, err).WithOperation("Delete")
	}

	return nil
}

// List retrieves all items with pagination
func (s *service) List(ctx context.Context, pagination *common.Pagination) (*common.PaginatedResult[Item], error) {
	items, total, err := s.repo.FindAll(ctx, pagination)
	if err != nil {
		return nil, errors.Internal(domainName, err).WithOperation("List")
	}

	return common.NewPaginatedResult(items, total, pagination), nil
}

// ListFiltered retrieves items with dynamic filtering, sorting, and pagination
func (s *service) ListFiltered(ctx context.Context, params *filter.Params) (*common.PaginatedResult[Item], error) {
	items, total, err := s.repo.FindAllFiltered(ctx, params)
	if err != nil {
		return nil, errors.Internal(domainName, err).WithOperation("ListFiltered")
	}

	return common.NewPaginatedResultFromFilter(items, total, params), nil
}
