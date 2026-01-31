package audit

import (
	"github.com/gofiber/fiber/v2"
	"github.com/voidmaindev/go-template/internal/common"
	"github.com/voidmaindev/go-template/internal/common/filter"
)

// Handler handles HTTP requests for audit logs
type Handler struct {
	service Service
}

// NewHandler creates a new audit handler
func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// List retrieves audit logs with filtering and pagination
// GET /audit/logs
func (h *Handler) List(c *fiber.Ctx) error {
	params := filter.ParseFromQuery(c)

	result, err := h.service.ListFiltered(c.Context(), params)
	if err != nil {
		return common.InternalServerErrorResponse(c)
	}

	// Convert to response DTOs
	responses := make([]AuditLogResponse, len(result.Data))
	for i, log := range result.Data {
		responses[i] = *log.ToResponse()
	}

	return common.SuccessResponse(c, common.NewFilteredResult(responses, result.Total, params))
}

// ListByUser retrieves audit logs for a specific user
// GET /audit/users/:id/logs
func (h *Handler) ListByUser(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil || id <= 0 {
		return common.BadRequestResponse(c, "invalid user ID")
	}

	pagination := common.PaginationFromQuery(c, "timestamp")
	pagination.ValidateSortWithConfig(AuditLog{}.FilterConfig())

	result, err := h.service.ListByUserID(c.Context(), uint(id), pagination)
	if err != nil {
		return common.InternalServerErrorResponse(c)
	}

	// Convert to response DTOs
	responses := make([]AuditLogResponse, len(result.Data))
	for i, log := range result.Data {
		responses[i] = *log.ToResponse()
	}

	return common.SuccessResponse(c, common.NewPaginatedResult(responses, result.Total, pagination))
}
