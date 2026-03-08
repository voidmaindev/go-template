package common

import (
	"math"
	"regexp"
	"strings"
	"sync"

	"github.com/gofiber/fiber/v2"
	"github.com/voidmaindev/go-template/internal/common/filter"
	"github.com/voidmaindev/go-template/pkg/ptr"
)

// DefaultPage is the default page number (not configurable)
const DefaultPage = 1

// paginationConfig holds the current pagination settings
// These can be overridden via InitPagination at startup
var (
	paginationMu        sync.RWMutex
	defaultPageSize     = 10  // Default items per page
	maxPageSize         = 100 // Maximum allowed page size
)

// InitPagination initializes pagination defaults from configuration.
// Should be called once at application startup after config is loaded.
func InitPagination(defPageSize, maxSize int) {
	paginationMu.Lock()
	defer paginationMu.Unlock()

	if defPageSize > 0 {
		defaultPageSize = defPageSize
	}
	if maxSize > 0 {
		maxPageSize = maxSize
	}
	// Ensure default doesn't exceed max
	if defaultPageSize > maxPageSize {
		defaultPageSize = maxPageSize
	}
}

// GetDefaultPageSize returns the configured default page size
func GetDefaultPageSize() int {
	paginationMu.RLock()
	defer paginationMu.RUnlock()
	return defaultPageSize
}

// GetMaxPageSize returns the configured maximum page size
func GetMaxPageSize() int {
	paginationMu.RLock()
	defer paginationMu.RUnlock()
	return maxPageSize
}

// sortFieldRegex validates sort field format (alphanumeric and underscores only)
var sortFieldRegex = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)

// Pagination holds pagination parameters
type Pagination struct {
	Page      int    `query:"page" json:"page"`
	PageSize  int    `query:"page_size" json:"page_size"`
	Sort      string `query:"sort" json:"sort"`
	Order     string `query:"order" json:"order"` // "asc" or "desc"
	validated bool   // cached validation flag to avoid redundant Validate() calls
}

// PaginatedResult wraps paginated data with metadata
type PaginatedResult[T any] struct {
	Data       []T   `json:"data"`
	Total      int64 `json:"total"`
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	TotalPages int   `json:"total_pages"`
	HasMore    bool  `json:"has_more"`
}

// NewPagination creates a new pagination with default values
func NewPagination() *Pagination {
	return &Pagination{
		Page:     DefaultPage,
		PageSize: GetDefaultPageSize(),
		Order:    "desc",
	}
}

// NewPaginationWithParams creates a new pagination with specified values
func NewPaginationWithParams(page, pageSize int, sort, order string) *Pagination {
	p := &Pagination{
		Page:     page,
		PageSize: pageSize,
		Sort:     sort,
		Order:    order,
	}
	p.Validate()
	return p
}

// PaginationFromQuery parses pagination parameters from Fiber query string.
// Supports optional default sort field for domain-specific defaults.
func PaginationFromQuery(c *fiber.Ctx, defaultSort ...string) *Pagination {
	sort := ""
	if len(defaultSort) > 0 {
		sort = defaultSort[0]
	}
	if s := c.Query("sort"); s != "" {
		sort = s
	}

	p := &Pagination{
		Page:     c.QueryInt("page", DefaultPage),
		PageSize: c.QueryInt("page_size", GetDefaultPageSize()),
		Sort:     sort,
		Order:    c.Query("order", "desc"),
	}
	p.Validate()
	return p
}

// PaginationFromOptional creates pagination from optional pointer values.
// Useful for OpenAPI generated params where page/pageSize are *int.
func PaginationFromOptional(page, pageSize *int, sort, order *string) *Pagination {
	p := &Pagination{
		Page:     ptr.DerefOr(page, DefaultPage),
		PageSize: ptr.DerefOr(pageSize, GetDefaultPageSize()),
		Sort:     ptr.DerefOr(sort, ""),
		Order:    ptr.DerefOr(order, "desc"),
	}
	p.Validate()
	return p
}

// Validate ensures pagination values are within acceptable ranges.
// Uses a cached flag so repeated calls are a no-op.
func (p *Pagination) Validate() {
	if p.validated {
		return
	}
	defer func() { p.validated = true }()

	if p.Page < 1 {
		p.Page = DefaultPage
	}

	defPageSize := GetDefaultPageSize()
	maxSize := GetMaxPageSize()

	if p.PageSize < 1 {
		p.PageSize = defPageSize
	}

	if p.PageSize > maxSize {
		p.PageSize = maxSize
	}

	if p.Order != "asc" && p.Order != "desc" {
		p.Order = "desc"
	}

	// Validate and sanitize sort field format (must be lowercase alphanumeric with underscores).
	// Actual field-level validation (is the field sortable?) is handled per-model
	// via IsSortValidForConfig / ValidateSortWithConfig.
	if p.Sort != "" {
		p.Sort = strings.ToLower(p.Sort)
		if !sortFieldRegex.MatchString(p.Sort) {
			p.Sort = "" // Reset invalid sort to default
		}
	}
}

// IsSortValidForConfig validates the sort field against a model's filter.Config.
// This is the preferred method as it uses per-model field definitions instead of
// the global AllowedSortFields whitelist.
func (p *Pagination) IsSortValidForConfig(config filter.Config) bool {
	if p.Sort == "" {
		return true
	}

	sort := strings.ToLower(p.Sort)

	// Check if field exists in config and is sortable
	fieldConfig, ok := config.Fields[sort]
	if !ok {
		return false
	}

	return fieldConfig.Sortable
}

// ValidateSortWithConfig validates and sanitizes the sort field using a model's
// filter.Config. If the sort field is not valid for the config, it is reset to empty.
// Returns true if the sort field was valid, false if it was reset.
func (p *Pagination) ValidateSortWithConfig(config filter.Config) bool {
	if !p.IsSortValidForConfig(config) {
		p.Sort = ""
		return false
	}
	return true
}

// GetOffset returns the offset for database queries
func (p *Pagination) GetOffset() int {
	p.Validate()
	return (p.Page - 1) * p.PageSize
}

// GetLimit returns the limit for database queries
func (p *Pagination) GetLimit() int {
	p.Validate()
	return p.PageSize
}

// GetOrderClause returns the order clause for database queries.
// Returns empty string if sort field is invalid or empty.
func (p *Pagination) GetOrderClause() string {
	p.Validate() // This sanitizes invalid sort fields

	if p.Sort == "" {
		return ""
	}

	// Use normalized lowercase for consistent SQL
	return strings.ToLower(p.Sort) + " " + p.Order
}

// NewPaginatedResult creates a new paginated result
func NewPaginatedResult[T any](data []T, total int64, pagination *Pagination) *PaginatedResult[T] {
	if pagination == nil {
		pagination = NewPagination()
	}
	pagination.Validate()

	totalPages := int(math.Ceil(float64(total) / float64(pagination.PageSize)))
	hasMore := pagination.Page < totalPages

	return &PaginatedResult[T]{
		Data:       data,
		Total:      total,
		Page:       pagination.Page,
		PageSize:   pagination.PageSize,
		TotalPages: totalPages,
		HasMore:    hasMore,
	}
}

// CalculateTotalPages computes the number of pages needed for the given total and page size.
func CalculateTotalPages(total int64, pageSize int) int {
	if pageSize <= 0 {
		return 0
	}
	return int(math.Ceil(float64(total) / float64(pageSize)))
}

// NewPaginatedResultFromFilter creates a PaginatedResult from filter.Params.
func NewPaginatedResultFromFilter[T any](data []T, total int64, params *filter.Params) *PaginatedResult[T] {
	if params == nil {
		params = filter.DefaultParams()
	}

	pageSize := params.Limit
	if pageSize < 1 {
		pageSize = GetDefaultPageSize()
	}

	page := params.Page
	if page < 1 {
		page = DefaultPage
	}

	totalPages := int(math.Ceil(float64(total) / float64(pageSize)))
	hasMore := page < totalPages

	return &PaginatedResult[T]{
		Data:       data,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
		HasMore:    hasMore,
	}
}
