package seeders

import (
	"fmt"
	"log/slog"

	"github.com/casbin/casbin/v2"
	gormadapter "github.com/casbin/gorm-adapter/v3"
	"github.com/voidmaindev/go-template/internal/config"
	"github.com/voidmaindev/go-template/internal/domain/rbac"
	"github.com/voidmaindev/go-template/internal/domain/user"
	"github.com/voidmaindev/go-template/pkg/utils"
	"gorm.io/gorm"
)

const (
	// DefaultSuperAdminPassword is the fallback password for development only.
	// In production, RBAC_SUPER_ADMIN_PASSWORD must be set via environment variable.
	DefaultSuperAdminPassword = "Ab123456"
)

// RBACSeeder creates system roles and super admin user
type RBACSeeder struct{}

// Name returns the seeder name.
func (s *RBACSeeder) Name() string {
	return "rbac_system_roles"
}

// Run creates system roles and super admin user.
// Configuration via Viper/environment variables:
//   - RBAC_SA_EMAIL: Super admin email (default: sa@admin.com)
//   - RBAC_SA_PASSWORD: Super admin password (REQUIRED in production)
//   - RBAC_SA_NAME: Super admin name (default: Super Admin)
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

	// Create super admin user
	saEmail := cfg.RBAC.SuperAdminEmail
	saName := cfg.RBAC.SuperAdminName
	saPassword := cfg.RBAC.SuperAdminPassword

	if saPassword == "" {
		slog.Warn("RBAC_SA_PASSWORD not set - using default password. Change this in production!")
		saPassword = DefaultSuperAdminPassword
	}

	hashedPassword, err := utils.HashPassword(saPassword)
	if err != nil {
		return fmt.Errorf("failed to hash super admin password: %w", err)
	}

	saUser := &user.User{
		Email:    saEmail,
		Password: hashedPassword,
		Name:     saName,
		Role:     user.RoleAdmin, // Keep the user model role field for backward compatibility
	}

	// Find or create super admin user
	var existingUser user.User
	result := db.Where("email = ?", saEmail).First(&existingUser)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			// Create new super admin user
			if err := db.Create(saUser).Error; err != nil {
				return fmt.Errorf("failed to create super admin user: %w", err)
			}
			existingUser = *saUser
			slog.Info("super admin user created", "email", saEmail)
		} else {
			return fmt.Errorf("failed to check for existing super admin: %w", result.Error)
		}
	} else {
		slog.Info("super admin user already exists", "email", saEmail)
	}

	// Assign admin role to super admin user via Casbin
	subject := rbac.FormatUserSubject(existingUser.ID)
	hasRole, _ := enforcer.HasRoleForUser(subject, rbac.RoleCodeAdmin)
	if !hasRole {
		if _, err := enforcer.AddRoleForUser(subject, rbac.RoleCodeAdmin); err != nil {
			return fmt.Errorf("failed to assign admin role to super admin: %w", err)
		}
		if err := enforcer.SavePolicy(); err != nil {
			return fmt.Errorf("failed to save admin role assignment: %w", err)
		}
		slog.Info("admin role assigned to super admin user")
	}

	return nil
}
