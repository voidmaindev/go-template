package example_city

import (
	"context"
	"errors"

	"github.com/voidmaindev/go-template/internal/common"
	"github.com/voidmaindev/go-template/internal/common/filter"
	"gorm.io/gorm"
)

// Repository defines the city repository interface.
// It extends common.Repository[City] with domain-specific queries,
// including methods that eager-load the related Country entity.
type Repository interface {
	common.Repository[City]

	// FindAllWithCountry retrieves all cities with the Country relation preloaded.
	// Returns paginated results along with the total count.
	FindAllWithCountry(ctx context.Context, pagination *common.Pagination) ([]City, int64, error)

	// FindAllFilteredWithCountry retrieves cities with dynamic filtering, sorting,
	// pagination, and the Country relation preloaded.
	FindAllFilteredWithCountry(ctx context.Context, params *filter.Params) ([]City, int64, error)

	// FindByCountryID retrieves all cities belonging to a specific example_country.
	// Returns paginated results along with the total count.
	FindByCountryID(ctx context.Context, countryID uint, pagination *common.Pagination) ([]City, int64, error)

	// FindByCountryIDWithCountry retrieves all cities belonging to a specific country
	// with the Country relation preloaded.
	FindByCountryIDWithCountry(ctx context.Context, countryID uint, pagination *common.Pagination) ([]City, int64, error)

	// FindByName retrieves a city by its unique name.
	// Returns a NotFound error if no city with the given name exists.
	FindByName(ctx context.Context, name string) (*City, error)

	// FindByIDWithCountry retrieves a city by ID with the Country relation preloaded.
	// Returns ErrCityNotFound if no city with the given ID exists.
	FindByIDWithCountry(ctx context.Context, id uint) (*City, error)
}

// repository implements the Repository interface
type repository struct {
	*common.BaseRepository[City]
}

// NewRepository creates a new city repository
func NewRepository(db *gorm.DB) Repository {
	return &repository{
		BaseRepository: common.NewBaseRepository[City](db, "city"),
	}
}

// FindAllWithCountry finds all cities with country preloaded
func (r *repository) FindAllWithCountry(ctx context.Context, pagination *common.Pagination) ([]City, int64, error) {
	return r.WithPreload("Country").FindAll(ctx, pagination)
}

// FindByCountryID finds all cities by country ID
func (r *repository) FindByCountryID(ctx context.Context, countryID uint, pagination *common.Pagination) ([]City, int64, error) {
	return r.FindByCondition(ctx, map[string]any{"country_id": countryID}, pagination)
}

// FindByCountryIDWithCountry finds all cities by country ID with country preloaded
func (r *repository) FindByCountryIDWithCountry(ctx context.Context, countryID uint, pagination *common.Pagination) ([]City, int64, error) {
	return r.WithPreload("Country").FindByCondition(ctx, map[string]any{"country_id": countryID}, pagination)
}

// FindByName finds a city by name
func (r *repository) FindByName(ctx context.Context, name string) (*City, error) {
	return r.FindOne(ctx, map[string]any{"name": name})
}

// FindByIDWithCountry finds a city by ID with country preloaded
func (r *repository) FindByIDWithCountry(ctx context.Context, id uint) (*City, error) {
	var city City
	err := r.DB().WithContext(ctx).Preload("Country").First(&city, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCityNotFound
		}
		return nil, err
	}
	return &city, nil
}

// FindAllFilteredWithCountry finds all cities with filtering, sorting, and country preloaded
func (r *repository) FindAllFilteredWithCountry(ctx context.Context, params *filter.Params) ([]City, int64, error) {
	return r.WithPreload("Country").FindAllFiltered(ctx, params)
}
