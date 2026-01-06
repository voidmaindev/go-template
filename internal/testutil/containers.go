package testutil

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/modules/redis"
	"github.com/testcontainers/testcontainers-go/wait"
	gormpg "gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	goredis "github.com/redis/go-redis/v9"
	internalredis "github.com/voidmaindev/go-template/internal/redis"
)

// PostgresContainer wraps a testcontainers PostgreSQL instance.
type PostgresContainer struct {
	*postgres.PostgresContainer
	DSN string
}

// RedisContainer wraps a testcontainers Redis instance.
type RedisContainer struct {
	*redis.RedisContainer
	Addr string
}

// TestContainers holds all test containers for integration tests.
type TestContainers struct {
	Postgres    *PostgresContainer
	Redis       *RedisContainer
	DB          *gorm.DB
	RedisClient *goredis.Client
}

// StartPostgres starts a PostgreSQL container for testing.
func StartPostgres(ctx context.Context, t *testing.T) *PostgresContainer {
	t.Helper()

	dbName := "testdb"
	dbUser := "testuser"
	dbPassword := "testpass"

	container, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase(dbName),
		postgres.WithUsername(dbUser),
		postgres.WithPassword(dbPassword),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(60*time.Second),
		),
	)
	if err != nil {
		t.Fatalf("failed to start postgres container: %v", err)
	}

	t.Cleanup(func() {
		if err := container.Terminate(context.Background()); err != nil {
			t.Logf("failed to terminate postgres container: %v", err)
		}
	})

	host, err := container.Host(ctx)
	if err != nil {
		t.Fatalf("failed to get postgres host: %v", err)
	}

	port, err := container.MappedPort(ctx, "5432")
	if err != nil {
		t.Fatalf("failed to get postgres port: %v", err)
	}

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port.Port(), dbUser, dbPassword, dbName)

	return &PostgresContainer{
		PostgresContainer: container,
		DSN:               dsn,
	}
}

// StartRedis starts a Redis container for testing.
func StartRedis(ctx context.Context, t *testing.T) *RedisContainer {
	t.Helper()

	container, err := redis.Run(ctx,
		"redis:7-alpine",
		testcontainers.WithWaitStrategy(
			wait.ForLog("Ready to accept connections").
				WithStartupTimeout(30*time.Second),
		),
	)
	if err != nil {
		t.Fatalf("failed to start redis container: %v", err)
	}

	t.Cleanup(func() {
		if err := container.Terminate(context.Background()); err != nil {
			t.Logf("failed to terminate redis container: %v", err)
		}
	})

	host, err := container.Host(ctx)
	if err != nil {
		t.Fatalf("failed to get redis host: %v", err)
	}

	port, err := container.MappedPort(ctx, "6379")
	if err != nil {
		t.Fatalf("failed to get redis port: %v", err)
	}

	addr := fmt.Sprintf("%s:%s", host, port.Port())

	return &RedisContainer{
		RedisContainer: container,
		Addr:           addr,
	}
}

// NewTestDB creates a GORM database connection from a PostgresContainer.
func NewTestDB(t *testing.T, dsn string) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(gormpg.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("failed to connect to test database: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("failed to get sql.DB: %v", err)
	}

	t.Cleanup(func() {
		if err := sqlDB.Close(); err != nil {
			t.Logf("failed to close test database: %v", err)
		}
	})

	return db
}

// NewTestRedis creates a Redis client from a RedisContainer.
func NewTestRedis(t *testing.T, addr string) *goredis.Client {
	t.Helper()

	client := goredis.NewClient(&goredis.Options{
		Addr: addr,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		t.Fatalf("failed to ping test redis: %v", err)
	}

	t.Cleanup(func() {
		if err := client.Close(); err != nil {
			t.Logf("failed to close test redis: %v", err)
		}
	})

	return client
}

// SetupIntegrationTest starts all required containers and returns connections.
// Use this in integration tests that need both database and Redis.
// Returns the internal redis.Client wrapper for use with container.New().
func SetupIntegrationTest(t *testing.T) (*gorm.DB, *internalredis.Client) {
	t.Helper()

	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx := TestContext(t)

	pgContainer := StartPostgres(ctx, t)
	redisContainer := StartRedis(ctx, t)

	db := NewTestDB(t, pgContainer.DSN)
	redisClient := NewTestRedisWrapped(t, redisContainer.Addr)

	return db, redisClient
}

// NewTestRedisWrapped creates an internal redis.Client wrapper from a Redis address.
// Use this when you need the wrapper type for container.New().
func NewTestRedisWrapped(t *testing.T, addr string) *internalredis.Client {
	t.Helper()

	client := goredis.NewClient(&goredis.Options{
		Addr: addr,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		t.Fatalf("failed to ping test redis: %v", err)
	}

	t.Cleanup(func() {
		if err := client.Close(); err != nil {
			t.Logf("failed to close test redis: %v", err)
		}
	})

	return internalredis.WrapClient(client)
}

// SetupPostgresOnly starts only PostgreSQL container.
// Use this for tests that don't need Redis.
func SetupPostgresOnly(t *testing.T) *gorm.DB {
	t.Helper()

	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx := TestContext(t)
	pgContainer := StartPostgres(ctx, t)
	return NewTestDB(t, pgContainer.DSN)
}

// SetupRedisOnly starts only Redis container.
// Use this for tests that don't need PostgreSQL.
func SetupRedisOnly(t *testing.T) *goredis.Client {
	t.Helper()

	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx := TestContext(t)
	redisContainer := StartRedis(ctx, t)
	return NewTestRedis(t, redisContainer.Addr)
}

// TruncateTable truncates a table in the test database.
func TruncateTable(t *testing.T, db *gorm.DB, tableName string) {
	t.Helper()
	if err := db.Exec(fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", tableName)).Error; err != nil {
		t.Fatalf("failed to truncate table %s: %v", tableName, err)
	}
}

// TruncateTables truncates multiple tables in the test database.
func TruncateTables(t *testing.T, db *gorm.DB, tableNames ...string) {
	t.Helper()
	for _, name := range tableNames {
		TruncateTable(t, db, name)
	}
}
