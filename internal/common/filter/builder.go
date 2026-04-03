package filter

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/jinzhu/inflection"
	"gorm.io/gorm"
)

// safeIdentifier validates that a SQL identifier (table or column name) contains
// only safe characters to prevent SQL injection, even from misconfigured developer configs.
var safeIdentifier = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)

// Apply applies filters, sorting, and pagination to a GORM query
func Apply(db *gorm.DB, config Config, params *Params) *gorm.DB {
	query := db

	// Track joined relations to avoid duplicate joins.
	// Shared between filters and sort so that sorting by a relation field
	// adds the join even if no filter references that relation.
	joinedRelations := make(map[string]bool)

	// Apply filters
	for _, f := range params.Filters {
		query = applyFilter(query, config, f, joinedRelations)
	}

	// Apply sorting
	for _, s := range params.Sort {
		query = applySort(query, config, s, joinedRelations)
	}

	// Apply pagination
	if params.Limit > 0 {
		query = query.Limit(params.Limit).Offset(params.Offset())
	}

	return query
}

// ApplyFiltersOnly applies only filters without pagination (useful for count queries)
func ApplyFiltersOnly(db *gorm.DB, config Config, params *Params) *gorm.DB {
	query := db
	joinedRelations := make(map[string]bool)

	for _, f := range params.Filters {
		query = applyFilter(query, config, f, joinedRelations)
	}

	return query
}

func applyFilter(db *gorm.DB, config Config, f FilterParam, joinedRelations map[string]bool) *gorm.DB {
	// Handle relation fields (e.g., "example_country.name" -> join Country, filter on name)
	if strings.Contains(f.Field, ".") {
		return applyRelationFilter(db, config, f, joinedRelations)
	}

	fieldConfig, ok := config.Fields[f.Field]
	if !ok {
		return db // Skip unknown fields
	}

	// Skip if it's a relation-only field (no DBColumn)
	if fieldConfig.DBColumn == "" && fieldConfig.Relation != "" {
		return db
	}

	if !isOperatorAllowed(fieldConfig, f.Operator) {
		return db // Skip disallowed operators
	}

	column := fmt.Sprintf("%s.%s", config.TableName, fieldConfig.DBColumn)
	return applyOperator(db, column, f.Operator, f.Value)
}

func applyRelationFilter(db *gorm.DB, config Config, f FilterParam, joinedRelations map[string]bool) *gorm.DB {
	// Parse "example_country.name" -> relation="country", field="name"
	parts := strings.SplitN(f.Field, ".", 2)
	if len(parts) != 2 {
		return db
	}

	relationName, fieldName := parts[0], parts[1]
	relationConfig, ok := config.Fields[relationName]
	if !ok || relationConfig.Relation == "" {
		return db // Not a valid relation
	}

	// Validate that the field is allowed for this relation (SQL injection protection)
	if !config.IsRelationFieldAllowed(relationName, fieldName) {
		return db // Skip disallowed relation fields
	}

	// Build join table name (pluralize relation name)
	joinTable := pluralize(strings.ToLower(relationConfig.Relation))

	// Defense-in-depth: validate identifiers contain only safe characters
	if !safeIdentifier.MatchString(joinTable) || !safeIdentifier.MatchString(fieldName) {
		return db
	}

	// Only join if not already joined
	if !joinedRelations[joinTable] {
		db = db.Joins(fmt.Sprintf("LEFT JOIN %s ON %s.%s = %s.id",
			joinTable, config.TableName, relationConfig.RelationFK, joinTable))
		joinedRelations[joinTable] = true
	}

	// Apply the filter on the joined table's column
	column := fmt.Sprintf("%s.%s", joinTable, fieldName)
	return applyOperator(db, column, f.Operator, f.Value)
}

func applyOperator(db *gorm.DB, column string, op Operator, value string) *gorm.DB {
	switch op {
	case OpEq:
		return db.Where(column+" = ?", value)
	case OpGt:
		return db.Where(column+" > ?", value)
	case OpLt:
		return db.Where(column+" < ?", value)
	case OpGte:
		return db.Where(column+" >= ?", value)
	case OpLte:
		return db.Where(column+" <= ?", value)
	case OpContains:
		return db.Where(column+" ILIKE ? ESCAPE '\\'", "%"+escapeLikeWildcards(value)+"%")
	case OpStartsWith:
		return db.Where(column+" ILIKE ? ESCAPE '\\'", escapeLikeWildcards(value)+"%")
	case OpEndsWith:
		return db.Where(column+" ILIKE ? ESCAPE '\\'", "%"+escapeLikeWildcards(value))
	case OpIn:
		values := strings.Split(value, ",")
		// Trim whitespace from each value
		for i := range values {
			values[i] = strings.TrimSpace(values[i])
		}
		return db.Where(column+" IN ?", values)
	case OpIsNull:
		return db.Where(column + " IS NULL")
	case OpIsNotNull:
		return db.Where(column + " IS NOT NULL")
	}
	return db
}

func applySort(db *gorm.DB, config Config, s SortParam, joinedRelations map[string]bool) *gorm.DB {
	// Handle relation sorting (e.g., "example_country.name")
	if strings.Contains(s.Field, ".") {
		parts := strings.SplitN(s.Field, ".", 2)
		if len(parts) != 2 {
			return db
		}
		relationName, fieldName := parts[0], parts[1]
		relationConfig, ok := config.Fields[relationName]
		if !ok || relationConfig.Relation == "" {
			return db
		}

		// Validate that the field is allowed for this relation (SQL injection protection)
		if !config.IsRelationFieldAllowed(relationName, fieldName) {
			return db // Skip disallowed relation fields
		}

		joinTable := pluralize(strings.ToLower(relationConfig.Relation))

		// Defense-in-depth: validate identifiers contain only safe characters
		if !safeIdentifier.MatchString(joinTable) || !safeIdentifier.MatchString(fieldName) {
			return db
		}

		// Add the join if not already joined by a filter
		if !joinedRelations[joinTable] {
			db = db.Joins(fmt.Sprintf("LEFT JOIN %s ON %s.%s = %s.id",
				joinTable, config.TableName, relationConfig.RelationFK, joinTable))
			joinedRelations[joinTable] = true
		}

		column := fmt.Sprintf("%s.%s", joinTable, fieldName)
		order := "ASC"
		if s.Desc {
			order = "DESC"
		}
		return db.Order(fmt.Sprintf("%s %s", column, order))
	}

	fieldConfig, ok := config.Fields[s.Field]
	if !ok || !fieldConfig.Sortable {
		return db // Skip invalid or non-sortable fields
	}

	order := "ASC"
	if s.Desc {
		order = "DESC"
	}
	return db.Order(fmt.Sprintf("%s.%s %s", config.TableName, fieldConfig.DBColumn, order))
}

func isOperatorAllowed(config FieldConfig, op Operator) bool {
	for _, allowed := range config.Operators {
		if allowed == op {
			return true
		}
	}
	return false
}

// escapeLikeWildcards escapes SQL LIKE wildcards (% and _) in user input
// so they are treated as literal characters rather than pattern wildcards.
func escapeLikeWildcards(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "%", "\\%")
	s = strings.ReplaceAll(s, "_", "\\_")
	return s
}

// pluralize converts a singular word to its plural form.
// Uses jinzhu/inflection for proper English inflection rules
// including irregular plurals (person->people, child->children).
func pluralize(word string) string {
	if word == "" {
		return word
	}
	return inflection.Plural(word)
}
