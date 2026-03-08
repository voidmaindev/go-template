package document

import (
	"context"

	"github.com/voidmaindev/go-template/internal/common"
	"github.com/voidmaindev/go-template/internal/common/errors"
	"github.com/voidmaindev/go-template/internal/common/filter"
	"github.com/voidmaindev/go-template/internal/domain/city"
	"github.com/voidmaindev/go-template/internal/domain/item"
	"github.com/voidmaindev/go-template/pkg/utils"
	"gorm.io/gorm"
)

// CityReader is the minimal interface for city lookups.
// Defines only the methods used by the document service.
type CityReader interface {
	FindByID(ctx context.Context, id uint) (*city.City, error)
}

// ItemReader is the minimal interface for item lookups.
// Defines only the methods used by the document service.
type ItemReader interface {
	FindByID(ctx context.Context, id uint) (*item.Item, error)
}

// Service defines the document service interface
type Service interface {
	// Document operations
	Create(ctx context.Context, req *CreateDocumentRequest) (*Document, error)
	GetByID(ctx context.Context, id uint) (*Document, error)
	GetByIDWithDetails(ctx context.Context, id uint) (*Document, error)
	Update(ctx context.Context, id uint, req *UpdateDocumentRequest) (*Document, error)
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context, pagination *common.Pagination) (*common.PaginatedResult[Document], error)
	ListFiltered(ctx context.Context, params *filter.Params) (*common.PaginatedResult[Document], error)

	// Document item operations
	AddItem(ctx context.Context, documentID uint, req *AddDocumentItemRequest) (*DocumentItem, error)
	UpdateItem(ctx context.Context, documentID, itemID uint, req *UpdateDocumentItemRequest) (*DocumentItem, error)
	RemoveItem(ctx context.Context, documentID, itemID uint) error
}

// service implements the Service interface
type service struct {
	repo        Repository
	itemRepo    ItemRepository
	cityRepo    CityReader
	productRepo ItemReader
}

// NewService creates a new document service
func NewService(repo Repository, itemRepo ItemRepository, cityRepo CityReader, productRepo ItemReader) Service {
	return &service{
		repo:        repo,
		itemRepo:    itemRepo,
		cityRepo:    cityRepo,
		productRepo: productRepo,
	}
}

// Create creates a new document with items
func (s *service) Create(ctx context.Context, req *CreateDocumentRequest) (*Document, error) {
	// Check if code already exists
	exists, err := s.repo.ExistsByCode(ctx, req.Code)
	if err != nil {
		return nil, errors.Internal(domainName, err).WithOperation("Create")
	}
	if exists {
		return nil, ErrDocumentCodeExists
	}

	// Validate city exists
	if _, err := s.cityRepo.FindByID(ctx, req.CityID); err != nil {
		if errors.IsNotFound(err) {
			return nil, ErrCityNotFound
		}
		return nil, errors.Internal(domainName, err).WithOperation("Create")
	}

	// Validate all items exist
	for _, itemReq := range req.Items {
		if _, err := s.productRepo.FindByID(ctx, itemReq.ItemID); err != nil {
			if errors.IsNotFound(err) {
				return nil, ErrItemNotFound
			}
			return nil, errors.Internal(domainName, err).WithOperation("Create")
		}
	}

	// Create document
	doc := &Document{
		Code:         req.Code,
		CityID:       req.CityID,
		DocumentDate: req.DocumentDate,
		TotalAmount:  0,
	}

	// Use transaction covering both document and item repos
	err = s.repo.RunTransaction(ctx, func(tx *gorm.DB) error {
		txDocRepo := s.repo.WithDocTx(tx)
		txItemRepo := s.itemRepo.WithItemTx(tx)

		if err := txDocRepo.Create(ctx, doc); err != nil {
			return errors.Internal(domainName, err).WithOperation("Create.Document")
		}

		// Create items
		var totalAmount int64
		for _, itemReq := range req.Items {
			docItem := &DocumentItem{
				DocumentID: doc.ID,
				ItemID:     itemReq.ItemID,
				Quantity:   itemReq.Quantity,
				Price:      itemReq.Price,
			}
			if err := txItemRepo.Create(ctx, docItem); err != nil {
				return errors.Internal(domainName, err).WithOperation("Create.DocumentItem")
			}
			totalAmount += docItem.GetLineTotal()
		}

		// Update total amount
		doc.TotalAmount = totalAmount
		if err := txDocRepo.UpdateTotalAmount(ctx, doc.ID, totalAmount); err != nil {
			return errors.Internal(domainName, err).WithOperation("Create.UpdateTotalAmount")
		}
		return nil
	})

	if err != nil {
		return nil, errors.Internal(domainName, err).WithOperation("Create")
	}

	// Return with details
	result, err := s.repo.FindByIDWithDetails(ctx, doc.ID)
	if err != nil {
		return nil, errors.Internal(domainName, err).WithOperation("Create")
	}
	return result, nil
}

// GetByID retrieves a document by ID
func (s *service) GetByID(ctx context.Context, id uint) (*Document, error) {
	doc, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, ErrDocumentNotFound
		}
		return nil, errors.Internal(domainName, err).WithOperation("GetByID")
	}
	return doc, nil
}

// GetByIDWithDetails retrieves a document with all related data
func (s *service) GetByIDWithDetails(ctx context.Context, id uint) (*Document, error) {
	doc, err := s.repo.FindByIDWithDetails(ctx, id)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, ErrDocumentNotFound
		}
		return nil, errors.Internal(domainName, err).WithOperation("GetByIDWithDetails")
	}
	return doc, nil
}

// Update updates a document
func (s *service) Update(ctx context.Context, id uint, req *UpdateDocumentRequest) (*Document, error) {
	doc, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, ErrDocumentNotFound
		}
		return nil, errors.Internal(domainName, err).WithOperation("Update")
	}

	// Handle Code validation separately (unique constraint)
	if req.Code != nil && *req.Code != doc.Code {
		existing, err := s.repo.FindByCode(ctx, *req.Code)
		if err != nil && !errors.IsNotFound(err) {
			return nil, errors.Internal(domainName, err).WithOperation("Update")
		}
		if existing != nil && existing.ID != id {
			return nil, ErrDocumentCodeExists
		}
		doc.Code = *req.Code
		req.Code = nil // Prevent copier from overwriting
	}

	// Handle CityID validation separately (FK constraint)
	if req.CityID != nil {
		_, err := s.cityRepo.FindByID(ctx, *req.CityID)
		if err != nil {
			if errors.IsNotFound(err) {
				return nil, ErrCityNotFound
			}
			return nil, errors.Internal(domainName, err).WithOperation("Update")
		}
		doc.CityID = *req.CityID
		req.CityID = nil // Prevent copier from overwriting
	}

	// Map remaining non-nil fields from request to model
	if err := utils.UpdateModel(doc, req); err != nil {
		return nil, errors.Internal(domainName, err).WithOperation("Update")
	}

	if err := s.repo.Update(ctx, doc); err != nil {
		return nil, errors.Internal(domainName, err).WithOperation("Update")
	}

	result, err := s.repo.FindByIDWithDetails(ctx, id)
	if err != nil {
		return nil, errors.Internal(domainName, err).WithOperation("Update")
	}
	return result, nil
}

// Delete soft-deletes a document and its items within a single transaction
func (s *service) Delete(ctx context.Context, id uint) error {
	_, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.IsNotFound(err) {
			return ErrDocumentNotFound
		}
		return errors.Internal(domainName, err).WithOperation("Delete")
	}

	return s.repo.RunTransaction(ctx, func(tx *gorm.DB) error {
		txItemRepo := s.itemRepo.WithItemTx(tx)
		txDocRepo := s.repo.WithDocTx(tx)

		if err := txItemRepo.DeleteByDocumentID(ctx, id); err != nil {
			return errors.Internal(domainName, err).WithOperation("Delete.Items")
		}
		if err := txDocRepo.Delete(ctx, id); err != nil {
			return errors.Internal(domainName, err).WithOperation("Delete.Document")
		}
		return nil
	})
}

// List retrieves all documents with pagination
func (s *service) List(ctx context.Context, pagination *common.Pagination) (*common.PaginatedResult[Document], error) {
	docs, total, err := s.repo.FindAllWithCity(ctx, pagination)
	if err != nil {
		return nil, errors.Internal(domainName, err).WithOperation("List")
	}

	return common.NewPaginatedResult(docs, total, pagination), nil
}

// ListFiltered retrieves documents with dynamic filtering, sorting, and pagination
func (s *service) ListFiltered(ctx context.Context, params *filter.Params) (*common.PaginatedResult[Document], error) {
	docs, total, err := s.repo.FindAllFilteredWithCity(ctx, params)
	if err != nil {
		return nil, errors.Internal(domainName, err).WithOperation("ListFiltered")
	}

	return common.NewPaginatedResultFromFilter(docs, total, params), nil
}

// AddItem adds an item to a document within a transaction
func (s *service) AddItem(ctx context.Context, documentID uint, req *AddDocumentItemRequest) (*DocumentItem, error) {
	// Validate document exists
	_, err := s.repo.FindByID(ctx, documentID)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, ErrDocumentNotFound
		}
		return nil, errors.Internal(domainName, err).WithOperation("AddItem")
	}

	// Validate item exists
	product, err := s.productRepo.FindByID(ctx, req.ItemID)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, ErrItemNotFound
		}
		return nil, errors.Internal(domainName, err).WithOperation("AddItem")
	}

	// Create document item
	docItem := &DocumentItem{
		DocumentID: documentID,
		ItemID:     req.ItemID,
		Quantity:   req.Quantity,
		Price:      req.Price,
	}

	// Create item and recalculate total atomically
	err = s.repo.RunTransaction(ctx, func(tx *gorm.DB) error {
		txItemRepo := s.itemRepo.WithItemTx(tx)
		txDocRepo := s.repo.WithDocTx(tx)

		if err := txItemRepo.Create(ctx, docItem); err != nil {
			return errors.Internal(domainName, err).WithOperation("AddItem.Create")
		}
		return s.recalculateTotalTx(ctx, txItemRepo, txDocRepo, documentID)
	})
	if err != nil {
		return nil, err
	}

	// Load item for response
	docItem.Item = *product
	return docItem, nil
}

// UpdateItem updates a document item within a transaction
func (s *service) UpdateItem(ctx context.Context, documentID, itemID uint, req *UpdateDocumentItemRequest) (*DocumentItem, error) {
	// Validate document exists
	_, err := s.repo.FindByID(ctx, documentID)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, ErrDocumentNotFound
		}
		return nil, errors.Internal(domainName, err).WithOperation("UpdateItem")
	}

	// Find document item
	docItem, err := s.itemRepo.FindByID(ctx, itemID)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, ErrDocumentItemNotFound
		}
		return nil, errors.Internal(domainName, err).WithOperation("UpdateItem")
	}
	if docItem.DocumentID != documentID {
		return nil, ErrDocumentItemNotFound
	}

	// Map non-nil fields from request to model
	if err := utils.UpdateModel(docItem, req); err != nil {
		return nil, errors.Internal(domainName, err).WithOperation("UpdateItem")
	}

	// Update item and recalculate total atomically
	err = s.repo.RunTransaction(ctx, func(tx *gorm.DB) error {
		txItemRepo := s.itemRepo.WithItemTx(tx)
		txDocRepo := s.repo.WithDocTx(tx)

		if err := txItemRepo.Update(ctx, docItem); err != nil {
			return errors.Internal(domainName, err).WithOperation("UpdateItem.Update")
		}
		return s.recalculateTotalTx(ctx, txItemRepo, txDocRepo, documentID)
	})
	if err != nil {
		return nil, err
	}

	return docItem, nil
}

// RemoveItem removes an item from a document within a transaction
func (s *service) RemoveItem(ctx context.Context, documentID, itemID uint) error {
	// Validate document exists
	_, err := s.repo.FindByID(ctx, documentID)
	if err != nil {
		if errors.IsNotFound(err) {
			return ErrDocumentNotFound
		}
		return errors.Internal(domainName, err).WithOperation("RemoveItem")
	}

	// Find document item
	docItem, err := s.itemRepo.FindByID(ctx, itemID)
	if err != nil {
		if errors.IsNotFound(err) {
			return ErrDocumentItemNotFound
		}
		return errors.Internal(domainName, err).WithOperation("RemoveItem")
	}
	if docItem.DocumentID != documentID {
		return ErrDocumentItemNotFound
	}

	// Delete item and recalculate total atomically
	return s.repo.RunTransaction(ctx, func(tx *gorm.DB) error {
		txItemRepo := s.itemRepo.WithItemTx(tx)
		txDocRepo := s.repo.WithDocTx(tx)

		if err := txItemRepo.Delete(ctx, itemID); err != nil {
			return errors.Internal(domainName, err).WithOperation("RemoveItem.Delete")
		}
		return s.recalculateTotalTx(ctx, txItemRepo, txDocRepo, documentID)
	})
}

// recalculateTotalTx recalculates the total amount of a document within a transaction
func (s *service) recalculateTotalTx(ctx context.Context, txItemRepo ItemRepository, txDocRepo Repository, documentID uint) error {
	items, err := txItemRepo.FindByDocumentID(ctx, documentID)
	if err != nil {
		return errors.Internal(domainName, err).WithOperation("recalculateTotal")
	}

	var total int64
	for _, item := range items {
		total += item.GetLineTotal()
	}

	if err := txDocRepo.UpdateTotalAmount(ctx, documentID, total); err != nil {
		return errors.Internal(domainName, err).WithOperation("recalculateTotal")
	}
	return nil
}
