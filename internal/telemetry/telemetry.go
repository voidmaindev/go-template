// Package telemetry provides OpenTelemetry tracing and metrics infrastructure.
package telemetry

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
)

// Config holds telemetry configuration.
type Config struct {
	Enabled       bool    `mapstructure:"enabled"`
	ServiceName   string  `mapstructure:"service_name"`
	ServiceVersion string  `mapstructure:"service_version"`
	Environment   string  `mapstructure:"environment"`
	OTLPEndpoint  string  `mapstructure:"otlp_endpoint"`
	OTLPInsecure  bool    `mapstructure:"otlp_insecure"`
	SamplingRatio float64 `mapstructure:"sampling_ratio"`
}

// DefaultConfig returns the default telemetry configuration.
func DefaultConfig() Config {
	return Config{
		Enabled:       false,
		ServiceName:   "go-template",
		ServiceVersion: "1.0.0",
		Environment:   "development",
		OTLPEndpoint:  "localhost:4318",
		OTLPInsecure:  true,
		SamplingRatio: 1.0,
	}
}

// Telemetry holds the OpenTelemetry components.
type Telemetry struct {
	config         Config
	tracerProvider *sdktrace.TracerProvider
	shutdown       func(context.Context) error
}

// New creates a new Telemetry instance.
func New(cfg Config) (*Telemetry, error) {
	if !cfg.Enabled {
		slog.Info("telemetry disabled")
		return &Telemetry{
			config:   cfg,
			shutdown: func(ctx context.Context) error { return nil },
		}, nil
	}

	ctx := context.Background()

	// Create resource
	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(cfg.ServiceName),
			semconv.ServiceVersion(cfg.ServiceVersion),
			attribute.String("environment", cfg.Environment),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Create OTLP exporter
	opts := []otlptracehttp.Option{
		otlptracehttp.WithEndpoint(cfg.OTLPEndpoint),
	}
	if cfg.OTLPInsecure {
		opts = append(opts, otlptracehttp.WithInsecure())
	}

	exporter, err := otlptracehttp.New(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create OTLP exporter: %w", err)
	}

	// Create sampler
	var sampler sdktrace.Sampler
	if cfg.SamplingRatio >= 1.0 {
		sampler = sdktrace.AlwaysSample()
	} else if cfg.SamplingRatio <= 0 {
		sampler = sdktrace.NeverSample()
	} else {
		sampler = sdktrace.TraceIDRatioBased(cfg.SamplingRatio)
	}

	// Create tracer provider
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sampler),
	)

	// Set global tracer provider
	otel.SetTracerProvider(tp)

	// Set global propagator
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	slog.Info("telemetry initialized",
		"service", cfg.ServiceName,
		"endpoint", cfg.OTLPEndpoint,
		"sampling_ratio", cfg.SamplingRatio)

	return &Telemetry{
		config:         cfg,
		tracerProvider: tp,
		shutdown: func(ctx context.Context) error {
			return tp.Shutdown(ctx)
		},
	}, nil
}

// Shutdown gracefully shuts down the telemetry components.
func (t *Telemetry) Shutdown(ctx context.Context) error {
	if t.shutdown == nil {
		return nil
	}

	shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return t.shutdown(shutdownCtx)
}

// TracerProvider returns the tracer provider.
func (t *Telemetry) TracerProvider() *sdktrace.TracerProvider {
	return t.tracerProvider
}

// IsEnabled returns whether telemetry is enabled.
func (t *Telemetry) IsEnabled() bool {
	return t.config.Enabled
}

// Setup initializes telemetry and returns a shutdown function.
// This is a convenience function for simple setup.
func Setup(cfg Config) (shutdown func(context.Context) error, err error) {
	t, err := New(cfg)
	if err != nil {
		return nil, err
	}
	return t.Shutdown, nil
}

// ShutdownAll shuts down multiple telemetry components.
func ShutdownAll(ctx context.Context, shutdowns ...func(context.Context) error) error {
	var errs []error
	for _, shutdown := range shutdowns {
		if shutdown != nil {
			if err := shutdown(ctx); err != nil {
				errs = append(errs, err)
			}
		}
	}
	return errors.Join(errs...)
}
