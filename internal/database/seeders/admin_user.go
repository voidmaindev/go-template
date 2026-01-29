package seeders

import (
	"log/slog"

	"github.com/voidmaindev/go-template/internal/config"
	"github.com/voidmaindev/go-template/internal/domain/user"
	"github.com/voidmaindev/go-template/pkg/utils"
	"gorm.io/gorm"
)

const (
	// DefaultAdminPassword is the fallback password for development only.
	// In production, SEED_ADMIN_PASSWORD must be set via environment variable.
	DefaultAdminPassword = "Ab123456"
)

// AdminUserSeeder creates a default admin user.
type AdminUserSeeder struct{}

// Name returns the seeder name.
func (s *AdminUserSeeder) Name() string {
	return "admin_user"
}

// Run creates the default admin user if it doesn't exist.
// Configuration via Viper/environment variables:
//   - SEED_ADMIN_EMAIL: Admin email (default: admin@admin.com)
//   - SEED_ADMIN_PASSWORD: Admin password (REQUIRED in production)
//   - SEED_ADMIN_NAME: Admin name (default: Administrator)
func (s *AdminUserSeeder) Run(db *gorm.DB, cfg *config.Config) error {
	email := cfg.Seed.AdminEmail
	name := cfg.Seed.AdminName
	password := cfg.Seed.AdminPassword

	if password == "" {
		slog.Warn("SEED_ADMIN_PASSWORD not set - using default password. Change this in production!")
		password = DefaultAdminPassword
	}

	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		return err
	}

	admin := &user.User{
		Email:    email,
		Password: &hashedPassword,
		Name:     name,
	}

	// FirstOrCreate ensures idempotency - won't duplicate if already exists
	return db.Where("email = ?", admin.Email).FirstOrCreate(admin).Error
}
