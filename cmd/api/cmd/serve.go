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
	"github.com/voidmaindev/go-template/internal/api"
	"github.com/voidmaindev/go-template/internal/app"
	"github.com/voidmaindev/go-template/internal/common"
	"github.com/voidmaindev/go-template/internal/config"
	"github.com/voidmaindev/go-template/internal/container"
	"github.com/voidmaindev/go-template/internal/database"
	"github.com/voidmaindev/go-template/internal/database/seeders"
	"github.com/voidmaindev/go-template/internal/docs"
	"github.com/voidmaindev/go-template/internal/domain/rbac"
	"github.com/voidmaindev/go-template/internal/health"
	"github.com/voidmaindev/go-template/internal/logger"
	"github.com/voidmaindev/go-template/internal/middleware"
	"github.com/voidmaindev/go-template/internal/redis"
	"github.com/voidmaindev/go-template/internal/telemetry"
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

	// Initialize pagination defaults from config
	common.InitPagination(cfg.Pagination.DefaultPageSize, cfg.Pagination.MaxPageSize)

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
	db, err := database.ConnectWithRetry(&cfg.Database, cfg.App.IsProduction(), cfg.Database.RetryAttempts, cfg.Database.RetryDelay)
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

	// Initialize rate limiter factory (Redis-based distributed rate limiting)
	rateLimiterFactory := middleware.NewRateLimiterFactory(redisClient, &cfg.RateLimit)
	middleware.RateLimiterFactoryKey.Set(c, rateLimiterFactory)
	if cfg.RateLimit.Enabled {
		slog.Info("Distributed rate limiting enabled",
			"auth_limit", cfg.RateLimit.AuthLimit,
			"api_read_limit", cfg.RateLimit.APIReadLimit,
			"api_write_limit", cfg.RateLimit.APIWriteLimit,
			"window_seconds", cfg.RateLimit.WindowSeconds,
		)
	}

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
	// This must happen before seeders because RBAC seeders need the casbin_rule table
	// which is created by the Casbin enforcer during registration
	c.RegisterAll()

	// Run pending seeders (ensures self_registered role exists for new user registration)
	slog.Info("Running pending seeders...")
	seederManager := seeders.DefaultSeederManager(db, cfg)
	if err := seederManager.Run(context.Background()); err != nil {
		slog.Error("Failed to run seeders", "error", err)
		os.Exit(1)
	}

	// Reload RBAC policies after seeding (policies were seeded after enforcer init)
	if enforcer, ok := rbac.EnforcerKey.Get(c); ok {
		if err := enforcer.LoadPolicy(); err != nil {
			slog.Error("Failed to reload RBAC policies", "error", err)
			os.Exit(1)
		}
		slog.Info("RBAC policies reloaded after seeding")
	}

	// Initialize Fiber app
	fiberApp := fiber.New(fiber.Config{
		AppName:               cfg.App.Name,
		ReadTimeout:           cfg.Server.ReadTimeout,
		WriteTimeout:          cfg.Server.WriteTimeout,
		IdleTimeout:           cfg.Server.IdleTimeout,
		DisableStartupMessage: !cfg.App.Debug,
		ErrorHandler:          customErrorHandler,
		// Security: Configure trusted proxy settings for accurate IP detection
		// In production behind a reverse proxy, enable proxy checking
		EnableTrustedProxyCheck: cfg.App.IsProduction(),
		// Use X-Forwarded-For header for client IP when behind proxy
		ProxyHeader: fiber.HeaderXForwardedFor,
		// Trust private network proxies (adjust for your infrastructure)
		TrustedProxies: []string{"10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16", "127.0.0.1"},
	})

	// Setup global middleware
	fiberApp.Use(middleware.RequestIDMiddleware()) // Add request ID first for tracing
	fiberApp.Use(middleware.SecurityHeaders())     // Add security headers
	middleware.SetupCORS(fiberApp, cfg)
	middleware.SetupSlogLogger(fiberApp)
	middleware.SetupCustomRecovery(fiberApp, cfg.App.IsDevelopment())

	// Setup health check routes (/healthz, /readyz, /health)
	healthChecker := health.DefaultHealthChecker(db, redisClient.Client, cfg.Telemetry.ServiceVersion)
	health.SetupHealthRoutes(fiberApp, healthChecker)

	// Setup Prometheus metrics endpoint
	fiberApp.Get("/metrics", telemetry.PrometheusHandler())

	// OpenAPI documentation endpoints
	fiberApp.Get("/docs", docs.ScalarHandler("/openapi.json"))
	fiberApp.Get("/openapi.json", func(ctx *fiber.Ctx) error {
		spec, err := api.GetSwagger()
		if err != nil {
			return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "failed to load OpenAPI spec",
			})
		}
		return ctx.JSON(spec)
	})

	// API v1 routes
	apiGroup := fiberApp.Group("/api/v1")

	// Register all domain routes (existing handlers)
	c.RegisterRoutes(apiGroup)

	// 404 handler
	fiberApp.Use(func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"error":   "endpoint not found",
		})
	})

	// Start server in goroutine with error channel for startup failures
	serverErr := make(chan error, 1)
	go func() {
		addr := cfg.Server.Addr()
		slog.Info("Server starting", "addr", addr)
		if err := fiberApp.Listen(addr); err != nil {
			serverErr <- err
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Wait for either shutdown signal or server error
	select {
	case err := <-serverErr:
		slog.Error("Server failed to start", "error", err)
		// Cleanup connections before exiting
		if err := redisClient.Close(); err != nil {
			slog.Error("Error closing Redis connection", "error", err)
		}
		if err := database.Close(db); err != nil {
			slog.Error("Error closing database connection", "error", err)
		}
		os.Exit(1)
	case <-quit:
		// Continue to graceful shutdown
	}

	slog.Info("Shutting down server...", "timeout", cfg.Server.ShutdownTimeout)

	// Create shutdown context with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer shutdownCancel()

	// Shutdown server with timeout
	shutdownComplete := make(chan struct{})
	go func() {
		if err := fiberApp.Shutdown(); err != nil {
			slog.Error("Error shutting down server", "error", err)
		}
		close(shutdownComplete)
	}()

	select {
	case <-shutdownComplete:
		slog.Info("Server shutdown completed gracefully")
	case <-shutdownCtx.Done():
		slog.Warn("Server shutdown timed out, forcing exit")
		// Force kill entire process including orphaned goroutines
		os.Exit(1)
	}

	// Shutdown domain services (LIFO order) before closing infrastructure
	if err := c.Shutdown(shutdownCtx); err != nil {
		slog.Error("Error shutting down domains", "error", err)
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
