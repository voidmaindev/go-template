package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/voidmaindev/go-template/internal/app"
	"github.com/voidmaindev/go-template/internal/config"
	"github.com/voidmaindev/go-template/internal/database"
	"github.com/voidmaindev/go-template/internal/database/migrations"
	"github.com/voidmaindev/go-template/internal/logger"
)

var (
	migrateTo string
)

// migrateCmd represents the migrate command
var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Database migration commands",
	Long:  `Manage database migrations: apply, revert, and check status.`,
}

// migrateUpCmd applies pending migrations
var migrateUpCmd = &cobra.Command{
	Use:   "up",
	Short: "Apply pending migrations",
	Long:  `Apply all pending database migrations, or up to a specific version with --to flag.`,
	Run:   runMigrateUp,
}

// migrateDownCmd reverts migrations
var migrateDownCmd = &cobra.Command{
	Use:   "down [steps]",
	Short: "Revert migrations",
	Long:  `Revert the last N migrations (default: 1).`,
	Args:  cobra.MaximumNArgs(1),
	Run:   runMigrateDown,
}

// migrateStatusCmd shows migration status
var migrateStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show migration status",
	Long:  `Show the status of all migrations (applied/pending).`,
	Run:   runMigrateStatus,
}

// migrateResetCmd reverts all migrations
var migrateResetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Revert all migrations",
	Long:  `Revert all applied migrations. Use with caution!`,
	Run:   runMigrateReset,
}

// migrateRefreshCmd reverts and re-applies all migrations
var migrateRefreshCmd = &cobra.Command{
	Use:   "refresh",
	Short: "Refresh all migrations",
	Long:  `Revert all migrations and re-apply them. Use with caution!`,
	Run:   runMigrateRefresh,
}

// migrateLegacyCmd provides backward compatibility with old migrate command
var migrateLegacyCmd = &cobra.Command{
	Use:    "legacy [app]",
	Short:  "Run legacy AutoMigrate (deprecated)",
	Long:   `Run GORM AutoMigrate for backward compatibility. Use 'migrate up' instead.`,
	Args:   cobra.ExactArgs(1),
	Hidden: true,
	Run:    runMigrateLegacy,
}

func init() {
	rootCmd.AddCommand(migrateCmd)

	migrateCmd.AddCommand(migrateUpCmd)
	migrateCmd.AddCommand(migrateDownCmd)
	migrateCmd.AddCommand(migrateStatusCmd)
	migrateCmd.AddCommand(migrateResetCmd)
	migrateCmd.AddCommand(migrateRefreshCmd)
	migrateCmd.AddCommand(migrateLegacyCmd)

	migrateUpCmd.Flags().StringVar(&migrateTo, "to", "", "Migrate up to specific version (e.g., 000003)")
}

func getMigrator() (*migrations.Migrator, func(), error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load config: %w", err)
	}

	logger.SetupFromEnv(cfg.App.Environment, cfg.App.Debug)

	db, err := database.ConnectWithRetry(&cfg.Database, cfg.App.IsProduction(), cfg.Database.RetryAttempts, cfg.Database.RetryDelay)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	cleanup := func() {
		database.Close(db)
	}

	migrator := migrations.DefaultMigrator(db)
	return migrator, cleanup, nil
}

func runMigrateUp(cmd *cobra.Command, args []string) {
	migrator, cleanup, err := getMigrator()
	if err != nil {
		slog.Error("Setup failed", "error", err)
		os.Exit(1)
	}
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	if migrateTo != "" {
		slog.Info("Applying migrations up to version", "version", migrateTo)
		if err := migrator.UpTo(ctx, migrateTo); err != nil {
			slog.Error("Migration failed", "error", err)
			os.Exit(1)
		}
	} else {
		slog.Info("Applying all pending migrations")
		if err := migrator.Up(ctx); err != nil {
			slog.Error("Migration failed", "error", err)
			os.Exit(1)
		}
	}

	slog.Info("Migrations completed successfully")
}

func runMigrateDown(cmd *cobra.Command, args []string) {
	migrator, cleanup, err := getMigrator()
	if err != nil {
		slog.Error("Setup failed", "error", err)
		os.Exit(1)
	}
	defer cleanup()

	steps := 1
	if len(args) > 0 {
		steps, err = strconv.Atoi(args[0])
		if err != nil || steps < 1 {
			slog.Error("Invalid steps argument", "value", args[0])
			os.Exit(1)
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	slog.Info("Reverting migrations", "steps", steps)
	if err := migrator.Down(ctx, steps); err != nil {
		slog.Error("Migration rollback failed", "error", err)
		os.Exit(1)
	}

	slog.Info("Rollback completed successfully")
}

func runMigrateStatus(cmd *cobra.Command, args []string) {
	migrator, cleanup, err := getMigrator()
	if err != nil {
		slog.Error("Setup failed", "error", err)
		os.Exit(1)
	}
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	statuses, err := migrator.Status(ctx)
	if err != nil {
		slog.Error("Failed to get migration status", "error", err)
		os.Exit(1)
	}

	fmt.Println("\nMigration Status:")
	fmt.Println(strings.Repeat("-", 80))
	fmt.Printf("%-10s %-35s %-10s %s\n", "Version", "Name", "Status", "Applied At")
	fmt.Println(strings.Repeat("-", 80))

	pendingCount := 0
	for _, status := range statuses {
		statusStr := "Pending"
		appliedAt := ""
		if status.Applied {
			statusStr = "Applied"
			if status.AppliedAt != nil {
				appliedAt = status.AppliedAt.Format("2006-01-02 15:04:05")
			}
		} else {
			pendingCount++
		}
		fmt.Printf("%-10s %-35s %-10s %s\n", status.Version, status.Name, statusStr, appliedAt)
	}

	fmt.Println(strings.Repeat("-", 80))
	fmt.Printf("Total: %d migrations, %d pending\n\n", len(statuses), pendingCount)
}

func runMigrateReset(cmd *cobra.Command, args []string) {
	migrator, cleanup, err := getMigrator()
	if err != nil {
		slog.Error("Setup failed", "error", err)
		os.Exit(1)
	}
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	slog.Warn("Resetting all migrations - this will drop all tables!")
	if err := migrator.Reset(ctx); err != nil {
		slog.Error("Reset failed", "error", err)
		os.Exit(1)
	}

	slog.Info("Reset completed successfully")
}

func runMigrateRefresh(cmd *cobra.Command, args []string) {
	migrator, cleanup, err := getMigrator()
	if err != nil {
		slog.Error("Setup failed", "error", err)
		os.Exit(1)
	}
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	slog.Warn("Refreshing all migrations - this will drop and recreate all tables!")
	if err := migrator.Refresh(ctx); err != nil {
		slog.Error("Refresh failed", "error", err)
		os.Exit(1)
	}

	slog.Info("Refresh completed successfully")
}

// runMigrateLegacy provides backward compatibility with the old migrate command
func runMigrateLegacy(cmd *cobra.Command, args []string) {
	appName := args[0]

	a := app.Get(appName)
	if a == nil {
		slog.Error("Unknown app", "name", appName, "available", strings.Join(app.List(), ", "))
		os.Exit(1)
	}

	cfg, err := config.Load()
	if err != nil {
		slog.Error("Failed to load config", "error", err)
		os.Exit(1)
	}

	logger.SetupFromEnv(cfg.App.Environment, cfg.App.Debug)

	slog.Warn("Using legacy AutoMigrate - consider using 'migrate up' instead")
	slog.Info("Running migrations", "app", appName, "description", a.Description)

	db, err := database.ConnectWithRetry(&cfg.Database, cfg.App.IsProduction(), cfg.Database.RetryAttempts, cfg.Database.RetryDelay)
	if err != nil {
		slog.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer database.Close(db)

	// Import container package for legacy support
	// Create container and register domains for this app
	// c := container.New(db, nil, cfg)
	// for _, d := range a.Domains() {
	// 	c.AddDomain(d)
	// }

	// Run legacy AutoMigrate
	if err := database.Migrate(db, a.Domains()[0].Models()...); err != nil {
		slog.Error("Failed to run migrations", "error", err)
		os.Exit(1)
	}

	slog.Info("Migrations completed successfully")
}
