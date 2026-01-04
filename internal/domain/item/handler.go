package item

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/voidmaindev/go-template/internal/common"
	"github.com/voidmaindev/go-template/internal/common/filter"
	"github.com/voidmaindev/go-template/pkg/validator"
)

// Handler handles HTTP requests for items
type Handler struct {
	service Service
}

// NewHandler creates a new item handler
func NewHandler(service Service) *Handler {
	return &Handler{
		service: service,
	}
}

// Create handles item creation
func (h *Handler) Create(c *fiber.Ctx) error {
	var req CreateItemRequest
	if err := c.BodyParser(&req); err != nil {
		return common.BadRequestResponse(c, "invalid request body")
	}

	if errs := validator.Validate(&req); errs != nil {
		return common.ValidationErrorResponse(c, errs)
	}

	item, err := h.service.Create(c.Context(), &req)
	if err != nil {
		return common.InternalServerErrorResponse(c)
	}

	return common.CreatedResponse(c, item.ToResponse())
}

// GetByID handles getting item by ID
func (h *Handler) GetByID(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return common.BadRequestResponse(c, "invalid item ID")
	}

	item, err := h.service.GetByID(c.Context(), uint(id))
	if err != nil {
		if errors.Is(err, ErrItemNotFound) {
			return common.NotFoundResponse(c, "item")
		}
		return common.InternalServerErrorResponse(c)
	}

	return common.SuccessResponse(c, item.ToResponse())
}

// Update handles item update
func (h *Handler) Update(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return common.BadRequestResponse(c, "invalid item ID")
	}

	var req UpdateItemRequest
	if err := c.BodyParser(&req); err != nil {
		return common.BadRequestResponse(c, "invalid request body")
	}

	if errs := validator.Validate(&req); errs != nil {
		return common.ValidationErrorResponse(c, errs)
	}

	item, err := h.service.Update(c.Context(), uint(id), &req)
	if err != nil {
		if errors.Is(err, ErrItemNotFound) {
			return common.NotFoundResponse(c, "item")
		}
		return common.InternalServerErrorResponse(c)
	}

	return common.SuccessResponse(c, item.ToResponse())
}

// Delete handles item deletion
func (h *Handler) Delete(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return common.BadRequestResponse(c, "invalid item ID")
	}

	if err := h.service.Delete(c.Context(), uint(id)); err != nil {
		if errors.Is(err, ErrItemNotFound) {
			return common.NotFoundResponse(c, "item")
		}
		return common.InternalServerErrorResponse(c)
	}

	return common.DeletedResponse(c)
}

// List handles listing all items with filtering and sorting
func (h *Handler) List(c *fiber.Ctx) error {
	params := filter.ParseFromQuery(c)

	result, err := h.service.ListFiltered(c.Context(), params)
	if err != nil {
		return common.InternalServerErrorResponse(c)
	}

	// Convert to response DTOs
	responses := make([]ItemResponse, len(result.Data))
	for i, item := range result.Data {
		responses[i] = *item.ToResponse()
	}

	return common.SuccessResponse(c, common.NewFilteredResult(responses, result.Total, params))
}
