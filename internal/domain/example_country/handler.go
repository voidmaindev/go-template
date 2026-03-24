package example_country

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/voidmaindev/go-template/internal/common"
	"github.com/voidmaindev/go-template/internal/common/filter"
)

// Handler handles HTTP requests for countries
type Handler struct {
	service Service
}

// NewHandler creates a new country handler
func NewHandler(service Service) *Handler {
	return &Handler{
		service: service,
	}
}

// Create handles country creation
func (h *Handler) Create(c *fiber.Ctx) error {
	req, err := common.ParseAndValidate[CreateCountryRequest](c)
	if err != nil {
		return nil
	}

	country, err := h.service.Create(c.Context(), req)
	if err != nil {
		if errors.Is(err, ErrCountryCodeExists) {
			return common.ConflictResponse(c, "country code already exists")
		}
		return common.HandleError(c, err)
	}

	return common.CreatedResponse(c, country.ToResponse())
}

// GetByID handles getting country by ID
func (h *Handler) GetByID(c *fiber.Ctx) error {
	id, err := common.ParseID(c, "id", "country")
	if err != nil {
		return nil
	}

	country, err := h.service.GetByID(c.Context(), id)
	if err != nil {
		return common.HandleError(c, err)
	}

	return common.SuccessResponse(c, country.ToResponse())
}

// Update handles country update
func (h *Handler) Update(c *fiber.Ctx) error {
	id, err := common.ParseID(c, "id", "country")
	if err != nil {
		return nil
	}

	req, err := common.ParseAndValidate[UpdateCountryRequest](c)
	if err != nil {
		return nil
	}

	country, err := h.service.Update(c.Context(), id, req)
	if err != nil {
		if errors.Is(err, ErrCountryCodeExists) {
			return common.ConflictResponse(c, "country code already exists")
		}
		return common.HandleError(c, err)
	}

	return common.SuccessResponse(c, country.ToResponse())
}

// Delete handles country deletion
func (h *Handler) Delete(c *fiber.Ctx) error {
	id, err := common.ParseID(c, "id", "country")
	if err != nil {
		return nil
	}

	if err := h.service.Delete(c.Context(), id); err != nil {
		return common.HandleError(c, err)
	}

	return common.DeletedResponse(c)
}

// List handles listing all countries with filtering and sorting
func (h *Handler) List(c *fiber.Ctx) error {
	params := filter.ParseFromQuery(c)

	result, err := h.service.ListFiltered(c.Context(), params)
	if err != nil {
		return common.HandleError(c, err)
	}

	responses := make([]CountryResponse, len(result.Data))
	for i, country := range result.Data {
		responses[i] = *country.ToResponse()
	}

	return common.SuccessResponse(c, common.NewPaginatedResultFromFilter(responses, result.Total, params))
}
