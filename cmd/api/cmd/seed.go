package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/voidmaindev/go-template/internal/config"
	"github.com/voidmaindev/go-template/internal/database"
	"github.com/voidmaindev/go-template/internal/database/seeders"
	"github.com/voidmaindev/go-template/internal/logger"
)

var (
	seedFresh bool
)

// seedCmd represents the seed command
var seedCmd = &cobra.Command{
	Use:   "seed",
	Short: "Run database seeders",
	Long:  `Run all pending database seeders to populate initial data.`,
	Run:   runSeed,
}

// seedStatusCmd shows seeder status
var seedStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show seeder status",
	Long:  `Show the status of all seeders (applied/pending).`,
	Run:   runSeedStatus,
}

// seedResetCmd clears seeder records
var seedResetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Reset seeder records",
	Long:  `Clear all seeder records to allow re-running seeders.`,
	Run:   runSeedReset,
}

func init() {
	rootCmd.AddCommand(seedCmd)

	seedCmd.AddCommand(seedStatusCmd)
	seedCmd.AddCommand(seedResetCmd)

	seedCmd.Flags().BoolVar(&seedFresh, "fresh", false, "Reset and re-run all seeders")
}

func getSeederManager() (*seeders.SeederManager, func(), error) {
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

	manager := seeders.DefaultSeederManager(db, cfg)
	return manager, cleanup, nil
}

func runSeed(cmd *cobra.Command, args []string) {
	manager, cleanup, err := getSeederManager()
	if err != nil {
		slog.Error("Setup failed", "error", err)
		os.Exit(1)
	}
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	if seedFresh {
		slog.Info("Running fresh seed (reset + run all)")
		if err := manager.Fresh(ctx); err != nil {
			slog.Error("Fresh seed failed", "error", err)
			os.Exit(1)
		}
	} else {
		slog.Info("Running pending seeders")
		if err := manager.Run(ctx); err != nil {
			slog.Error("Seeding failed", "error", err)
			os.Exit(1)
		}
	}

	slog.Info("Seeding completed successfully")
}

func runSeedStatus(cmd *cobra.Command, args []string) {
	manager, cleanup, err := getSeederManager()
	if err != nil {
		slog.Error("Setup failed", "error", err)
		os.Exit(1)
	}
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	statuses, err := manager.Status(ctx)
	if err != nil {
		slog.Error("Failed to get seeder status", "error", err)
		os.Exit(1)
	}

	fmt.Println("\nSeeder Status:")
	fmt.Println(strings.Repeat("-", 60))
	fmt.Printf("%-30s %-10s %s\n", "Name", "Status", "Seeded At")
	fmt.Println(strings.Repeat("-", 60))

	pendingCount := 0
	for _, status := range statuses {
		statusStr := "Pending"
		seededAt := ""
		if status.Applied {
			statusStr = "Applied"
			if status.SeededAt != nil {
				seededAt = status.SeededAt.Format("2006-01-02 15:04:05")
			}
		} else {
			pendingCount++
		}
		fmt.Printf("%-30s %-10s %s\n", status.Name, statusStr, seededAt)
	}

	fmt.Println(strings.Repeat("-", 60))
	fmt.Printf("Total: %d seeders, %d pending\n\n", len(statuses), pendingCount)
}

func runSeedReset(cmd *cobra.Command, args []string) {
	manager, cleanup, err := getSeederManager()
	if err != nil {
		slog.Error("Setup failed", "error", err)
		os.Exit(1)
	}
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	slog.Info("Resetting seeder records")
	if err := manager.Reset(ctx); err != nil {
		slog.Error("Reset failed", "error", err)
		os.Exit(1)
	}

	slog.Info("Seeder records cleared - seeders can be re-run")
}
