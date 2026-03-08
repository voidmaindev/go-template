package seeders

import (
	"errors"
	"fmt"
	"log/slog"

	gormadapter "github.com/casbin/gorm-adapter/v3"
	"github.com/voidmaindev/go-template/internal/config"
	"github.com/voidmaindev/go-template/internal/domain/rbac"
	"github.com/voidmaindev/go-template/internal/domain/user"
	"gorm.io/gorm"
)

// RBACSeeder creates system roles and assigns admin role to the default admin user.
type RBACSeeder struct{}

// Name returns the seeder name.
func (s *RBACSeeder) Name() string {
	return "rbac_system_roles"
}

// Run creates system roles and assigns admin role to the default admin user.
// The admin user must be created first by AdminUserSeeder.
func (s *RBACSeeder) Run(db *gorm.DB, cfg *config.Config) error {
	// Create system roles metadata
	systemRoles := []rbac.Role{
		{
			Code:        rbac.RoleCodeAdmin,
			Name:        "Administrator",
			Description: "Full system access - all permissions on all domains",
			IsSystem:    true,
		},
		{
			Code:        rbac.RoleCodeFullReader,
			Name:        "Full Reader",
			Description: "Read-only access to all domains",
			IsSystem:    true,
		},
		{
			Code:        rbac.RoleCodeFullWriter,
			Name:        "Full Writer",
			Description: "Full CRUD on non-protected domains, read-only on protected domains",
			IsSystem:    true,
		},
		{
			Code:        rbac.RoleCodeUser,
			Name:        "User",
			Description: "Read-only access to non-protected domains (default role for new users)",
			IsSystem:    true,
		},
	}

	for _, role := range systemRoles {
		result := db.Where("code = ?", role.Code).FirstOrCreate(&role)
		if result.Error != nil {
			return fmt.Errorf("failed to create system role %s: %w", role.Code, result.Error)
		}
	}

	slog.Info("system roles created/verified")

	// Insert admin wildcard policy directly via GORM (avoids Casbin's internal transaction)
	adminPolicy := gormadapter.CasbinRule{
		Ptype: "p",
		V0:    rbac.RoleCodeAdmin,
		V1:    "*",
		V2:    "*",
	}
	if err := db.Where("ptype = ? AND v0 = ? AND v1 = ? AND v2 = ?",
		"p", rbac.RoleCodeAdmin, "*", "*").
		FirstOrCreate(&adminPolicy).Error; err != nil {
		return fmt.Errorf("failed to create admin policy: %w", err)
	}

	slog.Info("system role policies created")

	// Find the admin user created by AdminUserSeeder
	adminEmail := cfg.Seed.AdminEmail
	var adminUser user.User
	result := db.Where("email = ?", adminEmail).First(&adminUser)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			slog.Warn("admin user not found - run AdminUserSeeder first", "email", adminEmail)
			return nil // Don't fail, admin user will be assigned role on next run
		}
		return fmt.Errorf("failed to find admin user: %w", result.Error)
	}

	// Assign admin role to the admin user via direct GORM insert
	subject := rbac.FormatUserSubject(adminUser.ID)
	roleAssignment := gormadapter.CasbinRule{
		Ptype: "g",
		V0:    subject,
		V1:    rbac.RoleCodeAdmin,
	}
	if err := db.Where("ptype = ? AND v0 = ? AND v1 = ?",
		"g", subject, rbac.RoleCodeAdmin).
		FirstOrCreate(&roleAssignment).Error; err != nil {
		return fmt.Errorf("failed to assign admin role: %w", err)
	}

	slog.Info("admin role assigned to admin user", "email", adminEmail)

	return nil
}
