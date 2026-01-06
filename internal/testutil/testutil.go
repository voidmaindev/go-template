// Package testutil provides common utilities for testing.
package testutil

import (
	"context"
	"testing"
	"time"
)

// TestContext returns a context with a reasonable timeout for tests.
func TestContext(t *testing.T) context.Context {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	t.Cleanup(cancel)
	return ctx
}

// TestContextWithTimeout returns a context with a custom timeout.
func TestContextWithTimeout(t *testing.T, timeout time.Duration) context.Context {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	t.Cleanup(cancel)
	return ctx
}

// RequireNoError fails the test immediately if err is not nil.
func RequireNoError(t *testing.T, err error, msgAndArgs ...any) {
	t.Helper()
	if err != nil {
		if len(msgAndArgs) > 0 {
			t.Fatalf("%v: %v", msgAndArgs[0], err)
		}
		t.Fatalf("unexpected error: %v", err)
	}
}

// RequireError fails the test if err is nil.
func RequireError(t *testing.T, err error, msgAndArgs ...any) {
	t.Helper()
	if err == nil {
		if len(msgAndArgs) > 0 {
			t.Fatalf("%v: expected error but got nil", msgAndArgs[0])
		}
		t.Fatal("expected error but got nil")
	}
}

// RequireEqual fails the test if expected != actual.
func RequireEqual[T comparable](t *testing.T, expected, actual T, msgAndArgs ...any) {
	t.Helper()
	if expected != actual {
		if len(msgAndArgs) > 0 {
			t.Fatalf("%v: expected %v, got %v", msgAndArgs[0], expected, actual)
		}
		t.Fatalf("expected %v, got %v", expected, actual)
	}
}

// RequireNotEqual fails the test if expected == actual.
func RequireNotEqual[T comparable](t *testing.T, expected, actual T, msgAndArgs ...any) {
	t.Helper()
	if expected == actual {
		if len(msgAndArgs) > 0 {
			t.Fatalf("%v: expected values to differ, both are %v", msgAndArgs[0], expected)
		}
		t.Fatalf("expected values to differ, both are %v", expected)
	}
}

// RequireNil fails the test if value is not nil.
func RequireNil(t *testing.T, value any, msgAndArgs ...any) {
	t.Helper()
	if value != nil {
		if len(msgAndArgs) > 0 {
			t.Fatalf("%v: expected nil, got %v", msgAndArgs[0], value)
		}
		t.Fatalf("expected nil, got %v", value)
	}
}

// RequireNotNil fails the test if value is nil.
func RequireNotNil(t *testing.T, value any, msgAndArgs ...any) {
	t.Helper()
	if value == nil {
		if len(msgAndArgs) > 0 {
			t.Fatalf("%v: expected non-nil value", msgAndArgs[0])
		}
		t.Fatal("expected non-nil value")
	}
}

// RequireTrue fails the test if condition is false.
func RequireTrue(t *testing.T, condition bool, msgAndArgs ...any) {
	t.Helper()
	if !condition {
		if len(msgAndArgs) > 0 {
			t.Fatalf("%v: expected true", msgAndArgs[0])
		}
		t.Fatal("expected true")
	}
}

// RequireFalse fails the test if condition is true.
func RequireFalse(t *testing.T, condition bool, msgAndArgs ...any) {
	t.Helper()
	if condition {
		if len(msgAndArgs) > 0 {
			t.Fatalf("%v: expected false", msgAndArgs[0])
		}
		t.Fatal("expected false")
	}
}

// RequireLen fails the test if the slice/map/string doesn't have the expected length.
func RequireLen[T any](t *testing.T, collection []T, expected int, msgAndArgs ...any) {
	t.Helper()
	if len(collection) != expected {
		if len(msgAndArgs) > 0 {
			t.Fatalf("%v: expected length %d, got %d", msgAndArgs[0], expected, len(collection))
		}
		t.Fatalf("expected length %d, got %d", expected, len(collection))
	}
}

// RequireContains fails if the string doesn't contain the substring.
func RequireContains(t *testing.T, s, substr string, msgAndArgs ...any) {
	t.Helper()
	if len(s) == 0 || len(substr) == 0 {
		if len(msgAndArgs) > 0 {
			t.Fatalf("%v: expected %q to contain %q", msgAndArgs[0], s, substr)
		}
		t.Fatalf("expected %q to contain %q", s, substr)
		return
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return
		}
	}
	if len(msgAndArgs) > 0 {
		t.Fatalf("%v: expected %q to contain %q", msgAndArgs[0], s, substr)
	}
	t.Fatalf("expected %q to contain %q", s, substr)
}

// AssertNoError logs error but doesn't fail immediately.
func AssertNoError(t *testing.T, err error, msgAndArgs ...any) bool {
	t.Helper()
	if err != nil {
		if len(msgAndArgs) > 0 {
			t.Errorf("%v: %v", msgAndArgs[0], err)
		} else {
			t.Errorf("unexpected error: %v", err)
		}
		return false
	}
	return true
}

// AssertEqual logs if values differ but doesn't fail immediately.
func AssertEqual[T comparable](t *testing.T, expected, actual T, msgAndArgs ...any) bool {
	t.Helper()
	if expected != actual {
		if len(msgAndArgs) > 0 {
			t.Errorf("%v: expected %v, got %v", msgAndArgs[0], expected, actual)
		} else {
			t.Errorf("expected %v, got %v", expected, actual)
		}
		return false
	}
	return true
}

// AssertTrue logs if condition is false but doesn't fail immediately.
func AssertTrue(t *testing.T, condition bool, msgAndArgs ...any) bool {
	t.Helper()
	if !condition {
		if len(msgAndArgs) > 0 {
			t.Errorf("%v: expected true", msgAndArgs[0])
		} else {
			t.Error("expected true")
		}
		return false
	}
	return true
}

// Ptr returns a pointer to the given value. Useful for optional fields in tests.
func Ptr[T any](v T) *T {
	return &v
}
