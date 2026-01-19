package user

import (
	"time"

	"github.com/voidmaindev/go-template/internal/common"
	"gorm.io/gorm"
)

// ================================
// User Domain Specifications
// ================================

// EmailSpec finds user by email
type EmailSpec struct {
	Email string
}

// ByEmail creates an email specification
func ByEmail(email string) EmailSpec {
	return EmailSpec{Email: email}
}

func (s EmailSpec) Apply(query any) any {
	return s.ApplyGorm(common.AsGormDB(query))
}

func (s EmailSpec) ApplyGorm(db *gorm.DB) *gorm.DB {
	return db.Where("email = ?", s.Email)
}

// ActiveAfterSpec finds users created after a date
type ActiveAfterSpec struct {
	After time.Time
}

// ActiveAfter creates a specification for users created after a date
func ActiveAfter(t time.Time) ActiveAfterSpec {
	return ActiveAfterSpec{After: t}
}

func (s ActiveAfterSpec) Apply(query any) any {
	return s.ApplyGorm(common.AsGormDB(query))
}

func (s ActiveAfterSpec) ApplyGorm(db *gorm.DB) *gorm.DB {
	return db.Where("created_at > ?", s.After)
}

// CreatedBeforeSpec finds users created before a date
type CreatedBeforeSpec struct {
	Before time.Time
}

// CreatedBefore creates a specification for users created before a date
func CreatedBefore(t time.Time) CreatedBeforeSpec {
	return CreatedBeforeSpec{Before: t}
}

func (s CreatedBeforeSpec) Apply(query any) any {
	return s.ApplyGorm(common.AsGormDB(query))
}

func (s CreatedBeforeSpec) ApplyGorm(db *gorm.DB) *gorm.DB {
	return db.Where("created_at < ?", s.Before)
}

// NameContainsSpec finds users whose name contains a string
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

// EmailContainsSpec finds users whose email contains a string
type EmailContainsSpec struct {
	Email string
}

// ByEmailContains creates an email contains specification
func ByEmailContains(email string) EmailContainsSpec {
	return EmailContainsSpec{Email: email}
}

func (s EmailContainsSpec) Apply(query any) any {
	return s.ApplyGorm(common.AsGormDB(query))
}

func (s EmailContainsSpec) ApplyGorm(db *gorm.DB) *gorm.DB {
	return db.Where("email LIKE ?", "%"+s.Email+"%")
}

// Ensure all specs implement GormSpecification
var (
	_ common.GormSpecification = EmailSpec{}
	_ common.GormSpecification = ActiveAfterSpec{}
	_ common.GormSpecification = CreatedBeforeSpec{}
	_ common.GormSpecification = NameContainsSpec{}
	_ common.GormSpecification = EmailContainsSpec{}
)
