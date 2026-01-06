package document

import (
	"time"

	"github.com/voidmaindev/go-template/internal/common"
	"gorm.io/gorm"
)

// ================================
// Document Domain Specifications
// ================================

// CodeSpec finds document by code
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

// CodeContainsSpec finds documents whose code contains a string
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

// CityIDSpec finds documents by city ID
type CityIDSpec struct {
	CityID uint
}

// ByCityID creates a city ID specification
func ByCityID(cityID uint) CityIDSpec {
	return CityIDSpec{CityID: cityID}
}

func (s CityIDSpec) Apply(query any) any {
	return s.ApplyGorm(common.AsGormDB(query))
}

func (s CityIDSpec) ApplyGorm(db *gorm.DB) *gorm.DB {
	return db.Where("city_id = ?", s.CityID)
}

// DateRangeSpec finds documents within a date range
type DateRangeSpec struct {
	Start time.Time
	End   time.Time
}

// ByDateRange creates a date range specification
func ByDateRange(start, end time.Time) DateRangeSpec {
	return DateRangeSpec{Start: start, End: end}
}

func (s DateRangeSpec) Apply(query any) any {
	return s.ApplyGorm(common.AsGormDB(query))
}

func (s DateRangeSpec) ApplyGorm(db *gorm.DB) *gorm.DB {
	return db.Where("document_date BETWEEN ? AND ?", s.Start, s.End)
}

// DateAfterSpec finds documents after a date
type DateAfterSpec struct {
	After time.Time
}

// ByDateAfter creates a date after specification
func ByDateAfter(after time.Time) DateAfterSpec {
	return DateAfterSpec{After: after}
}

func (s DateAfterSpec) Apply(query any) any {
	return s.ApplyGorm(common.AsGormDB(query))
}

func (s DateAfterSpec) ApplyGorm(db *gorm.DB) *gorm.DB {
	return db.Where("document_date > ?", s.After)
}

// DateBeforeSpec finds documents before a date
type DateBeforeSpec struct {
	Before time.Time
}

// ByDateBefore creates a date before specification
func ByDateBefore(before time.Time) DateBeforeSpec {
	return DateBeforeSpec{Before: before}
}

func (s DateBeforeSpec) Apply(query any) any {
	return s.ApplyGorm(common.AsGormDB(query))
}

func (s DateBeforeSpec) ApplyGorm(db *gorm.DB) *gorm.DB {
	return db.Where("document_date < ?", s.Before)
}

// TotalAmountAboveSpec finds documents with total amount above threshold
type TotalAmountAboveSpec struct {
	Amount int64
}

// ByTotalAmountAbove creates a total amount above specification
func ByTotalAmountAbove(amount int64) TotalAmountAboveSpec {
	return TotalAmountAboveSpec{Amount: amount}
}

func (s TotalAmountAboveSpec) Apply(query any) any {
	return s.ApplyGorm(common.AsGormDB(query))
}

func (s TotalAmountAboveSpec) ApplyGorm(db *gorm.DB) *gorm.DB {
	return db.Where("total_amount > ?", s.Amount)
}

// TotalAmountBelowSpec finds documents with total amount below threshold
type TotalAmountBelowSpec struct {
	Amount int64
}

// ByTotalAmountBelow creates a total amount below specification
func ByTotalAmountBelow(amount int64) TotalAmountBelowSpec {
	return TotalAmountBelowSpec{Amount: amount}
}

func (s TotalAmountBelowSpec) Apply(query any) any {
	return s.ApplyGorm(common.AsGormDB(query))
}

func (s TotalAmountBelowSpec) ApplyGorm(db *gorm.DB) *gorm.DB {
	return db.Where("total_amount < ?", s.Amount)
}

// HasItemsSpec finds documents that have at least one item
type HasItemsSpec struct{}

// HasItems returns documents with items
func HasItems() HasItemsSpec {
	return HasItemsSpec{}
}

func (s HasItemsSpec) Apply(query any) any {
	return s.ApplyGorm(common.AsGormDB(query))
}

func (s HasItemsSpec) ApplyGorm(db *gorm.DB) *gorm.DB {
	return db.Where("EXISTS (SELECT 1 FROM document_items WHERE document_items.document_id = documents.id)")
}

// Ensure all specs implement GormSpecification
var (
	_ common.GormSpecification = CodeSpec{}
	_ common.GormSpecification = CodeContainsSpec{}
	_ common.GormSpecification = CityIDSpec{}
	_ common.GormSpecification = DateRangeSpec{}
	_ common.GormSpecification = DateAfterSpec{}
	_ common.GormSpecification = DateBeforeSpec{}
	_ common.GormSpecification = TotalAmountAboveSpec{}
	_ common.GormSpecification = TotalAmountBelowSpec{}
	_ common.GormSpecification = HasItemsSpec{}
)
