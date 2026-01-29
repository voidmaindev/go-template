package migrations

import (
	"gorm.io/gorm"
)

// RenameRBACActions updates Casbin policy action names to standard CRUD terminology.
// Changes: write -> create, modify -> update
type RenameRBACActions struct{}

func (m *RenameRBACActions) Version() string { return "000008" }
func (m *RenameRBACActions) Name() string    { return "rename_rbac_actions" }

func (m *RenameRBACActions) Up(tx *gorm.DB) error {
	// Check if casbin_rule table exists
	if !tx.Migrator().HasTable("casbin_rule") {
		return nil
	}

	// Update 'write' actions to 'create'
	if err := tx.Exec("UPDATE casbin_rule SET v2 = 'create' WHERE v2 = 'write'").Error; err != nil {
		return err
	}

	// Update 'modify' actions to 'update'
	if err := tx.Exec("UPDATE casbin_rule SET v2 = 'update' WHERE v2 = 'modify'").Error; err != nil {
		return err
	}

	return nil
}

func (m *RenameRBACActions) Down(tx *gorm.DB) error {
	// Check if casbin_rule table exists
	if !tx.Migrator().HasTable("casbin_rule") {
		return nil
	}

	// Revert 'create' actions back to 'write'
	if err := tx.Exec("UPDATE casbin_rule SET v2 = 'write' WHERE v2 = 'create'").Error; err != nil {
		return err
	}

	// Revert 'update' actions back to 'modify'
	if err := tx.Exec("UPDATE casbin_rule SET v2 = 'modify' WHERE v2 = 'update'").Error; err != nil {
		return err
	}

	return nil
}
