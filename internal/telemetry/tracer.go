package telemetry

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

const (
	// TracerName is the default tracer name for the application.
	TracerName = "github.com/voidmaindev/go-template"
)

// Tracer returns the application tracer.
func Tracer() trace.Tracer {
	return otel.Tracer(TracerName)
}

// StartSpan starts a new span with the given name.
func StartSpan(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	return Tracer().Start(ctx, name, opts...)
}

// SpanFromContext returns the current span from context.
func SpanFromContext(ctx context.Context) trace.Span {
	return trace.SpanFromContext(ctx)
}

// SetSpanError records an error on the span and sets the status to error.
func SetSpanError(span trace.Span, err error) {
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
}

// SetSpanOK sets the span status to OK.
func SetSpanOK(span trace.Span) {
	span.SetStatus(codes.Ok, "")
}

// AddSpanEvent adds an event to the current span.
func AddSpanEvent(ctx context.Context, name string, attrs ...attribute.KeyValue) {
	span := SpanFromContext(ctx)
	span.AddEvent(name, trace.WithAttributes(attrs...))
}

// SetSpanAttributes sets attributes on the current span.
func SetSpanAttributes(ctx context.Context, attrs ...attribute.KeyValue) {
	span := SpanFromContext(ctx)
	span.SetAttributes(attrs...)
}

// TraceID returns the trace ID from context as a string.
func TraceID(ctx context.Context) string {
	span := SpanFromContext(ctx)
	if span == nil {
		return ""
	}
	traceID := span.SpanContext().TraceID()
	if !traceID.IsValid() {
		return ""
	}
	return traceID.String()
}

// SpanID returns the span ID from context as a string.
func SpanID(ctx context.Context) string {
	span := SpanFromContext(ctx)
	if span == nil {
		return ""
	}
	spanID := span.SpanContext().SpanID()
	if !spanID.IsValid() {
		return ""
	}
	return spanID.String()
}

// WithSpan executes the function within a new span.
func WithSpan[T any](ctx context.Context, name string, fn func(ctx context.Context) (T, error)) (T, error) {
	ctx, span := StartSpan(ctx, name)
	defer span.End()

	result, err := fn(ctx)
	if err != nil {
		SetSpanError(span, err)
	} else {
		SetSpanOK(span)
	}

	return result, err
}

// WithSpanVoid executes a void function within a new span.
func WithSpanVoid(ctx context.Context, name string, fn func(ctx context.Context) error) error {
	ctx, span := StartSpan(ctx, name)
	defer span.End()

	err := fn(ctx)
	if err != nil {
		SetSpanError(span, err)
	} else {
		SetSpanOK(span)
	}

	return err
}

// Common attribute keys for consistency
var (
	AttrUserID      = attribute.Key("user.id")
	AttrUserEmail   = attribute.Key("user.email")
	AttrRequestID   = attribute.Key("request.id")
	AttrHTTPMethod  = attribute.Key("http.method")
	AttrHTTPPath    = attribute.Key("http.path")
	AttrHTTPStatus  = attribute.Key("http.status_code")
	AttrDBOperation = attribute.Key("db.operation")
	AttrDBTable     = attribute.Key("db.table")
	AttrDBRowCount  = attribute.Key("db.row_count")
	AttrErrorType   = attribute.Key("error.type")
	AttrErrorMsg    = attribute.Key("error.message")
)
