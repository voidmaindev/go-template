package document

import (
	"context"

	"github.com/voidmaindev/go-template/internal/common"
	"github.com/voidmaindev/go-template/internal/common/errors"
	"github.com/voidmaindev/go-template/internal/common/filter"
	"github.com/voidmaindev/go-template/internal/domain/city"
	"github.com/voidmaindev/go-template/internal/domain/item"
	"github.com/voidmaindev/go-template/pkg/utils"
)

// Service defines the document service interface
type Service interface {
	// Document operations
	Create(ctx context.Context, req *CreateDocumentRequest) (*Document, error)
	GetByID(ctx context.Context, id uint) (*Document, error)
	GetByIDWithDetails(ctx context.Context, id uint) (*Document, error)
	Update(ctx context.Context, id uint, req *UpdateDocumentRequest) (*Document, error)
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context, pagination *common.Pagination) (*common.PaginatedResult[Document], error)
	ListFiltered(ctx context.Context, params *filter.Params) (*common.FilteredResult[Document], error)

	// Document item operations
	AddItem(ctx context.Context, documentID uint, req *AddDocumentItemRequest) (*DocumentItem, error)
	UpdateItem(ctx context.Context, documentID, itemID uint, req *UpdateDocumentItemRequest) (*DocumentItem, error)
	RemoveItem(ctx context.Context, documentID, itemID uint) error
}

// service implements the Service interface
type service struct {
	repo         Repository
	itemRepo     ItemRepository
	cityRepo     city.Repository
	productRepo  item.Repository
}

// NewService creates a new document service
func NewService(repo Repository, itemRepo ItemRepository, cityRepo city.Repository, productRepo item.Repository) Service {
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
	cityEntity, err := s.cityRepo.FindByID(ctx, req.CityID)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, ErrCityNotFound
		}
		return nil, errors.Internal(domainName, err).WithOperation("Create")
	}

	// Validate all items exist
	for _, itemReq := range req.Items {
		_, err := s.productRepo.FindByID(ctx, itemReq.ItemID)
		if err != nil {
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

	// Use transaction
	err = s.repo.Transaction(ctx, func(txRepo common.Repository[Document]) error {
		if err := txRepo.Create(ctx, doc); err != nil {
			return err
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
			if err := s.itemRepo.Create(ctx, docItem); err != nil {
				return err
			}
			totalAmount += docItem.GetLineTotal()
		}

		// Update total amount
		doc.TotalAmount = totalAmount
		return s.repo.UpdateTotalAmount(ctx, doc.ID, totalAmount)
	})

	if err != nil {
		return nil, errors.Internal(domainName, err).WithOperation("Create")
	}

	// Return with details
	result, err := s.repo.FindByIDWithDetails(ctx, doc.ID)
	if err != nil {
		return nil, errors.Internal(domainName, err).WithOperation("Create")
	}
	_ = cityEntity // Used for validation
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

// Delete soft-deletes a document and its items
func (s *service) Delete(ctx context.Context, id uint) error {
	_, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.IsNotFound(err) {
			return ErrDocumentNotFound
		}
		return errors.Internal(domainName, err).WithOperation("Delete")
	}

	// Delete items first
	if err := s.itemRepo.DeleteByDocumentID(ctx, id); err != nil {
		return errors.Internal(domainName, err).WithOperation("Delete")
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return errors.Internal(domainName, err).WithOperation("Delete")
	}
	return nil
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
func (s *service) ListFiltered(ctx context.Context, params *filter.Params) (*common.FilteredResult[Document], error) {
	docs, total, err := s.repo.FindAllFilteredWithCity(ctx, params)
	if err != nil {
		return nil, errors.Internal(domainName, err).WithOperation("ListFiltered")
	}

	return common.NewFilteredResult(docs, total, params), nil
}

// AddItem adds an item to a document
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

	if err := s.itemRepo.Create(ctx, docItem); err != nil {
		return nil, errors.Internal(domainName, err).WithOperation("AddItem")
	}

	// Recalculate total
	if err := s.recalculateTotal(ctx, documentID); err != nil {
		return nil, err
	}

	// Load item for response
	docItem.Item = *product
	return docItem, nil
}

// UpdateItem updates a document item
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

	if err := s.itemRepo.Update(ctx, docItem); err != nil {
		return nil, errors.Internal(domainName, err).WithOperation("UpdateItem")
	}

	// Recalculate total
	if err := s.recalculateTotal(ctx, documentID); err != nil {
		return nil, err
	}

	return docItem, nil
}

// RemoveItem removes an item from a document
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

	if err := s.itemRepo.Delete(ctx, itemID); err != nil {
		return errors.Internal(domainName, err).WithOperation("RemoveItem")
	}

	// Recalculate total
	return s.recalculateTotal(ctx, documentID)
}

// recalculateTotal recalculates the total amount of a document
func (s *service) recalculateTotal(ctx context.Context, documentID uint) error {
	items, err := s.itemRepo.FindByDocumentID(ctx, documentID)
	if err != nil {
		return errors.Internal(domainName, err).WithOperation("recalculateTotal")
	}

	var total int64
	for _, item := range items {
		total += item.GetLineTotal()
	}

	if err := s.repo.UpdateTotalAmount(ctx, documentID, total); err != nil {
		return errors.Internal(domainName, err).WithOperation("recalculateTotal")
	}
	return nil
}
