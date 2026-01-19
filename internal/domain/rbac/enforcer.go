package rbac

import (
	"fmt"

	"github.com/casbin/casbin/v3"
	gormadapter "github.com/casbin/gorm-adapter/v3"
	"github.com/voidmaindev/go-template/internal/config"
	"gorm.io/gorm"
)

// EnforcerKey is the key used to store the enforcer in the container
const EnforcerKey = "rbac.enforcer"

// NewEnforcer creates a new Casbin enforcer with GORM adapter
func NewEnforcer(db *gorm.DB, cfg *config.RBACConfig) (*casbin.Enforcer, error) {
	// Create GORM adapter (auto-creates casbin_rule table)
	adapter, err := gormadapter.NewAdapterByDB(db)
	if err != nil {
		return nil, fmt.Errorf("failed to create Casbin adapter: %w", err)
	}

	// Create enforcer with model file and adapter
	enforcer, err := casbin.NewEnforcer(cfg.ModelPath, adapter)
	if err != nil {
		return nil, fmt.Errorf("failed to create Casbin enforcer: %w", err)
	}

	// Load policies from database
	if err := enforcer.LoadPolicy(); err != nil {
		return nil, fmt.Errorf("failed to load Casbin policies: %w", err)
	}

	return enforcer, nil
}

// FormatUserSubject formats a user ID as a Casbin subject
func FormatUserSubject(userID uint) string {
	return fmt.Sprintf("user:%d", userID)
}

// ParseUserSubject parses a Casbin subject to extract user ID
func ParseUserSubject(subject string) (uint, bool) {
	var userID uint
	_, err := fmt.Sscanf(subject, "user:%d", &userID)
	return userID, err == nil
}
