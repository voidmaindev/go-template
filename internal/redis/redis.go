package redis

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/voidmaindev/GoTemplate/internal/config"
)

// Client wraps the Redis client with additional functionality
type Client struct {
	*redis.Client
}

// Connect establishes a connection to Redis
func Connect(cfg *config.RedisConfig) (*Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr(),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	// Ping to verify connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	slog.Info("Redis connection established")
	return &Client{Client: client}, nil
}

// ConnectWithRetry attempts to connect to Redis with retries
func ConnectWithRetry(cfg *config.RedisConfig, maxRetries int, delay time.Duration) (*Client, error) {
	var client *Client
	var err error

	for i := 0; i < maxRetries; i++ {
		client, err = Connect(cfg)
		if err == nil {
			return client, nil
		}

		slog.Warn("Failed to connect to Redis",
			"attempt", i+1,
			"max_retries", maxRetries,
			"error", err,
		)
		if i < maxRetries-1 {
			slog.Info("Retrying Redis connection", "delay", delay)
			time.Sleep(delay)
		}
	}

	return nil, fmt.Errorf("failed to connect to Redis after %d attempts: %w", maxRetries, err)
}

// Close closes the Redis connection
func (c *Client) Close() error {
	if err := c.Client.Close(); err != nil {
		return fmt.Errorf("failed to close redis connection: %w", err)
	}
	slog.Info("Redis connection closed")
	return nil
}

// HealthCheck verifies the Redis connection is healthy
func (c *Client) HealthCheck(ctx context.Context) error {
	if err := c.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("redis ping failed: %w", err)
	}
	return nil
}

// SetWithExpiry sets a key with an expiry time
func (c *Client) SetWithExpiry(ctx context.Context, key string, value interface{}, expiry time.Duration) error {
	return c.Set(ctx, key, value, expiry).Err()
}

// GetString gets a string value
func (c *Client) GetString(ctx context.Context, key string) (string, error) {
	val, err := c.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", nil
	}
	return val, err
}

// Exists checks if a key exists
func (c *Client) Exists(ctx context.Context, key string) (bool, error) {
	result, err := c.Client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return result > 0, nil
}

// DeleteKey deletes a key
func (c *Client) DeleteKey(ctx context.Context, key string) error {
	return c.Del(ctx, key).Err()
}

// SetNX sets a key only if it doesn't exist (useful for distributed locks)
func (c *Client) SetNX(ctx context.Context, key string, value interface{}, expiry time.Duration) (bool, error) {
	return c.Client.SetNX(ctx, key, value, expiry).Result()
}

// IncrementBy increments a key by a value
func (c *Client) IncrementBy(ctx context.Context, key string, value int64) (int64, error) {
	return c.IncrBy(ctx, key, value).Result()
}

// GetTTL returns the remaining time-to-live of a key
func (c *Client) GetTTL(ctx context.Context, key string) (time.Duration, error) {
	return c.TTL(ctx, key).Result()
}
