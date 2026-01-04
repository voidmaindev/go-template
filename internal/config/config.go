package config

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config holds all configuration for the application
type Config struct {
	App      AppConfig      `mapstructure:"app"`
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Redis    RedisConfig    `mapstructure:"redis"`
	JWT      JWTConfig      `mapstructure:"jwt"`
	CORS     CORSConfig     `mapstructure:"cors"`
}

// AppConfig holds application-level configuration
type AppConfig struct {
	Name        string `mapstructure:"name"`
	Environment string `mapstructure:"environment"`
	Debug       bool   `mapstructure:"debug"`
}

// ServerConfig holds HTTP server configuration
type ServerConfig struct {
	Host         string        `mapstructure:"host"`
	Port         int           `mapstructure:"port"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
	IdleTimeout  time.Duration `mapstructure:"idle_timeout"`
}

// DatabaseConfig holds PostgreSQL configuration
type DatabaseConfig struct {
	Host         string        `mapstructure:"host"`
	Port         int           `mapstructure:"port"`
	User         string        `mapstructure:"user"`
	Password     string        `mapstructure:"password"`
	DBName       string        `mapstructure:"dbname"`
	SSLMode      string        `mapstructure:"sslmode"`
	MaxIdleConns int           `mapstructure:"max_idle_conns"`
	MaxOpenConns int           `mapstructure:"max_open_conns"`
	MaxLifetime  time.Duration `mapstructure:"max_lifetime"`
}

// RedisConfig holds Redis configuration
type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

// JWTConfig holds JWT authentication configuration
type JWTConfig struct {
	SecretKey          string        `mapstructure:"secret_key"`
	AccessTokenExpiry  time.Duration `mapstructure:"access_token_expiry"`
	RefreshTokenExpiry time.Duration `mapstructure:"refresh_token_expiry"`
	Issuer             string        `mapstructure:"issuer"`
}

// CORSConfig holds CORS configuration
type CORSConfig struct {
	AllowedOrigins string `mapstructure:"allowed_origins"`
}

// Load loads configuration from config file and environment variables
func Load() (*Config, error) {
	// Set defaults
	setDefaults()

	// Enable environment variable reading
	viper.SetEnvPrefix("APP")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	// Bind environment variables to config keys
	bindEnvVars()

	// Unmarshal configuration into struct
	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &cfg, nil
}

// Validate checks that the configuration is valid
func (c *Config) Validate() error {
	// JWT secret validation
	if c.App.IsProduction() {
		if c.JWT.SecretKey == "" {
			return errors.New("JWT_SECRET is required in production")
		}
		if c.JWT.SecretKey == "your-super-secret-key-change-in-production-min-32-chars" {
			return errors.New("JWT_SECRET must be changed from the default value in production")
		}
	}

	// Ensure minimum JWT secret length
	if c.JWT.SecretKey != "" && len(c.JWT.SecretKey) < 32 {
		return errors.New("JWT_SECRET must be at least 32 characters")
	}

	// Validate database config
	if c.Database.Host == "" {
		return errors.New("database host is required")
	}

	return nil
}

// setDefaults sets default values for all configuration options
func setDefaults() {
	// App defaults
	viper.SetDefault("app.name", "go-template")
	viper.SetDefault("app.environment", "development")
	viper.SetDefault("app.debug", true)

	// Server defaults
	viper.SetDefault("server.host", "0.0.0.0")
	viper.SetDefault("server.port", 3000)
	viper.SetDefault("server.read_timeout", 10*time.Second)
	viper.SetDefault("server.write_timeout", 10*time.Second)
	viper.SetDefault("server.idle_timeout", 120*time.Second)

	// Database defaults
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 5432)
	viper.SetDefault("database.user", "postgres")
	viper.SetDefault("database.password", "postgres")
	viper.SetDefault("database.dbname", "go-template")
	viper.SetDefault("database.sslmode", "disable")
	viper.SetDefault("database.max_idle_conns", 10)
	viper.SetDefault("database.max_open_conns", 100)
	viper.SetDefault("database.max_lifetime", time.Hour)

	// Redis defaults
	viper.SetDefault("redis.host", "localhost")
	viper.SetDefault("redis.port", 6379)
	viper.SetDefault("redis.password", "")
	viper.SetDefault("redis.db", 0)

	// JWT defaults
	viper.SetDefault("jwt.secret_key", "your-super-secret-key-change-in-production-min-32-chars")
	viper.SetDefault("jwt.access_token_expiry", 15*time.Minute)
	viper.SetDefault("jwt.refresh_token_expiry", 7*24*time.Hour)
	viper.SetDefault("jwt.issuer", "go-template")

	// CORS defaults
	viper.SetDefault("cors.allowed_origins", "http://localhost:3000,http://localhost:5173")
}

// bindEnvVars binds environment variables to viper keys
// This allows using environment variables like APP_DATABASE_HOST instead of config file
func bindEnvVars() {
	// App
	viper.BindEnv("app.name", "APP_NAME")
	viper.BindEnv("app.environment", "APP_ENV")
	viper.BindEnv("app.debug", "APP_DEBUG")

	// Server
	viper.BindEnv("server.host", "SERVER_HOST")
	viper.BindEnv("server.port", "SERVER_PORT")
	viper.BindEnv("server.read_timeout", "SERVER_READ_TIMEOUT")
	viper.BindEnv("server.write_timeout", "SERVER_WRITE_TIMEOUT")
	viper.BindEnv("server.idle_timeout", "SERVER_IDLE_TIMEOUT")

	// Database
	viper.BindEnv("database.host", "DB_HOST")
	viper.BindEnv("database.port", "DB_PORT")
	viper.BindEnv("database.user", "DB_USER")
	viper.BindEnv("database.password", "DB_PASSWORD")
	viper.BindEnv("database.dbname", "DB_NAME")
	viper.BindEnv("database.sslmode", "DB_SSL_MODE")
	viper.BindEnv("database.max_idle_conns", "DB_MAX_IDLE_CONNS")
	viper.BindEnv("database.max_open_conns", "DB_MAX_OPEN_CONNS")
	viper.BindEnv("database.max_lifetime", "DB_MAX_LIFETIME")

	// Redis
	viper.BindEnv("redis.host", "REDIS_HOST")
	viper.BindEnv("redis.port", "REDIS_PORT")
	viper.BindEnv("redis.password", "REDIS_PASSWORD")
	viper.BindEnv("redis.db", "REDIS_DB")

	// JWT
	viper.BindEnv("jwt.secret_key", "JWT_SECRET")
	viper.BindEnv("jwt.access_token_expiry", "JWT_ACCESS_EXPIRY")
	viper.BindEnv("jwt.refresh_token_expiry", "JWT_REFRESH_EXPIRY")
	viper.BindEnv("jwt.issuer", "JWT_ISSUER")

	// CORS
	viper.BindEnv("cors.allowed_origins", "CORS_ALLOWED_ORIGINS")
}

// DSN returns the PostgreSQL connection string
func (c *DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.DBName, c.SSLMode,
	)
}

// Addr returns the Redis address
func (c *RedisConfig) Addr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// Addr returns the server address
func (c *ServerConfig) Addr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// IsDevelopment returns true if the app is in development mode
func (c *AppConfig) IsDevelopment() bool {
	return c.Environment == "development"
}

// IsProduction returns true if the app is in production mode
func (c *AppConfig) IsProduction() bool {
	return c.Environment == "production"
}
