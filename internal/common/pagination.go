package common

import (
	"math"
	"regexp"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/voidmaindev/go-template/internal/common/filter"
)

const (
	// DefaultPage is the default page number
	DefaultPage = 1
	// DefaultPageSize is the default number of items per page
	DefaultPageSize = 10
	// MaxPageSize is the maximum allowed page size
	MaxPageSize = 100
)

// AllowedSortFields defines valid sort column names to prevent SQL injection.
// Only lowercase snake_case field names are allowed.
var AllowedSortFields = map[string]bool{
	"id":           true,
	"created_at":   true,
	"updated_at":   true,
	"name":         true,
	"email":        true,
	"code":         true,
	"price":        true,
	"total_amount": true,
	"quantity":     true,
	"population":   true,
}

// sortFieldRegex validates sort field format (alphanumeric and underscores only)
var sortFieldRegex = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)

// Pagination holds pagination parameters
type Pagination struct {
	Page     int    `query:"page" json:"page"`
	PageSize int    `query:"page_size" json:"page_size"`
	Sort     string `query:"sort" json:"sort"`
	Order    string `query:"order" json:"order"` // "asc" or "desc"
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
		PageSize: DefaultPageSize,
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
		PageSize: c.QueryInt("page_size", DefaultPageSize),
		Sort:     sort,
		Order:    c.Query("order", "asc"),
	}
	p.Validate()
	return p
}

// PaginationFromOptional creates pagination from optional pointer values.
// Useful for OpenAPI generated params where page/pageSize are *int.
func PaginationFromOptional(page, pageSize *int, sort, order *string) *Pagination {
	p := &Pagination{
		Page:     intOrDefault(page, DefaultPage),
		PageSize: intOrDefault(pageSize, DefaultPageSize),
		Sort:     stringOrDefault(sort, ""),
		Order:    stringOrDefault(order, "asc"),
	}
	p.Validate()
	return p
}

func intOrDefault(v *int, def int) int {
	if v != nil {
		return *v
	}
	return def
}

func stringOrDefault(v *string, def string) string {
	if v != nil {
		return *v
	}
	return def
}

// Validate ensures pagination values are within acceptable ranges
func (p *Pagination) Validate() {
	if p.Page < 1 {
		p.Page = DefaultPage
	}

	if p.PageSize < 1 {
		p.PageSize = DefaultPageSize
	}

	if p.PageSize > MaxPageSize {
		p.PageSize = MaxPageSize
	}

	if p.Order != "asc" && p.Order != "desc" {
		p.Order = "desc"
	}

	// Validate and sanitize sort field
	if p.Sort != "" && !p.isValidSortField() {
		p.Sort = "" // Reset invalid sort to default
	}
}

// isValidSortField checks if the sort field is safe to use in SQL
func (p *Pagination) isValidSortField() bool {
	if p.Sort == "" {
		return true
	}

	// Normalize to lowercase
	sort := strings.ToLower(p.Sort)

	// Check against allowlist
	if !AllowedSortFields[sort] {
		return false
	}

	// Additional regex check to ensure format is safe
	return sortFieldRegex.MatchString(sort)
}

// IsSortValid returns whether the current sort field is valid
// Use this to check before processing if you want to return an error
func (p *Pagination) IsSortValid() bool {
	if p.Sort == "" {
		return true
	}
	return p.isValidSortField()
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

// FilteredResult wraps filtered data with metadata.
// Deprecated: Use PaginatedResult instead. This type is kept for backward compatibility
// and is structurally identical to PaginatedResult.
type FilteredResult[T any] struct {
	Data       []T   `json:"data"`
	Total      int64 `json:"total"`
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	TotalPages int   `json:"total_pages"`
	HasMore    bool  `json:"has_more"`
}

// CalculateTotalPages computes the number of pages needed for the given total and page size.
func CalculateTotalPages(total int64, pageSize int) int {
	if pageSize <= 0 {
		return 0
	}
	return int(math.Ceil(float64(total) / float64(pageSize)))
}

// NewFilteredResult creates a new filtered result from filter.Params
func NewFilteredResult[T any](data []T, total int64, params *filter.Params) *FilteredResult[T] {
	if params == nil {
		params = filter.DefaultParams()
	}

	pageSize := params.Limit
	if pageSize < 1 {
		pageSize = DefaultPageSize
	}

	page := params.Page
	if page < 1 {
		page = DefaultPage
	}

	totalPages := int(math.Ceil(float64(total) / float64(pageSize)))
	hasMore := page < totalPages

	return &FilteredResult[T]{
		Data:       data,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
		HasMore:    hasMore,
	}
}
