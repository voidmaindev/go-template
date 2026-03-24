package example_item

import (
	"github.com/gofiber/fiber/v2"
	"github.com/voidmaindev/go-template/internal/common"
	"github.com/voidmaindev/go-template/internal/common/filter"
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
	req, err := common.ParseAndValidate[CreateItemRequest](c)
	if err != nil {
		return nil
	}

	item, err := h.service.Create(c.Context(), req)
	if err != nil {
		return common.HandleError(c, err)
	}

	return common.CreatedResponse(c, item.ToResponse())
}

// GetByID handles getting item by ID
func (h *Handler) GetByID(c *fiber.Ctx) error {
	id, err := common.ParseID(c, "id", "item")
	if err != nil {
		return nil
	}

	item, err := h.service.GetByID(c.Context(), id)
	if err != nil {
		return common.HandleError(c, err)
	}

	return common.SuccessResponse(c, item.ToResponse())
}

// Update handles item update
func (h *Handler) Update(c *fiber.Ctx) error {
	id, err := common.ParseID(c, "id", "item")
	if err != nil {
		return nil
	}

	req, err := common.ParseAndValidate[UpdateItemRequest](c)
	if err != nil {
		return nil
	}

	item, err := h.service.Update(c.Context(), id, req)
	if err != nil {
		return common.HandleError(c, err)
	}

	return common.SuccessResponse(c, item.ToResponse())
}

// Delete handles item deletion
func (h *Handler) Delete(c *fiber.Ctx) error {
	id, err := common.ParseID(c, "id", "item")
	if err != nil {
		return nil
	}

	if err := h.service.Delete(c.Context(), id); err != nil {
		return common.HandleError(c, err)
	}

	return common.DeletedResponse(c)
}

// List handles listing all items with filtering and sorting
func (h *Handler) List(c *fiber.Ctx) error {
	params := filter.ParseFromQuery(c)

	result, err := h.service.ListFiltered(c.Context(), params)
	if err != nil {
		return common.HandleError(c, err)
	}

	return common.SuccessResponse(c, common.MapPaginatedResult(result, func(i Item) ItemResponse {
		return *i.ToResponse()
	}))
}
