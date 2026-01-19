package filter

// Operator represents a filter operation
type Operator string

const (
	OpEq         Operator = "eq"
	OpGt         Operator = "gt"
	OpLt         Operator = "lt"
	OpGte        Operator = "gte"
	OpLte        Operator = "lte"
	OpContains   Operator = "contains"
	OpStartsWith Operator = "starts_with"
	OpEndsWith   Operator = "ends_with"
	OpIn         Operator = "in"
	OpIsNull     Operator = "is_null"
	OpIsNotNull  Operator = "is_not_null"
)

// FieldType represents the data type of a field
type FieldType int

const (
	TypeString FieldType = iota
	TypeNumber
	TypeDate
	TypeBool
)

// FieldConfig defines how a field can be filtered/sorted
type FieldConfig struct {
	DBColumn   string     // Database column name
	Type       FieldType  // Field data type
	Operators  []Operator // Allowed filter operators
	Sortable   bool       // Can be used for sorting
	Relation   string     // If set, this is a relation field (e.g., "Country")
	RelationFK string     // Foreign key for joins (e.g., "country_id")
}

// Config holds all filterable/sortable fields for an entity
type Config struct {
	TableName             string
	Fields                map[string]FieldConfig
	AllowedRelationFields map[string][]string // Maps relation name to allowed field names
}

// Filterable interface - implement on models to enable filtering
type Filterable interface {
	FilterConfig() Config
}

// Predefined operator sets for common field types
var (
	StringOps = []Operator{OpEq, OpContains, OpStartsWith, OpEndsWith, OpIn, OpIsNull, OpIsNotNull}
	NumberOps = []Operator{OpEq, OpGt, OpLt, OpGte, OpLte, OpIn, OpIsNull, OpIsNotNull}
	DateOps   = []Operator{OpEq, OpGt, OpLt, OpGte, OpLte, OpIsNull, OpIsNotNull}
	BoolOps   = []Operator{OpEq, OpIsNull, OpIsNotNull}
)

// IsValid checks if the operator is a known valid operator
func (o Operator) IsValid() bool {
	switch o {
	case OpEq, OpGt, OpLt, OpGte, OpLte, OpContains, OpStartsWith, OpEndsWith, OpIn, OpIsNull, OpIsNotNull:
		return true
	}
	return false
}

// IsRelationFieldAllowed checks if a field is allowed for a given relation.
// Returns true if AllowedRelationFields is not configured (backward compatibility)
// or if the field is in the allowed list.
func (c Config) IsRelationFieldAllowed(relationName, fieldName string) bool {
	if c.AllowedRelationFields == nil {
		return false // Strict by default - must explicitly allow relation fields
	}
	allowedFields, ok := c.AllowedRelationFields[relationName]
	if !ok {
		return false // Relation not configured
	}
	for _, allowed := range allowedFields {
		if allowed == fieldName {
			return true
		}
	}
	return false
}
