package example_country

import (
	"github.com/voidmaindev/go-template/internal/common"
	"gorm.io/gorm"
)

// ================================
// Country Domain Specifications
// ================================

// CodeSpec finds country by code
type CodeSpec struct {
	Code string
}

// ByCode creates a code specification
func ByCode(code string) CodeSpec {
	return CodeSpec{Code: code}
}

func (s CodeSpec) Apply(query any) any {
	return s.ApplyGorm(common.AsGormDB(query))
}

func (s CodeSpec) ApplyGorm(db *gorm.DB) *gorm.DB {
	return db.Where("code = ?", s.Code)
}

// NameSpec finds country by exact name
type NameSpec struct {
	Name string
}

// ByName creates a name specification
func ByName(name string) NameSpec {
	return NameSpec{Name: name}
}

func (s NameSpec) Apply(query any) any {
	return s.ApplyGorm(common.AsGormDB(query))
}

func (s NameSpec) ApplyGorm(db *gorm.DB) *gorm.DB {
	return db.Where("name = ?", s.Name)
}

// NameContainsSpec finds countries whose name contains a string
type NameContainsSpec struct {
	Name string
}

// ByNameContains creates a name contains specification
func ByNameContains(name string) NameContainsSpec {
	return NameContainsSpec{Name: name}
}

func (s NameContainsSpec) Apply(query any) any {
	return s.ApplyGorm(common.AsGormDB(query))
}

func (s NameContainsSpec) ApplyGorm(db *gorm.DB) *gorm.DB {
	return db.Where("name LIKE ?", "%"+s.Name+"%")
}

// CodeContainsSpec finds countries whose code contains a string
type CodeContainsSpec struct {
	Code string
}

// ByCodeContains creates a code contains specification
func ByCodeContains(code string) CodeContainsSpec {
	return CodeContainsSpec{Code: code}
}

func (s CodeContainsSpec) Apply(query any) any {
	return s.ApplyGorm(common.AsGormDB(query))
}

func (s CodeContainsSpec) ApplyGorm(db *gorm.DB) *gorm.DB {
	return db.Where("code LIKE ?", "%"+s.Code+"%")
}

// Ensure all specs implement GormSpecification
var (
	_ common.GormSpecification = CodeSpec{}
	_ common.GormSpecification = NameSpec{}
	_ common.GormSpecification = NameContainsSpec{}
	_ common.GormSpecification = CodeContainsSpec{}
)
