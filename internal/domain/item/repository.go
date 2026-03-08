package item

import (
	"context"

	"github.com/voidmaindev/go-template/internal/common"
	"gorm.io/gorm"
)

// Repository defines the item repository interface.
// It extends common.Repository[Item] with domain-specific queries.
type Repository interface {
	common.Repository[Item]

	// FindByName retrieves an item by its unique name.
	// Returns a NotFound error if no item with the given name exists.
	FindByName(ctx context.Context, name string) (*Item, error)
}

// repository implements the Repository interface
type repository struct {
	*common.BaseRepository[Item]
}

// NewRepository creates a new item repository
func NewRepository(db *gorm.DB) Repository {
	return &repository{
		BaseRepository: common.NewBaseRepository[Item](db, "item"),
	}
}

// FindByName finds an item by name
func (r *repository) FindByName(ctx context.Context, name string) (*Item, error) {
	return r.FindOne(ctx, map[string]any{"name": name})
}
