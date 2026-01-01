package country

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/voidmaindev/GoTemplate/internal/common"
	"github.com/voidmaindev/GoTemplate/pkg/validator"
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
		return common.InternalServerErrorResponse(c)
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
		return common.InternalServerErrorResponse(c)
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
		return common.InternalServerErrorResponse(c)
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
		return common.InternalServerErrorResponse(c)
	}

	return common.DeletedResponse(c)
}

// List handles listing all countries
func (h *Handler) List(c *fiber.Ctx) error {
	pagination := &common.Pagination{
		Page:     c.QueryInt("page", 1),
		PageSize: c.QueryInt("page_size", 10),
		Sort:     c.Query("sort", "name"),
		Order:    c.Query("order", "asc"),
	}

	result, err := h.service.List(c.Context(), pagination)
	if err != nil {
		return common.InternalServerErrorResponse(c)
	}

	responses := make([]CountryResponse, len(result.Data))
	for i, country := range result.Data {
		responses[i] = *country.ToResponse()
	}

	return common.SuccessResponse(c, common.NewPaginatedResult(responses, result.Total, pagination))
}
