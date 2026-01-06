package city

import (
	"context"

	"github.com/voidmaindev/go-template/internal/common"
	"github.com/voidmaindev/go-template/internal/common/errors"
	"github.com/voidmaindev/go-template/internal/common/filter"
	"github.com/voidmaindev/go-template/internal/domain/country"
	"github.com/voidmaindev/go-template/pkg/utils"
)

// Service defines the city service interface
type Service interface {
	Create(ctx context.Context, req *CreateCityRequest) (*City, error)
	GetByID(ctx context.Context, id uint) (*City, error)
	GetByIDWithCountry(ctx context.Context, id uint) (*City, error)
	Update(ctx context.Context, id uint, req *UpdateCityRequest) (*City, error)
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context, pagination *common.Pagination) (*common.PaginatedResult[City], error)
	ListFiltered(ctx context.Context, params *filter.Params) (*common.FilteredResult[City], error)
	ListByCountry(ctx context.Context, countryID uint, pagination *common.Pagination) (*common.PaginatedResult[City], error)
}

// service implements the Service interface
type service struct {
	repo        Repository
	countryRepo country.Repository
}

// NewService creates a new city service
func NewService(repo Repository, countryRepo country.Repository) Service {
	return &service{
		repo:        repo,
		countryRepo: countryRepo,
	}
}

// Create creates a new city
func (s *service) Create(ctx context.Context, req *CreateCityRequest) (*City, error) {
	// Validate country exists
	countryEntity, err := s.countryRepo.FindByID(ctx, req.CountryID)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, ErrCountryNotFound
		}
		return nil, errors.Internal(domainName, err).WithOperation("Create")
	}

	city := &City{
		Name:      req.Name,
		CountryID: req.CountryID,
	}

	if err := s.repo.Create(ctx, city); err != nil {
		return nil, errors.Internal(domainName, err).WithOperation("Create")
	}

	// Load country for response
	city.Country = *countryEntity
	return city, nil
}

// GetByID retrieves a city by ID
func (s *service) GetByID(ctx context.Context, id uint) (*City, error) {
	city, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, ErrCityNotFound
		}
		return nil, errors.Internal(domainName, err).WithOperation("GetByID")
	}
	return city, nil
}

// GetByIDWithCountry retrieves a city by ID with country preloaded
func (s *service) GetByIDWithCountry(ctx context.Context, id uint) (*City, error) {
	city, err := s.repo.FindByIDWithCountry(ctx, id)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, ErrCityNotFound
		}
		return nil, errors.Internal(domainName, err).WithOperation("GetByIDWithCountry")
	}
	return city, nil
}

// Update updates a city
func (s *service) Update(ctx context.Context, id uint, req *UpdateCityRequest) (*City, error) {
	city, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, ErrCityNotFound
		}
		return nil, errors.Internal(domainName, err).WithOperation("Update")
	}

	// Handle CountryID validation separately (FK constraint)
	if req.CountryID != nil {
		countryEntity, err := s.countryRepo.FindByID(ctx, *req.CountryID)
		if err != nil {
			if errors.IsNotFound(err) {
				return nil, ErrCountryNotFound
			}
			return nil, errors.Internal(domainName, err).WithOperation("Update")
		}
		city.CountryID = *req.CountryID
		city.Country = *countryEntity
		req.CountryID = nil // Prevent copier from overwriting
	}

	// Map remaining non-nil fields from request to model
	if err := utils.UpdateModel(city, req); err != nil {
		return nil, errors.Internal(domainName, err).WithOperation("Update")
	}

	if err := s.repo.Update(ctx, city); err != nil {
		return nil, errors.Internal(domainName, err).WithOperation("Update")
	}

	return city, nil
}

// Delete soft-deletes a city
func (s *service) Delete(ctx context.Context, id uint) error {
	_, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.IsNotFound(err) {
			return ErrCityNotFound
		}
		return errors.Internal(domainName, err).WithOperation("Delete")
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return errors.Internal(domainName, err).WithOperation("Delete")
	}
	return nil
}

// List retrieves all cities with pagination
func (s *service) List(ctx context.Context, pagination *common.Pagination) (*common.PaginatedResult[City], error) {
	cities, total, err := s.repo.FindAllWithCountry(ctx, pagination)
	if err != nil {
		return nil, errors.Internal(domainName, err).WithOperation("List")
	}

	return common.NewPaginatedResult(cities, total, pagination), nil
}

// ListFiltered retrieves cities with dynamic filtering, sorting, and pagination
func (s *service) ListFiltered(ctx context.Context, params *filter.Params) (*common.FilteredResult[City], error) {
	cities, total, err := s.repo.FindAllFilteredWithCountry(ctx, params)
	if err != nil {
		return nil, errors.Internal(domainName, err).WithOperation("ListFiltered")
	}

	return common.NewFilteredResult(cities, total, params), nil
}

// ListByCountry retrieves all cities for a specific country
func (s *service) ListByCountry(ctx context.Context, countryID uint, pagination *common.Pagination) (*common.PaginatedResult[City], error) {
	// Validate country exists
	_, err := s.countryRepo.FindByID(ctx, countryID)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, ErrCountryNotFound
		}
		return nil, errors.Internal(domainName, err).WithOperation("ListByCountry")
	}

	cities, total, err := s.repo.FindByCountryIDWithCountry(ctx, countryID, pagination)
	if err != nil {
		return nil, errors.Internal(domainName, err).WithOperation("ListByCountry")
	}

	return common.NewPaginatedResult(cities, total, pagination), nil
}
