package audit

import "context"

// Logger is a minimal interface for audit logging.
// This interface allows domains to log audit events without
// importing the full audit package (to avoid circular dependencies).
type Logger interface {
	LogAsync(ctx context.Context, entry *AuditEntry)
}
