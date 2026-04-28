package sentry

import (
	"bytes"
	"log/slog"
	"strings"
	"sync"
	"testing"
	"time"

	sentrygo "github.com/getsentry/sentry-go"

	"github.com/voidmaindev/go-template/internal/config"
)

// withMockTransport binds a MockTransport to the current Sentry hub for the
// duration of the test and flips the package-level enabled flag.
func withMockTransport(t *testing.T) *sentrygo.MockTransport {
	t.Helper()
	prevEnabled := enabled

	mock := &sentrygo.MockTransport{}
	client, err := sentrygo.NewClient(sentrygo.ClientOptions{
		Dsn:       "https://public@example.com/1",
		Transport: mock,
	})
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	prevClient := sentrygo.CurrentHub().Client()
	sentrygo.CurrentHub().BindClient(client)
	enabled = true

	t.Cleanup(func() {
		enabled = prevEnabled
		sentrygo.CurrentHub().BindClient(prevClient)
	})
	return mock
}

func TestInit_EmptyDSN_NoOps(t *testing.T) {
	prev := enabled
	defer func() { enabled = prev }()

	ok, err := Init(config.SentryConfig{Enabled: true, DSN: ""})
	if err != nil {
		t.Fatalf("Init returned error: %v", err)
	}
	if ok {
		t.Fatal("expected Init to report not-initialized for empty DSN")
	}
	if Enabled() {
		t.Fatal("Enabled() should report false")
	}
}

func TestInit_DisabledFlag_NoOps(t *testing.T) {
	prev := enabled
	defer func() { enabled = prev }()

	ok, err := Init(config.SentryConfig{Enabled: false, DSN: "https://public@example.com/1"})
	if err != nil {
		t.Fatalf("Init returned error: %v", err)
	}
	if ok {
		t.Fatal("expected Init to report not-initialized when Enabled=false")
	}
}

func TestSlogHandler_ErrorMirrorsToSentry(t *testing.T) {
	mock := withMockTransport(t)

	var buf bytes.Buffer
	var mu sync.Mutex
	inner := slog.NewJSONHandler(&lockingWriter{w: &buf, mu: &mu}, &slog.HandlerOptions{Level: slog.LevelDebug})
	logger := slog.New(NewSlogHandler(inner))

	logger.Error("boom", "request_id", "abc-123", "count", 7)

	if !sentrygo.Flush(2 * time.Second) {
		t.Fatal("Sentry flush timed out")
	}

	events := mock.Events()
	if len(events) != 1 {
		t.Fatalf("expected 1 captured event, got %d", len(events))
	}
	ev := events[0]
	if ev.Message != "boom" {
		t.Errorf("expected message 'boom', got %q", ev.Message)
	}
	if got := ev.Tags["request_id"]; got != "abc-123" {
		t.Errorf("expected tag request_id=abc-123, got %q", got)
	}
	if !strings.Contains(buf.String(), `"msg":"boom"`) {
		t.Errorf("expected JSON log line on stdout buffer, got %q", buf.String())
	}
}

func TestSlogHandler_InfoDoesNotMirror(t *testing.T) {
	mock := withMockTransport(t)

	var buf bytes.Buffer
	var mu sync.Mutex
	inner := slog.NewJSONHandler(&lockingWriter{w: &buf, mu: &mu}, nil)
	logger := slog.New(NewSlogHandler(inner))

	logger.Info("just info", "request_id", "abc")

	if got := len(mock.Events()); got != 0 {
		t.Errorf("expected 0 sentry events for Info, got %d", got)
	}
	if !strings.Contains(buf.String(), `"msg":"just info"`) {
		t.Errorf("expected JSON log line on stdout buffer, got %q", buf.String())
	}
}

// lockingWriter avoids races when slog Handle and the test goroutine touch the
// same buffer.
type lockingWriter struct {
	w  *bytes.Buffer
	mu *sync.Mutex
}

func (lw *lockingWriter) Write(p []byte) (int, error) {
	lw.mu.Lock()
	defer lw.mu.Unlock()
	return lw.w.Write(p)
}
