package common

import (
	"testing"
)

func TestPagination_GetOrderClause_Sanitizes(t *testing.T) {
	tests := []struct {
		name          string
		sort          string
		order         string
		expectedOrder string
	}{
		{"valid sort returns clause", "id", "asc", "id asc"},
		{"valid sort desc", "name", "desc", "name desc"},
		{"invalid sort returns empty", "id; DROP TABLE", "asc", ""},
		{"empty sort returns empty", "", "asc", ""},
		{"invalid order defaults to desc", "id", "invalid", "id desc"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Pagination{Sort: tt.sort, Order: tt.order}
			got := p.GetOrderClause()
			if got != tt.expectedOrder {
				t.Errorf("GetOrderClause() = %q, want %q", got, tt.expectedOrder)
			}
		})
	}
}

func TestPagination_Validate(t *testing.T) {
	tests := []struct {
		name             string
		page             int
		pageSize         int
		order            string
		expectedPage     int
		expectedPageSize int
		expectedOrder    string
	}{
		{"valid values unchanged", 1, 10, "asc", 1, 10, "asc"},
		{"zero page becomes 1", 0, 10, "asc", 1, 10, "asc"},
		{"negative page becomes 1", -5, 10, "asc", 1, 10, "asc"},
		{"zero pageSize becomes default", 1, 0, "asc", 1, 10, "asc"},
		{"negative pageSize becomes default", 1, -5, "asc", 1, 10, "asc"},
		{"pageSize over max becomes max", 1, 500, "asc", 1, 100, "asc"},
		{"invalid order becomes desc", 1, 10, "invalid", 1, 10, "desc"},
		{"empty order becomes desc", 1, 10, "", 1, 10, "desc"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Pagination{Page: tt.page, PageSize: tt.pageSize, Order: tt.order}
			p.Validate()

			if p.Page != tt.expectedPage {
				t.Errorf("Page = %d, want %d", p.Page, tt.expectedPage)
			}
			if p.PageSize != tt.expectedPageSize {
				t.Errorf("PageSize = %d, want %d", p.PageSize, tt.expectedPageSize)
			}
			if p.Order != tt.expectedOrder {
				t.Errorf("Order = %s, want %s", p.Order, tt.expectedOrder)
			}
		})
	}
}

func TestPagination_Validate_SanitizesSort(t *testing.T) {
	p := &Pagination{
		Page:     1,
		PageSize: 10,
		Sort:     "id; DROP TABLE users;--",
		Order:    "asc",
	}

	p.Validate()

	if p.Sort != "" {
		t.Errorf("Sort was not sanitized, got %q, want empty string", p.Sort)
	}
}
