// Package integration provides integration tests for the application.
package integration

import (
	"context"
	"testing"

	"github.com/voidmaindev/go-template/internal/common"
	"github.com/voidmaindev/go-template/internal/config"
	"github.com/voidmaindev/go-template/internal/container"
	"github.com/voidmaindev/go-template/internal/domain/example_city"
	"github.com/voidmaindev/go-template/internal/domain/example_country"
	"github.com/voidmaindev/go-template/internal/domain/example_document"
	"github.com/voidmaindev/go-template/internal/domain/example_item"
	"github.com/voidmaindev/go-template/internal/domain/rbac"
	"github.com/voidmaindev/go-template/internal/domain/user"
	"github.com/voidmaindev/go-template/internal/redis"
	"github.com/voidmaindev/go-template/internal/testutil"
	"github.com/voidmaindev/go-template/pkg/ptr"
	"gorm.io/gorm"
)

// TestSuite holds all dependencies for integration tests.
type TestSuite struct {
	T           *testing.T
	Ctx         context.Context
	DB          *gorm.DB
	RedisClient *redis.Client
	Config      *config.Config
	Container   *container.Container
}

// SetupTestSuite creates a full integration test environment.
func SetupTestSuite(t *testing.T) *TestSuite {
	t.Helper()

	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx := testutil.TestContext(t)
	db, redisClient := testutil.SetupIntegrationTest(t)
	cfg := testutil.TestConfig()

	// Create container with test dependencies
	c := container.New(db, redisClient, cfg)

	// Register all domains (rbac must be before user, as user depends on rbac.Service)
	c.AddDomain(rbac.NewDomain())
	c.AddDomain(user.NewDomain())
	c.AddDomain(example_item.NewDomain())
	c.AddDomain(example_country.NewDomain())
	c.AddDomain(example_city.NewDomain())
	c.AddDomain(example_document.NewDomain())

	// Run migrations
	if err := db.AutoMigrate(c.GetAllModels()...); err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	// Register all domain components
	c.RegisterAll()

	return &TestSuite{
		T:           t,
		Ctx:         ctx,
		DB:          db,
		RedisClient: redisClient,
		Config:      cfg,
		Container:   c,
	}
}

// SetupTestSuiteDBOnly creates a test environment with only PostgreSQL.
func SetupTestSuiteDBOnly(t *testing.T) *TestSuite {
	t.Helper()

	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx := testutil.TestContext(t)
	db := testutil.SetupPostgresOnly(t)
	cfg := testutil.TestConfig()

	return &TestSuite{
		T:      t,
		Ctx:    ctx,
		DB:     db,
		Config: cfg,
	}
}

// CleanupTables truncates all domain tables.
func (s *TestSuite) CleanupTables() {
	s.T.Helper()
	testutil.TruncateTables(s.T, s.DB,
		"document_items",
		"documents",
		"cities",
		"countries",
		"items",
		"users",
	)
}

// UserService returns the user service from the container.
func (s *TestSuite) UserService() user.Service {
	return user.ServiceKey.MustGet(s.Container)
}

// UserRepository returns the user repository from the container.
func (s *TestSuite) UserRepository() user.Repository {
	return user.RepositoryKey.MustGet(s.Container)
}

// ItemService returns the item service from the container.
func (s *TestSuite) ItemService() example_item.Service {
	return example_item.ServiceKey.MustGet(s.Container)
}

// ItemRepository returns the item repository from the container.
func (s *TestSuite) ItemRepository() example_item.Repository {
	return example_item.RepositoryKey.MustGet(s.Container)
}

// CountryService returns the country service from the container.
func (s *TestSuite) CountryService() example_country.Service {
	return example_country.ServiceKey.MustGet(s.Container)
}

// CountryRepository returns the country repository from the container.
func (s *TestSuite) CountryRepository() example_country.Repository {
	return example_country.RepositoryKey.MustGet(s.Container)
}

// CityService returns the city service from the container.
func (s *TestSuite) CityService() example_city.Service {
	return example_city.ServiceKey.MustGet(s.Container)
}

// CityRepository returns the city repository from the container.
func (s *TestSuite) CityRepository() example_city.Repository {
	return example_city.RepositoryKey.MustGet(s.Container)
}

// DocumentService returns the document service from the container.
func (s *TestSuite) DocumentService() example_document.Service {
	return example_document.ServiceKey.MustGet(s.Container)
}

// DocumentRepository returns the document repository from the container.
func (s *TestSuite) DocumentRepository() example_document.Repository {
	return example_document.RepositoryKey.MustGet(s.Container)
}

// CreateTestUser creates a test user and returns it.
func (s *TestSuite) CreateTestUser(email, password, name string) *user.User {
	s.T.Helper()

	hashedPassword := testutil.HashPassword(s.T, password)
	u := &user.User{
		BaseModel: common.BaseModel{},
		Email:     email,
		Password:  ptr.To(hashedPassword),
		Name:      name,
	}

	if err := s.DB.Create(u).Error; err != nil {
		s.T.Fatalf("failed to create test user: %v", err)
	}

	return u
}

// CreateTestAdmin creates a test admin user and returns it.
// Note: The user is created in the database, but admin privileges
// are managed via RBAC role assignment, not a field on the user.
func (s *TestSuite) CreateTestAdmin(email, password, name string) *user.User {
	s.T.Helper()

	hashedPassword := testutil.HashPassword(s.T, password)
	u := &user.User{
		BaseModel: common.BaseModel{},
		Email:     email,
		Password:  ptr.To(hashedPassword),
		Name:      name,
	}

	if err := s.DB.Create(u).Error; err != nil {
		s.T.Fatalf("failed to create test admin: %v", err)
	}

	return u
}

// CreateTestItem creates a test item and returns it.
// Price is in cents (e.g., 1999 = $19.99).
func (s *TestSuite) CreateTestItem(name, description string, price int64) *example_item.Item {
	s.T.Helper()

	i := &example_item.Item{
		BaseModel:   common.BaseModel{},
		Name:        name,
		Description: description,
		Price:       price,
	}

	if err := s.DB.Create(i).Error; err != nil {
		s.T.Fatalf("failed to create test item: %v", err)
	}

	return i
}

// CreateTestCountry creates a test country and returns it.
func (s *TestSuite) CreateTestCountry(name, code string) *example_country.Country {
	s.T.Helper()

	c := &example_country.Country{
		BaseModel: common.BaseModel{},
		Name:      name,
		Code:      code,
	}

	if err := s.DB.Create(c).Error; err != nil {
		s.T.Fatalf("failed to create test country: %v", err)
	}

	return c
}

// CreateTestCity creates a test city and returns it.
func (s *TestSuite) CreateTestCity(name string, countryID uint) *example_city.City {
	s.T.Helper()

	c := &example_city.City{
		BaseModel: common.BaseModel{},
		Name:      name,
		CountryID: countryID,
	}

	if err := s.DB.Create(c).Error; err != nil {
		s.T.Fatalf("failed to create test city: %v", err)
	}

	return c
}
