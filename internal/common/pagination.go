package common

import (
	"math"
)

const (
	// DefaultPage is the default page number
	DefaultPage = 1
	// DefaultPageSize is the default number of items per page
	DefaultPageSize = 10
	// MaxPageSize is the maximum allowed page size
	MaxPageSize = 100
)

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

// GetOrderClause returns the order clause for database queries
func (p *Pagination) GetOrderClause() string {
	if p.Sort == "" {
		return ""
	}
	p.Validate()
	return p.Sort + " " + p.Order
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
