package city

import (
	"context"
	"errors"

	"github.com/voidmaindev/GoTemplate/internal/common"
	"github.com/voidmaindev/GoTemplate/internal/domain/country"
	"github.com/voidmaindev/GoTemplate/pkg/utils"
)

var (
	ErrCityNotFound    = errors.New("city not found")
	ErrCountryNotFound = errors.New("country not found")
)

// Service defines the city service interface
type Service interface {
	Create(ctx context.Context, req *CreateCityRequest) (*City, error)
	GetByID(ctx context.Context, id uint) (*City, error)
	GetByIDWithCountry(ctx context.Context, id uint) (*City, error)
	Update(ctx context.Context, id uint, req *UpdateCityRequest) (*City, error)
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context, pagination *common.Pagination) (*common.PaginatedResult[City], error)
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
		return nil, err
	}
	if countryEntity == nil {
		return nil, ErrCountryNotFound
	}

	city := &City{
		Name:      req.Name,
		CountryID: req.CountryID,
	}

	if err := s.repo.Create(ctx, city); err != nil {
		return nil, err
	}

	// Load country for response
	city.Country = *countryEntity
	return city, nil
}

// GetByID retrieves a city by ID
func (s *service) GetByID(ctx context.Context, id uint) (*City, error) {
	city, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if city == nil {
		return nil, ErrCityNotFound
	}
	return city, nil
}

// GetByIDWithCountry retrieves a city by ID with country preloaded
func (s *service) GetByIDWithCountry(ctx context.Context, id uint) (*City, error) {
	city, err := s.repo.FindByIDWithCountry(ctx, id)
	if err != nil {
		return nil, err
	}
	if city == nil {
		return nil, ErrCityNotFound
	}
	return city, nil
}

// Update updates a city
func (s *service) Update(ctx context.Context, id uint, req *UpdateCityRequest) (*City, error) {
	city, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if city == nil {
		return nil, ErrCityNotFound
	}

	// Handle CountryID validation separately (FK constraint)
	if req.CountryID != nil {
		countryEntity, err := s.countryRepo.FindByID(ctx, *req.CountryID)
		if err != nil {
			return nil, err
		}
		if countryEntity == nil {
			return nil, ErrCountryNotFound
		}
		city.CountryID = *req.CountryID
		city.Country = *countryEntity
		req.CountryID = nil // Prevent copier from overwriting
	}

	// Map remaining non-nil fields from request to model
	if err := utils.UpdateModel(city, req); err != nil {
		return nil, err
	}

	if err := s.repo.Update(ctx, city); err != nil {
		return nil, err
	}

	return city, nil
}

// Delete soft-deletes a city
func (s *service) Delete(ctx context.Context, id uint) error {
	city, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if city == nil {
		return ErrCityNotFound
	}

	return s.repo.Delete(ctx, id)
}

// List retrieves all cities with pagination
func (s *service) List(ctx context.Context, pagination *common.Pagination) (*common.PaginatedResult[City], error) {
	cities, total, err := s.repo.WithPreload("Country").(Repository).FindAll(ctx, pagination)
	if err != nil {
		return nil, err
	}

	return common.NewPaginatedResult(cities, total, pagination), nil
}

// ListByCountry retrieves all cities for a specific country
func (s *service) ListByCountry(ctx context.Context, countryID uint, pagination *common.Pagination) (*common.PaginatedResult[City], error) {
	// Validate country exists
	countryEntity, err := s.countryRepo.FindByID(ctx, countryID)
	if err != nil {
		return nil, err
	}
	if countryEntity == nil {
		return nil, ErrCountryNotFound
	}

	cities, total, err := s.repo.FindByCountryID(ctx, countryID, pagination)
	if err != nil {
		return nil, err
	}

	return common.NewPaginatedResult(cities, total, pagination), nil
}
