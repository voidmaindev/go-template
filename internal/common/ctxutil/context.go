package ctxutil

import "context"

// contextKey is a type for context keys
type contextKey string

// Context keys
const (
	requestIDKey contextKey = "request_id"
	traceIDKey   contextKey = "trace_id"
	spanIDKey    contextKey = "span_id"
	userIDKey    contextKey = "user_id"
	userEmailKey contextKey = "user_email"
	userRoleKey  contextKey = "user_role"
	sessionIDKey contextKey = "session_id"
)

// ================================
// Request/Trace Context
// ================================

// WithRequestID adds request ID to context
func WithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, requestIDKey, id)
}

// GetRequestID retrieves request ID from context
func GetRequestID(ctx context.Context) string {
	if id, ok := ctx.Value(requestIDKey).(string); ok {
		return id
	}
	return ""
}

// WithTraceID adds trace ID to context
func WithTraceID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, traceIDKey, id)
}

// GetTraceID retrieves trace ID from context
func GetTraceID(ctx context.Context) string {
	if id, ok := ctx.Value(traceIDKey).(string); ok {
		return id
	}
	return ""
}

// WithSpanID adds span ID to context
func WithSpanID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, spanIDKey, id)
}

// GetSpanID retrieves span ID from context
func GetSpanID(ctx context.Context) string {
	if id, ok := ctx.Value(spanIDKey).(string); ok {
		return id
	}
	return ""
}

// ================================
// User Context
// ================================

// WithUserID adds user ID to context
func WithUserID(ctx context.Context, id uint) context.Context {
	return context.WithValue(ctx, userIDKey, id)
}

// GetUserID retrieves user ID from context
func GetUserID(ctx context.Context) uint {
	if id, ok := ctx.Value(userIDKey).(uint); ok {
		return id
	}
	return 0
}

// HasUserID checks if user ID is present in context
func HasUserID(ctx context.Context) bool {
	_, ok := ctx.Value(userIDKey).(uint)
	return ok
}

// WithUserEmail adds user email to context
func WithUserEmail(ctx context.Context, email string) context.Context {
	return context.WithValue(ctx, userEmailKey, email)
}

// GetUserEmail retrieves user email from context
func GetUserEmail(ctx context.Context) string {
	if email, ok := ctx.Value(userEmailKey).(string); ok {
		return email
	}
	return ""
}

// WithUserRole adds user role to context
func WithUserRole(ctx context.Context, role string) context.Context {
	return context.WithValue(ctx, userRoleKey, role)
}

// GetUserRole retrieves user role from context
func GetUserRole(ctx context.Context) string {
	if role, ok := ctx.Value(userRoleKey).(string); ok {
		return role
	}
	return ""
}

// IsAdmin checks if user is admin
func IsAdmin(ctx context.Context) bool {
	return GetUserRole(ctx) == "admin"
}

// ================================
// Session Context
// ================================

// WithSessionID adds session ID to context
func WithSessionID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, sessionIDKey, id)
}

// GetSessionID retrieves session ID from context
func GetSessionID(ctx context.Context) string {
	if id, ok := ctx.Value(sessionIDKey).(string); ok {
		return id
	}
	return ""
}

// ================================
// Builder Pattern
// ================================

// ContextBuilder provides a fluent interface for building context
type ContextBuilder struct {
	ctx context.Context
}

// NewContextBuilder creates a new ContextBuilder
func NewContextBuilder(ctx context.Context) *ContextBuilder {
	return &ContextBuilder{ctx: ctx}
}

// WithRequestID adds request ID
func (b *ContextBuilder) WithRequestID(id string) *ContextBuilder {
	b.ctx = WithRequestID(b.ctx, id)
	return b
}

// WithTraceID adds trace ID
func (b *ContextBuilder) WithTraceID(id string) *ContextBuilder {
	b.ctx = WithTraceID(b.ctx, id)
	return b
}

// WithSpanID adds span ID
func (b *ContextBuilder) WithSpanID(id string) *ContextBuilder {
	b.ctx = WithSpanID(b.ctx, id)
	return b
}

// WithUserID adds user ID
func (b *ContextBuilder) WithUserID(id uint) *ContextBuilder {
	b.ctx = WithUserID(b.ctx, id)
	return b
}

// WithUserEmail adds user email
func (b *ContextBuilder) WithUserEmail(email string) *ContextBuilder {
	b.ctx = WithUserEmail(b.ctx, email)
	return b
}

// WithUserRole adds user role
func (b *ContextBuilder) WithUserRole(role string) *ContextBuilder {
	b.ctx = WithUserRole(b.ctx, role)
	return b
}

// WithSessionID adds session ID
func (b *ContextBuilder) WithSessionID(id string) *ContextBuilder {
	b.ctx = WithSessionID(b.ctx, id)
	return b
}

// Build returns the built context
func (b *ContextBuilder) Build() context.Context {
	return b.ctx
}

// ================================
// Context Info Extraction
// ================================

// ContextInfo holds all context information
type ContextInfo struct {
	RequestID string
	TraceID   string
	SpanID    string
	UserID    uint
	UserEmail string
	UserRole  string
	SessionID string
}

// Extract extracts all context information
func Extract(ctx context.Context) ContextInfo {
	return ContextInfo{
		RequestID: GetRequestID(ctx),
		TraceID:   GetTraceID(ctx),
		SpanID:    GetSpanID(ctx),
		UserID:    GetUserID(ctx),
		UserEmail: GetUserEmail(ctx),
		UserRole:  GetUserRole(ctx),
		SessionID: GetSessionID(ctx),
	}
}

// ToLogArgs converts context info to log arguments
func (i ContextInfo) ToLogArgs() []any {
	args := make([]any, 0, 14)

	if i.RequestID != "" {
		args = append(args, "request_id", i.RequestID)
	}
	if i.TraceID != "" {
		args = append(args, "trace_id", i.TraceID)
	}
	if i.SpanID != "" {
		args = append(args, "span_id", i.SpanID)
	}
	if i.UserID != 0 {
		args = append(args, "user_id", i.UserID)
	}
	if i.UserEmail != "" {
		args = append(args, "user_email", i.UserEmail)
	}
	if i.UserRole != "" {
		args = append(args, "user_role", i.UserRole)
	}
	if i.SessionID != "" {
		args = append(args, "session_id", i.SessionID)
	}

	return args
}
