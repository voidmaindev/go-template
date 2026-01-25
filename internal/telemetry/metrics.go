package telemetry

import (
	"sync"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Metrics holds all application metrics.
type Metrics struct {
	// HTTP metrics
	HTTPRequestsTotal    *prometheus.CounterVec
	HTTPRequestDuration  *prometheus.HistogramVec
	HTTPRequestsInFlight prometheus.Gauge

	// Database metrics
	DBQueriesTotal    *prometheus.CounterVec
	DBQueryDuration   *prometheus.HistogramVec
	DBConnectionsOpen prometheus.Gauge

	// Redis metrics
	RedisOperationsTotal   *prometheus.CounterVec
	RedisOperationDuration *prometheus.HistogramVec

	// Business metrics
	UsersRegistered  prometheus.Counter
	UsersLoggedIn    prometheus.Counter
	UsersUpdated     prometheus.Counter
	UsersDeleted     prometheus.Counter
	DocumentsCreated prometheus.Counter

	// Security metrics
	AuthFailures   *prometheus.CounterVec
	RateLimitHits  *prometheus.CounterVec
}

// NewMetrics creates and registers all application metrics.
func NewMetrics(namespace string) *Metrics {
	if namespace == "" {
		namespace = "go_template"
	}

	return &Metrics{
		// HTTP metrics
		HTTPRequestsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "http_requests_total",
				Help:      "Total number of HTTP requests",
			},
			[]string{"method", "path", "status"},
		),
		HTTPRequestDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "http_request_duration_seconds",
				Help:      "HTTP request duration in seconds",
				Buckets:   []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
			},
			[]string{"method", "path", "status"},
		),
		HTTPRequestsInFlight: promauto.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "http_requests_in_flight",
				Help:      "Current number of HTTP requests being processed",
			},
		),

		// Database metrics
		DBQueriesTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "db_queries_total",
				Help:      "Total number of database queries",
			},
			[]string{"operation", "table", "status"},
		),
		DBQueryDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "db_query_duration_seconds",
				Help:      "Database query duration in seconds",
				Buckets:   []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5},
			},
			[]string{"operation", "table"},
		),
		DBConnectionsOpen: promauto.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "db_connections_open",
				Help:      "Current number of open database connections",
			},
		),

		// Redis metrics
		RedisOperationsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "redis_operations_total",
				Help:      "Total number of Redis operations",
			},
			[]string{"operation", "status"},
		),
		RedisOperationDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "redis_operation_duration_seconds",
				Help:      "Redis operation duration in seconds",
				Buckets:   []float64{.0001, .0005, .001, .005, .01, .025, .05, .1},
			},
			[]string{"operation"},
		),

		// Business metrics
		UsersRegistered: promauto.NewCounter(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "users_registered_total",
				Help:      "Total number of users registered",
			},
		),
		UsersLoggedIn: promauto.NewCounter(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "users_logged_in_total",
				Help:      "Total number of user logins",
			},
		),
		UsersUpdated: promauto.NewCounter(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "users_updated_total",
				Help:      "Total number of user profile updates",
			},
		),
		UsersDeleted: promauto.NewCounter(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "users_deleted_total",
				Help:      "Total number of users deleted",
			},
		),
		DocumentsCreated: promauto.NewCounter(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "documents_created_total",
				Help:      "Total number of documents created",
			},
		),

		// Security metrics
		AuthFailures: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "auth_failures_total",
				Help:      "Total number of authentication failures",
			},
			[]string{"reason"},
		),
		RateLimitHits: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "rate_limit_hits_total",
				Help:      "Total number of rate limit hits",
			},
			[]string{"tier"},
		),
	}
}

// PrometheusHandler returns a Fiber handler for the /metrics endpoint.
func PrometheusHandler() fiber.Handler {
	return adaptor.HTTPHandler(promhttp.Handler())
}

// DefaultMetrics is the global metrics instance.
var (
	DefaultMetrics *Metrics
	metricsOnce    sync.Once
)

// InitMetrics initializes the default metrics (safe to call multiple times).
func InitMetrics(namespace string) {
	metricsOnce.Do(func() {
		DefaultMetrics = NewMetrics(namespace)
	})
}

// GetMetrics returns the default metrics instance.
func GetMetrics() *Metrics {
	InitMetrics("")
	return DefaultMetrics
}

// RecordHTTPRequest records an HTTP request metric.
func RecordHTTPRequest(method, path, status string, duration float64) {
	m := GetMetrics()
	m.HTTPRequestsTotal.WithLabelValues(method, path, status).Inc()
	m.HTTPRequestDuration.WithLabelValues(method, path, status).Observe(duration)
}

// RecordDBQuery records a database query metric.
func RecordDBQuery(operation, table, status string, duration float64) {
	m := GetMetrics()
	m.DBQueriesTotal.WithLabelValues(operation, table, status).Inc()
	m.DBQueryDuration.WithLabelValues(operation, table).Observe(duration)
}

// RecordRedisOperation records a Redis operation metric.
func RecordRedisOperation(operation, status string, duration float64) {
	m := GetMetrics()
	m.RedisOperationsTotal.WithLabelValues(operation, status).Inc()
	m.RedisOperationDuration.WithLabelValues(operation).Observe(duration)
}

// IncrementUsersRegistered increments the users registered counter.
func IncrementUsersRegistered() {
	GetMetrics().UsersRegistered.Inc()
}

// IncrementUsersLoggedIn increments the users logged in counter.
func IncrementUsersLoggedIn() {
	GetMetrics().UsersLoggedIn.Inc()
}

// IncrementDocumentsCreated increments the documents created counter.
func IncrementDocumentsCreated() {
	GetMetrics().DocumentsCreated.Inc()
}

// IncrementUsersUpdated increments the users updated counter.
func IncrementUsersUpdated() {
	GetMetrics().UsersUpdated.Inc()
}

// IncrementUsersDeleted increments the users deleted counter.
func IncrementUsersDeleted() {
	GetMetrics().UsersDeleted.Inc()
}

// IncrementAuthFailures increments the auth failures counter with reason label.
func IncrementAuthFailures(reason string) {
	GetMetrics().AuthFailures.WithLabelValues(reason).Inc()
}

// IncrementRateLimitHits increments the rate limit hits counter with tier label.
func IncrementRateLimitHits(tier string) {
	GetMetrics().RateLimitHits.WithLabelValues(tier).Inc()
}
