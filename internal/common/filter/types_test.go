package filter

import (
	"testing"
)

func TestOperator_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		operator Operator
		want     bool
	}{
		{"eq is valid", OpEq, true},
		{"gt is valid", OpGt, true},
		{"lt is valid", OpLt, true},
		{"gte is valid", OpGte, true},
		{"lte is valid", OpLte, true},
		{"contains is valid", OpContains, true},
		{"starts_with is valid", OpStartsWith, true},
		{"ends_with is valid", OpEndsWith, true},
		{"in is valid", OpIn, true},
		{"is_null is valid", OpIsNull, true},
		{"is_not_null is valid", OpIsNotNull, true},
		{"invalid operator", Operator("invalid"), false},
		{"empty operator", Operator(""), false},
		{"like is invalid", Operator("like"), false},
		{"ne is invalid", Operator("ne"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.operator.IsValid(); got != tt.want {
				t.Errorf("Operator.IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPredefinedOperatorSets(t *testing.T) {
	t.Run("StringOps contains expected operators", func(t *testing.T) {
		expected := map[Operator]bool{
			OpEq:         true,
			OpContains:   true,
			OpStartsWith: true,
			OpEndsWith:   true,
			OpIn:         true,
			OpIsNull:     true,
			OpIsNotNull:  true,
		}

		if len(StringOps) != len(expected) {
			t.Errorf("StringOps has %d operators, want %d", len(StringOps), len(expected))
		}

		for _, op := range StringOps {
			if !expected[op] {
				t.Errorf("StringOps contains unexpected operator: %s", op)
			}
		}
	})

	t.Run("NumberOps contains expected operators", func(t *testing.T) {
		expected := map[Operator]bool{
			OpEq:        true,
			OpGt:        true,
			OpLt:        true,
			OpGte:       true,
			OpLte:       true,
			OpIn:        true,
			OpIsNull:    true,
			OpIsNotNull: true,
		}

		if len(NumberOps) != len(expected) {
			t.Errorf("NumberOps has %d operators, want %d", len(NumberOps), len(expected))
		}

		for _, op := range NumberOps {
			if !expected[op] {
				t.Errorf("NumberOps contains unexpected operator: %s", op)
			}
		}
	})

	t.Run("DateOps contains expected operators", func(t *testing.T) {
		expected := map[Operator]bool{
			OpEq:        true,
			OpGt:        true,
			OpLt:        true,
			OpGte:       true,
			OpLte:       true,
			OpIsNull:    true,
			OpIsNotNull: true,
		}

		if len(DateOps) != len(expected) {
			t.Errorf("DateOps has %d operators, want %d", len(DateOps), len(expected))
		}

		for _, op := range DateOps {
			if !expected[op] {
				t.Errorf("DateOps contains unexpected operator: %s", op)
			}
		}
	})

	t.Run("BoolOps contains expected operators", func(t *testing.T) {
		expected := map[Operator]bool{
			OpEq:        true,
			OpIsNull:    true,
			OpIsNotNull: true,
		}

		if len(BoolOps) != len(expected) {
			t.Errorf("BoolOps has %d operators, want %d", len(BoolOps), len(expected))
		}

		for _, op := range BoolOps {
			if !expected[op] {
				t.Errorf("BoolOps contains unexpected operator: %s", op)
			}
		}
	})
}

func TestFieldType_Values(t *testing.T) {
	// Ensure field types have expected values
	if TypeString != 0 {
		t.Errorf("TypeString = %d, want 0", TypeString)
	}
	if TypeNumber != 1 {
		t.Errorf("TypeNumber = %d, want 1", TypeNumber)
	}
	if TypeDate != 2 {
		t.Errorf("TypeDate = %d, want 2", TypeDate)
	}
	if TypeBool != 3 {
		t.Errorf("TypeBool = %d, want 3", TypeBool)
	}
}

func TestFieldConfig_Structure(t *testing.T) {
	config := FieldConfig{
		DBColumn:   "name",
		Type:       TypeString,
		Operators:  StringOps,
		Sortable:   true,
		Relation:   "Country",
		RelationFK: "country_id",
	}

	if config.DBColumn != "name" {
		t.Errorf("DBColumn = %s, want name", config.DBColumn)
	}
	if config.Type != TypeString {
		t.Errorf("Type = %d, want TypeString", config.Type)
	}
	if !config.Sortable {
		t.Error("Sortable should be true")
	}
	if config.Relation != "Country" {
		t.Errorf("Relation = %s, want Country", config.Relation)
	}
	if config.RelationFK != "country_id" {
		t.Errorf("RelationFK = %s, want country_id", config.RelationFK)
	}
}

func TestConfig_Structure(t *testing.T) {
	config := Config{
		TableName: "cities",
		Fields: map[string]FieldConfig{
			"id":   {DBColumn: "id", Type: TypeNumber, Operators: NumberOps, Sortable: true},
			"name": {DBColumn: "name", Type: TypeString, Operators: StringOps, Sortable: true},
		},
	}

	if config.TableName != "cities" {
		t.Errorf("TableName = %s, want cities", config.TableName)
	}
	if len(config.Fields) != 2 {
		t.Errorf("Fields count = %d, want 2", len(config.Fields))
	}
	if _, ok := config.Fields["id"]; !ok {
		t.Error("Fields should contain 'id'")
	}
	if _, ok := config.Fields["name"]; !ok {
		t.Error("Fields should contain 'name'")
	}
}
