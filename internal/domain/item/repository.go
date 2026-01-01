package item

import (
	"context"

	"github.com/voidmaindev/GoTemplate/internal/common"
	"gorm.io/gorm"
)

// Repository defines the item repository interface
type Repository interface {
	common.Repository[Item]

	// FindByName finds an item by name
	FindByName(ctx context.Context, name string) (*Item, error)
}

// repository implements the Repository interface
type repository struct {
	*common.BaseRepository[Item]
}

// NewRepository creates a new item repository
func NewRepository(db *gorm.DB) Repository {
	return &repository{
		BaseRepository: common.NewBaseRepository[Item](db),
	}
}

// FindByName finds an item by name
func (r *repository) FindByName(ctx context.Context, name string) (*Item, error) {
	return r.FindOne(ctx, map[string]interface{}{"name": name})
}
