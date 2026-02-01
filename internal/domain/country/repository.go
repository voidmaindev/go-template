package country

import (
	"context"

	"github.com/voidmaindev/go-template/internal/common"
	"gorm.io/gorm"
)

// Repository defines the country repository interface.
// It extends common.Repository[Country] with domain-specific queries.
type Repository interface {
	common.Repository[Country]

	// FindByCode retrieves a country by its unique ISO code (e.g., "US", "DE").
	// Returns a NotFound error if no country with the given code exists.
	FindByCode(ctx context.Context, code string) (*Country, error)

	// FindByName retrieves a country by its unique name.
	// Returns a NotFound error if no country with the given name exists.
	FindByName(ctx context.Context, name string) (*Country, error)

	// ExistsByCode checks if a country with the given ISO code exists.
	ExistsByCode(ctx context.Context, code string) (bool, error)
}

// repository implements the Repository interface
type repository struct {
	*common.BaseRepository[Country]
}

// NewRepository creates a new country repository
func NewRepository(db *gorm.DB) Repository {
	return &repository{
		BaseRepository: common.NewBaseRepository[Country](db),
	}
}

// FindByCode finds a country by code
func (r *repository) FindByCode(ctx context.Context, code string) (*Country, error) {
	return r.FindOne(ctx, map[string]any{"code": code})
}

// FindByName finds a country by name
func (r *repository) FindByName(ctx context.Context, name string) (*Country, error) {
	return r.FindOne(ctx, map[string]any{"name": name})
}

// ExistsByCode checks if a country with the given code exists
func (r *repository) ExistsByCode(ctx context.Context, code string) (bool, error) {
	return r.Exists(ctx, map[string]any{"code": code})
}
