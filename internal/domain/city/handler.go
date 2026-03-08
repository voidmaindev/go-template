package city

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/voidmaindev/go-template/internal/common"
	"github.com/voidmaindev/go-template/internal/common/filter"
	"github.com/voidmaindev/go-template/pkg/validator"
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
	var req CreateCityRequest
	if err := c.BodyParser(&req); err != nil {
		return common.BadRequestResponse(c, "invalid request body")
	}

	if errs := validator.Validate(&req); errs != nil {
		return common.ValidationErrorResponse(c, errs)
	}

	city, err := h.service.Create(c.Context(), &req)
	if err != nil {
		if errors.Is(err, ErrCountryNotFound) {
			return common.BadRequestResponse(c, "country not found")
		}
		return common.HandleError(c, err)
	}

	return common.CreatedResponse(c, city.ToResponse())
}

// GetByID handles getting city by ID
func (h *Handler) GetByID(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return common.BadRequestResponse(c, "invalid city ID")
	}

	city, err := h.service.GetByIDWithCountry(c.Context(), uint(id))
	if err != nil {
		if errors.Is(err, ErrCityNotFound) {
			return common.NotFoundResponse(c, "city")
		}
		return common.HandleError(c, err)
	}

	return common.SuccessResponse(c, city.ToResponse())
}

// Update handles city update
func (h *Handler) Update(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return common.BadRequestResponse(c, "invalid city ID")
	}

	var req UpdateCityRequest
	if err := c.BodyParser(&req); err != nil {
		return common.BadRequestResponse(c, "invalid request body")
	}

	if errs := validator.Validate(&req); errs != nil {
		return common.ValidationErrorResponse(c, errs)
	}

	city, err := h.service.Update(c.Context(), uint(id), &req)
	if err != nil {
		if errors.Is(err, ErrCityNotFound) {
			return common.NotFoundResponse(c, "city")
		}
		if errors.Is(err, ErrCountryNotFound) {
			return common.BadRequestResponse(c, "country not found")
		}
		return common.HandleError(c, err)
	}

	return common.SuccessResponse(c, city.ToResponse())
}

// Delete handles city deletion
func (h *Handler) Delete(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return common.BadRequestResponse(c, "invalid city ID")
	}

	if err := h.service.Delete(c.Context(), uint(id)); err != nil {
		if errors.Is(err, ErrCityNotFound) {
			return common.NotFoundResponse(c, "city")
		}
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
	countryID, err := c.ParamsInt("countryId")
	if err != nil {
		return common.BadRequestResponse(c, "invalid country ID")
	}

	pagination := common.PaginationFromQuery(c, "name")

	result, err := h.service.ListByCountry(c.Context(), uint(countryID), pagination)
	if err != nil {
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
