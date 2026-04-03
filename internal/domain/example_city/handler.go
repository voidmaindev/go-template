package example_city

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/voidmaindev/go-template/internal/common"
	"github.com/voidmaindev/go-template/internal/common/filter"
)

// Handler handles HTTP requests for cities
type Handler struct {
	service Service
}

// NewHandler creates a new city handler
func NewHandler(service Service) *Handler {
	return &Handler{
		service: service,
	}
}

// Create handles city creation
func (h *Handler) Create(c *fiber.Ctx) error {
	req, err := common.ParseAndValidate[CreateCityRequest](c)
	if err != nil {
		return nil // response already sent
	}

	city, err := h.service.Create(c.Context(), req)
	if err != nil {
		// ErrCountryNotFound → 400: the referenced country ID in the request is invalid
		if errors.Is(err, ErrCountryNotFound) {
			return common.BadRequestResponse(c, "country not found")
		}
		return common.HandleError(c, err)
	}

	return common.CreatedResponse(c, city.ToResponse())
}

// GetByID handles getting city by ID
func (h *Handler) GetByID(c *fiber.Ctx) error {
	id, err := common.ParseID(c, "id", "city")
	if err != nil {
		return nil // response already sent
	}

	city, err := h.service.GetByIDWithCountry(c.Context(), id)
	if err != nil {
		return common.HandleError(c, err)
	}

	return common.SuccessResponse(c, city.ToResponse())
}

// Update handles city update
func (h *Handler) Update(c *fiber.Ctx) error {
	id, err := common.ParseID(c, "id", "city")
	if err != nil {
		return nil // response already sent
	}

	req, err := common.ParseAndValidate[UpdateCityRequest](c)
	if err != nil {
		return nil // response already sent
	}

	city, err := h.service.Update(c.Context(), id, req)
	if err != nil {
		// ErrCountryNotFound → 400: the referenced country ID in the request is invalid
		if errors.Is(err, ErrCountryNotFound) {
			return common.BadRequestResponse(c, "country not found")
		}
		return common.HandleError(c, err)
	}

	return common.SuccessResponse(c, city.ToResponse())
}

// Delete handles city deletion
func (h *Handler) Delete(c *fiber.Ctx) error {
	id, err := common.ParseID(c, "id", "city")
	if err != nil {
		return nil // response already sent
	}

	if err := h.service.Delete(c.Context(), id); err != nil {
		return common.HandleError(c, err)
	}

	return common.DeletedResponse(c)
}

// List handles listing all cities with filtering and sorting
func (h *Handler) List(c *fiber.Ctx) error {
	params := filter.ParseFromQuery(c)

	result, err := h.service.ListFiltered(c.Context(), params)
	if err != nil {
		return common.HandleError(c, err)
	}

	responses := make([]CityResponse, len(result.Data))
	for i, city := range result.Data {
		responses[i] = *city.ToResponse()
	}

	return common.SuccessResponse(c, common.NewPaginatedResultFromFilter(responses, result.Total, params))
}

// ListByCountry handles listing cities by country
func (h *Handler) ListByCountry(c *fiber.Ctx) error {
	countryID, err := common.ParseID(c, "countryId", "country")
	if err != nil {
		return nil // response already sent
	}

	pagination := common.PaginationFromQuery(c, "name")

	result, err := h.service.ListByCountry(c.Context(), countryID, pagination)
	if err != nil {
		// ErrCountryNotFound → 404: the country resource itself doesn't exist
		if errors.Is(err, ErrCountryNotFound) {
			return common.NotFoundResponse(c, "country")
		}
		return common.HandleError(c, err)
	}

	responses := make([]CityResponse, len(result.Data))
	for i, city := range result.Data {
		responses[i] = *city.ToResponse()
	}

	return common.SuccessResponse(c, common.NewPaginatedResult(responses, result.Total, pagination))
}
