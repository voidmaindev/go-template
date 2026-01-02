package config

import (
	"strings"
	"testing"
	"time"
)

func TestConfig_Validate_Development(t *testing.T) {
	tests := []struct {
		name      string
		config    Config
		wantError bool
		errorMsg  string
	}{
		{
			name: "valid development config with default secret",
			config: Config{
				App: AppConfig{
					Environment: "development",
				},
				JWT: JWTConfig{
					SecretKey: "your-super-secret-key-change-in-production-min-32-chars",
				},
				Database: DatabaseConfig{
					Host: "localhost",
				},
			},
			wantError: false,
		},
		{
			name: "valid development config with custom secret",
			config: Config{
				App: AppConfig{
					Environment: "development",
				},
				JWT: JWTConfig{
					SecretKey: "my-custom-secret-key-at-least-32-characters",
				},
				Database: DatabaseConfig{
					Host: "localhost",
				},
			},
			wantError: false,
		},
		{
			name: "development with short secret",
			config: Config{
				App: AppConfig{
					Environment: "development",
				},
				JWT: JWTConfig{
					SecretKey: "short",
				},
				Database: DatabaseConfig{
					Host: "localhost",
				},
			},
			wantError: true,
			errorMsg:  "at least 32 characters",
		},
		{
			name: "missing database host",
			config: Config{
				App: AppConfig{
					Environment: "development",
				},
				JWT: JWTConfig{
					SecretKey: "my-custom-secret-key-at-least-32-characters",
				},
				Database: DatabaseConfig{
					Host: "",
				},
			},
			wantError: true,
			errorMsg:  "database host is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()

			if tt.wantError {
				if err == nil {
					t.Error("Validate() expected error, got nil")
				} else if tt.errorMsg != "" && !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Validate() error = %v, want error containing %q", err, tt.errorMsg)
				}
			} else {
				if err != nil {
					t.Errorf("Validate() unexpected error = %v", err)
				}
			}
		})
	}
}

func TestConfig_Validate_Production(t *testing.T) {
	tests := []struct {
		name      string
		config    Config
		wantError bool
		errorMsg  string
	}{
		{
			name: "valid production config",
			config: Config{
				App: AppConfig{
					Environment: "production",
				},
				JWT: JWTConfig{
					SecretKey: "my-production-secret-key-at-least-32-chars",
				},
				Database: DatabaseConfig{
					Host: "db.example.com",
				},
			},
			wantError: false,
		},
		{
			name: "production with empty secret",
			config: Config{
				App: AppConfig{
					Environment: "production",
				},
				JWT: JWTConfig{
					SecretKey: "",
				},
				Database: DatabaseConfig{
					Host: "db.example.com",
				},
			},
			wantError: true,
			errorMsg:  "JWT_SECRET is required in production",
		},
		{
			name: "production with default secret",
			config: Config{
				App: AppConfig{
					Environment: "production",
				},
				JWT: JWTConfig{
					SecretKey: "your-super-secret-key-change-in-production-min-32-chars",
				},
				Database: DatabaseConfig{
					Host: "db.example.com",
				},
			},
			wantError: true,
			errorMsg:  "must be changed from the default",
		},
		{
			name: "production with short secret",
			config: Config{
				App: AppConfig{
					Environment: "production",
				},
				JWT: JWTConfig{
					SecretKey: "short-secret",
				},
				Database: DatabaseConfig{
					Host: "db.example.com",
				},
			},
			wantError: true,
			errorMsg:  "at least 32 characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()

			if tt.wantError {
				if err == nil {
					t.Error("Validate() expected error, got nil")
				} else if tt.errorMsg != "" && !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Validate() error = %v, want error containing %q", err, tt.errorMsg)
				}
			} else {
				if err != nil {
					t.Errorf("Validate() unexpected error = %v", err)
				}
			}
		})
	}
}

func TestAppConfig_IsDevelopment(t *testing.T) {
	tests := []struct {
		env      string
		expected bool
	}{
		{"development", true},
		{"production", false},
		{"staging", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.env, func(t *testing.T) {
			cfg := &AppConfig{Environment: tt.env}
			if got := cfg.IsDevelopment(); got != tt.expected {
				t.Errorf("IsDevelopment() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestAppConfig_IsProduction(t *testing.T) {
	tests := []struct {
		env      string
		expected bool
	}{
		{"production", true},
		{"development", false},
		{"staging", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.env, func(t *testing.T) {
			cfg := &AppConfig{Environment: tt.env}
			if got := cfg.IsProduction(); got != tt.expected {
				t.Errorf("IsProduction() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestDatabaseConfig_DSN(t *testing.T) {
	cfg := &DatabaseConfig{
		Host:     "localhost",
		Port:     5432,
		User:     "postgres",
		Password: "secret",
		DBName:   "testdb",
		SSLMode:  "disable",
	}

	dsn := cfg.DSN()

	if !strings.Contains(dsn, "host=localhost") {
		t.Error("DSN should contain host")
	}
	if !strings.Contains(dsn, "port=5432") {
		t.Error("DSN should contain port")
	}
	if !strings.Contains(dsn, "user=postgres") {
		t.Error("DSN should contain user")
	}
	if !strings.Contains(dsn, "password=secret") {
		t.Error("DSN should contain password")
	}
	if !strings.Contains(dsn, "dbname=testdb") {
		t.Error("DSN should contain dbname")
	}
	if !strings.Contains(dsn, "sslmode=disable") {
		t.Error("DSN should contain sslmode")
	}
}

func TestRedisConfig_Addr(t *testing.T) {
	cfg := &RedisConfig{
		Host: "localhost",
		Port: 6379,
	}

	addr := cfg.Addr()
	expected := "localhost:6379"

	if addr != expected {
		t.Errorf("Addr() = %v, want %v", addr, expected)
	}
}

func TestServerConfig_Addr(t *testing.T) {
	cfg := &ServerConfig{
		Host: "0.0.0.0",
		Port: 3000,
	}

	addr := cfg.Addr()
	expected := "0.0.0.0:3000"

	if addr != expected {
		t.Errorf("Addr() = %v, want %v", addr, expected)
	}
}

func TestJWTConfig_Defaults(t *testing.T) {
	// Test that default values are reasonable
	cfg := &JWTConfig{
		SecretKey:          "test-secret-key-at-least-32-characters",
		AccessTokenExpiry:  15 * time.Minute,
		RefreshTokenExpiry: 7 * 24 * time.Hour,
		Issuer:             "test",
	}

	if cfg.AccessTokenExpiry <= 0 {
		t.Error("AccessTokenExpiry should be positive")
	}
	if cfg.RefreshTokenExpiry <= cfg.AccessTokenExpiry {
		t.Error("RefreshTokenExpiry should be greater than AccessTokenExpiry")
	}
}
