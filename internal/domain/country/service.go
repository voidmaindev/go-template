package country

import (
	"context"
	"errors"
	"strings"

	"github.com/voidmaindev/GoTemplate/internal/common"
	"github.com/voidmaindev/GoTemplate/pkg/utils"
)

var (
	ErrCountryNotFound     = errors.New("country not found")
	ErrCountryCodeExists   = errors.New("country code already exists")
)

// Service defines the country service interface
type Service interface {
	Create(ctx context.Context, req *CreateCountryRequest) (*Country, error)
	GetByID(ctx context.Context, id uint) (*Country, error)
	GetByCode(ctx context.Context, code string) (*Country, error)
	Update(ctx context.Context, id uint, req *UpdateCountryRequest) (*Country, error)
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context, pagination *common.Pagination) (*common.PaginatedResult[Country], error)
}

// service implements the Service interface
type service struct {
	repo Repository
}

// NewService creates a new country service
func NewService(repo Repository) Service {
	return &service{
		repo: repo,
	}
}

// Create creates a new country
func (s *service) Create(ctx context.Context, req *CreateCountryRequest) (*Country, error) {
	// Normalize code to uppercase
	code := strings.ToUpper(req.Code)

	// Check if code already exists
	exists, err := s.repo.ExistsByCode(ctx, code)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrCountryCodeExists
	}

	country := &Country{
		Name: req.Name,
		Code: code,
	}

	if err := s.repo.Create(ctx, country); err != nil {
		return nil, err
	}

	return country, nil
}

// GetByID retrieves a country by ID
func (s *service) GetByID(ctx context.Context, id uint) (*Country, error) {
	country, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if country == nil {
		return nil, ErrCountryNotFound
	}
	return country, nil
}

// GetByCode retrieves a country by code
func (s *service) GetByCode(ctx context.Context, code string) (*Country, error) {
	country, err := s.repo.FindByCode(ctx, strings.ToUpper(code))
	if err != nil {
		return nil, err
	}
	if country == nil {
		return nil, ErrCountryNotFound
	}
	return country, nil
}

// Update updates a country
func (s *service) Update(ctx context.Context, id uint, req *UpdateCountryRequest) (*Country, error) {
	country, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if country == nil {
		return nil, ErrCountryNotFound
	}

	// Handle Code validation separately (unique constraint + uppercase normalization)
	if req.Code != nil {
		code := strings.ToUpper(*req.Code)
		existing, err := s.repo.FindByCode(ctx, code)
		if err != nil {
			return nil, err
		}
		if existing != nil && existing.ID != id {
			return nil, ErrCountryCodeExists
		}
		country.Code = code
		req.Code = nil // Prevent copier from overwriting
	}

	// Map remaining non-nil fields from request to model
	if err := utils.UpdateModel(country, req); err != nil {
		return nil, err
	}

	if err := s.repo.Update(ctx, country); err != nil {
		return nil, err
	}

	return country, nil
}

// Delete soft-deletes a country
func (s *service) Delete(ctx context.Context, id uint) error {
	country, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if country == nil {
		return ErrCountryNotFound
	}

	return s.repo.Delete(ctx, id)
}

// List retrieves all countries with pagination
func (s *service) List(ctx context.Context, pagination *common.Pagination) (*common.PaginatedResult[Country], error) {
	countries, total, err := s.repo.FindAll(ctx, pagination)
	if err != nil {
		return nil, err
	}

	return common.NewPaginatedResult(countries, total, pagination), nil
}
