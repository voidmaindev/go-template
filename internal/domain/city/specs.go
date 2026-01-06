package city

import (
	"github.com/voidmaindev/go-template/internal/common"
	"gorm.io/gorm"
)

// ================================
// City Domain Specifications
// ================================

// NameSpec finds city by exact name
type NameSpec struct {
	Name string
}

// ByName creates a name specification
func ByName(name string) NameSpec {
	return NameSpec{Name: name}
}

func (s NameSpec) Apply(query any) any {
	return s.ApplyGorm(query.(*gorm.DB))
}

func (s NameSpec) ApplyGorm(db *gorm.DB) *gorm.DB {
	return db.Where("name = ?", s.Name)
}

// NameContainsSpec finds cities whose name contains a string
type NameContainsSpec struct {
	Name string
}

// ByNameContains creates a name contains specification
func ByNameContains(name string) NameContainsSpec {
	return NameContainsSpec{Name: name}
}

func (s NameContainsSpec) Apply(query any) any {
	return s.ApplyGorm(query.(*gorm.DB))
}

func (s NameContainsSpec) ApplyGorm(db *gorm.DB) *gorm.DB {
	return db.Where("name LIKE ?", "%"+s.Name+"%")
}

// CountryIDSpec finds cities by country ID
type CountryIDSpec struct {
	CountryID uint
}

// ByCountryID creates a country ID specification
func ByCountryID(countryID uint) CountryIDSpec {
	return CountryIDSpec{CountryID: countryID}
}

func (s CountryIDSpec) Apply(query any) any {
	return s.ApplyGorm(query.(*gorm.DB))
}

func (s CountryIDSpec) ApplyGorm(db *gorm.DB) *gorm.DB {
	return db.Where("country_id = ?", s.CountryID)
}

// CountryCodeSpec finds cities by country code (requires join)
type CountryCodeSpec struct {
	CountryCode string
}

// ByCountryCode creates a country code specification
func ByCountryCode(code string) CountryCodeSpec {
	return CountryCodeSpec{CountryCode: code}
}

func (s CountryCodeSpec) Apply(query any) any {
	return s.ApplyGorm(query.(*gorm.DB))
}

func (s CountryCodeSpec) ApplyGorm(db *gorm.DB) *gorm.DB {
	return db.Joins("Country").Where("Country.code = ?", s.CountryCode)
}

// Ensure all specs implement GormSpecification
var (
	_ common.GormSpecification = NameSpec{}
	_ common.GormSpecification = NameContainsSpec{}
	_ common.GormSpecification = CountryIDSpec{}
	_ common.GormSpecification = CountryCodeSpec{}
)
