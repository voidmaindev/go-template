package document

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/voidmaindev/go-template/internal/common"
	"github.com/voidmaindev/go-template/internal/common/filter"
	"github.com/voidmaindev/go-template/pkg/validator"
)

// Handler handles HTTP requests for documents
type Handler struct {
	service Service
}

// NewHandler creates a new document handler
func NewHandler(service Service) *Handler {
	return &Handler{
		service: service,
	}
}

// Create handles document creation
func (h *Handler) Create(c *fiber.Ctx) error {
	var req CreateDocumentRequest
	if err := c.BodyParser(&req); err != nil {
		return common.BadRequestResponse(c, "invalid request body")
	}

	if errs := validator.Validate(&req); errs != nil {
		return common.ValidationErrorResponse(c, errs)
	}

	doc, err := h.service.Create(c.Context(), &req)
	if err != nil {
		switch {
		case errors.Is(err, ErrDocumentCodeExists):
			return common.ConflictResponse(c, "document code already exists")
		case errors.Is(err, ErrCityNotFound):
			return common.BadRequestResponse(c, "city not found")
		case errors.Is(err, ErrItemNotFound):
			return common.BadRequestResponse(c, "one or more items not found")
		default:
			return common.HandleError(c, err)
		}
	}

	return common.CreatedResponse(c, doc.ToResponse())
}

// GetByID handles getting document by ID
func (h *Handler) GetByID(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return common.BadRequestResponse(c, "invalid document ID")
	}

	doc, err := h.service.GetByIDWithDetails(c.Context(), uint(id))
	if err != nil {
		if errors.Is(err, ErrDocumentNotFound) {
			return common.NotFoundResponse(c, "document")
		}
		return common.HandleError(c, err)
	}

	return common.SuccessResponse(c, doc.ToResponse())
}

// Update handles document update
func (h *Handler) Update(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return common.BadRequestResponse(c, "invalid document ID")
	}

	var req UpdateDocumentRequest
	if err := c.BodyParser(&req); err != nil {
		return common.BadRequestResponse(c, "invalid request body")
	}

	if errs := validator.Validate(&req); errs != nil {
		return common.ValidationErrorResponse(c, errs)
	}

	doc, err := h.service.Update(c.Context(), uint(id), &req)
	if err != nil {
		switch {
		case errors.Is(err, ErrDocumentNotFound):
			return common.NotFoundResponse(c, "document")
		case errors.Is(err, ErrDocumentCodeExists):
			return common.ConflictResponse(c, "document code already exists")
		case errors.Is(err, ErrCityNotFound):
			return common.BadRequestResponse(c, "city not found")
		default:
			return common.HandleError(c, err)
		}
	}

	return common.SuccessResponse(c, doc.ToResponse())
}

// Delete handles document deletion
func (h *Handler) Delete(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return common.BadRequestResponse(c, "invalid document ID")
	}

	if err := h.service.Delete(c.Context(), uint(id)); err != nil {
		if errors.Is(err, ErrDocumentNotFound) {
			return common.NotFoundResponse(c, "document")
		}
		return common.HandleError(c, err)
	}

	return common.DeletedResponse(c)
}

// List handles listing all documents with filtering and sorting
func (h *Handler) List(c *fiber.Ctx) error {
	params := filter.ParseFromQuery(c)

	result, err := h.service.ListFiltered(c.Context(), params)
	if err != nil {
		return common.HandleError(c, err)
	}

	responses := make([]DocumentResponse, len(result.Data))
	for i, doc := range result.Data {
		responses[i] = *doc.ToResponse()
	}

	return common.SuccessResponse(c, common.NewPaginatedResultFromFilter(responses, result.Total, params))
}

// AddItem handles adding item to document
func (h *Handler) AddItem(c *fiber.Ctx) error {
	documentID, err := c.ParamsInt("id")
	if err != nil {
		return common.BadRequestResponse(c, "invalid document ID")
	}

	var req AddDocumentItemRequest
	if err := c.BodyParser(&req); err != nil {
		return common.BadRequestResponse(c, "invalid request body")
	}

	if errs := validator.Validate(&req); errs != nil {
		return common.ValidationErrorResponse(c, errs)
	}

	docItem, err := h.service.AddItem(c.Context(), uint(documentID), &req)
	if err != nil {
		switch {
		case errors.Is(err, ErrDocumentNotFound):
			return common.NotFoundResponse(c, "document")
		case errors.Is(err, ErrItemNotFound):
			return common.BadRequestResponse(c, "item not found")
		default:
			return common.HandleError(c, err)
		}
	}

	return common.CreatedResponse(c, docItem.ToResponse())
}

// UpdateItem handles updating document item
func (h *Handler) UpdateItem(c *fiber.Ctx) error {
	documentID, err := c.ParamsInt("id")
	if err != nil {
		return common.BadRequestResponse(c, "invalid document ID")
	}

	itemID, err := c.ParamsInt("itemId")
	if err != nil {
		return common.BadRequestResponse(c, "invalid item ID")
	}

	var req UpdateDocumentItemRequest
	if err := c.BodyParser(&req); err != nil {
		return common.BadRequestResponse(c, "invalid request body")
	}

	if errs := validator.Validate(&req); errs != nil {
		return common.ValidationErrorResponse(c, errs)
	}

	docItem, err := h.service.UpdateItem(c.Context(), uint(documentID), uint(itemID), &req)
	if err != nil {
		switch {
		case errors.Is(err, ErrDocumentNotFound):
			return common.NotFoundResponse(c, "document")
		case errors.Is(err, ErrDocumentItemNotFound):
			return common.NotFoundResponse(c, "document item")
		default:
			return common.HandleError(c, err)
		}
	}

	return common.SuccessResponse(c, docItem.ToResponse())
}

// RemoveItem handles removing item from document
func (h *Handler) RemoveItem(c *fiber.Ctx) error {
	documentID, err := c.ParamsInt("id")
	if err != nil {
		return common.BadRequestResponse(c, "invalid document ID")
	}

	itemID, err := c.ParamsInt("itemId")
	if err != nil {
		return common.BadRequestResponse(c, "invalid item ID")
	}

	if err := h.service.RemoveItem(c.Context(), uint(documentID), uint(itemID)); err != nil {
		switch {
		case errors.Is(err, ErrDocumentNotFound):
			return common.NotFoundResponse(c, "document")
		case errors.Is(err, ErrDocumentItemNotFound):
			return common.NotFoundResponse(c, "document item")
		default:
			return common.HandleError(c, err)
		}
	}

	return common.DeletedResponse(c)
}
