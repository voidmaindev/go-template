package filter

import (
	"testing"
)

func TestDefaultParams(t *testing.T) {
	params := DefaultParams()

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
}

func TestParams_HasFilters(t *testing.T) {
	tests := []struct {
		name    string
		params  *Params
		want    bool
	}{
		{
			name:    "no filters",
			params:  &Params{},
			want:    false,
		},
		{
			name:    "empty filters slice",
			params:  &Params{Filters: []FilterParam{}},
			want:    false,
		},
		{
			name: "has one filter",
			params: &Params{
				Filters: []FilterParam{
					{Field: "name", Operator: OpEq, Value: "test"},
				},
			},
			want: true,
		},
		{
			name: "has multiple filters",
			params: &Params{
				Filters: []FilterParam{
					{Field: "name", Operator: OpContains, Value: "test"},
					{Field: "id", Operator: OpGt, Value: "5"},
				},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.params.HasFilters(); got != tt.want {
				t.Errorf("HasFilters() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParams_HasSort(t *testing.T) {
	tests := []struct {
		name   string
		params *Params
		want   bool
	}{
		{
			name:   "no sort",
			params: &Params{},
			want:   false,
		},
		{
			name:   "empty sort slice",
			params: &Params{Sort: []SortParam{}},
			want:   false,
		},
		{
			name: "has one sort",
			params: &Params{
				Sort: []SortParam{
					{Field: "name", Desc: false},
				},
			},
			want: true,
		},
		{
			name: "has multiple sorts",
			params: &Params{
				Sort: []SortParam{
					{Field: "name", Desc: true},
					{Field: "created_at", Desc: false},
				},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.params.HasSort(); got != tt.want {
				t.Errorf("HasSort() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParams_Offset(t *testing.T) {
	tests := []struct {
		name   string
		params *Params
		want   int
	}{
		{
			name:   "page 1, limit 10",
			params: &Params{Page: 1, Limit: 10},
			want:   0,
		},
		{
			name:   "page 2, limit 10",
			params: &Params{Page: 2, Limit: 10},
			want:   10,
		},
		{
			name:   "page 3, limit 10",
			params: &Params{Page: 3, Limit: 10},
			want:   20,
		},
		{
			name:   "page 1, limit 25",
			params: &Params{Page: 1, Limit: 25},
			want:   0,
		},
		{
			name:   "page 5, limit 25",
			params: &Params{Page: 5, Limit: 25},
			want:   100,
		},
		{
			name:   "page 0 (invalid), should return 0",
			params: &Params{Page: 0, Limit: 10},
			want:   0,
		},
		{
			name:   "negative page, should return 0",
			params: &Params{Page: -1, Limit: 10},
			want:   0,
		},
		{
			name:   "page 10, limit 50",
			params: &Params{Page: 10, Limit: 50},
			want:   450,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.params.Offset(); got != tt.want {
				t.Errorf("Offset() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestFilterParam_Structure(t *testing.T) {
	param := FilterParam{
		Field:    "country.name",
		Operator: OpContains,
		Value:    "Germany",
	}

	if param.Field != "country.name" {
		t.Errorf("Field = %s, want country.name", param.Field)
	}
	if param.Operator != OpContains {
		t.Errorf("Operator = %s, want contains", param.Operator)
	}
	if param.Value != "Germany" {
		t.Errorf("Value = %s, want Germany", param.Value)
	}
}

func TestSortParam_Structure(t *testing.T) {
	t.Run("ascending", func(t *testing.T) {
		param := SortParam{Field: "name", Desc: false}
		if param.Field != "name" {
			t.Errorf("Field = %s, want name", param.Field)
		}
		if param.Desc {
			t.Error("Desc should be false")
		}
	})

	t.Run("descending", func(t *testing.T) {
		param := SortParam{Field: "created_at", Desc: true}
		if param.Field != "created_at" {
			t.Errorf("Field = %s, want created_at", param.Field)
		}
		if !param.Desc {
			t.Error("Desc should be true")
		}
	})
}
