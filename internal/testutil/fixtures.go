package testutil

import (
	"context"
	"testing"
	"time"

	"github.com/voidmaindev/go-template/internal/common"
	"github.com/voidmaindev/go-template/internal/config"
	"github.com/voidmaindev/go-template/pkg/utils"
	"gorm.io/gorm"
)

// FixtureBuilder helps build test fixtures with fluent API.
type FixtureBuilder struct {
	t  *testing.T
	db *gorm.DB
}

// NewFixtureBuilder creates a new fixture builder.
func NewFixtureBuilder(t *testing.T, db *gorm.DB) *FixtureBuilder {
	t.Helper()
	return &FixtureBuilder{t: t, db: db}
}

// UserFixture represents test user data.
type UserFixture struct {
	ID        uint
	Email     string
	Password  string // Plain text password for testing
	Name      string
	Role      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// DefaultUserFixture returns a default user fixture.
func DefaultUserFixture() UserFixture {
	return UserFixture{
		Email:    "test@example.com",
		Password: "password123",
		Name:     "Test User",
		Role:     "user",
	}
}

// AdminUserFixture returns an admin user fixture.
func AdminUserFixture() UserFixture {
	return UserFixture{
		Email:    "admin@example.com",
		Password: "adminpass123",
		Name:     "Admin User",
		Role:     "admin",
	}
}

// ItemFixture represents test item data.
type ItemFixture struct {
	ID          uint
	Name        string
	Description string
	Price       int64 // Price in cents (e.g., 9999 = $99.99)
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// DefaultItemFixture returns a default item fixture.
func DefaultItemFixture() ItemFixture {
	return ItemFixture{
		Name:        "Test Item",
		Description: "A test item for testing",
		Price:       9999, // $99.99 in cents
	}
}

// CountryFixture represents test country data.
type CountryFixture struct {
	ID        uint
	Name      string
	Code      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// DefaultCountryFixture returns a default country fixture.
func DefaultCountryFixture() CountryFixture {
	return CountryFixture{
		Name: "Test Country",
		Code: "TC",
	}
}

// CityFixture represents test city data.
type CityFixture struct {
	ID        uint
	Name      string
	CountryID uint
	CreatedAt time.Time
	UpdatedAt time.Time
}

// DefaultCityFixture returns a default city fixture.
func DefaultCityFixture(countryID uint) CityFixture {
	return CityFixture{
		Name:      "Test City",
		CountryID: countryID,
	}
}

// DocumentFixture represents test document data.
type DocumentFixture struct {
	ID        uint
	Title     string
	CityID    uint
	CreatedAt time.Time
	UpdatedAt time.Time
}

// DefaultDocumentFixture returns a default document fixture.
func DefaultDocumentFixture(cityID uint) DocumentFixture {
	return DocumentFixture{
		Title:  "Test Document",
		CityID: cityID,
	}
}

// TestJWTConfig returns a JWT config suitable for testing.
func TestJWTConfig() *config.JWTConfig {
	return &config.JWTConfig{
		SecretKey:          "test-secret-key-at-least-32-characters-long!!",
		AccessTokenExpiry:  15 * time.Minute,
		RefreshTokenExpiry: 7 * 24 * time.Hour,
		Issuer:             "test-issuer",
	}
}

// TestConfig returns a full config suitable for testing.
func TestConfig() *config.Config {
	return &config.Config{
		App: config.AppConfig{
			Name:        "go-template-test",
			Environment: "test",
			Debug:       true,
		},
		Server: config.ServerConfig{
			Host:            "localhost",
			Port:            3000,
			ReadTimeout:     10 * time.Second,
			WriteTimeout:    10 * time.Second,
			IdleTimeout:     120 * time.Second,
			ShutdownTimeout: 30 * time.Second,
		},
		Database: config.DatabaseConfig{
			Host:               "localhost",
			Port:               5432,
			User:               "testuser",
			Password:           "testpass",
			DBName:             "testdb",
			SSLMode:            "disable",
			MaxIdleConns:       5,
			MaxOpenConns:       10,
			MaxLifetime:        time.Hour,
			SlowQueryThreshold: 200 * time.Millisecond,
			RetryAttempts:      3,
			RetryDelay:         time.Second,
		},
		Redis: config.RedisConfig{
			Host:          "localhost",
			Port:          6379,
			Password:      "",
			DB:            0,
			RetryAttempts: 3,
			RetryDelay:    time.Second,
		},
		JWT: *TestJWTConfig(),
		CORS: config.CORSConfig{
			AllowedOrigins: "*",
		},
	}
}

// HashPassword is a test helper that hashes a password.
func HashPassword(t *testing.T, password string) string {
	t.Helper()
	hash, err := utils.HashPassword(password)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}
	return hash
}

// GenerateTestAccessToken generates an access token for testing.
func GenerateTestAccessToken(t *testing.T, userID uint, email string) string {
	t.Helper()
	cfg := TestJWTConfig()
	jwtCfg := &utils.JWTConfig{
		SecretKey:          cfg.SecretKey,
		AccessTokenExpiry:  cfg.AccessTokenExpiry,
		RefreshTokenExpiry: cfg.RefreshTokenExpiry,
		Issuer:             cfg.Issuer,
	}
	token, err := utils.GenerateAccessToken(userID, email, jwtCfg)
	if err != nil {
		t.Fatalf("failed to generate access token: %v", err)
	}
	return token
}

// GenerateTestRefreshToken generates a refresh token for testing.
func GenerateTestRefreshToken(t *testing.T, userID uint, email string) string {
	t.Helper()
	cfg := TestJWTConfig()
	jwtCfg := &utils.JWTConfig{
		SecretKey:          cfg.SecretKey,
		AccessTokenExpiry:  cfg.AccessTokenExpiry,
		RefreshTokenExpiry: cfg.RefreshTokenExpiry,
		Issuer:             cfg.Issuer,
	}
	token, err := utils.GenerateRefreshToken(userID, email, jwtCfg)
	if err != nil {
		t.Fatalf("failed to generate refresh token: %v", err)
	}
	return token
}

// MockBaseModel returns a common.BaseModel with test values.
func MockBaseModel(id uint) common.BaseModel {
	now := time.Now()
	return common.BaseModel{
		ID:        id,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// CreateTestPagination creates pagination params for testing.
func CreateTestPagination(page, pageSize int) *common.Pagination {
	return &common.Pagination{
		Page:     page,
		PageSize: pageSize,
	}
}

// SeedTestData is a helper to seed test data using raw SQL or GORM.
type SeedTestData struct {
	t  *testing.T
	db *gorm.DB
}

// NewSeedTestData creates a new test data seeder.
func NewSeedTestData(t *testing.T, db *gorm.DB) *SeedTestData {
	return &SeedTestData{t: t, db: db}
}

// Exec executes raw SQL for seeding.
func (s *SeedTestData) Exec(sql string, values ...any) *SeedTestData {
	s.t.Helper()
	if err := s.db.Exec(sql, values...).Error; err != nil {
		s.t.Fatalf("failed to execute seed SQL: %v", err)
	}
	return s
}

// Create creates a record using GORM.
func (s *SeedTestData) Create(value any) *SeedTestData {
	s.t.Helper()
	if err := s.db.Create(value).Error; err != nil {
		s.t.Fatalf("failed to create seed record: %v", err)
	}
	return s
}

// WaitForCondition waits for a condition to be true with timeout.
func WaitForCondition(t *testing.T, timeout time.Duration, condition func() bool) bool {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if condition() {
			return true
		}
		time.Sleep(100 * time.Millisecond)
	}
	return false
}

// ContextWithTestValue adds a test-specific value to context.
func ContextWithTestValue(ctx context.Context, key, value any) context.Context {
	return context.WithValue(ctx, key, value)
}
