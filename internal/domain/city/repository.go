package city

import (
	"context"

	"github.com/voidmaindev/go-template/internal/common"
	"gorm.io/gorm"
)

// Repository defines the city repository interface
type Repository interface {
	common.Repository[City]

	// FindByCountryID finds all cities by country ID
	FindByCountryID(ctx context.Context, countryID uint, pagination *common.Pagination) ([]City, int64, error)

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

// FindByCountryID finds all cities by country ID
func (r *repository) FindByCountryID(ctx context.Context, countryID uint, pagination *common.Pagination) ([]City, int64, error) {
	return r.FindByCondition(ctx, map[string]any{"country_id": countryID}, pagination)
}

// FindByName finds a city by name
func (r *repository) FindByName(ctx context.Context, name string) (*City, error) {
	return r.FindOne(ctx, map[string]any{"name": name})
}

// FindByIDWithCountry finds a city by ID with country preloaded
func (r *repository) FindByIDWithCountry(ctx context.Context, id uint) (*City, error) {
	return r.WithPreload("Country").(*repository).BaseRepository.FindByID(ctx, id)
}
