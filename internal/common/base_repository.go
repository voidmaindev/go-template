package common

import (
	"context"
	"errors"

	"github.com/voidmaindev/go-template/internal/common/filter"
	"gorm.io/gorm"
)

// BaseRepository provides a generic implementation of Repository interface
type BaseRepository[T any] struct {
	db       *gorm.DB
	preloads []string
}

// NewBaseRepository creates a new instance of BaseRepository
func NewBaseRepository[T any](db *gorm.DB) *BaseRepository[T] {
	return &BaseRepository[T]{
		db:       db,
		preloads: []string{},
	}
}

// Create inserts a new entity into the database
func (r *BaseRepository[T]) Create(ctx context.Context, entity *T) error {
	return r.db.WithContext(ctx).Create(entity).Error
}

// CreateBatch inserts multiple entities in batches
func (r *BaseRepository[T]) CreateBatch(ctx context.Context, entities []T, batchSize int) error {
	if batchSize <= 0 {
		batchSize = 100
	}
	return r.db.WithContext(ctx).CreateInBatches(entities, batchSize).Error
}

// FindByID retrieves an entity by its primary key.
// Returns ErrNotFound if the entity doesn't exist.
func (r *BaseRepository[T]) FindByID(ctx context.Context, id uint) (*T, error) {
	var entity T
	query := r.applyPreloads(r.db.WithContext(ctx))
	if err := query.First(&entity, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &entity, nil
}

// FindAll retrieves all entities with pagination
func (r *BaseRepository[T]) FindAll(ctx context.Context, pagination *Pagination) ([]T, int64, error) {
	var entities []T
	var total int64

	query := r.db.WithContext(ctx).Model(new(T))

	// Count total records
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply preloads and pagination
	query = r.applyPreloads(query)
	query = r.applyPagination(query, pagination)

	if err := query.Find(&entities).Error; err != nil {
		return nil, 0, err
	}

	return entities, total, nil
}

// FindByCondition retrieves entities matching conditions
func (r *BaseRepository[T]) FindByCondition(ctx context.Context, condition map[string]any, pagination *Pagination) ([]T, int64, error) {
	var entities []T
	var total int64

	query := r.db.WithContext(ctx).Model(new(T)).Where(condition)

	// Count total records
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply preloads and pagination
	query = r.applyPreloads(query)
	query = r.applyPagination(query, pagination)

	if err := query.Find(&entities).Error; err != nil {
		return nil, 0, err
	}

	return entities, total, nil
}

// FindAllFiltered retrieves entities with dynamic filtering, sorting, and pagination
// The entity type T must implement filter.Filterable interface
func (r *BaseRepository[T]) FindAllFiltered(ctx context.Context, params *filter.Params) ([]T, int64, error) {
	var entities []T
	var total int64
	var entity T

	// Get filter config if entity implements Filterable
	filterable, ok := any(entity).(filter.Filterable)
	if !ok {
		return nil, 0, errors.New("entity does not implement filter.Filterable interface")
	}
	config := filterable.FilterConfig()

	// Count query (without pagination)
	countQuery := r.db.WithContext(ctx).Model(&entity)
	countQuery = filter.ApplyFiltersOnly(countQuery, config, params)
	if err := countQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Data query (with pagination and sorting)
	query := r.db.WithContext(ctx).Model(&entity)
	query = r.applyPreloads(query)
	query = filter.Apply(query, config, params)

	if err := query.Find(&entities).Error; err != nil {
		return nil, 0, err
	}

	return entities, total, nil
}

// FindOne retrieves a single entity matching conditions.
// Returns ErrNotFound if no entity matches.
func (r *BaseRepository[T]) FindOne(ctx context.Context, condition map[string]any) (*T, error) {
	var entity T
	query := r.applyPreloads(r.db.WithContext(ctx))
	if err := query.Where(condition).First(&entity).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &entity, nil
}

// Update updates an existing entity
func (r *BaseRepository[T]) Update(ctx context.Context, entity *T) error {
	return r.db.WithContext(ctx).Save(entity).Error
}

// UpdateFields updates specific fields of an entity
func (r *BaseRepository[T]) UpdateFields(ctx context.Context, id uint, fields map[string]any) error {
	return r.db.WithContext(ctx).Model(new(T)).Where("id = ?", id).Updates(fields).Error
}

// Delete soft-deletes an entity by ID
func (r *BaseRepository[T]) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(new(T), id).Error
}

// HardDelete permanently removes an entity
func (r *BaseRepository[T]) HardDelete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Unscoped().Delete(new(T), id).Error
}

// Exists checks if an entity exists with given conditions
func (r *BaseRepository[T]) Exists(ctx context.Context, condition map[string]any) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(new(T)).Where(condition).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// Count returns the count of entities matching conditions
func (r *BaseRepository[T]) Count(ctx context.Context, condition map[string]any) (int64, error) {
	var count int64
	query := r.db.WithContext(ctx).Model(new(T))
	if len(condition) > 0 {
		query = query.Where(condition)
	}
	if err := query.Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

// WithTx returns a new repository instance using the provided transaction
func (r *BaseRepository[T]) WithTx(tx *gorm.DB) Repository[T] {
	return &BaseRepository[T]{
		db:       tx,
		preloads: r.preloads,
	}
}

// WithPreload returns a repository with eager loading
func (r *BaseRepository[T]) WithPreload(preloads ...string) Repository[T] {
	newPreloads := make([]string, len(r.preloads)+len(preloads))
	copy(newPreloads, r.preloads)
	copy(newPreloads[len(r.preloads):], preloads)
	return &BaseRepository[T]{
		db:       r.db,
		preloads: newPreloads,
	}
}

// Transaction executes operations within a transaction
func (r *BaseRepository[T]) Transaction(ctx context.Context, fn func(txRepo Repository[T]) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txRepo := r.WithTx(tx)
		return fn(txRepo)
	})
}

// GetDB returns the underlying database connection
func (r *BaseRepository[T]) GetDB() *gorm.DB {
	return r.db
}

// applyPreloads applies eager loading to the query
func (r *BaseRepository[T]) applyPreloads(query *gorm.DB) *gorm.DB {
	for _, preload := range r.preloads {
		query = query.Preload(preload)
	}
	return query
}

// applyPagination applies pagination and sorting to the query
func (r *BaseRepository[T]) applyPagination(query *gorm.DB, pagination *Pagination) *gorm.DB {
	if pagination == nil {
		pagination = NewPagination()
	}

	pagination.Validate()

	if orderClause := pagination.GetOrderClause(); orderClause != "" {
		query = query.Order(orderClause)
	} else {
		// Default ordering by ID descending
		query = query.Order("id DESC")
	}

	return query.Offset(pagination.GetOffset()).Limit(pagination.GetLimit())
}
