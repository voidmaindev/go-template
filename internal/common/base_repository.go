package common

import (
	"context"
	"database/sql"
	"errors"

	commonerrors "github.com/voidmaindev/go-template/internal/common/errors"
	"github.com/voidmaindev/go-template/internal/common/filter"
	"gorm.io/gorm"
)

const repositoryDomain = "repository"

// BaseRepository provides a generic implementation of Repository interface
type BaseRepository[T any] struct {
	db         *gorm.DB
	preloads   []string
	entityName string
}

// NewBaseRepository creates a new instance of BaseRepository
func NewBaseRepository[T any](db *gorm.DB, entityName string) *BaseRepository[T] {
	return &BaseRepository[T]{
		db:         db,
		preloads:   []string{},
		entityName: entityName,
	}
}

// Create inserts a new entity into the database
func (r *BaseRepository[T]) Create(ctx context.Context, entity *T) error {
	if err := r.db.WithContext(ctx).Create(entity).Error; err != nil {
		return r.wrapError("Create", err)
	}
	return nil
}

// CreateBatch inserts multiple entities in batches
func (r *BaseRepository[T]) CreateBatch(ctx context.Context, entities []T, batchSize int) error {
	if batchSize <= 0 {
		batchSize = 100
	}
	if err := r.db.WithContext(ctx).CreateInBatches(entities, batchSize).Error; err != nil {
		return r.wrapError("CreateBatch", err)
	}
	return nil
}

// FindByID retrieves an entity by its primary key.
// Returns a typed not found error if the entity doesn't exist.
func (r *BaseRepository[T]) FindByID(ctx context.Context, id uint) (*T, error) {
	var entity T
	query := r.applyPreloads(r.db.WithContext(ctx))
	if err := query.First(&entity, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, commonerrors.NotFound(repositoryDomain, r.entityName)
		}
		return nil, r.wrapError("FindByID", err)
	}
	return &entity, nil
}

// FindAll retrieves all entities with pagination
func (r *BaseRepository[T]) FindAll(ctx context.Context, pagination *Pagination) ([]T, int64, error) {
	var entities []T
	var total int64

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		query := tx.Model(new(T))

		// Count total records
		if err := query.Count(&total).Error; err != nil {
			return err
		}

		// Apply preloads and pagination
		query = r.applyPreloads(query)
		query = r.applyPagination(query, pagination)

		return query.Find(&entities).Error
	}, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return nil, 0, r.wrapError("FindAll", err)
	}

	return entities, total, nil
}

// FindByCondition retrieves entities matching conditions
func (r *BaseRepository[T]) FindByCondition(ctx context.Context, condition map[string]any, pagination *Pagination) ([]T, int64, error) {
	var entities []T
	var total int64

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		query := tx.Model(new(T)).Where(condition)

		// Count total records
		if err := query.Count(&total).Error; err != nil {
			return err
		}

		// Apply preloads and pagination
		query = r.applyPreloads(query)
		query = r.applyPagination(query, pagination)

		return query.Find(&entities).Error
	}, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return nil, 0, r.wrapError("FindByCondition", err)
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

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Count query (without pagination)
		countQuery := tx.Model(&entity)
		countQuery = filter.ApplyFiltersOnly(countQuery, config, params)
		if err := countQuery.Count(&total).Error; err != nil {
			return err
		}

		// Data query (with pagination and sorting)
		query := tx.Model(&entity)
		query = r.applyPreloads(query)
		query = filter.Apply(query, config, params)

		return query.Find(&entities).Error
	}, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return nil, 0, r.wrapError("FindAllFiltered", err)
	}

	return entities, total, nil
}

// FindOne retrieves a single entity matching conditions.
// Returns a typed not found error if no entity matches.
func (r *BaseRepository[T]) FindOne(ctx context.Context, condition map[string]any) (*T, error) {
	var entity T
	query := r.applyPreloads(r.db.WithContext(ctx))
	if err := query.Where(condition).First(&entity).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, commonerrors.NotFound(repositoryDomain, r.entityName)
		}
		return nil, r.wrapError("FindOne", err)
	}
	return &entity, nil
}

// Update updates an existing entity
func (r *BaseRepository[T]) Update(ctx context.Context, entity *T) error {
	if err := r.db.WithContext(ctx).Save(entity).Error; err != nil {
		return r.wrapError("Update", err)
	}
	return nil
}

// UpdateFields updates specific fields of an entity
func (r *BaseRepository[T]) UpdateFields(ctx context.Context, id uint, fields map[string]any) error {
	if err := r.db.WithContext(ctx).Model(new(T)).Where("id = ?", id).Updates(fields).Error; err != nil {
		return r.wrapError("UpdateFields", err)
	}
	return nil
}

// Delete soft-deletes an entity by ID
func (r *BaseRepository[T]) Delete(ctx context.Context, id uint) error {
	if err := r.db.WithContext(ctx).Delete(new(T), id).Error; err != nil {
		return r.wrapError("Delete", err)
	}
	return nil
}

// HardDelete permanently removes an entity
func (r *BaseRepository[T]) HardDelete(ctx context.Context, id uint) error {
	if err := r.db.WithContext(ctx).Unscoped().Delete(new(T), id).Error; err != nil {
		return r.wrapError("HardDelete", err)
	}
	return nil
}

// Exists checks if an entity exists with given conditions.
// Uses SELECT 1 LIMIT 1 instead of COUNT(*) for better performance on large tables.
func (r *BaseRepository[T]) Exists(ctx context.Context, condition map[string]any) (bool, error) {
	var exists bool
	err := r.db.WithContext(ctx).Model(new(T)).
		Select("1").
		Where(condition).
		Limit(1).
		Scan(&exists).Error
	if err != nil {
		return false, r.wrapError("Exists", err)
	}
	return exists, nil
}

// Count returns the count of entities matching conditions
func (r *BaseRepository[T]) Count(ctx context.Context, condition map[string]any) (int64, error) {
	var count int64
	query := r.db.WithContext(ctx).Model(new(T))
	if len(condition) > 0 {
		query = query.Where(condition)
	}
	if err := query.Count(&count).Error; err != nil {
		return 0, r.wrapError("Count", err)
	}
	return count, nil
}

// WithTx returns a new repository instance using the provided transaction
func (r *BaseRepository[T]) WithTx(tx *gorm.DB) Repository[T] {
	return &BaseRepository[T]{
		db:         tx,
		preloads:   r.preloads,
		entityName: r.entityName,
	}
}

// WithPreload returns a repository with eager loading
func (r *BaseRepository[T]) WithPreload(preloads ...string) Repository[T] {
	newPreloads := make([]string, len(r.preloads)+len(preloads))
	copy(newPreloads, r.preloads)
	copy(newPreloads[len(r.preloads):], preloads)
	return &BaseRepository[T]{
		db:         r.db,
		preloads:   newPreloads,
		entityName: r.entityName,
	}
}

// Transaction executes operations within a transaction
func (r *BaseRepository[T]) Transaction(ctx context.Context, fn func(txRepo Repository[T]) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txRepo := r.WithTx(tx)
		return fn(txRepo)
	})
}

// DB returns the underlying database connection for use by embedding repositories.
// This is intentionally not part of the Repository[T] interface to avoid leaking
// infrastructure concerns to consumers.
func (r *BaseRepository[T]) DB() *gorm.DB {
	return r.db
}

// wrapError wraps a database error with repository context (domain, operation, entity name).
// Preserves existing DomainErrors and wraps raw errors as Internal.
func (r *BaseRepository[T]) wrapError(operation string, err error) error {
	if err == nil {
		return nil
	}
	return commonerrors.Internal(repositoryDomain, err).
		WithOperation(operation).
		WithDetail("entity", r.entityName)
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
