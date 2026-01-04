package city

import (
	"context"

	"github.com/voidmaindev/go-template/internal/common"
	"github.com/voidmaindev/go-template/internal/common/filter"
	"gorm.io/gorm"
)

// Repository defines the city repository interface
type Repository interface {
	common.Repository[City]

	// FindAllWithCountry finds all cities with country preloaded
	FindAllWithCountry(ctx context.Context, pagination *common.Pagination) ([]City, int64, error)

	// FindAllFilteredWithCountry finds all cities with filtering, sorting, and country preloaded
	FindAllFilteredWithCountry(ctx context.Context, params *filter.Params) ([]City, int64, error)

	// FindByCountryID finds all cities by country ID
	FindByCountryID(ctx context.Context, countryID uint, pagination *common.Pagination) ([]City, int64, error)

	// FindByCountryIDWithCountry finds all cities by country ID with country preloaded
	FindByCountryIDWithCountry(ctx context.Context, countryID uint, pagination *common.Pagination) ([]City, int64, error)

	// FindByName finds a city by name
	FindByName(ctx context.Context, name string) (*City, error)

	// FindByIDWithCountry finds a city by ID with country preloaded
	FindByIDWithCountry(ctx context.Context, id uint) (*City, error)
}

// repository implements the Repository interface
type repository struct {
	*common.BaseRepository[City]
}

// NewRepository creates a new city repository
func NewRepository(db *gorm.DB) Repository {
	return &repository{
		BaseRepository: common.NewBaseRepository[City](db),
	}
}

// FindAllWithCountry finds all cities with country preloaded
func (r *repository) FindAllWithCountry(ctx context.Context, pagination *common.Pagination) ([]City, int64, error) {
	var cities []City
	var total int64

	query := r.GetDB().WithContext(ctx).Model(&City{})

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	query = query.Preload("Country")
	if pagination != nil {
		pagination.Validate()
		if orderClause := pagination.GetOrderClause(); orderClause != "" {
			query = query.Order(orderClause)
		} else {
			query = query.Order("id DESC")
		}
		query = query.Offset(pagination.GetOffset()).Limit(pagination.GetLimit())
	}

	if err := query.Find(&cities).Error; err != nil {
		return nil, 0, err
	}

	return cities, total, nil
}

// FindByCountryID finds all cities by country ID
func (r *repository) FindByCountryID(ctx context.Context, countryID uint, pagination *common.Pagination) ([]City, int64, error) {
	return r.FindByCondition(ctx, map[string]any{"country_id": countryID}, pagination)
}

// FindByCountryIDWithCountry finds all cities by country ID with country preloaded
func (r *repository) FindByCountryIDWithCountry(ctx context.Context, countryID uint, pagination *common.Pagination) ([]City, int64, error) {
	var cities []City
	var total int64

	query := r.GetDB().WithContext(ctx).Model(&City{}).Where("country_id = ?", countryID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	query = query.Preload("Country")
	if pagination != nil {
		pagination.Validate()
		if orderClause := pagination.GetOrderClause(); orderClause != "" {
			query = query.Order(orderClause)
		} else {
			query = query.Order("id DESC")
		}
		query = query.Offset(pagination.GetOffset()).Limit(pagination.GetLimit())
	}

	if err := query.Find(&cities).Error; err != nil {
		return nil, 0, err
	}

	return cities, total, nil
}

// FindByName finds a city by name
func (r *repository) FindByName(ctx context.Context, name string) (*City, error) {
	return r.FindOne(ctx, map[string]any{"name": name})
}

// FindByIDWithCountry finds a city by ID with country preloaded
func (r *repository) FindByIDWithCountry(ctx context.Context, id uint) (*City, error) {
	var city City
	err := r.GetDB().WithContext(ctx).Preload("Country").First(&city, id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, common.ErrNotFound
		}
		return nil, err
	}
	return &city, nil
}

// FindAllFilteredWithCountry finds all cities with filtering, sorting, and country preloaded
func (r *repository) FindAllFilteredWithCountry(ctx context.Context, params *filter.Params) ([]City, int64, error) {
	var cities []City
	var total int64
	var city City

	config := city.FilterConfig()

	// Count query (without pagination)
	countQuery := r.GetDB().WithContext(ctx).Model(&City{})
	countQuery = filter.ApplyFiltersOnly(countQuery, config, params)
	if err := countQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Data query (with pagination, sorting, and preload)
	query := r.GetDB().WithContext(ctx).Model(&City{}).Preload("Country")
	query = filter.Apply(query, config, params)

	if err := query.Find(&cities).Error; err != nil {
		return nil, 0, err
	}

	return cities, total, nil
}
