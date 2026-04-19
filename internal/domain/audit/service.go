package audit

import (
	"context"

	"github.com/voidmaindev/go-template/internal/common"
	"github.com/voidmaindev/go-template/internal/common/ctxutil"
	"github.com/voidmaindev/go-template/internal/common/errors"
	"github.com/voidmaindev/go-template/internal/common/filter"
	"github.com/voidmaindev/go-template/internal/common/logging"
)

const domainName = "audit"

// Service defines the audit service interface
type Service interface {
	// Log creates an audit log entry synchronously
	Log(ctx context.Context, entry *AuditEntry) error
	// LogAsync creates an audit log entry asynchronously (fire-and-forget)
	LogAsync(ctx context.Context, entry *AuditEntry)
	// List retrieves audit logs with pagination
	List(ctx context.Context, pagination *common.Pagination) (*common.PaginatedResult[AuditLog], error)
	// ListFiltered retrieves audit logs with dynamic filtering
	ListFiltered(ctx context.Context, params *filter.Params) (*common.PaginatedResult[AuditLog], error)
	// ListByUserID retrieves audit logs for a specific user
	ListByUserID(ctx context.Context, userID uint, pagination *common.Pagination) (*common.PaginatedResult[AuditLog], error)
	// Shutdown drains the audit queue and waits for the worker to finish
	Shutdown(ctx context.Context) error
}

// queuedEntry bundles an audit entry with the request/trace correlation IDs
// captured from the caller's context at enqueue time. The request context
// itself cannot cross the goroutine boundary (it's typically cancelled when
// the request completes), so we preserve only the correlation metadata.
type queuedEntry struct {
	entry     *AuditEntry
	requestID string
	traceID   string
}

// service implements the Service interface
type service struct {
	repo   Repository
	logger *logging.Logger
	queue  chan queuedEntry
	done   chan struct{}
}

// NewService creates a new audit service with a background worker for async logging
func NewService(repo Repository) Service {
	s := &service{
		repo:   repo,
		logger: logging.New(domainName),
		queue:  make(chan queuedEntry, 1000),
		done:   make(chan struct{}),
	}
	go s.worker()
	return s
}

// Log creates an audit log entry synchronously
func (s *service) Log(ctx context.Context, entry *AuditEntry) error {
	log := entry.ToAuditLog()
	return s.repo.Create(ctx, log)
}

// LogAsync queues an audit log entry for asynchronous processing.
// If the queue is full, the entry is dropped and a warning is logged.
func (s *service) LogAsync(ctx context.Context, entry *AuditEntry) {
	q := queuedEntry{
		entry:     entry,
		requestID: ctxutil.GetRequestID(ctx),
		traceID:   ctxutil.GetTraceID(ctx),
	}
	select {
	case s.queue <- q:
		// queued
	default:
		s.logger.Warn(ctx, "audit queue full, dropping entry",
			"action", entry.Action,
			"userID", entry.UserID,
		)
	}
}

// worker processes audit entries from the queue until the channel is closed.
// Each entry runs on a fresh background context seeded with the request/trace
// IDs captured at enqueue time, so logs remain correlatable to the original
// request without being cancelled when that request ends.
func (s *service) worker() {
	defer close(s.done)
	for q := range s.queue {
		ctx := context.Background()
		if q.requestID != "" {
			ctx = ctxutil.WithRequestID(ctx, q.requestID)
		}
		if q.traceID != "" {
			ctx = ctxutil.WithTraceID(ctx, q.traceID)
		}
		if err := s.Log(ctx, q.entry); err != nil {
			s.logger.Error(ctx, "failed to create audit log", err,
				"action", q.entry.Action,
				"userID", q.entry.UserID,
			)
		}
	}
}

// Shutdown drains the audit queue and waits for the worker to finish.
func (s *service) Shutdown(ctx context.Context) error {
	remaining := len(s.queue)
	if remaining > 0 {
		s.logger.Info(ctx, "draining audit queue", "remaining", remaining)
	}
	close(s.queue)
	select {
	case <-s.done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// List retrieves audit logs with pagination
func (s *service) List(ctx context.Context, pagination *common.Pagination) (*common.PaginatedResult[AuditLog], error) {
	logs, total, err := s.repo.FindAll(ctx, pagination)
	if err != nil {
		return nil, errors.Internal(domainName, err).WithOperation("List")
	}

	return common.NewPaginatedResult(logs, total, pagination), nil
}

// ListFiltered retrieves audit logs with dynamic filtering
func (s *service) ListFiltered(ctx context.Context, params *filter.Params) (*common.PaginatedResult[AuditLog], error) {
	logs, total, err := s.repo.FindAllFiltered(ctx, params)
	if err != nil {
		return nil, errors.Internal(domainName, err).WithOperation("ListFiltered")
	}

	return common.NewPaginatedResultFromFilter(logs, total, params), nil
}

// ListByUserID retrieves audit logs for a specific user
func (s *service) ListByUserID(ctx context.Context, userID uint, pagination *common.Pagination) (*common.PaginatedResult[AuditLog], error) {
	logs, total, err := s.repo.FindByUserID(ctx, userID, pagination)
	if err != nil {
		return nil, errors.Internal(domainName, err).WithOperation("ListByUserID")
	}

	return common.NewPaginatedResult(logs, total, pagination), nil
}
