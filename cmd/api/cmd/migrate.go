package cmd

import (
	"log/slog"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/voidmaindev/go-template/internal/app"
	"github.com/voidmaindev/go-template/internal/config"
	"github.com/voidmaindev/go-template/internal/container"
	"github.com/voidmaindev/go-template/internal/database"
	"github.com/voidmaindev/go-template/internal/logger"
)

// migrateCmd represents the migrate command
var migrateCmd = &cobra.Command{
	Use:   "migrate [app]",
	Short: "Run database migrations for an app",
	Long:  `Run database migrations to create or update tables for the specified app.`,
	Args:  cobra.ExactArgs(1),
	Run:   runMigrate,
}

func init() {
	rootCmd.AddCommand(migrateCmd)
}

func runMigrate(cmd *cobra.Command, args []string) {
	appName := args[0]

	// Get app from registry
	a := app.Get(appName)
	if a == nil {
		slog.Error("Unknown app", "name", appName, "available", strings.Join(app.List(), ", "))
		os.Exit(1)
	}

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		slog.Error("Failed to load config", "error", err)
		os.Exit(1)
	}

	// Setup logger
	logger.SetupFromEnv(cfg.App.Environment, cfg.App.Debug)

	slog.Info("Running migrations", "app", appName, "description", a.Description)

	// Connect to database
	db, err := database.ConnectWithRetry(&cfg.Database, cfg.App.IsProduction(), cfg.Database.RetryAttempts, cfg.Database.RetryDelay)
	if err != nil {
		slog.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer database.Close(db)

	// Create container and register domains for this app
	c := container.New(db, nil, cfg)
	for _, d := range a.Domains() {
		c.AddDomain(d)
	}

	// Run migrations using models from app's domains
	if err := database.Migrate(db, c.GetAllModels()...); err != nil {
		slog.Error("Failed to run migrations", "error", err)
		os.Exit(1)
	}

	slog.Info("Migrations completed successfully")
}
