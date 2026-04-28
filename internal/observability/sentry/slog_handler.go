package sentry

import (
	"context"
	"log/slog"

	sentrygo "github.com/getsentry/sentry-go"
)

// slogHandler wraps an inner slog.Handler and mirrors records at LevelError
// (or higher) into Sentry as events. The inner handler always runs, so log
// lines still reach stdout / Loki via Alloy.
type slogHandler struct {
	inner slog.Handler
}

// NewSlogHandler wraps an existing slog.Handler so that Error-level records
// are also forwarded to Sentry. Cheap to call when Sentry is disabled —
// records simply pass through to the inner handler.
func NewSlogHandler(inner slog.Handler) slog.Handler {
	return &slogHandler{inner: inner}
}

func (h *slogHandler) Enabled(ctx context.Context, lvl slog.Level) bool {
	// Always allow Error-and-above through, even if the inner handler's level
	// is set higher — we don't want the Sentry mirror to silently drop errors.
	if lvl >= slog.LevelError {
		return true
	}
	return h.inner.Enabled(ctx, lvl)
}

func (h *slogHandler) Handle(ctx context.Context, r slog.Record) error {
	if err := h.inner.Handle(ctx, r); err != nil {
		return err
	}

	if r.Level < slog.LevelError || !enabled {
		return nil
	}

	hub := sentrygo.GetHubFromContext(ctx)
	if hub == nil {
		hub = sentrygo.CurrentHub()
	}

	event := sentrygo.NewEvent()
	event.Level = sentrygo.LevelError
	event.Message = r.Message
	event.Timestamp = r.Time

	tags := make(map[string]string)
	extra := make(sentrygo.Context)

	r.Attrs(func(a slog.Attr) bool {
		routeAttr(a, "", tags, extra)
		return true
	})

	// Merge into existing scope so request_id/route from the Hub middleware survive.
	hub.WithScope(func(scope *sentrygo.Scope) {
		for k, v := range tags {
			scope.SetTag(k, v)
		}
		if len(extra) > 0 {
			scope.SetContext("attrs", extra)
		}
		hub.CaptureEvent(event)
	})

	return nil
}

func (h *slogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &slogHandler{inner: h.inner.WithAttrs(attrs)}
}

func (h *slogHandler) WithGroup(name string) slog.Handler {
	return &slogHandler{inner: h.inner.WithGroup(name)}
}

// routeAttr places primitive attrs into Sentry tags (low-cardinality and
// searchable) and complex values into the "attrs" context.
func routeAttr(a slog.Attr, prefix string, tags map[string]string, extra sentrygo.Context) {
	key := a.Key
	if prefix != "" {
		key = prefix + "." + key
	}
	v := a.Value.Resolve()

	switch v.Kind() {
	case slog.KindString:
		tags[key] = v.String()
	case slog.KindInt64, slog.KindUint64, slog.KindFloat64, slog.KindBool, slog.KindDuration, slog.KindTime:
		tags[key] = v.String()
	case slog.KindGroup:
		for _, ga := range v.Group() {
			routeAttr(ga, key, tags, extra)
		}
	default:
		extra[key] = v.Any()
	}
}
