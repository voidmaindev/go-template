package country

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/voidmaindev/go-template/internal/common"
	"github.com/voidmaindev/go-template/internal/common/filter"
	"github.com/voidmaindev/go-template/pkg/validator"
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
	var req CreateCountryRequest
	if err := c.BodyParser(&req); err != nil {
		return common.BadRequestResponse(c, "invalid request body")
	}

	if errs := validator.Validate(&req); errs != nil {
		return common.ValidationErrorResponse(c, errs)
	}

	country, err := h.service.Create(c.Context(), &req)
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
	id, err := c.ParamsInt("id")
	if err != nil {
		return common.BadRequestResponse(c, "invalid country ID")
	}

	country, err := h.service.GetByID(c.Context(), uint(id))
	if err != nil {
		if errors.Is(err, ErrCountryNotFound) {
			return common.NotFoundResponse(c, "country")
		}
		return common.HandleError(c, err)
	}

	return common.SuccessResponse(c, country.ToResponse())
}

// Update handles country update
func (h *Handler) Update(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return common.BadRequestResponse(c, "invalid country ID")
	}

	var req UpdateCountryRequest
	if err := c.BodyParser(&req); err != nil {
		return common.BadRequestResponse(c, "invalid request body")
	}

	if errs := validator.Validate(&req); errs != nil {
		return common.ValidationErrorResponse(c, errs)
	}

	country, err := h.service.Update(c.Context(), uint(id), &req)
	if err != nil {
		if errors.Is(err, ErrCountryNotFound) {
			return common.NotFoundResponse(c, "country")
		}
		if errors.Is(err, ErrCountryCodeExists) {
			return common.ConflictResponse(c, "country code already exists")
		}
		return common.HandleError(c, err)
	}

	return common.SuccessResponse(c, country.ToResponse())
}

// Delete handles country deletion
func (h *Handler) Delete(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return common.BadRequestResponse(c, "invalid country ID")
	}

	if err := h.service.Delete(c.Context(), uint(id)); err != nil {
		if errors.Is(err, ErrCountryNotFound) {
			return common.NotFoundResponse(c, "country")
		}
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

	return common.SuccessResponse(c, common.NewFilteredResult(responses, result.Total, params))
}
