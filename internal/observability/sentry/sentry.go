// Package sentry wires the getsentry/sentry-go SDK into the application.
//
// Init returns a no-op when DSN is empty so freshly cloned templates work
// without external configuration. CaptureDomainError is exposed as the
// hook bound to common.OnInternalError in serve.go, keeping internal/common
// free of any Sentry import.
package sentry

import (
	"context"
	"fmt"
	"log/slog"
	"regexp"
	"time"

	sentrygo "github.com/getsentry/sentry-go"

	"github.com/voidmaindev/go-template/internal/common/errors"
	"github.com/voidmaindev/go-template/internal/config"
)

// Skip events for these routes — high-volume liveness/scrape paths would
// otherwise burn through Sentry quota during a flapping incident.
var noisyRoutes = regexp.MustCompile(`^/(healthz|readyz|metrics)$`)

var enabled bool

// Enabled reports whether Sentry was successfully initialized.
func Enabled() bool { return enabled }

// Init configures the global Sentry client. With Enabled=false or empty DSN,
// returns (false, nil) and leaves Sentry as a no-op — calls to CaptureException
// from any other package will silently do nothing.
func Init(cfg config.SentryConfig) (bool, error) {
	if !cfg.Enabled || cfg.DSN == "" {
		slog.Warn("Sentry disabled or DSN empty, skipping init",
			"enabled", cfg.Enabled,
			"dsn_set", cfg.DSN != "",
		)
		enabled = false
		return false, nil
	}

	err := sentrygo.Init(sentrygo.ClientOptions{
		Dsn:              cfg.DSN,
		Environment:      cfg.Environment,
		Release:          cfg.Release,
		AttachStacktrace: cfg.AttachStacktrace,
		Debug:            cfg.Debug,
		TracesSampleRate: cfg.TracesSampleRate,
		BeforeSend:       beforeSend,
	})
	if err != nil {
		return false, fmt.Errorf("sentry init: %w", err)
	}

	enabled = true
	slog.Info("Sentry initialized",
		"environment", cfg.Environment,
		"release", cfg.Release,
	)
	return true, nil
}

// Flush blocks until queued events are sent or the timeout elapses.
// Returns false if the timeout was hit.
func Flush(timeout time.Duration) bool {
	if !enabled {
		return true
	}
	return sentrygo.Flush(timeout)
}

// beforeSend drops events from noisy routes and scrubs sensitive headers.
func beforeSend(event *sentrygo.Event, hint *sentrygo.EventHint) *sentrygo.Event {
	if route, ok := event.Tags["route"]; ok && noisyRoutes.MatchString(route) {
		return nil
	}
	if event.Request != nil && event.Request.Headers != nil {
		for k := range event.Request.Headers {
			lk := normalizeHeader(k)
			if lk == "authorization" || lk == "cookie" || lk == "set-cookie" {
				event.Request.Headers[k] = "[redacted]"
			}
		}
	}
	return event
}

func normalizeHeader(h string) string {
	out := make([]byte, len(h))
	for i := 0; i < len(h); i++ {
		c := h[i]
		if c >= 'A' && c <= 'Z' {
			c += 'a' - 'A'
		}
		out[i] = c
	}
	return string(out)
}

// CaptureDomainError reports a CodeInternal domain error to Sentry with
// request/domain context. Bound to common.OnInternalError at startup so
// internal/common stays free of any Sentry import.
func CaptureDomainError(ctx context.Context, de *errors.DomainError, requestID string) {
	if !enabled || de == nil {
		return
	}

	hub := sentrygo.GetHubFromContext(ctx)
	if hub == nil {
		hub = sentrygo.CurrentHub()
	}

	hub.WithScope(func(scope *sentrygo.Scope) {
		scope.SetTag("domain", de.Domain)
		scope.SetTag("code", string(de.Code))
		if de.Operation != "" {
			scope.SetTag("operation", de.Operation)
		}
		if requestID != "" {
			scope.SetTag("request_id", requestID)
		}
		if len(de.Details) > 0 {
			scope.SetContext("details", sentrygo.Context(sanitizeDetails(de.Details)))
		}
		// Group by domain+code in Sentry's issue list
		scope.SetFingerprint([]string{de.Domain, string(de.Code)})

		if de.Cause != nil {
			hub.CaptureException(de.Cause)
			return
		}
		hub.CaptureMessage(de.Error())
	})
}

// sanitizeDetails strips obvious secret-looking keys before sending to Sentry.
func sanitizeDetails(in map[string]any) map[string]any {
	out := make(map[string]any, len(in))
	for k, v := range in {
		lk := normalizeHeader(k)
		if lk == "password" || lk == "token" || lk == "secret" || lk == "authorization" {
			out[k] = "[redacted]"
			continue
		}
		out[k] = v
	}
	return out
}
