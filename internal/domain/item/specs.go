package item

import (
	"github.com/voidmaindev/go-template/internal/common"
	"gorm.io/gorm"
)

// ================================
// Item Domain Specifications
// ================================

// NameSpec finds item by exact name
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

// NameContainsSpec finds items whose name contains a string
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

// PriceRangeSpec finds items within a price range
type PriceRangeSpec struct {
	Min int64
	Max int64
}

// ByPriceRange creates a price range specification
func ByPriceRange(min, max int64) PriceRangeSpec {
	return PriceRangeSpec{Min: min, Max: max}
}

func (s PriceRangeSpec) Apply(query any) any {
	return s.ApplyGorm(common.AsGormDB(query))
}

func (s PriceRangeSpec) ApplyGorm(db *gorm.DB) *gorm.DB {
	return db.Where("price BETWEEN ? AND ?", s.Min, s.Max)
}

// PriceAboveSpec finds items with price above threshold
type PriceAboveSpec struct {
	Price int64
}

// ByPriceAbove creates a price above specification
func ByPriceAbove(price int64) PriceAboveSpec {
	return PriceAboveSpec{Price: price}
}

func (s PriceAboveSpec) Apply(query any) any {
	return s.ApplyGorm(common.AsGormDB(query))
}

func (s PriceAboveSpec) ApplyGorm(db *gorm.DB) *gorm.DB {
	return db.Where("price > ?", s.Price)
}

// PriceBelowSpec finds items with price below threshold
type PriceBelowSpec struct {
	Price int64
}

// ByPriceBelow creates a price below specification
func ByPriceBelow(price int64) PriceBelowSpec {
	return PriceBelowSpec{Price: price}
}

func (s PriceBelowSpec) Apply(query any) any {
	return s.ApplyGorm(common.AsGormDB(query))
}

func (s PriceBelowSpec) ApplyGorm(db *gorm.DB) *gorm.DB {
	return db.Where("price < ?", s.Price)
}

// HasDescriptionSpec finds items with non-empty description
type HasDescriptionSpec struct{}

// HasDescription returns items with description
func HasDescription() HasDescriptionSpec {
	return HasDescriptionSpec{}
}

func (s HasDescriptionSpec) Apply(query any) any {
	return s.ApplyGorm(common.AsGormDB(query))
}

func (s HasDescriptionSpec) ApplyGorm(db *gorm.DB) *gorm.DB {
	return db.Where("description IS NOT NULL AND description != ''")
}

// Ensure all specs implement GormSpecification
var (
	_ common.GormSpecification = NameSpec{}
	_ common.GormSpecification = NameContainsSpec{}
	_ common.GormSpecification = PriceRangeSpec{}
	_ common.GormSpecification = PriceAboveSpec{}
	_ common.GormSpecification = PriceBelowSpec{}
	_ common.GormSpecification = HasDescriptionSpec{}
)
