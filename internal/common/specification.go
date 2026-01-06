package common

import (
	"fmt"

	"gorm.io/gorm"
)

// AsGormDB safely asserts query to *gorm.DB with a descriptive panic message.
// Use this in Apply() methods instead of raw type assertion.
func AsGormDB(query any) *gorm.DB {
	db, ok := query.(*gorm.DB)
	if !ok {
		panic(fmt.Sprintf("specification: expected *gorm.DB, got %T", query))
	}
	return db
}

// Specification defines a query specification interface
type Specification interface {
	// Apply applies the specification to a query
	Apply(query any) any
}

// GormSpecification is a specification that works with GORM
type GormSpecification interface {
	Specification
	// ApplyGorm applies the specification to a GORM query
	ApplyGorm(db *gorm.DB) *gorm.DB
}

// ================================
// Basic Specifications
// ================================

// IDSpec finds by ID
type IDSpec struct {
	ID uint
}

// ByID creates an ID specification
func ByID(id uint) IDSpec {
	return IDSpec{ID: id}
}

func (s IDSpec) Apply(query any) any {
	return s.ApplyGorm(AsGormDB(query))
}

func (s IDSpec) ApplyGorm(db *gorm.DB) *gorm.DB {
	return db.Where("id = ?", s.ID)
}

// FieldSpec finds by field value
type FieldSpec struct {
	Field string
	Value any
}

// ByField creates a field specification
func ByField(field string, value any) FieldSpec {
	return FieldSpec{Field: field, Value: value}
}

func (s FieldSpec) Apply(query any) any {
	return s.ApplyGorm(AsGormDB(query))
}

func (s FieldSpec) ApplyGorm(db *gorm.DB) *gorm.DB {
	return db.Where(s.Field+" = ?", s.Value)
}

// FieldContainsSpec finds by field containing value
type FieldContainsSpec struct {
	Field string
	Value string
}

// ByFieldContains creates a contains specification
func ByFieldContains(field, value string) FieldContainsSpec {
	return FieldContainsSpec{Field: field, Value: value}
}

func (s FieldContainsSpec) Apply(query any) any {
	return s.ApplyGorm(AsGormDB(query))
}

func (s FieldContainsSpec) ApplyGorm(db *gorm.DB) *gorm.DB {
	return db.Where(s.Field+" LIKE ?", "%"+s.Value+"%")
}

// FieldInSpec finds by field in list
type FieldInSpec struct {
	Field  string
	Values []any
}

// ByFieldIn creates an in specification
func ByFieldIn(field string, values ...any) FieldInSpec {
	return FieldInSpec{Field: field, Values: values}
}

func (s FieldInSpec) Apply(query any) any {
	return s.ApplyGorm(AsGormDB(query))
}

func (s FieldInSpec) ApplyGorm(db *gorm.DB) *gorm.DB {
	return db.Where(s.Field+" IN ?", s.Values)
}

// ================================
// Composite Specifications
// ================================

// AndSpec combines specifications with AND
type AndSpec struct {
	Specs []Specification
}

// And creates a composite AND specification
func And(specs ...Specification) AndSpec {
	return AndSpec{Specs: specs}
}

func (s AndSpec) Apply(query any) any {
	return s.ApplyGorm(AsGormDB(query))
}

func (s AndSpec) ApplyGorm(db *gorm.DB) *gorm.DB {
	for _, spec := range s.Specs {
		if gs, ok := spec.(GormSpecification); ok {
			db = gs.ApplyGorm(db)
		}
	}
	return db
}

// OrSpec combines specifications with OR
type OrSpec struct {
	Specs []Specification
}

// Or creates a composite OR specification
func Or(specs ...Specification) OrSpec {
	return OrSpec{Specs: specs}
}

func (s OrSpec) Apply(query any) any {
	return s.ApplyGorm(AsGormDB(query))
}

func (s OrSpec) ApplyGorm(db *gorm.DB) *gorm.DB {
	if len(s.Specs) == 0 {
		return db
	}

	// Build OR clause
	combined := db.Session(&gorm.Session{NewDB: true})
	for i, spec := range s.Specs {
		if gs, ok := spec.(GormSpecification); ok {
			if i == 0 {
				combined = gs.ApplyGorm(combined)
			} else {
				combined = combined.Or(gs.ApplyGorm(db.Session(&gorm.Session{NewDB: true})))
			}
		}
	}
	return db.Where(combined)
}

// NotSpec negates a specification
type NotSpec struct {
	Spec Specification
}

// Not creates a negation specification
func Not(spec Specification) NotSpec {
	return NotSpec{Spec: spec}
}

func (s NotSpec) Apply(query any) any {
	return s.ApplyGorm(AsGormDB(query))
}

func (s NotSpec) ApplyGorm(db *gorm.DB) *gorm.DB {
	if gs, ok := s.Spec.(GormSpecification); ok {
		sub := gs.ApplyGorm(db.Session(&gorm.Session{NewDB: true}))
		return db.Not(sub)
	}
	return db
}

// ================================
// Range Specifications
// ================================

// FieldGTSpec finds by field greater than value
type FieldGTSpec struct {
	Field string
	Value any
}

// ByFieldGT creates a greater than specification
func ByFieldGT(field string, value any) FieldGTSpec {
	return FieldGTSpec{Field: field, Value: value}
}

func (s FieldGTSpec) Apply(query any) any {
	return s.ApplyGorm(AsGormDB(query))
}

func (s FieldGTSpec) ApplyGorm(db *gorm.DB) *gorm.DB {
	return db.Where(s.Field+" > ?", s.Value)
}

// FieldGTESpec finds by field greater than or equal
type FieldGTESpec struct {
	Field string
	Value any
}

// ByFieldGTE creates a greater than or equal specification
func ByFieldGTE(field string, value any) FieldGTESpec {
	return FieldGTESpec{Field: field, Value: value}
}

func (s FieldGTESpec) Apply(query any) any {
	return s.ApplyGorm(AsGormDB(query))
}

func (s FieldGTESpec) ApplyGorm(db *gorm.DB) *gorm.DB {
	return db.Where(s.Field+" >= ?", s.Value)
}

// FieldLTSpec finds by field less than value
type FieldLTSpec struct {
	Field string
	Value any
}

// ByFieldLT creates a less than specification
func ByFieldLT(field string, value any) FieldLTSpec {
	return FieldLTSpec{Field: field, Value: value}
}

func (s FieldLTSpec) Apply(query any) any {
	return s.ApplyGorm(AsGormDB(query))
}

func (s FieldLTSpec) ApplyGorm(db *gorm.DB) *gorm.DB {
	return db.Where(s.Field+" < ?", s.Value)
}

// FieldLTESpec finds by field less than or equal
type FieldLTESpec struct {
	Field string
	Value any
}

// ByFieldLTE creates a less than or equal specification
func ByFieldLTE(field string, value any) FieldLTESpec {
	return FieldLTESpec{Field: field, Value: value}
}

func (s FieldLTESpec) Apply(query any) any {
	return s.ApplyGorm(AsGormDB(query))
}

func (s FieldLTESpec) ApplyGorm(db *gorm.DB) *gorm.DB {
	return db.Where(s.Field+" <= ?", s.Value)
}

// FieldBetweenSpec finds by field between two values
type FieldBetweenSpec struct {
	Field string
	Min   any
	Max   any
}

// ByFieldBetween creates a between specification
func ByFieldBetween(field string, min, max any) FieldBetweenSpec {
	return FieldBetweenSpec{Field: field, Min: min, Max: max}
}

func (s FieldBetweenSpec) Apply(query any) any {
	return s.ApplyGorm(AsGormDB(query))
}

func (s FieldBetweenSpec) ApplyGorm(db *gorm.DB) *gorm.DB {
	return db.Where(s.Field+" BETWEEN ? AND ?", s.Min, s.Max)
}

// ================================
// Null Specifications
// ================================

// FieldNullSpec finds by field being null
type FieldNullSpec struct {
	Field string
}

// ByFieldNull creates a null specification
func ByFieldNull(field string) FieldNullSpec {
	return FieldNullSpec{Field: field}
}

func (s FieldNullSpec) Apply(query any) any {
	return s.ApplyGorm(AsGormDB(query))
}

func (s FieldNullSpec) ApplyGorm(db *gorm.DB) *gorm.DB {
	return db.Where(s.Field + " IS NULL")
}

// FieldNotNullSpec finds by field not being null
type FieldNotNullSpec struct {
	Field string
}

// ByFieldNotNull creates a not null specification
func ByFieldNotNull(field string) FieldNotNullSpec {
	return FieldNotNullSpec{Field: field}
}

func (s FieldNotNullSpec) Apply(query any) any {
	return s.ApplyGorm(AsGormDB(query))
}

func (s FieldNotNullSpec) ApplyGorm(db *gorm.DB) *gorm.DB {
	return db.Where(s.Field + " IS NOT NULL")
}

// ================================
// Empty Specification (matches all)
// ================================

// TrueSpec always matches (no-op)
type TrueSpec struct{}

// True returns a specification that always matches
func True() TrueSpec {
	return TrueSpec{}
}

func (s TrueSpec) Apply(query any) any {
	return query
}

func (s TrueSpec) ApplyGorm(db *gorm.DB) *gorm.DB {
	return db
}

// ================================
// Helper to apply specification
// ================================

// ApplySpecification applies a specification to a GORM query
func ApplySpecification(db *gorm.DB, spec Specification) *gorm.DB {
	if spec == nil {
		return db
	}
	if gs, ok := spec.(GormSpecification); ok {
		return gs.ApplyGorm(db)
	}
	return db
}
