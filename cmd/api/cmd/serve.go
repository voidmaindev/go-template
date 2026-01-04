package cmd

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/gofiber/fiber/v2"
	"github.com/spf13/cobra"
	"github.com/voidmaindev/go-template/internal/app"
	"github.com/voidmaindev/go-template/internal/config"
	"github.com/voidmaindev/go-template/internal/container"
	"github.com/voidmaindev/go-template/internal/database"
	"github.com/voidmaindev/go-template/internal/logger"
	"github.com/voidmaindev/go-template/internal/middleware"
	"github.com/voidmaindev/go-template/internal/redis"
)


// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve [app]",
	Short: "Start an app server",
	Long:  `Start an HTTP API server for the specified app with its configured domains.`,
	Args:  cobra.ExactArgs(1),
	Run:   runServe,
}

func init() {
	rootCmd.AddCommand(serveCmd)

	// Local flags for serve command
	serveCmd.Flags().IntP("port", "p", 3000, "Port to run the server on")
	serveCmd.Flags().StringP("host", "H", "0.0.0.0", "Host to bind the server to")
}

func runServe(cmd *cobra.Command, args []string) {
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

	// Setup logger based on environment
	logger.SetupFromEnv(cfg.App.Environment, cfg.App.Debug)

	// Override from flags if provided
	if port, _ := cmd.Flags().GetInt("port"); port != 0 {
		cfg.Server.Port = port
	}
	if host, _ := cmd.Flags().GetString("host"); host != "" {
		cfg.Server.Host = host
	}

	slog.Info("Starting server",
		"appName", appName,
		"description", a.Description,
		"env", cfg.App.Environment,
		"debug", cfg.App.Debug,
	)

	// Connect to database
	slog.Info("Connecting to database...")
	db, err := database.ConnectWithRetry(&cfg.Database, cfg.Database.RetryAttempts, cfg.Database.RetryDelay)
	if err != nil {
		slog.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}

	// Connect to Redis
	slog.Info("Connecting to Redis...")
	redisClient, err := redis.ConnectWithRetry(&cfg.Redis, cfg.Redis.RetryAttempts, cfg.Redis.RetryDelay)
	if err != nil {
		slog.Error("Failed to connect to Redis", "error", err)
		os.Exit(1)
	}

	// Create dependency container
	c := container.New(db, redisClient, cfg)

	// Register domains for this app
	for _, d := range a.Domains() {
		c.AddDomain(d)
	}

	// Run database migrations
	slog.Info("Running migrations...")
	if err := database.Migrate(db, c.GetAllModels()...); err != nil {
		slog.Error("Failed to run migrations", "error", err)
		os.Exit(1)
	}

	// Register all domain components (repos, services, handlers)
	c.RegisterAll()

	// Initialize Fiber app
	app := fiber.New(fiber.Config{
		AppName:               cfg.App.Name,
		ReadTimeout:           cfg.Server.ReadTimeout,
		WriteTimeout:          cfg.Server.WriteTimeout,
		IdleTimeout:           cfg.Server.IdleTimeout,
		DisableStartupMessage: !cfg.App.Debug,
		ErrorHandler:          customErrorHandler,
	})

	// Setup global middleware
	middleware.SetupCORS(app, cfg)
	middleware.SetupSlogLogger(app)
	middleware.SetupCustomRecovery(app, cfg.App.IsDevelopment())

	// Health check endpoint with DB and Redis verification
	app.Get("/health", func(c *fiber.Ctx) error {
		// Check database connectivity
		if err := database.HealthCheck(db); err != nil {
			slog.Error("Health check failed: database", "error", err)
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"status":  "unhealthy",
				"service": cfg.App.Name,
				"error":   "database connection failed",
			})
		}

		// Check Redis connectivity
		if err := redisClient.HealthCheck(c.Context()); err != nil {
			slog.Error("Health check failed: redis", "error", err)
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"status":  "unhealthy",
				"service": cfg.App.Name,
				"error":   "redis connection failed",
			})
		}

		return c.JSON(fiber.Map{
			"status":  "healthy",
			"service": cfg.App.Name,
			"env":     cfg.App.Environment,
		})
	})

	// API v1 routes
	api := app.Group("/api/v1")

	// Register all domain routes
	c.RegisterRoutes(api)

	// 404 handler
	app.Use(func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"error":   "endpoint not found",
		})
	})

	// Start server in goroutine
	go func() {
		addr := cfg.Server.Addr()
		slog.Info("Server starting", "addr", addr)
		if err := app.Listen(addr); err != nil {
			slog.Error("Failed to start server", "error", err)
			os.Exit(1)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("Shutting down server...", "timeout", cfg.Server.ShutdownTimeout)

	// Create shutdown context with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer shutdownCancel()

	// Shutdown server with timeout
	shutdownComplete := make(chan struct{})
	go func() {
		if err := app.Shutdown(); err != nil {
			slog.Error("Error shutting down server", "error", err)
		}
		close(shutdownComplete)
	}()

	select {
	case <-shutdownComplete:
		slog.Info("Server shutdown completed gracefully")
	case <-shutdownCtx.Done():
		slog.Warn("Server shutdown timed out, forcing exit")
	}

	if err := redisClient.Close(); err != nil {
		slog.Error("Error closing Redis connection", "error", err)
	}

	if err := database.Close(db); err != nil {
		slog.Error("Error closing database connection", "error", err)
	}

	slog.Info("Server shutdown complete")
}

func customErrorHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError
	message := "an internal error occurred"

	// Only expose error details for Fiber errors (safe, intentional errors)
	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
		message = e.Message
	} else {
		// Log the actual error for debugging but don't expose to client
		slog.Error("Internal server error",
			"error", err,
			"path", c.Path(),
			"method", c.Method(),
		)
	}

	return c.Status(code).JSON(fiber.Map{
		"success": false,
		"error":   message,
	})
}
