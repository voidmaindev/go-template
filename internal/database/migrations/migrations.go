// Package migrations provides versioned database migrations using GORM.
package migrations

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sort"
	"time"

	"gorm.io/gorm"
)

// Migration defines the interface for database migrations.
type Migration interface {
	// Version returns the migration version (e.g., "000001").
	Version() string
	// Name returns a human-readable name for the migration.
	Name() string
	// Up applies the migration.
	Up(tx *gorm.DB) error
	// Down reverts the migration.
	Down(tx *gorm.DB) error
}

// MigrationRecord tracks applied migrations in the database.
type MigrationRecord struct {
	ID        uint      `gorm:"primarykey"`
	Version   string    `gorm:"uniqueIndex;size:14;not null"`
	Name      string    `gorm:"size:255;not null"`
	AppliedAt time.Time `gorm:"not null"`
}

// TableName returns the table name for migration records.
func (MigrationRecord) TableName() string {
	return "schema_migrations"
}

// MigrationStatus represents the status of a migration.
type MigrationStatus struct {
	Version   string
	Name      string
	Applied   bool
	AppliedAt *time.Time
}

// Migrator handles database migrations.
type Migrator struct {
	db         *gorm.DB
	migrations []Migration
}

// NewMigrator creates a new Migrator instance.
func NewMigrator(db *gorm.DB) *Migrator {
	return &Migrator{
		db:         db,
		migrations: make([]Migration, 0),
	}
}

// Register adds migrations to the migrator.
func (m *Migrator) Register(migrations ...Migration) {
	m.migrations = append(m.migrations, migrations...)
	// Sort migrations by version
	sort.Slice(m.migrations, func(i, j int) bool {
		return m.migrations[i].Version() < m.migrations[j].Version()
	})
}

// ensureMigrationTable creates the migration tracking table if it doesn't exist.
func (m *Migrator) ensureMigrationTable() error {
	return m.db.AutoMigrate(&MigrationRecord{})
}

// getAppliedVersions returns a map of applied migration versions.
func (m *Migrator) getAppliedVersions(ctx context.Context) (map[string]MigrationRecord, error) {
	var records []MigrationRecord
	if err := m.db.WithContext(ctx).Find(&records).Error; err != nil {
		return nil, fmt.Errorf("failed to get applied migrations: %w", err)
	}

	result := make(map[string]MigrationRecord)
	for _, r := range records {
		result[r.Version] = r
	}
	return result, nil
}

// Up applies all pending migrations.
func (m *Migrator) Up(ctx context.Context) error {
	if err := m.ensureMigrationTable(); err != nil {
		return fmt.Errorf("failed to ensure migration table: %w", err)
	}

	applied, err := m.getAppliedVersions(ctx)
	if err != nil {
		return err
	}

	for _, migration := range m.migrations {
		if _, ok := applied[migration.Version()]; ok {
			continue // Already applied
		}

		slog.Info("applying migration",
			"version", migration.Version(),
			"name", migration.Name())

		if err := m.applyMigration(ctx, migration); err != nil {
			return fmt.Errorf("migration %s failed: %w", migration.Version(), err)
		}

		slog.Info("migration applied successfully",
			"version", migration.Version(),
			"name", migration.Name())
	}

	return nil
}

// UpTo applies migrations up to and including the specified version.
func (m *Migrator) UpTo(ctx context.Context, version string) error {
	if err := m.ensureMigrationTable(); err != nil {
		return fmt.Errorf("failed to ensure migration table: %w", err)
	}

	applied, err := m.getAppliedVersions(ctx)
	if err != nil {
		return err
	}

	for _, migration := range m.migrations {
		if migration.Version() > version {
			break
		}

		if _, ok := applied[migration.Version()]; ok {
			continue
		}

		slog.Info("applying migration",
			"version", migration.Version(),
			"name", migration.Name())

		if err := m.applyMigration(ctx, migration); err != nil {
			return fmt.Errorf("migration %s failed: %w", migration.Version(), err)
		}

		slog.Info("migration applied successfully",
			"version", migration.Version(),
			"name", migration.Name())
	}

	return nil
}

// applyMigration applies a single migration within a transaction.
func (m *Migrator) applyMigration(ctx context.Context, migration Migration) error {
	return m.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := migration.Up(tx); err != nil {
			return err
		}

		record := MigrationRecord{
			Version:   migration.Version(),
			Name:      migration.Name(),
			AppliedAt: time.Now(),
		}
		return tx.Create(&record).Error
	})
}

// Down reverts the last N migrations.
func (m *Migrator) Down(ctx context.Context, steps int) error {
	if err := m.ensureMigrationTable(); err != nil {
		return fmt.Errorf("failed to ensure migration table: %w", err)
	}

	if steps <= 0 {
		steps = 1
	}

	applied, err := m.getAppliedVersions(ctx)
	if err != nil {
		return err
	}

	// Get applied migrations in reverse order
	var appliedMigrations []Migration
	for i := len(m.migrations) - 1; i >= 0; i-- {
		if _, ok := applied[m.migrations[i].Version()]; ok {
			appliedMigrations = append(appliedMigrations, m.migrations[i])
		}
	}

	if len(appliedMigrations) == 0 {
		slog.Info("no migrations to revert")
		return nil
	}

	toRevert := appliedMigrations
	if steps < len(toRevert) {
		toRevert = appliedMigrations[:steps]
	}

	for _, migration := range toRevert {
		slog.Info("reverting migration",
			"version", migration.Version(),
			"name", migration.Name())

		if err := m.revertMigration(ctx, migration); err != nil {
			return fmt.Errorf("failed to revert migration %s: %w", migration.Version(), err)
		}

		slog.Info("migration reverted successfully",
			"version", migration.Version(),
			"name", migration.Name())
	}

	return nil
}

// revertMigration reverts a single migration within a transaction.
func (m *Migrator) revertMigration(ctx context.Context, migration Migration) error {
	return m.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := migration.Down(tx); err != nil {
			return err
		}

		return tx.Where("version = ?", migration.Version()).Delete(&MigrationRecord{}).Error
	})
}

// Status returns the status of all registered migrations.
func (m *Migrator) Status(ctx context.Context) ([]MigrationStatus, error) {
	if err := m.ensureMigrationTable(); err != nil {
		return nil, fmt.Errorf("failed to ensure migration table: %w", err)
	}

	applied, err := m.getAppliedVersions(ctx)
	if err != nil {
		return nil, err
	}

	statuses := make([]MigrationStatus, len(m.migrations))
	for i, migration := range m.migrations {
		status := MigrationStatus{
			Version: migration.Version(),
			Name:    migration.Name(),
			Applied: false,
		}

		if record, ok := applied[migration.Version()]; ok {
			status.Applied = true
			status.AppliedAt = &record.AppliedAt
		}

		statuses[i] = status
	}

	return statuses, nil
}

// Reset reverts all migrations.
func (m *Migrator) Reset(ctx context.Context) error {
	return m.Down(ctx, len(m.migrations))
}

// Refresh reverts all migrations and re-applies them.
func (m *Migrator) Refresh(ctx context.Context) error {
	if err := m.Reset(ctx); err != nil {
		return fmt.Errorf("reset failed: %w", err)
	}
	return m.Up(ctx)
}

// HasPending returns true if there are pending migrations.
func (m *Migrator) HasPending(ctx context.Context) (bool, error) {
	if err := m.ensureMigrationTable(); err != nil {
		return false, err
	}

	applied, err := m.getAppliedVersions(ctx)
	if err != nil {
		return false, err
	}

	for _, migration := range m.migrations {
		if _, ok := applied[migration.Version()]; !ok {
			return true, nil
		}
	}

	return false, nil
}

// GetPending returns all pending migrations.
func (m *Migrator) GetPending(ctx context.Context) ([]Migration, error) {
	if err := m.ensureMigrationTable(); err != nil {
		return nil, err
	}

	applied, err := m.getAppliedVersions(ctx)
	if err != nil {
		return nil, err
	}

	var pending []Migration
	for _, migration := range m.migrations {
		if _, ok := applied[migration.Version()]; !ok {
			pending = append(pending, migration)
		}
	}

	return pending, nil
}

// Validate checks that all migrations can be found and are properly ordered.
func (m *Migrator) Validate() error {
	if len(m.migrations) == 0 {
		return errors.New("no migrations registered")
	}

	seen := make(map[string]bool)
	for _, migration := range m.migrations {
		if migration.Version() == "" {
			return fmt.Errorf("migration has empty version")
		}
		if migration.Name() == "" {
			return fmt.Errorf("migration %s has empty name", migration.Version())
		}
		if seen[migration.Version()] {
			return fmt.Errorf("duplicate migration version: %s", migration.Version())
		}
		seen[migration.Version()] = true
	}

	return nil
}

// DefaultMigrator returns a migrator with all default migrations registered.
func DefaultMigrator(db *gorm.DB) *Migrator {
	m := NewMigrator(db)
	m.Register(
		&CreateUsersTable{},
		&CreateItemsTable{},
		&CreateCountriesTable{},
		&CreateCitiesTable{},
		&CreateDocumentsTable{},
		&CreateDocumentItemsTable{},
		&DropUserRoleColumn{},
	)
	return m
}
