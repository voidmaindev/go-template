package errors

import (
	"fmt"
	"runtime"
)

// DomainError represents a structured domain error with context
type DomainError struct {
	Code      ErrorCode
	Message   string
	Domain    string
	Operation string
	Cause     error
	Details   map[string]any
	RequestID string
	TraceID   string
	stack     []uintptr
}

// New creates a new DomainError for a specific domain and error code.
// Stack traces are not captured by default for performance; use WithStack()
// to opt in for unexpected/internal errors where a stack trace is valuable.
func New(domain string, code ErrorCode) *DomainError {
	return &DomainError{
		Domain: domain,
		Code:   code,
	}
}

// WithStack captures a stack trace at the call site and attaches it to the error.
// Use this for unexpected/internal errors where a stack trace aids debugging.
// Expected errors (NotFound, Unauthorized, ValidationError, etc.) should not capture stacks.
func (e *DomainError) WithStack() *DomainError {
	e.stack = captureStack()
	return e
}

// WithMessage sets the error message
func (e *DomainError) WithMessage(msg string) *DomainError {
	e.Message = msg
	return e
}

// WithMessagef sets a formatted error message
func (e *DomainError) WithMessagef(format string, args ...any) *DomainError {
	e.Message = fmt.Sprintf(format, args...)
	return e
}

// WithOperation sets the operation name where the error occurred
func (e *DomainError) WithOperation(op string) *DomainError {
	e.Operation = op
	return e
}

// WithCause sets the underlying cause of this error
func (e *DomainError) WithCause(err error) *DomainError {
	e.Cause = err
	return e
}

// WithDetails adds contextual details to the error
func (e *DomainError) WithDetails(details map[string]any) *DomainError {
	e.Details = details
	return e
}

// WithDetail adds a single detail key-value pair
func (e *DomainError) WithDetail(key string, value any) *DomainError {
	if e.Details == nil {
		e.Details = make(map[string]any)
	}
	e.Details[key] = value
	return e
}

// WithContext adds request and trace IDs for debugging
func (e *DomainError) WithContext(requestID, traceID string) *DomainError {
	e.RequestID = requestID
	e.TraceID = traceID
	return e
}

// WithRequestID adds request ID for debugging
func (e *DomainError) WithRequestID(requestID string) *DomainError {
	e.RequestID = requestID
	return e
}

// WithTraceID adds trace ID for debugging
func (e *DomainError) WithTraceID(traceID string) *DomainError {
	e.TraceID = traceID
	return e
}

// Error implements the error interface
func (e *DomainError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return string(e.Code)
}

// Unwrap returns the underlying cause for errors.Unwrap support
func (e *DomainError) Unwrap() error {
	return e.Cause
}

// HTTPStatus returns the HTTP status code for this error
func (e *DomainError) HTTPStatus() int {
	return e.Code.HTTPStatus()
}

// Is implements errors.Is support for error comparison
func (e *DomainError) Is(target error) bool {
	if t, ok := target.(*DomainError); ok {
		return e.Code == t.Code && e.Domain == t.Domain
	}
	return false
}

// Stack returns the stack trace where the error was created
func (e *DomainError) Stack() []uintptr {
	return e.stack
}

// StackTrace returns the formatted stack trace
func (e *DomainError) StackTrace() string {
	if len(e.stack) == 0 {
		return ""
	}

	frames := runtime.CallersFrames(e.stack)
	var trace string
	for {
		frame, more := frames.Next()
		trace += fmt.Sprintf("%s\n\t%s:%d\n", frame.Function, frame.File, frame.Line)
		if !more {
			break
		}
	}
	return trace
}

// Clone creates a shallow copy of the error, preserving the existing stack trace.
// Details are deep-copied to prevent mutation of the original.
func (e *DomainError) Clone() *DomainError {
	clone := &DomainError{
		Code:      e.Code,
		Message:   e.Message,
		Domain:    e.Domain,
		Operation: e.Operation,
		Cause:     e.Cause,
		RequestID: e.RequestID,
		TraceID:   e.TraceID,
	}
	if len(e.stack) > 0 {
		clone.stack = make([]uintptr, len(e.stack))
		copy(clone.stack, e.stack)
	}
	if e.Details != nil {
		clone.Details = make(map[string]any, len(e.Details))
		for k, v := range e.Details {
			clone.Details[k] = v
		}
	}
	return clone
}

func captureStack() []uintptr {
	pcs := make([]uintptr, 32)
	n := runtime.Callers(3, pcs)
	return pcs[:n]
}
