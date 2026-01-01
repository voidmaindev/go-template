package document

import (
	"context"

	"github.com/voidmaindev/GoTemplate/internal/common"
	"gorm.io/gorm"
)

// Repository defines the document repository interface
type Repository interface {
	common.Repository[Document]

	// FindByCode finds a document by code
	FindByCode(ctx context.Context, code string) (*Document, error)

	// FindByIDWithDetails finds a document with all related data
	FindByIDWithDetails(ctx context.Context, id uint) (*Document, error)

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
	return r.FindOne(ctx, map[string]interface{}{"code": code})
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
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &doc, nil
}

// FindByCityID finds all documents by city ID
func (r *repository) FindByCityID(ctx context.Context, cityID uint, pagination *common.Pagination) ([]Document, int64, error) {
	return r.FindByCondition(ctx, map[string]interface{}{"city_id": cityID}, pagination)
}

// ExistsByCode checks if a document with the given code exists
func (r *repository) ExistsByCode(ctx context.Context, code string) (bool, error) {
	return r.Exists(ctx, map[string]interface{}{"code": code})
}

// UpdateTotalAmount updates the total amount of a document
func (r *repository) UpdateTotalAmount(ctx context.Context, id uint, totalAmount int64) error {
	return r.UpdateFields(ctx, id, map[string]interface{}{"total_amount": totalAmount})
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
