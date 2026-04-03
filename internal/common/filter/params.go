package filter

import "sync"

// paginationDefaults holds configurable page size limits.
// Initialized via InitPagination at startup; safe for concurrent reads after init.
var (
	paginationMu     sync.RWMutex
	defaultPageSize  = 10
	maxPageSize      = 100
)

// InitPagination sets default and maximum page sizes for the filter package.
// Should be called once at startup after config is loaded.
func InitPagination(defSize, maxSize int) {
	paginationMu.Lock()
	defer paginationMu.Unlock()
	if defSize > 0 {
		defaultPageSize = defSize
	}
	if maxSize > 0 {
		maxPageSize = maxSize
	}
	if defaultPageSize > maxPageSize {
		defaultPageSize = maxPageSize
	}
}

// getDefaultPageSize returns the configured default page size.
func getDefaultPageSize() int {
	paginationMu.RLock()
	defer paginationMu.RUnlock()
	return defaultPageSize
}

// getMaxPageSize returns the configured maximum page size.
func getMaxPageSize() int {
	paginationMu.RLock()
	defer paginationMu.RUnlock()
	return maxPageSize
}

// FilterParam represents a single filter condition
type FilterParam struct {
	Field    string   // Field name (e.g., "name", "example_country.name")
	Operator Operator // Filter operator
	Value    string   // Filter value (as string, will be converted based on field type)
}

// SortParam represents a sort directive
type SortParam struct {
	Field string
	Desc  bool
}

// Params holds all parsed filter and sort parameters
type Params struct {
	Filters []FilterParam
	Sort    []SortParam
	Page    int
	Limit   int
}

// DefaultParams returns default pagination parameters
func DefaultParams() *Params {
	return &Params{
		Page:  1,
		Limit: getDefaultPageSize(),
	}
}

// HasFilters returns true if there are any filter parameters
func (p *Params) HasFilters() bool {
	return len(p.Filters) > 0
}

// HasSort returns true if there are any sort parameters
func (p *Params) HasSort() bool {
	return len(p.Sort) > 0
}

// Offset calculates the offset for pagination
func (p *Params) Offset() int {
	if p.Page < 1 {
		return 0
	}
	return (p.Page - 1) * p.Limit
}
