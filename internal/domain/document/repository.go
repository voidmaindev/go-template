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

	// RunTransaction runs a function within a database transaction, providing the
	// raw *gorm.DB for cross-repository transactions (e.g., document + item repos).
	RunTransaction(ctx context.Context, fn func(tx *gorm.DB) error) error

	// WithDocTx returns a document repository scoped to the given transaction.
	WithDocTx(tx *gorm.DB) Repository
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

	// WithItemTx returns an item repository scoped to the given transaction.
	WithItemTx(tx *gorm.DB) ItemRepository
}

// repository implements the Repository interface
type repository struct {
	*common.BaseRepository[Document]
}

// itemRepository implements the ItemRepository interface
type itemRepository struct {
	*common.BaseRepository[DocumentItem]
}

// NewRepository creates a new document repository
func NewRepository(db *gorm.DB) Repository {
	return &repository{
		BaseRepository: common.NewBaseRepository[Document](db, "document"),
	}
}

// NewItemRepository creates a new document item repository
func NewItemRepository(db *gorm.DB) ItemRepository {
	return &itemRepository{
		BaseRepository: common.NewBaseRepository[DocumentItem](db, "document item"),
	}
}

// FindByCode finds a document by code
func (r *repository) FindByCode(ctx context.Context, code string) (*Document, error) {
	return r.FindOne(ctx, map[string]any{"code": code})
}

// FindByIDWithDetails finds a document with all related data
func (r *repository) FindByIDWithDetails(ctx context.Context, id uint) (*Document, error) {
	var doc Document
	err := r.DB().WithContext(ctx).
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
	return r.WithPreload("City").FindAll(ctx, pagination)
}

// FindAllFilteredWithCity finds all documents with filtering, sorting, and city preloaded
func (r *repository) FindAllFilteredWithCity(ctx context.Context, params *filter.Params) ([]Document, int64, error) {
	return r.WithPreload("City").FindAllFiltered(ctx, params)
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
	err := r.DB().WithContext(ctx).Where("document_id = ?", documentID).Find(&items).Error
	return items, err
}

// FindByDocumentIDWithItem finds all items with item preloaded
func (r *itemRepository) FindByDocumentIDWithItem(ctx context.Context, documentID uint) ([]DocumentItem, error) {
	var items []DocumentItem
	err := r.DB().WithContext(ctx).Preload("Item").Where("document_id = ?", documentID).Find(&items).Error
	return items, err
}

// DeleteByDocumentID deletes all items for a document
func (r *itemRepository) DeleteByDocumentID(ctx context.Context, documentID uint) error {
	return r.DB().WithContext(ctx).Where("document_id = ?", documentID).Delete(&DocumentItem{}).Error
}

// RunTransaction runs a function within a database transaction, providing the
// raw *gorm.DB for cross-repository transactions.
func (r *repository) RunTransaction(ctx context.Context, fn func(tx *gorm.DB) error) error {
	return r.DB().WithContext(ctx).Transaction(fn)
}

// WithDocTx returns a document repository scoped to the given transaction.
func (r *repository) WithDocTx(tx *gorm.DB) Repository {
	return &repository{
		BaseRepository: common.NewBaseRepository[Document](tx, "document"),
	}
}

// WithItemTx returns an item repository scoped to the given transaction.
func (r *itemRepository) WithItemTx(tx *gorm.DB) ItemRepository {
	return &itemRepository{
		BaseRepository: common.NewBaseRepository[DocumentItem](tx, "document item"),
	}
}
