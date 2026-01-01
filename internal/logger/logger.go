package logger

import (
	"io"
	"log/slog"
	"os"
)

// Config holds logger configuration
type Config struct {
	Level       string // debug, info, warn, error
	Format      string // text, json
	AddSource   bool   // add source file info
	Development bool   // development mode (pretty output)
}

// Setup initializes the global slog logger
func Setup(cfg *Config) {
	var level slog.Level
	switch cfg.Level {
	case "debug":
		level = slog.LevelDebug
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level:     level,
		AddSource: cfg.AddSource,
	}

	var handler slog.Handler
	if cfg.Format == "json" {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	slog.SetDefault(slog.New(handler))
}

// SetupFromEnv configures logger based on environment
func SetupFromEnv(environment string, debug bool) {
	cfg := &Config{
		Level:       "info",
		Format:      "json",
		AddSource:   false,
		Development: false,
	}

	if environment == "development" || debug {
		cfg.Level = "debug"
		cfg.Format = "text"
		cfg.Development = true
	}

	Setup(cfg)
}

// NewWithOutput creates a new logger with custom output
func NewWithOutput(w io.Writer, format string) *slog.Logger {
	var handler slog.Handler
	opts := &slog.HandlerOptions{Level: slog.LevelInfo}

	if format == "json" {
		handler = slog.NewJSONHandler(w, opts)
	} else {
		handler = slog.NewTextHandler(w, opts)
	}

	return slog.New(handler)
}
