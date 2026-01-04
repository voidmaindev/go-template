package database

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/voidmaindev/go-template/internal/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// SlowQueryThreshold defines the duration above which queries are logged as slow
const SlowQueryThreshold = 200 * time.Millisecond

// slogAdapter adapts slog to GORM's logger interface
type slogAdapter struct{}

func (s *slogAdapter) LogMode(level logger.LogLevel) logger.Interface {
	return s
}

func (s *slogAdapter) Info(ctx context.Context, msg string, data ...any) {
	slog.Info(fmt.Sprintf(msg, data...))
}

func (s *slogAdapter) Warn(ctx context.Context, msg string, data ...any) {
	slog.Warn(fmt.Sprintf(msg, data...))
}

func (s *slogAdapter) Error(ctx context.Context, msg string, data ...any) {
	slog.Error(fmt.Sprintf(msg, data...))
}

func (s *slogAdapter) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	elapsed := time.Since(begin)
	sql, rows := fc()

	if err != nil {
		slog.Error("SQL error",
			"error", err,
			"elapsed", elapsed,
			"rows", rows,
			"sql", sql,
		)
		return
	}

	if elapsed > SlowQueryThreshold {
		slog.Warn("Slow SQL query",
			"elapsed", elapsed,
			"rows", rows,
			"sql", sql,
		)
		return
	}

	slog.Debug("SQL query",
		"elapsed", elapsed,
		"rows", rows,
		"sql", sql,
	)
}

// Connect establishes a connection to the PostgreSQL database
func Connect(cfg *config.DatabaseConfig) (*gorm.DB, error) {
	// Open database connection
	db, err := gorm.Open(postgres.Open(cfg.DSN()), &gorm.Config{
		Logger:                 &slogAdapter{},
		PrepareStmt:            true,
		SkipDefaultTransaction: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Get underlying SQL DB
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// Configure connection pool
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(cfg.MaxLifetime)

	// Ping to verify connection
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	slog.Info("Database connection established")
	return db, nil
}

// ConnectWithRetry attempts to connect to the database with retries
func ConnectWithRetry(cfg *config.DatabaseConfig, maxRetries int, delay time.Duration) (*gorm.DB, error) {
	var db *gorm.DB
	var err error

	for i := 0; i < maxRetries; i++ {
		db, err = Connect(cfg)
		if err == nil {
			return db, nil
		}

		slog.Warn("Failed to connect to database",
			"attempt", i+1,
			"max_retries", maxRetries,
			"error", err,
		)
		if i < maxRetries-1 {
			slog.Info("Retrying database connection", "delay", delay)
			time.Sleep(delay)
		}
	}

	return nil, fmt.Errorf("failed to connect to database after %d attempts: %w", maxRetries, err)
}

// Close closes the database connection
func Close(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	if err := sqlDB.Close(); err != nil {
		return fmt.Errorf("failed to close database connection: %w", err)
	}

	slog.Info("Database connection closed")
	return nil
}

// HealthCheck verifies the database connection is healthy
func HealthCheck(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}

	return nil
}
