// Package seeders provides database seeding functionality.
package seeders

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/voidmaindev/go-template/internal/config"
	"gorm.io/gorm"
)

// Seeder defines the interface for database seeders.
type Seeder interface {
	// Name returns a unique name for the seeder.
	Name() string
	// Run executes the seeder logic.
	Run(db *gorm.DB, cfg *config.Config) error
}

// SeederRecord tracks applied seeders in the database.
type SeederRecord struct {
	ID       uint      `gorm:"primarykey"`
	Name     string    `gorm:"uniqueIndex;size:255;not null"`
	SeededAt time.Time `gorm:"not null"`
}

// TableName returns the table name for seeder records.
func (SeederRecord) TableName() string {
	return "seeders"
}

// SeederStatus represents the status of a seeder.
type SeederStatus struct {
	Name     string
	Applied  bool
	SeededAt *time.Time
}

// SeederManager handles database seeding.
type SeederManager struct {
	db      *gorm.DB
	cfg     *config.Config
	seeders []Seeder
}

// NewSeederManager creates a new SeederManager instance.
func NewSeederManager(db *gorm.DB, cfg *config.Config) *SeederManager {
	return &SeederManager{
		db:      db,
		cfg:     cfg,
		seeders: make([]Seeder, 0),
	}
}

// Register adds seeders to the manager.
func (m *SeederManager) Register(seeders ...Seeder) {
	m.seeders = append(m.seeders, seeders...)
}

// ensureSeederTable creates the seeder tracking table if it doesn't exist.
func (m *SeederManager) ensureSeederTable() error {
	return m.db.AutoMigrate(&SeederRecord{})
}

// getAppliedSeeders returns a map of applied seeder names.
func (m *SeederManager) getAppliedSeeders(ctx context.Context) (map[string]SeederRecord, error) {
	var records []SeederRecord
	if err := m.db.WithContext(ctx).Find(&records).Error; err != nil {
		return nil, fmt.Errorf("failed to get applied seeders: %w", err)
	}

	result := make(map[string]SeederRecord)
	for _, r := range records {
		result[r.Name] = r
	}
	return result, nil
}

// Run executes all pending seeders.
func (m *SeederManager) Run(ctx context.Context) error {
	if err := m.ensureSeederTable(); err != nil {
		return fmt.Errorf("failed to ensure seeder table: %w", err)
	}

	applied, err := m.getAppliedSeeders(ctx)
	if err != nil {
		return err
	}

	for _, seeder := range m.seeders {
		if _, ok := applied[seeder.Name()]; ok {
			slog.Debug("seeder already applied, skipping",
				"name", seeder.Name())
			continue
		}

		slog.Info("running seeder",
			"name", seeder.Name())

		if err := m.runSeeder(ctx, seeder); err != nil {
			return fmt.Errorf("seeder %s failed: %w", seeder.Name(), err)
		}

		slog.Info("seeder completed successfully",
			"name", seeder.Name())
	}

	return nil
}

// runSeeder executes a single seeder within a transaction.
func (m *SeederManager) runSeeder(ctx context.Context, seeder Seeder) error {
	return m.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := seeder.Run(tx, m.cfg); err != nil {
			return err
		}

		record := SeederRecord{
			Name:     seeder.Name(),
			SeededAt: time.Now(),
		}
		return tx.Create(&record).Error
	})
}

// Reset clears all seeder records, allowing seeders to be re-run.
func (m *SeederManager) Reset(ctx context.Context) error {
	if err := m.ensureSeederTable(); err != nil {
		return fmt.Errorf("failed to ensure seeder table: %w", err)
	}

	if err := m.db.WithContext(ctx).Where("1=1").Delete(&SeederRecord{}).Error; err != nil {
		return fmt.Errorf("failed to reset seeders: %w", err)
	}

	slog.Info("seeder records cleared")
	return nil
}

// Fresh resets and re-runs all seeders.
func (m *SeederManager) Fresh(ctx context.Context) error {
	if err := m.Reset(ctx); err != nil {
		return err
	}
	return m.Run(ctx)
}

// Status returns the status of all registered seeders.
func (m *SeederManager) Status(ctx context.Context) ([]SeederStatus, error) {
	if err := m.ensureSeederTable(); err != nil {
		return nil, fmt.Errorf("failed to ensure seeder table: %w", err)
	}

	applied, err := m.getAppliedSeeders(ctx)
	if err != nil {
		return nil, err
	}

	statuses := make([]SeederStatus, len(m.seeders))
	for i, seeder := range m.seeders {
		status := SeederStatus{
			Name:    seeder.Name(),
			Applied: false,
		}

		if record, ok := applied[seeder.Name()]; ok {
			status.Applied = true
			status.SeededAt = &record.SeededAt
		}

		statuses[i] = status
	}

	return statuses, nil
}

// HasPending returns true if there are pending seeders.
func (m *SeederManager) HasPending(ctx context.Context) (bool, error) {
	if err := m.ensureSeederTable(); err != nil {
		return false, err
	}

	applied, err := m.getAppliedSeeders(ctx)
	if err != nil {
		return false, err
	}

	for _, seeder := range m.seeders {
		if _, ok := applied[seeder.Name()]; !ok {
			return true, nil
		}
	}

	return false, nil
}

// DefaultSeederManager returns a seeder manager with all default seeders registered.
func DefaultSeederManager(db *gorm.DB, cfg *config.Config) *SeederManager {
	m := NewSeederManager(db, cfg)
	m.Register(
		&AdminUserSeeder{},
		&RBACSeeder{},
	)
	return m
}
