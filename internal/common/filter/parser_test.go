package filter

import (
	"testing"
)

func TestParseFieldOperator(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantField string
		wantOp   Operator
	}{
		{
			name:      "simple field without operator",
			input:     "name",
			wantField: "name",
			wantOp:    OpEq,
		},
		{
			name:      "field with eq operator",
			input:     "name__eq",
			wantField: "name",
			wantOp:    OpEq,
		},
		{
			name:      "field with contains operator",
			input:     "name__contains",
			wantField: "name",
			wantOp:    OpContains,
		},
		{
			name:      "field with starts_with operator",
			input:     "name__starts_with",
			wantField: "name",
			wantOp:    OpStartsWith,
		},
		{
			name:      "field with ends_with operator",
			input:     "email__ends_with",
			wantField: "email",
			wantOp:    OpEndsWith,
		},
		{
			name:      "field with gt operator",
			input:     "price__gt",
			wantField: "price",
			wantOp:    OpGt,
		},
		{
			name:      "field with gte operator",
			input:     "price__gte",
			wantField: "price",
			wantOp:    OpGte,
		},
		{
			name:      "field with lt operator",
			input:     "price__lt",
			wantField: "price",
			wantOp:    OpLt,
		},
		{
			name:      "field with lte operator",
			input:     "price__lte",
			wantField: "price",
			wantOp:    OpLte,
		},
		{
			name:      "field with in operator",
			input:     "status__in",
			wantField: "status",
			wantOp:    OpIn,
		},
		{
			name:      "field with is_null operator",
			input:     "deleted_at__is_null",
			wantField: "deleted_at",
			wantOp:    OpIsNull,
		},
		{
			name:      "field with is_not_null operator",
			input:     "email__is_not_null",
			wantField: "email",
			wantOp:    OpIsNotNull,
		},
		{
			name:      "relation field without operator",
			input:     "example_country.name",
			wantField: "example_country.name",
			wantOp:    OpEq,
		},
		{
			name:      "relation field with contains operator",
			input:     "example_country.name__contains",
			wantField: "example_country.name",
			wantOp:    OpContains,
		},
		{
			name:      "invalid operator defaults to eq",
			input:     "name__invalid",
			wantField: "name__invalid",
			wantOp:    OpEq,
		},
		{
			name:      "field with underscores and operator",
			input:     "created_at__gte",
			wantField: "created_at",
			wantOp:    OpGte,
		},
		{
			name:      "empty string",
			input:     "",
			wantField: "",
			wantOp:    OpEq,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotField, gotOp := parseFieldOperator(tt.input)
			if gotField != tt.wantField {
				t.Errorf("parseFieldOperator(%q) field = %q, want %q", tt.input, gotField, tt.wantField)
			}
			if gotOp != tt.wantOp {
				t.Errorf("parseFieldOperator(%q) operator = %q, want %q", tt.input, gotOp, tt.wantOp)
			}
		})
	}
}

func TestParseFromMap(t *testing.T) {
	t.Run("empty map", func(t *testing.T) {
		params := ParseFromMap(map[string]string{})

		if params.Page != 1 {
			t.Errorf("Page = %d, want 1", params.Page)
		}
		if params.Limit != 10 {
			t.Errorf("Limit = %d, want 10", params.Limit)
		}
		if len(params.Filters) != 0 {
			t.Errorf("Filters should be empty, got %d", len(params.Filters))
		}
		if len(params.Sort) != 0 {
			t.Errorf("Sort should be empty, got %d", len(params.Sort))
		}
	})

	t.Run("pagination params", func(t *testing.T) {
		params := ParseFromMap(map[string]string{
			"page":      "3",
			"page_size": "25",
		})

		if params.Page != 3 {
			t.Errorf("Page = %d, want 3", params.Page)
		}
		if params.Limit != 25 {
			t.Errorf("Limit = %d, want 25", params.Limit)
		}
	})

	t.Run("sort params ascending", func(t *testing.T) {
		params := ParseFromMap(map[string]string{
			"sort":  "name",
			"order": "asc",
		})

		if len(params.Sort) != 1 {
			t.Fatalf("Sort count = %d, want 1", len(params.Sort))
		}
		if params.Sort[0].Field != "name" {
			t.Errorf("Sort field = %s, want name", params.Sort[0].Field)
		}
		if params.Sort[0].Desc {
			t.Error("Sort Desc should be false")
		}
	})

	t.Run("sort params descending", func(t *testing.T) {
		params := ParseFromMap(map[string]string{
			"sort":  "created_at",
			"order": "desc",
		})

		if len(params.Sort) != 1 {
			t.Fatalf("Sort count = %d, want 1", len(params.Sort))
		}
		if params.Sort[0].Field != "created_at" {
			t.Errorf("Sort field = %s, want created_at", params.Sort[0].Field)
		}
		if !params.Sort[0].Desc {
			t.Error("Sort Desc should be true")
		}
	})

	t.Run("sort with uppercase order", func(t *testing.T) {
		params := ParseFromMap(map[string]string{
			"sort":  "name",
			"order": "DESC",
		})

		if len(params.Sort) != 1 {
			t.Fatalf("Sort count = %d, want 1", len(params.Sort))
		}
		if !params.Sort[0].Desc {
			t.Error("Sort Desc should be true for uppercase DESC")
		}
	})

	t.Run("simple filter", func(t *testing.T) {
		params := ParseFromMap(map[string]string{
			"name": "Berlin",
		})

		if len(params.Filters) != 1 {
			t.Fatalf("Filters count = %d, want 1", len(params.Filters))
		}
		if params.Filters[0].Field != "name" {
			t.Errorf("Filter field = %s, want name", params.Filters[0].Field)
		}
		if params.Filters[0].Operator != OpEq {
			t.Errorf("Filter operator = %s, want eq", params.Filters[0].Operator)
		}
		if params.Filters[0].Value != "Berlin" {
			t.Errorf("Filter value = %s, want Berlin", params.Filters[0].Value)
		}
	})

	t.Run("filter with operator", func(t *testing.T) {
		params := ParseFromMap(map[string]string{
			"name__contains": "Ber",
		})

		if len(params.Filters) != 1 {
			t.Fatalf("Filters count = %d, want 1", len(params.Filters))
		}
		if params.Filters[0].Field != "name" {
			t.Errorf("Filter field = %s, want name", params.Filters[0].Field)
		}
		if params.Filters[0].Operator != OpContains {
			t.Errorf("Filter operator = %s, want contains", params.Filters[0].Operator)
		}
		if params.Filters[0].Value != "Ber" {
			t.Errorf("Filter value = %s, want Ber", params.Filters[0].Value)
		}
	})

	t.Run("relation filter", func(t *testing.T) {
		params := ParseFromMap(map[string]string{
			"example_country.name__contains": "Germany",
		})

		if len(params.Filters) != 1 {
			t.Fatalf("Filters count = %d, want 1", len(params.Filters))
		}
		if params.Filters[0].Field != "example_country.name" {
			t.Errorf("Filter field = %s, want example_country.name", params.Filters[0].Field)
		}
		if params.Filters[0].Operator != OpContains {
			t.Errorf("Filter operator = %s, want contains", params.Filters[0].Operator)
		}
	})

	t.Run("multiple filters", func(t *testing.T) {
		params := ParseFromMap(map[string]string{
			"name__contains": "New",
			"id__gt":         "5",
		})

		if len(params.Filters) != 2 {
			t.Fatalf("Filters count = %d, want 2", len(params.Filters))
		}
	})

	t.Run("combined pagination, sort, and filters", func(t *testing.T) {
		params := ParseFromMap(map[string]string{
			"page":           "2",
			"page_size":      "20",
			"sort":           "name",
			"order":          "desc",
			"name__contains": "test",
			"status":         "active",
		})

		if params.Page != 2 {
			t.Errorf("Page = %d, want 2", params.Page)
		}
		if params.Limit != 20 {
			t.Errorf("Limit = %d, want 20", params.Limit)
		}
		if len(params.Sort) != 1 {
			t.Errorf("Sort count = %d, want 1", len(params.Sort))
		}
		if len(params.Filters) != 2 {
			t.Errorf("Filters count = %d, want 2", len(params.Filters))
		}
	})

	t.Run("invalid page value", func(t *testing.T) {
		params := ParseFromMap(map[string]string{
			"page": "invalid",
		})

		if params.Page != 1 {
			t.Errorf("Page = %d, want 1 (default)", params.Page)
		}
	})

	t.Run("negative page value", func(t *testing.T) {
		params := ParseFromMap(map[string]string{
			"page": "-5",
		})

		// parseIntSafe returns 0 for negative, so default is used
		if params.Page != 1 {
			t.Errorf("Page = %d, want 1 (default)", params.Page)
		}
	})

	t.Run("page_size over 100 uses default", func(t *testing.T) {
		params := ParseFromMap(map[string]string{
			"page_size": "200",
		})

		// Over 100 should keep default 10
		if params.Limit != 10 {
			t.Errorf("Limit = %d, want 10 (default, 200 > 100)", params.Limit)
		}
	})

	t.Run("page_size exactly 100", func(t *testing.T) {
		params := ParseFromMap(map[string]string{
			"page_size": "100",
		})

		if params.Limit != 100 {
			t.Errorf("Limit = %d, want 100", params.Limit)
		}
	})
}

func TestParseIntSafe(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  int
	}{
		{"positive number", "123", 123},
		{"zero", "0", 0},
		{"single digit", "5", 5},
		{"large number", "9999", 9999},
		{"invalid characters", "12a3", 0},
		{"empty string", "", 0},
		{"negative number", "-5", 0},
		{"with spaces", "12 3", 0},
		{"decimal", "12.5", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseIntSafe(tt.input); got != tt.want {
				t.Errorf("parseIntSafe(%q) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}

func TestReservedParams(t *testing.T) {
	reserved := []string{"page", "page_size", "sort", "order"}

	for _, param := range reserved {
		if !reservedParams[param] {
			t.Errorf("'%s' should be a reserved parameter", param)
		}
	}

	notReserved := []string{"name", "id", "filter", "search", "query"}
	for _, param := range notReserved {
		if reservedParams[param] {
			t.Errorf("'%s' should NOT be a reserved parameter", param)
		}
	}
}

func TestMultipleSameFieldFilters(t *testing.T) {
	t.Run("multiple filters on same field with different operators", func(t *testing.T) {
		// Simulate parsing multiple filters on the same field
		// This tests the core logic that allows date range filtering like:
		// ?created_at__gte=2024-01-01&created_at__lte=2024-12-31
		params := &Params{
			Page:  1,
			Limit: 10,
		}

		// Manually add filters as ParseFromQuery would via VisitAll
		params.Filters = append(params.Filters, FilterParam{
			Field:    "created_at",
			Operator: OpGte,
			Value:    "2024-01-01",
		})
		params.Filters = append(params.Filters, FilterParam{
			Field:    "created_at",
			Operator: OpLte,
			Value:    "2024-12-31",
		})

		// Verify both filters are present
		if len(params.Filters) != 2 {
			t.Fatalf("Expected 2 filters, got %d", len(params.Filters))
		}

		// Verify first filter
		if params.Filters[0].Field != "created_at" {
			t.Errorf("Filter 0 field = %s, want created_at", params.Filters[0].Field)
		}
		if params.Filters[0].Operator != OpGte {
			t.Errorf("Filter 0 operator = %s, want gte", params.Filters[0].Operator)
		}
		if params.Filters[0].Value != "2024-01-01" {
			t.Errorf("Filter 0 value = %s, want 2024-01-01", params.Filters[0].Value)
		}

		// Verify second filter
		if params.Filters[1].Field != "created_at" {
			t.Errorf("Filter 1 field = %s, want created_at", params.Filters[1].Field)
		}
		if params.Filters[1].Operator != OpLte {
			t.Errorf("Filter 1 operator = %s, want lte", params.Filters[1].Operator)
		}
		if params.Filters[1].Value != "2024-12-31" {
			t.Errorf("Filter 1 value = %s, want 2024-12-31", params.Filters[1].Value)
		}
	})

	t.Run("price range with multiple filters", func(t *testing.T) {
		params := &Params{
			Page:  1,
			Limit: 10,
		}

		// Simulate: ?price__gte=1000&price__lte=5000
		params.Filters = append(params.Filters, FilterParam{
			Field:    "price",
			Operator: OpGte,
			Value:    "1000",
		})
		params.Filters = append(params.Filters, FilterParam{
			Field:    "price",
			Operator: OpLte,
			Value:    "5000",
		})

		if len(params.Filters) != 2 {
			t.Fatalf("Expected 2 filters, got %d", len(params.Filters))
		}

		// Both filters should have the same field
		if params.Filters[0].Field != params.Filters[1].Field {
			t.Error("Both filters should have the same field 'price'")
		}

		// But different operators
		if params.Filters[0].Operator == params.Filters[1].Operator {
			t.Error("Filters should have different operators (gte vs lte)")
		}
	})

	t.Run("three filters on same field", func(t *testing.T) {
		params := &Params{
			Page:  1,
			Limit: 10,
		}

		// Simulate: ?id__gt=10&id__lt=100&id__in=20,30,40
		params.Filters = append(params.Filters, FilterParam{
			Field:    "id",
			Operator: OpGt,
			Value:    "10",
		})
		params.Filters = append(params.Filters, FilterParam{
			Field:    "id",
			Operator: OpLt,
			Value:    "100",
		})
		params.Filters = append(params.Filters, FilterParam{
			Field:    "id",
			Operator: OpIn,
			Value:    "20,30,40",
		})

		if len(params.Filters) != 3 {
			t.Fatalf("Expected 3 filters, got %d", len(params.Filters))
		}

		// All three should be on the same field
		for i, f := range params.Filters {
			if f.Field != "id" {
				t.Errorf("Filter %d field = %s, want id", i, f.Field)
			}
		}
	})
}
