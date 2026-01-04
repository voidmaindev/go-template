package database

import (
	"log/slog"

	"gorm.io/gorm"
)

// Migrator interface for models that need migration
type Migrator interface {
	TableName() string
}

// Migrate runs auto-migrations for all registered models
func Migrate(db *gorm.DB, models ...any) error {
	slog.Info("Running database migrations...")

	if err := db.AutoMigrate(models...); err != nil {
		return err
	}

	slog.Info("Database migrations completed successfully")
	return nil
}

// MigrateWithIndexes runs migrations and creates custom indexes
func MigrateWithIndexes(db *gorm.DB, models ...any) error {
	if err := Migrate(db, models...); err != nil {
		return err
	}

	// Add any custom indexes here if needed
	// Example:
	// db.Exec("CREATE INDEX IF NOT EXISTS idx_users_email ON users (email)")

	return nil
}

// DropTables drops all tables (USE WITH CAUTION - for testing only)
func DropTables(db *gorm.DB, models ...any) error {
	slog.Warn("Dropping all tables...")

	for _, model := range models {
		if err := db.Migrator().DropTable(model); err != nil {
			slog.Error("Failed to drop table", "model", model, "error", err)
		}
	}

	slog.Info("All tables dropped")
	return nil
}

// HasTable checks if a table exists
func HasTable(db *gorm.DB, model any) bool {
	return db.Migrator().HasTable(model)
}

// CreateTableIfNotExists creates a table only if it doesn't exist
func CreateTableIfNotExists(db *gorm.DB, model any) error {
	if !HasTable(db, model) {
		return db.Migrator().CreateTable(model)
	}
	return nil
}
