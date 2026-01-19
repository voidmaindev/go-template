package migrations

import (
	"gorm.io/gorm"
)

// DropUserRoleColumn removes the role column from users table.
// Authorization is now handled exclusively by RBAC (Casbin).
type DropUserRoleColumn struct{}

func (m *DropUserRoleColumn) Version() string { return "000007" }
func (m *DropUserRoleColumn) Name() string    { return "drop_user_role_column" }

func (m *DropUserRoleColumn) Up(tx *gorm.DB) error {
	// Check if column exists before dropping
	if tx.Migrator().HasColumn("users", "role") {
		return tx.Migrator().DropColumn("users", "role")
	}
	return nil
}

func (m *DropUserRoleColumn) Down(tx *gorm.DB) error {
	// Re-add the role column with default value
	return tx.Exec("ALTER TABLE users ADD COLUMN role VARCHAR(20) NOT NULL DEFAULT 'user'").Error
}
