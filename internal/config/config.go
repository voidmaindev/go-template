package config

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config holds all configuration for the application
type Config struct {
	App              AppConfig              `mapstructure:"app"`
	Server           ServerConfig           `mapstructure:"server"`
	Database         DatabaseConfig         `mapstructure:"database"`
	Redis            RedisConfig            `mapstructure:"redis"`
	JWT              JWTConfig              `mapstructure:"jwt"`
	CORS             CORSConfig             `mapstructure:"cors"`
	Telemetry        TelemetryConfig        `mapstructure:"telemetry"`
	Seed             SeedConfig             `mapstructure:"seed"`
	RBAC             RBACConfig             `mapstructure:"rbac"`
	RateLimit        RateLimitConfig        `mapstructure:"ratelimit"`
	Pagination       PaginationConfig       `mapstructure:"pagination"`
	SelfRegistration SelfRegistrationConfig `mapstructure:"self_registration"`
	Email            EmailConfig            `mapstructure:"email"`
	OAuth            OAuthConfig            `mapstructure:"oauth"`
	Security         SecurityConfig         `mapstructure:"security"`
	Sentry           SentryConfig           `mapstructure:"sentry"`
}

// AppConfig holds application-level configuration
type AppConfig struct {
	Name        string `mapstructure:"name"`
	Environment string `mapstructure:"environment"`
	Debug       bool   `mapstructure:"debug"`
}

// ServerConfig holds HTTP server configuration
type ServerConfig struct {
	Host            string        `mapstructure:"host"`
	Port            int           `mapstructure:"port"`
	ReadTimeout     time.Duration `mapstructure:"read_timeout"`
	WriteTimeout    time.Duration `mapstructure:"write_timeout"`
	IdleTimeout     time.Duration `mapstructure:"idle_timeout"`
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout"`
	BodyLimit       int           `mapstructure:"body_limit"` // Maximum request body size in bytes (default 4MB)
}

// DatabaseConfig holds PostgreSQL configuration
type DatabaseConfig struct {
	Host               string        `mapstructure:"host"`
	Port               int           `mapstructure:"port"`
	User               string        `mapstructure:"user"`
	Password           string        `mapstructure:"password"`
	DBName             string        `mapstructure:"dbname"`
	SSLMode            string        `mapstructure:"sslmode"`
	MaxIdleConns       int           `mapstructure:"max_idle_conns"`
	MaxOpenConns       int           `mapstructure:"max_open_conns"`
	MaxLifetime        time.Duration `mapstructure:"max_lifetime"`
	SlowQueryThreshold time.Duration `mapstructure:"slow_query_threshold"`
	RetryAttempts      int           `mapstructure:"retry_attempts"`
	RetryDelay         time.Duration `mapstructure:"retry_delay"`
}

// RedisConfig holds Redis configuration
type RedisConfig struct {
	Host          string        `mapstructure:"host"`
	Port          int           `mapstructure:"port"`
	Password      string        `mapstructure:"password"`
	DB            int           `mapstructure:"db"`
	RetryAttempts int           `mapstructure:"retry_attempts"`
	RetryDelay    time.Duration `mapstructure:"retry_delay"`
}

// JWTConfig holds JWT authentication configuration
type JWTConfig struct {
	SecretKey               string        `mapstructure:"secret_key"`
	AccessTokenExpiry       time.Duration `mapstructure:"access_token_expiry"`
	RefreshTokenExpiry      time.Duration `mapstructure:"refresh_token_expiry"`
	Issuer                  string        `mapstructure:"issuer"`
	MinPasswordResponseTime time.Duration `mapstructure:"min_password_response_time"` // Timing attack protection delay
}

// CORSConfig holds CORS configuration
type CORSConfig struct {
	AllowedOrigins string `mapstructure:"allowed_origins"`
}

// TelemetryConfig holds telemetry configuration
type TelemetryConfig struct {
	Enabled        bool    `mapstructure:"enabled"`
	ServiceName    string  `mapstructure:"service_name"`
	ServiceVersion string  `mapstructure:"service_version"`
	OTLPEndpoint   string  `mapstructure:"otlp_endpoint"`
	OTLPInsecure   bool    `mapstructure:"otlp_insecure"`
	SamplingRatio  float64 `mapstructure:"sampling_ratio"`
}

// SeedConfig holds database seeding configuration
type SeedConfig struct {
	AdminEmail    string `mapstructure:"admin_email"`
	AdminPassword string `mapstructure:"admin_password"`
	AdminName     string `mapstructure:"admin_name"`
}

// RBACConfig holds RBAC configuration
type RBACConfig struct {
	ModelPath string `mapstructure:"model_path"`
}

// RateLimitConfig holds distributed rate limiting configuration
type RateLimitConfig struct {
	Enabled        bool `mapstructure:"enabled"`
	AuthLimit      int  `mapstructure:"auth_limit"`       // Public auth endpoints (login, register)
	AuthUserLimit  int  `mapstructure:"auth_user_limit"`  // Authenticated auth ops (logout, password)
	RBACAdminLimit int  `mapstructure:"rbac_admin_limit"` // RBAC admin operations
	APIWriteLimit  int  `mapstructure:"api_write_limit"`  // POST, PUT, DELETE operations
	APIReadLimit   int  `mapstructure:"api_read_limit"`   // GET operations
	GlobalLimit    int  `mapstructure:"global_limit"`     // Fallback catch-all
	WindowSeconds  int  `mapstructure:"window_seconds"`   // Rate limit window in seconds
}

// PaginationConfig holds pagination defaults
type PaginationConfig struct {
	DefaultPageSize int `mapstructure:"default_page_size"` // Default items per page
	MaxPageSize     int `mapstructure:"max_page_size"`     // Maximum allowed page size
}

// SelfRegistrationConfig holds self-registration settings
type SelfRegistrationConfig struct {
	Enabled                  bool          `mapstructure:"enabled"`
	RequireEmailVerification bool          `mapstructure:"require_email_verification"`
	VerificationTokenExpiry  time.Duration `mapstructure:"verification_token_expiry"`
	PasswordResetTokenExpiry time.Duration `mapstructure:"password_reset_token_expiry"`
	DefaultRole              string        `mapstructure:"default_role"`
	BaseURL                  string        `mapstructure:"base_url"` // Frontend URL for email links
}

// EmailConfig holds email provider settings
type EmailConfig struct {
	From     string         `mapstructure:"from"`
	FromName string         `mapstructure:"from_name"`
	SendGrid SendGridConfig `mapstructure:"sendgrid"`
}

// SendGridConfig holds SendGrid-specific settings
type SendGridConfig struct {
	APIKey string `mapstructure:"api_key"`
}

// OAuthConfig holds OAuth provider settings
type OAuthConfig struct {
	Google   OAuthProviderConfig `mapstructure:"google"`
	Facebook OAuthProviderConfig `mapstructure:"facebook"`
	Apple    OAuthProviderConfig `mapstructure:"apple"`
}

// OAuthProviderConfig holds settings for a single OAuth provider
type OAuthProviderConfig struct {
	Enabled      bool   `mapstructure:"enabled"`
	ClientID     string `mapstructure:"client_id"`
	ClientSecret string `mapstructure:"client_secret"`
	RedirectURL  string `mapstructure:"redirect_url"`
}

// SentryConfig holds Sentry error-tracking configuration.
// Empty DSN with Enabled=true is a no-op (warns at startup) so a freshly
// cloned template runs without requiring a DSN.
type SentryConfig struct {
	Enabled          bool    `mapstructure:"enabled"`
	DSN              string  `mapstructure:"dsn"`
	Environment      string  `mapstructure:"environment"`
	Release          string  `mapstructure:"release"`
	TracesSampleRate float64 `mapstructure:"traces_sample_rate"`
	AttachStacktrace bool    `mapstructure:"attach_stacktrace"`
	Debug            bool    `mapstructure:"debug"`
}

// SecurityConfig holds security-related settings
type SecurityConfig struct {
	LoginRateLimitPerEmail int           `mapstructure:"login_rate_limit_per_email"` // Max failed login attempts per email
	LoginRateLimitPerIP    int           `mapstructure:"login_rate_limit_per_ip"`    // Max failed login attempts per IP
	LoginLockoutDuration   time.Duration `mapstructure:"login_lockout_duration"`     // Lockout duration after max attempts
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
	if err := bindEnvVars(); err != nil {
		return nil, fmt.Errorf("failed to bind environment variables: %w", err)
	}

	// Unmarshal configuration into struct
	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Sentry environment defaults to app environment when unset
	if cfg.Sentry.Environment == "" {
		cfg.Sentry.Environment = cfg.App.Environment
	}
	// BUILD_SHA env var (set via Dockerfile ldflags or CI) overrides the literal "dev"
	if sha := os.Getenv("BUILD_SHA"); sha != "" {
		cfg.Sentry.Release = sha
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

	// Database SSL validation in production
	if c.App.IsProduction() && c.Database.SSLMode == "disable" {
		return errors.New("database SSL cannot be disabled in production")
	}

	// Seed password validation in production
	if c.App.IsProduction() && c.Seed.AdminPassword == "" {
		return errors.New("SEED_ADMIN_PASSWORD is required in production")
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
	viper.SetDefault("server.shutdown_timeout", 30*time.Second)
	viper.SetDefault("server.body_limit", 4*1024*1024) // 4MB default

	// Database defaults
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 5432)
	viper.SetDefault("database.user", "postgres")
	viper.SetDefault("database.password", "postgres")
	viper.SetDefault("database.dbname", "go-template")
	viper.SetDefault("database.sslmode", "require")
	viper.SetDefault("database.max_idle_conns", 10)
	viper.SetDefault("database.max_open_conns", 100)
	viper.SetDefault("database.max_lifetime", time.Hour)
	viper.SetDefault("database.slow_query_threshold", 200*time.Millisecond)
	viper.SetDefault("database.retry_attempts", 5)
	viper.SetDefault("database.retry_delay", 5*time.Second)

	// Redis defaults
	viper.SetDefault("redis.host", "localhost")
	viper.SetDefault("redis.port", 6379)
	viper.SetDefault("redis.password", "")
	viper.SetDefault("redis.db", 0)
	viper.SetDefault("redis.retry_attempts", 5)
	viper.SetDefault("redis.retry_delay", 5*time.Second)

	// JWT defaults
	viper.SetDefault("jwt.secret_key", "your-super-secret-key-change-in-production-min-32-chars")
	viper.SetDefault("jwt.access_token_expiry", 15*time.Minute)
	viper.SetDefault("jwt.refresh_token_expiry", 7*24*time.Hour)
	viper.SetDefault("jwt.issuer", "go-template")
	viper.SetDefault("jwt.min_password_response_time", 200*time.Millisecond)

	// CORS defaults
	viper.SetDefault("cors.allowed_origins", "http://localhost:3000,http://localhost:5173")

	// Telemetry defaults
	viper.SetDefault("telemetry.enabled", false)
	viper.SetDefault("telemetry.service_name", "go-template")
	viper.SetDefault("telemetry.service_version", "1.0.0")
	viper.SetDefault("telemetry.otlp_endpoint", "localhost:4318")
	viper.SetDefault("telemetry.otlp_insecure", true)
	viper.SetDefault("telemetry.sampling_ratio", 1.0)

	// Sentry defaults — Enabled=true with empty DSN is a deliberate no-op
	viper.SetDefault("sentry.enabled", true)
	viper.SetDefault("sentry.dsn", "")
	viper.SetDefault("sentry.environment", "") // resolved from app.environment after unmarshal
	viper.SetDefault("sentry.release", "dev")
	viper.SetDefault("sentry.traces_sample_rate", 0.0) // errors only; performance stays in OTel
	viper.SetDefault("sentry.attach_stacktrace", true)
	viper.SetDefault("sentry.debug", false)

	// Seed defaults (password intentionally empty - must be set in production)
	viper.SetDefault("seed.admin_email", "admin@admin.com")
	viper.SetDefault("seed.admin_password", "") // Empty by default, required in production
	viper.SetDefault("seed.admin_name", "Administrator")

	// RBAC defaults
	viper.SetDefault("rbac.model_path", "config/rbac_model.conf")

	// Rate limit defaults (sliding window algorithm)
	viper.SetDefault("ratelimit.enabled", true)
	viper.SetDefault("ratelimit.auth_limit", 5)        // 5 requests/min for login, register
	viper.SetDefault("ratelimit.auth_user_limit", 10)  // 10 requests/min for logout, password change
	viper.SetDefault("ratelimit.rbac_admin_limit", 30) // 30 requests/min for RBAC management
	viper.SetDefault("ratelimit.api_write_limit", 60)  // 60 requests/min for POST, PUT, DELETE
	viper.SetDefault("ratelimit.api_read_limit", 200)  // 200 requests/min for GET
	viper.SetDefault("ratelimit.global_limit", 1000)   // 1000 requests/min fallback
	viper.SetDefault("ratelimit.window_seconds", 60)   // 1 minute window

	// Pagination defaults
	viper.SetDefault("pagination.default_page_size", 10) // Default items per page
	viper.SetDefault("pagination.max_page_size", 100)    // Maximum allowed page size

	// Self-registration defaults
	viper.SetDefault("self_registration.enabled", false)
	viper.SetDefault("self_registration.require_email_verification", true)
	viper.SetDefault("self_registration.verification_token_expiry", 24*time.Hour)
	viper.SetDefault("self_registration.password_reset_token_expiry", 1*time.Hour)
	viper.SetDefault("self_registration.default_role", "self_registered")
	viper.SetDefault("self_registration.base_url", "http://localhost:5173")

	// Email provider defaults (SendGrid only)
	viper.SetDefault("email.from", "noreply@example.com")
	viper.SetDefault("email.from_name", "Go Template")
	viper.SetDefault("email.sendgrid.api_key", "")

	// Security defaults
	viper.SetDefault("security.login_rate_limit_per_email", 5)      // 5 failed attempts per email
	viper.SetDefault("security.login_rate_limit_per_ip", 20)        // 20 failed attempts per IP
	viper.SetDefault("security.login_lockout_duration", 15*time.Minute) // 15 minute lockout

	// OAuth defaults (all disabled by default)
	viper.SetDefault("oauth.google.enabled", false)
	viper.SetDefault("oauth.google.client_id", "")
	viper.SetDefault("oauth.google.client_secret", "")
	viper.SetDefault("oauth.google.redirect_url", "http://localhost:3000/api/auth/oauth/google/callback")

	viper.SetDefault("oauth.facebook.enabled", false)
	viper.SetDefault("oauth.facebook.client_id", "")
	viper.SetDefault("oauth.facebook.client_secret", "")
	viper.SetDefault("oauth.facebook.redirect_url", "http://localhost:3000/api/auth/oauth/facebook/callback")

	viper.SetDefault("oauth.apple.enabled", false)
	viper.SetDefault("oauth.apple.client_id", "")
	viper.SetDefault("oauth.apple.client_secret", "")
	viper.SetDefault("oauth.apple.redirect_url", "http://localhost:3000/api/auth/oauth/apple/callback")
}

// mustBindEnv binds an environment variable to a viper key, collecting any errors.
func mustBindEnv(errs *[]error, key, envVar string) {
	if err := viper.BindEnv(key, envVar); err != nil {
		*errs = append(*errs, fmt.Errorf("failed to bind %s to %s: %w", envVar, key, err))
	}
}

// bindEnvVars binds environment variables to viper keys.
// This allows using environment variables like APP_DATABASE_HOST instead of config file.
// Returns an error if any binding fails.
func bindEnvVars() error {
	var errs []error

	// App
	mustBindEnv(&errs, "app.name", "APP_NAME")
	mustBindEnv(&errs, "app.environment", "APP_ENV")
	mustBindEnv(&errs, "app.debug", "APP_DEBUG")

	// Server
	mustBindEnv(&errs, "server.host", "SERVER_HOST")
	mustBindEnv(&errs, "server.port", "SERVER_PORT")
	mustBindEnv(&errs, "server.read_timeout", "SERVER_READ_TIMEOUT")
	mustBindEnv(&errs, "server.write_timeout", "SERVER_WRITE_TIMEOUT")
	mustBindEnv(&errs, "server.idle_timeout", "SERVER_IDLE_TIMEOUT")
	mustBindEnv(&errs, "server.shutdown_timeout", "SERVER_SHUTDOWN_TIMEOUT")
	mustBindEnv(&errs, "server.body_limit", "SERVER_BODY_LIMIT")

	// Database
	mustBindEnv(&errs, "database.host", "DB_HOST")
	mustBindEnv(&errs, "database.port", "DB_PORT")
	mustBindEnv(&errs, "database.user", "DB_USER")
	mustBindEnv(&errs, "database.password", "DB_PASSWORD")
	mustBindEnv(&errs, "database.dbname", "DB_NAME")
	mustBindEnv(&errs, "database.sslmode", "DB_SSL_MODE")
	mustBindEnv(&errs, "database.max_idle_conns", "DB_MAX_IDLE_CONNS")
	mustBindEnv(&errs, "database.max_open_conns", "DB_MAX_OPEN_CONNS")
	mustBindEnv(&errs, "database.max_lifetime", "DB_MAX_LIFETIME")
	mustBindEnv(&errs, "database.slow_query_threshold", "DB_SLOW_QUERY_THRESHOLD")
	mustBindEnv(&errs, "database.retry_attempts", "DB_RETRY_ATTEMPTS")
	mustBindEnv(&errs, "database.retry_delay", "DB_RETRY_DELAY")

	// Redis
	mustBindEnv(&errs, "redis.host", "REDIS_HOST")
	mustBindEnv(&errs, "redis.port", "REDIS_PORT")
	mustBindEnv(&errs, "redis.password", "REDIS_PASSWORD")
	mustBindEnv(&errs, "redis.db", "REDIS_DB")
	mustBindEnv(&errs, "redis.retry_attempts", "REDIS_RETRY_ATTEMPTS")
	mustBindEnv(&errs, "redis.retry_delay", "REDIS_RETRY_DELAY")

	// JWT
	mustBindEnv(&errs, "jwt.secret_key", "JWT_SECRET")
	mustBindEnv(&errs, "jwt.access_token_expiry", "JWT_ACCESS_EXPIRY")
	mustBindEnv(&errs, "jwt.refresh_token_expiry", "JWT_REFRESH_EXPIRY")
	mustBindEnv(&errs, "jwt.issuer", "JWT_ISSUER")
	mustBindEnv(&errs, "jwt.min_password_response_time", "JWT_MIN_PASSWORD_RESPONSE_TIME")

	// CORS
	mustBindEnv(&errs, "cors.allowed_origins", "CORS_ALLOWED_ORIGINS")

	// Telemetry
	mustBindEnv(&errs, "telemetry.enabled", "TELEMETRY_ENABLED")
	mustBindEnv(&errs, "telemetry.service_name", "TELEMETRY_SERVICE_NAME")
	mustBindEnv(&errs, "telemetry.service_version", "TELEMETRY_SERVICE_VERSION")
	mustBindEnv(&errs, "telemetry.otlp_endpoint", "OTLP_ENDPOINT")
	mustBindEnv(&errs, "telemetry.otlp_insecure", "OTLP_INSECURE")
	mustBindEnv(&errs, "telemetry.sampling_ratio", "TELEMETRY_SAMPLING_RATIO")

	// Sentry — bare SENTRY_* names match the SDK convention (sentry-go reads
	// SENTRY_DSN natively); deliberately not under the APP_ prefix.
	mustBindEnv(&errs, "sentry.enabled", "SENTRY_ENABLED")
	mustBindEnv(&errs, "sentry.dsn", "SENTRY_DSN")
	mustBindEnv(&errs, "sentry.environment", "SENTRY_ENVIRONMENT")
	mustBindEnv(&errs, "sentry.release", "SENTRY_RELEASE")
	mustBindEnv(&errs, "sentry.traces_sample_rate", "SENTRY_TRACES_SAMPLE_RATE")
	mustBindEnv(&errs, "sentry.attach_stacktrace", "SENTRY_ATTACH_STACKTRACE")
	mustBindEnv(&errs, "sentry.debug", "SENTRY_DEBUG")

	// Seed
	mustBindEnv(&errs, "seed.admin_email", "SEED_ADMIN_EMAIL")
	mustBindEnv(&errs, "seed.admin_password", "SEED_ADMIN_PASSWORD")
	mustBindEnv(&errs, "seed.admin_name", "SEED_ADMIN_NAME")

	// RBAC
	mustBindEnv(&errs, "rbac.model_path", "RBAC_MODEL_PATH")

	// Rate limit
	mustBindEnv(&errs, "ratelimit.enabled", "RATELIMIT_ENABLED")
	mustBindEnv(&errs, "ratelimit.auth_limit", "RATELIMIT_AUTH_LIMIT")
	mustBindEnv(&errs, "ratelimit.auth_user_limit", "RATELIMIT_AUTH_USER_LIMIT")
	mustBindEnv(&errs, "ratelimit.rbac_admin_limit", "RATELIMIT_RBAC_ADMIN_LIMIT")
	mustBindEnv(&errs, "ratelimit.api_write_limit", "RATELIMIT_API_WRITE_LIMIT")
	mustBindEnv(&errs, "ratelimit.api_read_limit", "RATELIMIT_API_READ_LIMIT")
	mustBindEnv(&errs, "ratelimit.global_limit", "RATELIMIT_GLOBAL_LIMIT")
	mustBindEnv(&errs, "ratelimit.window_seconds", "RATELIMIT_WINDOW_SECONDS")

	// Pagination
	mustBindEnv(&errs, "pagination.default_page_size", "PAGINATION_DEFAULT_PAGE_SIZE")
	mustBindEnv(&errs, "pagination.max_page_size", "PAGINATION_MAX_PAGE_SIZE")

	// Self-registration
	mustBindEnv(&errs, "self_registration.enabled", "SELF_REGISTRATION_ENABLED")
	mustBindEnv(&errs, "self_registration.require_email_verification", "SELF_REGISTRATION_REQUIRE_EMAIL_VERIFICATION")
	mustBindEnv(&errs, "self_registration.verification_token_expiry", "SELF_REGISTRATION_VERIFICATION_TOKEN_EXPIRY")
	mustBindEnv(&errs, "self_registration.password_reset_token_expiry", "SELF_REGISTRATION_PASSWORD_RESET_TOKEN_EXPIRY")
	mustBindEnv(&errs, "self_registration.default_role", "SELF_REGISTRATION_DEFAULT_ROLE")
	mustBindEnv(&errs, "self_registration.base_url", "SELF_REGISTRATION_BASE_URL")

	// Email provider (SendGrid only)
	mustBindEnv(&errs, "email.from", "EMAIL_FROM")
	mustBindEnv(&errs, "email.from_name", "EMAIL_FROM_NAME")
	mustBindEnv(&errs, "email.sendgrid.api_key", "SENDGRID_API_KEY")

	// OAuth - Google
	mustBindEnv(&errs, "oauth.google.enabled", "OAUTH_GOOGLE_ENABLED")
	mustBindEnv(&errs, "oauth.google.client_id", "OAUTH_GOOGLE_CLIENT_ID")
	mustBindEnv(&errs, "oauth.google.client_secret", "OAUTH_GOOGLE_CLIENT_SECRET")
	mustBindEnv(&errs, "oauth.google.redirect_url", "OAUTH_GOOGLE_REDIRECT_URL")

	// OAuth - Facebook
	mustBindEnv(&errs, "oauth.facebook.enabled", "OAUTH_FACEBOOK_ENABLED")
	mustBindEnv(&errs, "oauth.facebook.client_id", "OAUTH_FACEBOOK_CLIENT_ID")
	mustBindEnv(&errs, "oauth.facebook.client_secret", "OAUTH_FACEBOOK_CLIENT_SECRET")
	mustBindEnv(&errs, "oauth.facebook.redirect_url", "OAUTH_FACEBOOK_REDIRECT_URL")

	// OAuth - Apple
	mustBindEnv(&errs, "oauth.apple.enabled", "OAUTH_APPLE_ENABLED")
	mustBindEnv(&errs, "oauth.apple.client_id", "OAUTH_APPLE_CLIENT_ID")
	mustBindEnv(&errs, "oauth.apple.client_secret", "OAUTH_APPLE_CLIENT_SECRET")
	mustBindEnv(&errs, "oauth.apple.redirect_url", "OAUTH_APPLE_REDIRECT_URL")

	// Security
	mustBindEnv(&errs, "security.login_rate_limit_per_email", "SECURITY_LOGIN_RATE_LIMIT_PER_EMAIL")
	mustBindEnv(&errs, "security.login_rate_limit_per_ip", "SECURITY_LOGIN_RATE_LIMIT_PER_IP")
	mustBindEnv(&errs, "security.login_lockout_duration", "SECURITY_LOGIN_LOCKOUT_DURATION")

	return errors.Join(errs...)
}

// DSN returns the PostgreSQL connection string.
// Contains the password in plaintext — pass only to the database driver.
// For logs or error messages, use SafeDSN instead.
func (c *DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.DBName, c.SSLMode,
	)
}

// SafeDSN returns the PostgreSQL connection string with the password redacted.
// Safe to include in logs, error messages, and diagnostics.
func (c *DatabaseConfig) SafeDSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=*** dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.DBName, c.SSLMode,
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
