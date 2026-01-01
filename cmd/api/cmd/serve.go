package cmd

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/spf13/cobra"
	"github.com/voidmaindev/GoTemplate/internal/config"
	"github.com/voidmaindev/GoTemplate/internal/database"
	"github.com/voidmaindev/GoTemplate/internal/logger"
	"github.com/voidmaindev/GoTemplate/internal/middleware"
	"github.com/voidmaindev/GoTemplate/internal/redis"

	"github.com/voidmaindev/GoTemplate/internal/domain/city"
	"github.com/voidmaindev/GoTemplate/internal/domain/country"
	"github.com/voidmaindev/GoTemplate/internal/domain/document"
	"github.com/voidmaindev/GoTemplate/internal/domain/item"
	"github.com/voidmaindev/GoTemplate/internal/domain/user"
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the API server",
	Long:  `Start the HTTP API server with all configured routes and middleware.`,
	Run:   runServe,
}

func init() {
	rootCmd.AddCommand(serveCmd)

	// Local flags for serve command
	serveCmd.Flags().IntP("port", "p", 3000, "Port to run the server on")
	serveCmd.Flags().StringP("host", "H", "0.0.0.0", "Host to bind the server to")
}

func runServe(cmd *cobra.Command, args []string) {
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
		"app", cfg.App.Name,
		"env", cfg.App.Environment,
		"debug", cfg.App.Debug,
	)

	// Connect to database
	slog.Info("Connecting to database...")
	db, err := database.ConnectWithRetry(&cfg.Database, 5, 5*time.Second)
	if err != nil {
		slog.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}

	// Connect to Redis
	slog.Info("Connecting to Redis...")
	redisClient, err := redis.ConnectWithRetry(&cfg.Redis, 5, 5*time.Second)
	if err != nil {
		slog.Error("Failed to connect to Redis", "error", err)
		os.Exit(1)
	}

	// Run database migrations
	slog.Info("Running migrations...")
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

	// Health check endpoint
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "healthy",
			"service": cfg.App.Name,
			"env":     cfg.App.Environment,
		})
	})

	// API v1 routes
	api := app.Group("/api/v1")

	// Initialize token store
	tokenStore := user.NewTokenStore(redisClient)

	// Initialize repositories
	userRepo := user.NewRepository(db)
	itemRepo := item.NewRepository(db)
	countryRepo := country.NewRepository(db)
	cityRepo := city.NewRepository(db)
	documentRepo := document.NewRepository(db)
	documentItemRepo := document.NewItemRepository(db)

	// Initialize services
	userService := user.NewService(userRepo, tokenStore, &cfg.JWT)
	itemService := item.NewService(itemRepo)
	countryService := country.NewService(countryRepo)
	cityService := city.NewService(cityRepo, countryRepo)
	documentService := document.NewService(documentRepo, documentItemRepo, cityRepo, itemRepo)

	// Initialize handlers
	userHandler := user.NewHandler(userService)
	itemHandler := item.NewHandler(itemService)
	countryHandler := country.NewHandler(countryService)
	cityHandler := city.NewHandler(cityService)
	documentHandler := document.NewHandler(documentService)

	// Register routes
	user.RegisterRoutes(api, userHandler, &cfg.JWT, tokenStore)
	item.RegisterRoutes(api, itemHandler, &cfg.JWT, tokenStore)
	country.RegisterRoutes(api, countryHandler, &cfg.JWT, tokenStore)
	city.RegisterRoutes(api, cityHandler, &cfg.JWT, tokenStore)
	document.RegisterRoutes(api, documentHandler, &cfg.JWT, tokenStore)

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

	slog.Info("Shutting down server...")

	if err := app.Shutdown(); err != nil {
		slog.Error("Error shutting down server", "error", err)
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
	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
	}
	return c.Status(code).JSON(fiber.Map{
		"success": false,
		"error":   err.Error(),
	})
}
