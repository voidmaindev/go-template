package filter

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
		Limit: 10,
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
