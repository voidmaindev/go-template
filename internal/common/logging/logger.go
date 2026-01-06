package logging

import (
	"context"
	"log/slog"

	"github.com/voidmaindev/go-template/internal/common/ctxutil"
)

// Logger provides structured logging with context
type Logger struct {
	logger *slog.Logger
	domain string
}

// New creates a new Logger for a domain
func New(domain string) *Logger {
	return &Logger{
		logger: slog.Default(),
		domain: domain,
	}
}

// NewWithLogger creates a Logger with custom slog.Logger
func NewWithLogger(logger *slog.Logger, domain string) *Logger {
	return &Logger{
		logger: logger,
		domain: domain,
	}
}

// WithOperation returns a new logger with operation context
func (l *Logger) WithOperation(operation string) *OperationLogger {
	return &OperationLogger{
		logger:    l.logger,
		domain:    l.domain,
		operation: operation,
	}
}

// Info logs at info level with domain context
func (l *Logger) Info(ctx context.Context, msg string, args ...any) {
	attrs := l.baseAttrs(ctx)
	attrs = append(attrs, args...)
	l.logger.InfoContext(ctx, msg, attrs...)
}

// Warn logs at warn level with domain context
func (l *Logger) Warn(ctx context.Context, msg string, args ...any) {
	attrs := l.baseAttrs(ctx)
	attrs = append(attrs, args...)
	l.logger.WarnContext(ctx, msg, attrs...)
}

// Error logs at error level with domain context
func (l *Logger) Error(ctx context.Context, msg string, err error, args ...any) {
	attrs := l.baseAttrs(ctx)
	if err != nil {
		attrs = append(attrs, "error", err.Error())
	}
	attrs = append(attrs, args...)
	l.logger.ErrorContext(ctx, msg, attrs...)
}

// Debug logs at debug level with domain context
func (l *Logger) Debug(ctx context.Context, msg string, args ...any) {
	attrs := l.baseAttrs(ctx)
	attrs = append(attrs, args...)
	l.logger.DebugContext(ctx, msg, attrs...)
}

// baseAttrs returns common attributes
func (l *Logger) baseAttrs(ctx context.Context) []any {
	attrs := []any{"domain", l.domain}
	attrs = append(attrs, ctxutil.Extract(ctx).ToLogArgs()...)
	return attrs
}

// ================================
// OperationLogger
// ================================

// OperationLogger adds operation context to logs
type OperationLogger struct {
	logger     *slog.Logger
	domain     string
	operation  string
	entityType string
	entityID   any
	attrs      []any
}

// WithEntity adds entity context
func (l *OperationLogger) WithEntity(entityType string, id any) *OperationLogger {
	l.entityType = entityType
	l.entityID = id
	return l
}

// With adds additional attributes
func (l *OperationLogger) With(args ...any) *OperationLogger {
	l.attrs = append(l.attrs, args...)
	return l
}

// baseAttrs returns common attributes
func (l *OperationLogger) baseAttrs(ctx context.Context) []any {
	attrs := []any{
		"domain", l.domain,
		"operation", l.operation,
	}

	// Add context values
	attrs = append(attrs, ctxutil.Extract(ctx).ToLogArgs()...)

	// Add entity context
	if l.entityType != "" {
		attrs = append(attrs, "entity_type", l.entityType)
		if l.entityID != nil {
			attrs = append(attrs, "entity_id", l.entityID)
		}
	}

	// Add custom attributes
	attrs = append(attrs, l.attrs...)

	return attrs
}

// Info logs at info level
func (l *OperationLogger) Info(ctx context.Context, msg string, args ...any) {
	attrs := append(l.baseAttrs(ctx), args...)
	l.logger.InfoContext(ctx, msg, attrs...)
}

// Warn logs at warn level
func (l *OperationLogger) Warn(ctx context.Context, msg string, args ...any) {
	attrs := append(l.baseAttrs(ctx), args...)
	l.logger.WarnContext(ctx, msg, attrs...)
}

// Error logs at error level with error
func (l *OperationLogger) Error(ctx context.Context, msg string, err error, args ...any) {
	attrs := append(l.baseAttrs(ctx), args...)
	if err != nil {
		attrs = append(attrs, "error", err.Error())
	}
	l.logger.ErrorContext(ctx, msg, attrs...)
}

// Debug logs at debug level
func (l *OperationLogger) Debug(ctx context.Context, msg string, args ...any) {
	attrs := append(l.baseAttrs(ctx), args...)
	l.logger.DebugContext(ctx, msg, attrs...)
}

// ================================
// Convenience Functions
// ================================

// ForDomain returns a logger for a specific domain
func ForDomain(domain string) *Logger {
	return New(domain)
}

// StartOperation starts a new operation logger
func StartOperation(domain, operation string) *OperationLogger {
	return New(domain).WithOperation(operation)
}
