package common

import (
	"github.com/voidmaindev/go-template/internal/common/filter"
	"gorm.io/gorm"
)

// QueryOptions holds all query modifiers
type QueryOptions struct {
	Preloads   []string
	Pagination *Pagination
	Filter     *filter.Params
	OrderBy    string
	OrderDir   string
	Limit      int
	Offset     int
}

// QueryOption is a function that modifies QueryOptions
type QueryOption func(*QueryOptions)

// WithPreload adds eager loading for relations
func WithPreload(relations ...string) QueryOption {
	return func(o *QueryOptions) {
		o.Preloads = append(o.Preloads, relations...)
	}
}

// WithPagination adds pagination
func WithPagination(p *Pagination) QueryOption {
	return func(o *QueryOptions) {
		o.Pagination = p
	}
}

// WithFilter adds dynamic filtering
func WithFilter(params *filter.Params) QueryOption {
	return func(o *QueryOptions) {
		o.Filter = params
	}
}

// WithSort adds sorting
func WithSort(field, direction string) QueryOption {
	return func(o *QueryOptions) {
		o.OrderBy = field
		o.OrderDir = direction
	}
}

// WithLimit adds a limit
func WithLimit(limit int) QueryOption {
	return func(o *QueryOptions) {
		o.Limit = limit
	}
}

// WithOffset adds an offset
func WithOffset(offset int) QueryOption {
	return func(o *QueryOptions) {
		o.Offset = offset
	}
}

// ApplyOptions collects all options into QueryOptions
func ApplyOptions(opts ...QueryOption) *QueryOptions {
	options := &QueryOptions{}
	for _, opt := range opts {
		opt(options)
	}
	return options
}

// ApplyToGorm applies options to a GORM query
func (o *QueryOptions) ApplyToGorm(db *gorm.DB) *gorm.DB {
	query := db

	// Apply preloads
	for _, preload := range o.Preloads {
		query = query.Preload(preload)
	}

	// Apply sorting
	if o.OrderBy != "" {
		dir := "ASC"
		if o.OrderDir == "desc" {
			dir = "DESC"
		}
		query = query.Order(o.OrderBy + " " + dir)
	}

	// Apply limit (non-pagination)
	if o.Limit > 0 {
		query = query.Limit(o.Limit)
	}

	// Apply offset (non-pagination)
	if o.Offset > 0 {
		query = query.Offset(o.Offset)
	}

	return query
}

// ApplyPaginationToGorm applies pagination to a GORM query
func (o *QueryOptions) ApplyPaginationToGorm(db *gorm.DB) *gorm.DB {
	if o.Pagination == nil {
		return db
	}
	offset := (o.Pagination.Page - 1) * o.Pagination.PageSize
	return db.Offset(offset).Limit(o.Pagination.PageSize)
}

// HasPagination returns true if pagination is set
func (o *QueryOptions) HasPagination() bool {
	return o.Pagination != nil
}

// HasPreloads returns true if preloads are set
func (o *QueryOptions) HasPreloads() bool {
	return len(o.Preloads) > 0
}

// HasSort returns true if sorting is set
func (o *QueryOptions) HasSort() bool {
	return o.OrderBy != ""
}

// HasFilter returns true if filter is set
func (o *QueryOptions) HasFilter() bool {
	return o.Filter != nil
}

// Clone creates a copy of QueryOptions
func (o *QueryOptions) Clone() *QueryOptions {
	clone := &QueryOptions{
		Preloads:   make([]string, len(o.Preloads)),
		Pagination: o.Pagination,
		Filter:     o.Filter,
		OrderBy:    o.OrderBy,
		OrderDir:   o.OrderDir,
		Limit:      o.Limit,
		Offset:     o.Offset,
	}
	copy(clone.Preloads, o.Preloads)
	return clone
}

// ================================
// Builder pattern for QueryOptions
// ================================

// QueryBuilder provides a fluent interface for building queries
type QueryBuilder struct {
	options *QueryOptions
}

// NewQueryBuilder creates a new QueryBuilder
func NewQueryBuilder() *QueryBuilder {
	return &QueryBuilder{
		options: &QueryOptions{},
	}
}

// Preload adds eager loading
func (b *QueryBuilder) Preload(relations ...string) *QueryBuilder {
	b.options.Preloads = append(b.options.Preloads, relations...)
	return b
}

// Paginate adds pagination
func (b *QueryBuilder) Paginate(page, pageSize int) *QueryBuilder {
	b.options.Pagination = &Pagination{
		Page:     page,
		PageSize: pageSize,
	}
	return b
}

// Sort adds sorting
func (b *QueryBuilder) Sort(field string, desc bool) *QueryBuilder {
	b.options.OrderBy = field
	if desc {
		b.options.OrderDir = "desc"
	} else {
		b.options.OrderDir = "asc"
	}
	return b
}

// Limit sets the limit
func (b *QueryBuilder) Limit(limit int) *QueryBuilder {
	b.options.Limit = limit
	return b
}

// Offset sets the offset
func (b *QueryBuilder) Offset(offset int) *QueryBuilder {
	b.options.Offset = offset
	return b
}

// Filter sets the filter params
func (b *QueryBuilder) Filter(params *filter.Params) *QueryBuilder {
	b.options.Filter = params
	return b
}

// Build returns the QueryOptions
func (b *QueryBuilder) Build() *QueryOptions {
	return b.options
}

// Apply applies the options to a GORM query
func (b *QueryBuilder) Apply(db *gorm.DB) *gorm.DB {
	return b.options.ApplyToGorm(db)
}
