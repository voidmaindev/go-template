package common

import (
	"context"

	"github.com/voidmaindev/go-template/internal/common/filter"
	"gorm.io/gorm"
)

// Repository defines the generic repository interface for CRUD operations
type Repository[T any] interface {
	// Create inserts a new entity
	Create(ctx context.Context, entity *T) error

	// CreateBatch inserts multiple entities in batches
	CreateBatch(ctx context.Context, entities []T, batchSize int) error

	// FindByID retrieves an entity by its primary key
	FindByID(ctx context.Context, id uint) (*T, error)

	// FindAll retrieves all entities with pagination
	FindAll(ctx context.Context, pagination *Pagination) ([]T, int64, error)

	// FindByCondition retrieves entities matching conditions
	FindByCondition(ctx context.Context, condition map[string]any, pagination *Pagination) ([]T, int64, error)

	// FindAllFiltered retrieves entities with dynamic filtering, sorting, and pagination
	FindAllFiltered(ctx context.Context, params *filter.Params) ([]T, int64, error)

	// FindOne retrieves a single entity matching conditions
	FindOne(ctx context.Context, condition map[string]any) (*T, error)

	// Update updates an existing entity
	Update(ctx context.Context, entity *T) error

	// UpdateFields updates specific fields of an entity
	UpdateFields(ctx context.Context, id uint, fields map[string]any) error

	// Delete soft-deletes an entity by ID
	Delete(ctx context.Context, id uint) error

	// HardDelete permanently removes an entity
	HardDelete(ctx context.Context, id uint) error

	// Exists checks if an entity exists with given conditions
	Exists(ctx context.Context, condition map[string]any) (bool, error)

	// Count returns the count of entities matching conditions
	Count(ctx context.Context, condition map[string]any) (int64, error)

	// WithTx returns a new repository instance using the provided transaction
	WithTx(tx *gorm.DB) Repository[T]

	// WithPreload returns a repository with eager loading
	WithPreload(preloads ...string) Repository[T]

	// Transaction executes operations within a transaction
	Transaction(ctx context.Context, fn func(txRepo Repository[T]) error) error
}
