# Observability: Sentry + Loki

This template ships with two error/log signals on top of the existing OpenTelemetry tracing and Prometheus metrics:

- **Sentry** — captures panics, internal (5xx) domain errors, and any `slog.Error` call. Adds per-request context (request_id, route, method, user_id, env, release).
- **Loki via Grafana Alloy** — a sidecar tails Docker container stdout, parses the existing slog JSON, and ships log lines to a self-hosted Loki. **No log-shipping code in the app.**

Both subsystems are **enabled by default** but degrade gracefully:

- Empty `SENTRY_DSN` → Sentry initializes as a no-op (warns once at startup, never tries to send).
- Loki / Alloy run only when the `observability` Docker Compose profile is active.

## Quick start

```bash
# Bring up the full observability stack alongside the API
docker compose --profile observability up -d

# Open Grafana
#   http://localhost:3001  (admin / admin)
# Explore -> Loki -> {service="api"}
```

To enable Sentry, set `SENTRY_DSN` in `.env` (or the environment) and restart the API. With an empty DSN the integration is wired but silent.

## Environment variables

### Sentry

| Variable                    | Default                | Notes                                             |
| --------------------------- | ---------------------- | ------------------------------------------------- |
| `SENTRY_ENABLED`            | `true`                 | Set to `false` to skip init entirely              |
| `SENTRY_DSN`                | _empty_                | When empty, Sentry is a no-op                     |
| `SENTRY_ENVIRONMENT`        | falls back to `APP_ENV`| `development` / `staging` / `production` …        |
| `SENTRY_RELEASE`            | `dev`                  | Overridden by `BUILD_SHA` env var or ldflags      |
| `SENTRY_TRACES_SAMPLE_RATE` | `0.0`                  | Errors only — performance stays in OTel           |
| `SENTRY_ATTACH_STACKTRACE`  | `true`                 | Attach stacktrace to all events                   |
| `SENTRY_DEBUG`              | `false`                | Verbose SDK logging                               |

These names match the official sentry-go convention (no `APP_` prefix); the SDK reads `SENTRY_DSN` natively.

### Logging / Loki

The app emits structured slog JSON to stdout. There is **no Loki client in the app** — log shipping is handled entirely by the Alloy sidecar. To turn logs off, do it at the Docker layer (drop the `observability` profile or stop the `alloy` service).

## How it fits together

```
[Fiber app] ──► slog JSON ──► stdout ──► [Alloy] ──► [Loki]
       │                                            │
       └─ panic / 5xx / slog.Error ──► [Sentry] ◄───┘  (Grafana datasource)
```

- `internal/observability/sentry/sentry.go` — `Init`, `Flush`, `CaptureDomainError`
- `internal/observability/sentry/slog_handler.go` — wraps the active slog handler so `slog.Error` is mirrored to Sentry without losing the stdout line
- `internal/observability/sentry/middleware.go` — Fiber middleware that attaches a per-request hub with request_id / route / method / user_id
- `cmd/api/cmd/serve.go` — wiring (init, slog wrap, `OnInternalError` hook, hub middleware, `Flush` on shutdown)
- `internal/middleware/recovery.go` — `hub.RecoverWithContext` on panic
- `internal/common/response.go` — `OnInternalError` hook fires for `CodeInternal` only
- `config/alloy/config.alloy` — Alloy pipeline
- `grafana/provisioning/datasources/datasource.yml` — Loki datasource

## Querying logs

In Grafana → Explore → Loki:

```logql
# All API logs
{service="api"}

# Errors only
{service="api", level="error"}

# Specific request (request_id stays in the body, not as a label)
{service="api"} | json | request_id="9b6f…"

# Panics
{service="api"} |= "Panic recovered"
```

## Testing Sentry locally

The cheapest end-to-end smoke test is a free sentry.io project:

```bash
SENTRY_DSN="https://xxx@oXXX.ingest.sentry.io/XXX" \
SENTRY_ENVIRONMENT=local \
go run ./cmd/api serve api

# Trigger a panic via any handler that calls panic(...)
# Or test slog.Error mirroring:
#   slog.Error("manual smoke", "request_id", "test-1")

# Event should appear in your Sentry project within ~30s
```

To test the no-op path (no DSN required):

```bash
go test ./internal/observability/sentry/...
```

## Cardinality discipline

Loki labels are kept low-cardinality on purpose: `service`, `container_name`, `env`, `level`. Per-request fields (`request_id`, `user_id`) stay in the JSON body and are queried via `| json | request_id="…"`. Promoting either to a label would explode the Loki index.

Sentry quota is protected by `BeforeSend`, which drops events tagged with `route` matching `/healthz`, `/readyz`, or `/metrics`. A flapping liveness endpoint won't burn through your quota.

## Production notes

- **Loki storage**: the bundled Loki uses BoltDB + filesystem (Loki's shipped `local-config.yaml`). Acceptable for the template; in production swap to S3/GCS/Azure Blob.
- **Alloy + Docker socket**: `/var/run/docker.sock` is mounted read-only into Alloy. This is privileged-equivalent and fine for local dev. In production prefer a per-host agent (DaemonSet on Kubernetes) or file-based tailing.
- **Self-hosted Sentry**: not bundled. The footprint is large (10+ services). Either point at sentry.io / a managed instance, or run Sentry in a separate compose file.
- **PII**: sentry-go's `SendDefaultPII` is off by default — keep it that way. `BeforeSend` defensively strips `Authorization`, `Cookie`, and `Set-Cookie` headers. Domain-error `Details` are also scrubbed for `password` / `token` / `secret` keys.
