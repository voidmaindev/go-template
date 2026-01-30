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

	// Create partial unique index on users.email (GORM can't express this via tags)
	// Only applies to non-deleted records, allowing soft-deleted emails to be reused
	if err := db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email ON users (email) WHERE deleted_at IS NULL").Error; err != nil {
		slog.Warn("Failed to create users email index", "error", err)
	}

	slog.Info("Database migrations completed successfully")
	return nil
}

// MigrateWithIndexes runs migrations and creates custom indexes
func MigrateWithIndexes(db *gorm.DB, models ...any) error {
	if err := Migrate(db, models...); err != nil {
		return err
	}

	// Create indexes for foreign key columns to optimize JOIN queries
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_cities_country_id ON cities(country_id)",
		"CREATE INDEX IF NOT EXISTS idx_documents_city_id ON documents(city_id)",
		"CREATE INDEX IF NOT EXISTS idx_document_items_document_id ON document_items(document_id)",
		"CREATE INDEX IF NOT EXISTS idx_document_items_item_id ON document_items(item_id)",
	}

	for _, idx := range indexes {
		if err := db.Exec(idx).Error; err != nil {
			slog.Warn("Failed to create index", "sql", idx, "error", err)
			// Continue with other indexes even if one fails
		}
	}

	slog.Info("Database indexes created/verified")
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
