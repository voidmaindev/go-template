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
	viper.BindEnv("server.shutdown_timeout", "SERVER_SHUTDOWN_TIMEOUT")

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
	viper.BindEnv("database.slow_query_threshold", "DB_SLOW_QUERY_THRESHOLD")
	viper.BindEnv("database.retry_attempts", "DB_RETRY_ATTEMPTS")
	viper.BindEnv("database.retry_delay", "DB_RETRY_DELAY")

	// Redis
	viper.BindEnv("redis.host", "REDIS_HOST")
	viper.BindEnv("redis.port", "REDIS_PORT")
	viper.BindEnv("redis.password", "REDIS_PASSWORD")
	viper.BindEnv("redis.db", "REDIS_DB")
	viper.BindEnv("redis.retry_attempts", "REDIS_RETRY_ATTEMPTS")
	viper.BindEnv("redis.retry_delay", "REDIS_RETRY_DELAY")

	// JWT
	viper.BindEnv("jwt.secret_key", "JWT_SECRET")
	viper.BindEnv("jwt.access_token_expiry", "JWT_ACCESS_EXPIRY")
	viper.BindEnv("jwt.refresh_token_expiry", "JWT_REFRESH_EXPIRY")
	viper.BindEnv("jwt.issuer", "JWT_ISSUER")
	viper.BindEnv("jwt.min_password_response_time", "JWT_MIN_PASSWORD_RESPONSE_TIME")

	// CORS
	viper.BindEnv("cors.allowed_origins", "CORS_ALLOWED_ORIGINS")

	// Telemetry
	viper.BindEnv("telemetry.enabled", "TELEMETRY_ENABLED")
	viper.BindEnv("telemetry.service_name", "TELEMETRY_SERVICE_NAME")
	viper.BindEnv("telemetry.service_version", "TELEMETRY_SERVICE_VERSION")
	viper.BindEnv("telemetry.otlp_endpoint", "OTLP_ENDPOINT")
	viper.BindEnv("telemetry.otlp_insecure", "OTLP_INSECURE")
	viper.BindEnv("telemetry.sampling_ratio", "TELEMETRY_SAMPLING_RATIO")

	// Seed
	viper.BindEnv("seed.admin_email", "SEED_ADMIN_EMAIL")
	viper.BindEnv("seed.admin_password", "SEED_ADMIN_PASSWORD")
	viper.BindEnv("seed.admin_name", "SEED_ADMIN_NAME")

	// RBAC
	viper.BindEnv("rbac.model_path", "RBAC_MODEL_PATH")

	// Rate limit
	viper.BindEnv("ratelimit.enabled", "RATELIMIT_ENABLED")
	viper.BindEnv("ratelimit.auth_limit", "RATELIMIT_AUTH_LIMIT")
	viper.BindEnv("ratelimit.auth_user_limit", "RATELIMIT_AUTH_USER_LIMIT")
	viper.BindEnv("ratelimit.rbac_admin_limit", "RATELIMIT_RBAC_ADMIN_LIMIT")
	viper.BindEnv("ratelimit.api_write_limit", "RATELIMIT_API_WRITE_LIMIT")
	viper.BindEnv("ratelimit.api_read_limit", "RATELIMIT_API_READ_LIMIT")
	viper.BindEnv("ratelimit.global_limit", "RATELIMIT_GLOBAL_LIMIT")
	viper.BindEnv("ratelimit.window_seconds", "RATELIMIT_WINDOW_SECONDS")

	// Pagination
	viper.BindEnv("pagination.default_page_size", "PAGINATION_DEFAULT_PAGE_SIZE")
	viper.BindEnv("pagination.max_page_size", "PAGINATION_MAX_PAGE_SIZE")

	// Self-registration
	viper.BindEnv("self_registration.enabled", "SELF_REGISTRATION_ENABLED")
	viper.BindEnv("self_registration.require_email_verification", "SELF_REGISTRATION_REQUIRE_EMAIL_VERIFICATION")
	viper.BindEnv("self_registration.verification_token_expiry", "SELF_REGISTRATION_VERIFICATION_TOKEN_EXPIRY")
	viper.BindEnv("self_registration.password_reset_token_expiry", "SELF_REGISTRATION_PASSWORD_RESET_TOKEN_EXPIRY")
	viper.BindEnv("self_registration.default_role", "SELF_REGISTRATION_DEFAULT_ROLE")
	viper.BindEnv("self_registration.base_url", "SELF_REGISTRATION_BASE_URL")

	// Email provider (SendGrid only)
	viper.BindEnv("email.from", "EMAIL_FROM")
	viper.BindEnv("email.from_name", "EMAIL_FROM_NAME")
	viper.BindEnv("email.sendgrid.api_key", "SENDGRID_API_KEY")

	// OAuth - Google
	viper.BindEnv("oauth.google.enabled", "OAUTH_GOOGLE_ENABLED")
	viper.BindEnv("oauth.google.client_id", "OAUTH_GOOGLE_CLIENT_ID")
	viper.BindEnv("oauth.google.client_secret", "OAUTH_GOOGLE_CLIENT_SECRET")
	viper.BindEnv("oauth.google.redirect_url", "OAUTH_GOOGLE_REDIRECT_URL")

	// OAuth - Facebook
	viper.BindEnv("oauth.facebook.enabled", "OAUTH_FACEBOOK_ENABLED")
	viper.BindEnv("oauth.facebook.client_id", "OAUTH_FACEBOOK_CLIENT_ID")
	viper.BindEnv("oauth.facebook.client_secret", "OAUTH_FACEBOOK_CLIENT_SECRET")
	viper.BindEnv("oauth.facebook.redirect_url", "OAUTH_FACEBOOK_REDIRECT_URL")

	// OAuth - Apple
	viper.BindEnv("oauth.apple.enabled", "OAUTH_APPLE_ENABLED")
	viper.BindEnv("oauth.apple.client_id", "OAUTH_APPLE_CLIENT_ID")
	viper.BindEnv("oauth.apple.client_secret", "OAUTH_APPLE_CLIENT_SECRET")
	viper.BindEnv("oauth.apple.redirect_url", "OAUTH_APPLE_REDIRECT_URL")

	// Security
	viper.BindEnv("security.login_rate_limit_per_email", "SECURITY_LOGIN_RATE_LIMIT_PER_EMAIL")
	viper.BindEnv("security.login_rate_limit_per_ip", "SECURITY_LOGIN_RATE_LIMIT_PER_IP")
	viper.BindEnv("security.login_lockout_duration", "SECURITY_LOGIN_LOCKOUT_DURATION")
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
