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

// RoleSpec finds users by role
type RoleSpec struct {
	Role Role
}

// ByRole creates a role specification
func ByRole(role Role) RoleSpec {
	return RoleSpec{Role: role}
}

func (s RoleSpec) Apply(query any) any {
	return s.ApplyGorm(common.AsGormDB(query))
}

func (s RoleSpec) ApplyGorm(db *gorm.DB) *gorm.DB {
	return db.Where("role = ?", s.Role)
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

// AdminUsersSpec finds only admin users
type AdminUsersSpec struct{}

// AdminUsers returns all admin users
func AdminUsers() AdminUsersSpec {
	return AdminUsersSpec{}
}

func (s AdminUsersSpec) Apply(query any) any {
	return s.ApplyGorm(common.AsGormDB(query))
}

func (s AdminUsersSpec) ApplyGorm(db *gorm.DB) *gorm.DB {
	return db.Where("role = ?", RoleAdmin)
}

// RegularUsersSpec finds only regular users
type RegularUsersSpec struct{}

// RegularUsers returns all regular users
func RegularUsers() RegularUsersSpec {
	return RegularUsersSpec{}
}

func (s RegularUsersSpec) Apply(query any) any {
	return s.ApplyGorm(common.AsGormDB(query))
}

func (s RegularUsersSpec) ApplyGorm(db *gorm.DB) *gorm.DB {
	return db.Where("role = ?", RoleUser)
}

// Ensure all specs implement GormSpecification
var (
	_ common.GormSpecification = EmailSpec{}
	_ common.GormSpecification = RoleSpec{}
	_ common.GormSpecification = ActiveAfterSpec{}
	_ common.GormSpecification = CreatedBeforeSpec{}
	_ common.GormSpecification = NameContainsSpec{}
	_ common.GormSpecification = EmailContainsSpec{}
	_ common.GormSpecification = AdminUsersSpec{}
	_ common.GormSpecification = RegularUsersSpec{}
)
