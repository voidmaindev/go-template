package seeders

import (
	"fmt"
	"log/slog"

	gormadapter "github.com/casbin/gorm-adapter/v3"
	"github.com/voidmaindev/go-template/internal/config"
	"github.com/voidmaindev/go-template/internal/domain/rbac"
	"gorm.io/gorm"
)

// SelfRegisteredRoleSeeder creates the self_registered role with limited permissions.
type SelfRegisteredRoleSeeder struct{}

// Name returns the seeder name.
func (s *SelfRegisteredRoleSeeder) Name() string {
	return "self_registered_role"
}

// Run creates the self_registered role with limited permissions.
// Self-registered users get read-only access to non-protected domains.
func (s *SelfRegisteredRoleSeeder) Run(db *gorm.DB, cfg *config.Config) error {
	// Create self_registered role metadata
	selfRegisteredRole := rbac.Role{
		Code:        rbac.RoleCodeSelfRegistered,
		Name:        "Self-Registered User",
		Description: "Limited read access for self-registered users (email or OAuth)",
		IsSystem:    true,
	}

	result := db.Where("code = ?", selfRegisteredRole.Code).FirstOrCreate(&selfRegisteredRole)
	if result.Error != nil {
		return fmt.Errorf("failed to create self_registered role: %w", result.Error)
	}

	slog.Info("self_registered role created/verified")

	// Add read-only policies for non-protected domains
	// We'll give read access to common domains like item, city, country, document
	nonProtectedDomains := []string{"item", "city", "country", "document"}

	for _, domain := range nonProtectedDomains {
		policy := gormadapter.CasbinRule{
			Ptype: "p",
			V0:    rbac.RoleCodeSelfRegistered,
			V1:    domain,
			V2:    rbac.ActionRead,
		}
		if err := db.Where("ptype = ? AND v0 = ? AND v1 = ? AND v2 = ?",
			"p", rbac.RoleCodeSelfRegistered, domain, rbac.ActionRead).
			FirstOrCreate(&policy).Error; err != nil {
			return fmt.Errorf("failed to create self_registered policy for %s: %w", domain, err)
		}
	}

	slog.Info("self_registered role policies created")
	return nil
}
