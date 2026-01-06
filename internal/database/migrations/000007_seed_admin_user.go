package migrations

import (
	"time"

	"github.com/voidmaindev/go-template/pkg/utils"
	"gorm.io/gorm"
)

// SeedAdminUser creates a default admin user.
type SeedAdminUser struct{}

func (m *SeedAdminUser) Version() string { return "000007" }
func (m *SeedAdminUser) Name() string    { return "seed_admin_user" }

func (m *SeedAdminUser) Up(tx *gorm.DB) error {
	// Hash the default password
	hashedPassword, err := utils.HashPassword("Ab123456")
	if err != nil {
		return err
	}

	now := time.Now()

	// Insert admin user using raw SQL to avoid model dependencies
	return tx.Exec(`
		INSERT INTO users (email, password, name, role, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT (email) DO NOTHING
	`, "admin@admin.com", hashedPassword, "Administrator", "admin", now, now).Error
}

func (m *SeedAdminUser) Down(tx *gorm.DB) error {
	return tx.Exec(`DELETE FROM users WHERE email = ?`, "admin@admin.com").Error
}
