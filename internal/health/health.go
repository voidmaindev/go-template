// Package health provides health check functionality for Kubernetes probes.
package health

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	goredis "github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// Status represents the health status of a component.
type Status string

const (
	StatusHealthy   Status = "healthy"
	StatusUnhealthy Status = "unhealthy"
	StatusDegraded  Status = "degraded"

	// checkTimeout bounds every individual check. A hung dependency (DB, Redis,
	// custom probe) must not stall the readiness probe — kubelet would otherwise
	// fail to mark the pod unready and traffic keeps flowing.
	checkTimeout = 5 * time.Second
)

// CheckFunc is a function that performs a health check.
type CheckFunc func(ctx context.Context) error

// CheckResult represents the result of a health check.
type CheckResult struct {
	Status    Status        `json:"status"`
	Latency   string        `json:"latency"`
	Details   any           `json:"details,omitempty"`
	Error     string        `json:"error,omitempty"`
	Timestamp time.Time     `json:"timestamp"`
}

// HealthResponse represents the overall health response.
type HealthResponse struct {
	Status    Status                  `json:"status"`
	Checks    map[string]CheckResult  `json:"checks"`
	Version   string                  `json:"version,omitempty"`
	Uptime    string                  `json:"uptime,omitempty"`
	Timestamp time.Time               `json:"timestamp"`
}

// HealthChecker manages health checks for the application.
type HealthChecker struct {
	mu        sync.RWMutex
	checks    map[string]CheckFunc
	version   string
	startTime time.Time
}

// NewHealthChecker creates a new HealthChecker.
func NewHealthChecker(version string) *HealthChecker {
	return &HealthChecker{
		checks:    make(map[string]CheckFunc),
		version:   version,
		startTime: time.Now(),
	}
}

// Register adds a health check.
func (h *HealthChecker) Register(name string, check CheckFunc) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.checks[name] = check
}

// Liveness returns true if the application is alive.
// This is a simple check that the app is running.
func (h *HealthChecker) Liveness(ctx context.Context) bool {
	// Application is alive if this code is running
	return true
}

// Readiness checks if the application is ready to serve traffic.
// This checks all registered dependencies.
func (h *HealthChecker) Readiness(ctx context.Context) HealthResponse {
	h.mu.RLock()
	checks := make(map[string]CheckFunc, len(h.checks))
	for k, v := range h.checks {
		checks[k] = v
	}
	h.mu.RUnlock()

	results := make(map[string]CheckResult)
	overallStatus := StatusHealthy

	var wg sync.WaitGroup
	var resultsMu sync.Mutex

	for name, check := range checks {
		wg.Add(1)
		go func(name string, check CheckFunc) {
			defer wg.Done()

			checkCtx, cancel := context.WithTimeout(ctx, checkTimeout)
			defer cancel()

			start := time.Now()
			err := check(checkCtx)
			latency := time.Since(start)

			result := CheckResult{
				Status:    StatusHealthy,
				Latency:   latency.String(),
				Timestamp: time.Now(),
			}

			if err != nil {
				result.Status = StatusUnhealthy
				result.Error = err.Error()
			}

			resultsMu.Lock()
			results[name] = result
			if result.Status != StatusHealthy && overallStatus == StatusHealthy {
				overallStatus = StatusUnhealthy
			}
			resultsMu.Unlock()
		}(name, check)
	}

	wg.Wait()

	return HealthResponse{
		Status:    overallStatus,
		Checks:    results,
		Version:   h.version,
		Uptime:    time.Since(h.startTime).String(),
		Timestamp: time.Now(),
	}
}

// Health returns the combined health status (backward compatible).
func (h *HealthChecker) Health(ctx context.Context) HealthResponse {
	return h.Readiness(ctx)
}

// RegisterDatabase adds a database health check.
func (h *HealthChecker) RegisterDatabase(db *gorm.DB) {
	h.Register("database", func(ctx context.Context) error {
		sqlDB, err := db.DB()
		if err != nil {
			return err
		}
		return sqlDB.PingContext(ctx)
	})
}

// RegisterRedis adds a Redis health check.
func (h *HealthChecker) RegisterRedis(client *goredis.Client) {
	h.Register("redis", func(ctx context.Context) error {
		return client.Ping(ctx).Err()
	})
}

// RegisterCustom adds a custom health check.
func (h *HealthChecker) RegisterCustom(name string, check CheckFunc) {
	h.Register(name, check)
}

// MemoryCheck returns a check function for memory usage.
// Note: This is a simplified check that monitors Go heap allocation.
// For comprehensive system memory monitoring, consider using cgroups metrics
// or platform-specific APIs.
//
// Usage:
//
//	checker.Register("memory", MemoryCheck(80.0)) // Alert at 80% heap usage
func MemoryCheck(maxHeapMB float64) CheckFunc {
	return func(ctx context.Context) error {
		var m runtime.MemStats
		runtime.ReadMemStats(&m)

		// Convert to MB for readability
		heapAllocMB := float64(m.HeapAlloc) / 1024 / 1024

		// If max is set to 0 or negative, skip the check (monitoring only)
		if maxHeapMB <= 0 {
			return nil
		}

		if heapAllocMB > maxHeapMB {
			return fmt.Errorf("heap allocation %.2f MB exceeds threshold %.2f MB", heapAllocMB, maxHeapMB)
		}
		return nil
	}
}

// SetupHealthRoutes registers health check routes on the Fiber app.
func SetupHealthRoutes(app *fiber.App, checker *HealthChecker) {
	// Liveness probe - is the app alive?
	app.Get("/healthz", func(c *fiber.Ctx) error {
		if checker.Liveness(c.Context()) {
			return c.JSON(fiber.Map{
				"status": "healthy",
			})
		}
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"status": "unhealthy",
		})
	})

	// Readiness probe - is the app ready to serve traffic?
	app.Get("/readyz", func(c *fiber.Ctx) error {
		response := checker.Readiness(c.Context())
		status := fiber.StatusOK
		if response.Status != StatusHealthy {
			status = fiber.StatusServiceUnavailable
		}
		return c.Status(status).JSON(response)
	})

	// Combined health endpoint (backward compatible)
	app.Get("/health", func(c *fiber.Ctx) error {
		response := checker.Health(c.Context())
		status := fiber.StatusOK
		if response.Status != StatusHealthy {
			status = fiber.StatusServiceUnavailable
		}
		return c.Status(status).JSON(response)
	})
}

// DefaultHealthChecker creates a health checker with database and Redis checks.
func DefaultHealthChecker(db *gorm.DB, redisClient *goredis.Client, version string) *HealthChecker {
	checker := NewHealthChecker(version)

	if db != nil {
		checker.RegisterDatabase(db)
	}

	if redisClient != nil {
		checker.RegisterRedis(redisClient)
	}

	return checker
}
