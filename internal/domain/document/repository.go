package document

import (
	"context"
	"errors"

	"github.com/voidmaindev/go-template/internal/common"
	"github.com/voidmaindev/go-template/internal/common/filter"
	"gorm.io/gorm"
)

// Repository defines the document repository interface
type Repository interface {
	common.Repository[Document]

	// FindByCode finds a document by code
	FindByCode(ctx context.Context, code string) (*Document, error)

	// FindByIDWithDetails finds a document with all related data
	FindByIDWithDetails(ctx context.Context, id uint) (*Document, error)

	// FindAllWithCity finds all documents with city preloaded
	FindAllWithCity(ctx context.Context, pagination *common.Pagination) ([]Document, int64, error)

	// FindAllFilteredWithCity finds all documents with filtering, sorting, and city preloaded
	FindAllFilteredWithCity(ctx context.Context, params *filter.Params) ([]Document, int64, error)

	// FindByCityID finds all documents by city ID
	FindByCityID(ctx context.Context, cityID uint, pagination *common.Pagination) ([]Document, int64, error)

	// ExistsByCode checks if a document with the given code exists
	ExistsByCode(ctx context.Context, code string) (bool, error)

	// UpdateTotalAmount updates the total amount of a document
	UpdateTotalAmount(ctx context.Context, id uint, totalAmount int64) error
}

// ItemRepository defines the document item repository interface
type ItemRepository interface {
	common.Repository[DocumentItem]

	// FindByDocumentID finds all items for a document
	FindByDocumentID(ctx context.Context, documentID uint) ([]DocumentItem, error)

	// FindByDocumentIDWithItem finds all items with item preloaded
	FindByDocumentIDWithItem(ctx context.Context, documentID uint) ([]DocumentItem, error)

	// DeleteByDocumentID deletes all items for a document
	DeleteByDocumentID(ctx context.Context, documentID uint) error
}

// repository implements the Repository interface
type repository struct {
	*common.BaseRepository[Document]
	db *gorm.DB
}

// itemRepository implements the ItemRepository interface
type itemRepository struct {
	*common.BaseRepository[DocumentItem]
	db *gorm.DB
}

// NewRepository creates a new document repository
func NewRepository(db *gorm.DB) Repository {
	return &repository{
		BaseRepository: common.NewBaseRepository[Document](db),
		db:             db,
	}
}

// NewItemRepository creates a new document item repository
func NewItemRepository(db *gorm.DB) ItemRepository {
	return &itemRepository{
		BaseRepository: common.NewBaseRepository[DocumentItem](db),
		db:             db,
	}
}

// FindByCode finds a document by code
func (r *repository) FindByCode(ctx context.Context, code string) (*Document, error) {
	return r.FindOne(ctx, map[string]any{"code": code})
}

// FindByIDWithDetails finds a document with all related data
func (r *repository) FindByIDWithDetails(ctx context.Context, id uint) (*Document, error) {
	var doc Document
	err := r.db.WithContext(ctx).
		Preload("City").
		Preload("City.Country").
		Preload("Items").
		Preload("Items.Item").
		First(&doc, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrDocumentNotFound
		}
		return nil, err
	}
	return &doc, nil
}

// FindAllWithCity finds all documents with city preloaded
func (r *repository) FindAllWithCity(ctx context.Context, pagination *common.Pagination) ([]Document, int64, error) {
	var docs []Document
	var total int64

	query := r.db.WithContext(ctx).Model(&Document{})

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	query = query.Preload("City")
	if pagination != nil {
		pagination.Validate()
		if orderClause := pagination.GetOrderClause(); orderClause != "" {
			query = query.Order(orderClause)
		} else {
			query = query.Order("id DESC")
		}
		query = query.Offset(pagination.GetOffset()).Limit(pagination.GetLimit())
	}

	if err := query.Find(&docs).Error; err != nil {
		return nil, 0, err
	}

	return docs, total, nil
}

// FindAllFilteredWithCity finds all documents with filtering, sorting, and city preloaded
func (r *repository) FindAllFilteredWithCity(ctx context.Context, params *filter.Params) ([]Document, int64, error) {
	var docs []Document
	var total int64
	var doc Document

	config := doc.FilterConfig()

	// Count query (without pagination)
	countQuery := r.db.WithContext(ctx).Model(&Document{})
	countQuery = filter.ApplyFiltersOnly(countQuery, config, params)
	if err := countQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Data query (with pagination, sorting, and preload)
	query := r.db.WithContext(ctx).Model(&Document{}).Preload("City")
	query = filter.Apply(query, config, params)

	if err := query.Find(&docs).Error; err != nil {
		return nil, 0, err
	}

	return docs, total, nil
}

// FindByCityID finds all documents by city ID
func (r *repository) FindByCityID(ctx context.Context, cityID uint, pagination *common.Pagination) ([]Document, int64, error) {
	return r.FindByCondition(ctx, map[string]any{"city_id": cityID}, pagination)
}

// ExistsByCode checks if a document with the given code exists
func (r *repository) ExistsByCode(ctx context.Context, code string) (bool, error) {
	return r.Exists(ctx, map[string]any{"code": code})
}

// UpdateTotalAmount updates the total amount of a document
func (r *repository) UpdateTotalAmount(ctx context.Context, id uint, totalAmount int64) error {
	return r.UpdateFields(ctx, id, map[string]any{"total_amount": totalAmount})
}

// FindByDocumentID finds all items for a document
func (r *itemRepository) FindByDocumentID(ctx context.Context, documentID uint) ([]DocumentItem, error) {
	var items []DocumentItem
	err := r.db.WithContext(ctx).Where("document_id = ?", documentID).Find(&items).Error
	return items, err
}

// FindByDocumentIDWithItem finds all items with item preloaded
func (r *itemRepository) FindByDocumentIDWithItem(ctx context.Context, documentID uint) ([]DocumentItem, error) {
	var items []DocumentItem
	err := r.db.WithContext(ctx).Preload("Item").Where("document_id = ?", documentID).Find(&items).Error
	return items, err
}

// DeleteByDocumentID deletes all items for a document
func (r *itemRepository) DeleteByDocumentID(ctx context.Context, documentID uint) error {
	return r.db.WithContext(ctx).Where("document_id = ?", documentID).Delete(&DocumentItem{}).Error
}
