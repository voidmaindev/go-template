package seeders

import (
	"fmt"
	"log/slog"

	"github.com/casbin/casbin/v3"
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
	}

	for _, role := range systemRoles {
		result := db.Where("code = ?", role.Code).FirstOrCreate(&role)
		if result.Error != nil {
			return fmt.Errorf("failed to create system role %s: %w", role.Code, result.Error)
		}
	}

	slog.Info("system roles created/verified")

	// Create Casbin enforcer to add policies
	adapter, err := gormadapter.NewAdapterByDB(db)
	if err != nil {
		return fmt.Errorf("failed to create Casbin adapter: %w", err)
	}

	enforcer, err := casbin.NewEnforcer(cfg.RBAC.ModelPath, adapter)
	if err != nil {
		return fmt.Errorf("failed to create Casbin enforcer: %w", err)
	}

	// Add admin wildcard policy
	if _, err := enforcer.AddPolicy(rbac.RoleCodeAdmin, "*", "*"); err != nil {
		slog.Warn("failed to add admin policy (may already exist)", "error", err)
	}

	if err := enforcer.SavePolicy(); err != nil {
		return fmt.Errorf("failed to save Casbin policies: %w", err)
	}

	slog.Info("system role policies created")

	// Find the admin user created by AdminUserSeeder
	adminEmail := cfg.Seed.AdminEmail
	var adminUser user.User
	result := db.Where("email = ?", adminEmail).First(&adminUser)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			slog.Warn("admin user not found - run AdminUserSeeder first", "email", adminEmail)
			return nil // Don't fail, admin user will be assigned role on next run
		}
		return fmt.Errorf("failed to find admin user: %w", result.Error)
	}

	// Assign admin role to the admin user via Casbin
	subject := rbac.FormatUserSubject(adminUser.ID)
	hasRole, _ := enforcer.HasRoleForUser(subject, rbac.RoleCodeAdmin)
	if !hasRole {
		if _, err := enforcer.AddRoleForUser(subject, rbac.RoleCodeAdmin); err != nil {
			return fmt.Errorf("failed to assign admin role: %w", err)
		}
		if err := enforcer.SavePolicy(); err != nil {
			return fmt.Errorf("failed to save admin role assignment: %w", err)
		}
		slog.Info("admin role assigned to admin user", "email", adminEmail)
	} else {
		slog.Info("admin user already has admin role", "email", adminEmail)
	}

	return nil
}
