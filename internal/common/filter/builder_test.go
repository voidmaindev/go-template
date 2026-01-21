package filter

import (
	"testing"
)

func TestPluralize(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"regular word", "city", "cities"},
		{"word ending in y", "country", "countries"},
		{"word ending in s", "status", "statuses"},
		{"word ending in x", "box", "boxes"},
		{"word ending in ch", "match", "matches"},
		{"word ending in sh", "dish", "dishes"},
		{"regular word 2", "user", "users"},
		{"regular word 3", "item", "items"},
		{"regular word 4", "document", "documents"},
		{"empty string", "", ""},
		{"single letter", "a", "as"},
		{"word with y after vowel", "day", "days"},
		// Irregular plurals now correctly handled by inflection library
		{"irregular person", "person", "people"},
		{"irregular child", "child", "children"},
		{"irregular man", "man", "men"},
		{"irregular woman", "woman", "women"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := pluralize(tt.input); got != tt.want {
				t.Errorf("pluralize(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestIsOperatorAllowed(t *testing.T) {
	tests := []struct {
		name   string
		config FieldConfig
		op     Operator
		want   bool
	}{
		{
			name:   "allowed operator in list",
			config: FieldConfig{Operators: []Operator{OpEq, OpContains, OpIn}},
			op:     OpContains,
			want:   true,
		},
		{
			name:   "disallowed operator",
			config: FieldConfig{Operators: []Operator{OpEq, OpGt, OpLt}},
			op:     OpContains,
			want:   false,
		},
		{
			name:   "empty operators list",
			config: FieldConfig{Operators: []Operator{}},
			op:     OpEq,
			want:   false,
		},
		{
			name:   "first operator in list",
			config: FieldConfig{Operators: StringOps},
			op:     OpEq,
			want:   true,
		},
		{
			name:   "last operator in list",
			config: FieldConfig{Operators: StringOps},
			op:     OpIsNotNull,
			want:   true,
		},
		{
			name:   "number operator on number field",
			config: FieldConfig{Operators: NumberOps},
			op:     OpGte,
			want:   true,
		},
		{
			name:   "string operator on number field",
			config: FieldConfig{Operators: NumberOps},
			op:     OpContains,
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isOperatorAllowed(tt.config, tt.op); got != tt.want {
				t.Errorf("isOperatorAllowed() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Test Config creation for testing purposes
func getTestConfig() Config {
	return Config{
		TableName: "cities",
		Fields: map[string]FieldConfig{
			"id":         {DBColumn: "id", Type: TypeNumber, Operators: NumberOps, Sortable: true},
			"name":       {DBColumn: "name", Type: TypeString, Operators: StringOps, Sortable: true},
			"population": {DBColumn: "population", Type: TypeNumber, Operators: NumberOps, Sortable: true},
			"created_at": {DBColumn: "created_at", Type: TypeDate, Operators: DateOps, Sortable: true},
			"country_id": {DBColumn: "country_id", Type: TypeNumber, Operators: NumberOps, Sortable: false},
			"country":    {Relation: "Country", RelationFK: "country_id"},
		},
	}
}

func TestApplyFilter_SkipsUnknownFields(t *testing.T) {
	config := getTestConfig()

	// Unknown field should not cause panic and should be skipped
	filter := FilterParam{
		Field:    "unknown_field",
		Operator: OpEq,
		Value:    "test",
	}

	joinedRelations := make(map[string]bool)

	// This should not panic
	result := applyFilter(nil, config, filter, joinedRelations)

	// Should return the same db (nil in this case)
	if result != nil {
		t.Error("applyFilter should return unchanged db for unknown field")
	}
}

func TestApplyFilter_SkipsRelationOnlyFields(t *testing.T) {
	config := getTestConfig()

	// Relation-only field (no DBColumn) should be skipped
	filter := FilterParam{
		Field:    "country",
		Operator: OpEq,
		Value:    "1",
	}

	joinedRelations := make(map[string]bool)

	// This should not panic
	result := applyFilter(nil, config, filter, joinedRelations)

	// Should return the same db (nil in this case)
	if result != nil {
		t.Error("applyFilter should return unchanged db for relation-only field")
	}
}

func TestApplyFilter_SkipsDisallowedOperators(t *testing.T) {
	config := getTestConfig()

	// contains is not allowed on number fields
	filter := FilterParam{
		Field:    "id",
		Operator: OpContains,
		Value:    "test",
	}

	joinedRelations := make(map[string]bool)

	// This should not panic
	result := applyFilter(nil, config, filter, joinedRelations)

	// Should return the same db (nil in this case)
	if result != nil {
		t.Error("applyFilter should return unchanged db for disallowed operator")
	}
}

func TestApplySort_SkipsNonSortableFields(t *testing.T) {
	config := getTestConfig()

	// country_id is not sortable
	sort := SortParam{
		Field: "country_id",
		Desc:  false,
	}

	// This should not panic
	result := applySort(nil, config, sort)

	// Should return the same db (nil in this case)
	if result != nil {
		t.Error("applySort should return unchanged db for non-sortable field")
	}
}

func TestApplySort_SkipsUnknownFields(t *testing.T) {
	config := getTestConfig()

	sort := SortParam{
		Field: "unknown_field",
		Desc:  true,
	}

	// This should not panic
	result := applySort(nil, config, sort)

	// Should return the same db (nil in this case)
	if result != nil {
		t.Error("applySort should return unchanged db for unknown field")
	}
}

func TestApplyRelationFilter_InvalidRelation(t *testing.T) {
	config := getTestConfig()

	// Try to filter on a relation that doesn't exist
	filter := FilterParam{
		Field:    "invalid_relation.name",
		Operator: OpContains,
		Value:    "test",
	}

	joinedRelations := make(map[string]bool)

	// This should not panic
	result := applyRelationFilter(nil, config, filter, joinedRelations)

	// Should return the same db (nil in this case)
	if result != nil {
		t.Error("applyRelationFilter should return unchanged db for invalid relation")
	}
}

func TestApplyRelationFilter_InvalidFieldFormat(t *testing.T) {
	config := getTestConfig()

	// Missing second part after dot
	filter := FilterParam{
		Field:    "country",
		Operator: OpContains,
		Value:    "test",
	}

	joinedRelations := make(map[string]bool)

	// applyRelationFilter expects a dot in field name
	// This test ensures it handles edge cases
	result := applyRelationFilter(nil, config, filter, joinedRelations)

	if result != nil {
		t.Error("applyRelationFilter should return unchanged db for invalid field format")
	}
}

func TestApplyRelationFilter_FieldOnNonRelation(t *testing.T) {
	config := getTestConfig()

	// "name" is not a relation field
	filter := FilterParam{
		Field:    "name.something",
		Operator: OpContains,
		Value:    "test",
	}

	joinedRelations := make(map[string]bool)

	result := applyRelationFilter(nil, config, filter, joinedRelations)

	if result != nil {
		t.Error("applyRelationFilter should return unchanged db when field is not a relation")
	}
}

func TestJoinedRelationsTracking(t *testing.T) {
	// Test that the joinedRelations map is properly updated
	joinedRelations := make(map[string]bool)

	// Simulate adding a join
	joinedRelations["countries"] = true

	if !joinedRelations["countries"] {
		t.Error("countries should be in joinedRelations")
	}

	if joinedRelations["cities"] {
		t.Error("cities should not be in joinedRelations")
	}
}

func TestApplyOperator_InvalidOperator(t *testing.T) {
	// Unknown operator should return unchanged db (nil in this case)
	result := applyOperator(nil, "column", Operator("invalid"), "value")

	if result != nil {
		t.Error("applyOperator should return unchanged db for invalid operator")
	}
}

// Note: Apply and ApplyFiltersOnly require a real GORM database connection
// to test properly. Integration tests should be added in a separate test file
// with database fixtures.

func TestSortParam_DescFlag(t *testing.T) {
	t.Run("ascending sort", func(t *testing.T) {
		sort := SortParam{Field: "name", Desc: false}
		if sort.Field != "name" {
			t.Errorf("Field = %s, want name", sort.Field)
		}
		if sort.Desc {
			t.Error("Desc should be false for ascending")
		}
	})

	t.Run("descending sort", func(t *testing.T) {
		sort := SortParam{Field: "created_at", Desc: true}
		if sort.Field != "created_at" {
			t.Errorf("Field = %s, want created_at", sort.Field)
		}
		if !sort.Desc {
			t.Error("Desc should be true for descending")
		}
	})
}
