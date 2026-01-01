package cmd

import (
	"log/slog"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/voidmaindev/GoTemplate/internal/config"
	"github.com/voidmaindev/GoTemplate/internal/database"
	"github.com/voidmaindev/GoTemplate/internal/logger"

	"github.com/voidmaindev/GoTemplate/internal/domain/city"
	"github.com/voidmaindev/GoTemplate/internal/domain/country"
	"github.com/voidmaindev/GoTemplate/internal/domain/document"
	"github.com/voidmaindev/GoTemplate/internal/domain/item"
	"github.com/voidmaindev/GoTemplate/internal/domain/user"
)

// migrateCmd represents the migrate command
var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Run database migrations",
	Long:  `Run database migrations to create or update tables.`,
	Run:   runMigrate,
}

func init() {
	rootCmd.AddCommand(migrateCmd)
}

func runMigrate(cmd *cobra.Command, args []string) {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		slog.Error("Failed to load config", "error", err)
		os.Exit(1)
	}

	// Setup logger
	logger.SetupFromEnv(cfg.App.Environment, cfg.App.Debug)

	slog.Info("Connecting to database...")

	// Connect to database
	db, err := database.ConnectWithRetry(&cfg.Database, 5, 5*time.Second)
	if err != nil {
		slog.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer database.Close(db)

	slog.Info("Running migrations...")

	// Run migrations
	if err := database.Migrate(db,
		&user.User{},
		&item.Item{},
		&country.Country{},
		&city.City{},
		&document.Document{},
		&document.DocumentItem{},
	); err != nil {
		slog.Error("Failed to run migrations", "error", err)
		os.Exit(1)
	}

	slog.Info("Migrations completed successfully")
}
