package audit

import (
	"context"

	"github.com/voidmaindev/go-template/internal/common"
	"github.com/voidmaindev/go-template/internal/common/filter"
	"github.com/voidmaindev/go-template/internal/common/logging"
)

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
}

// service implements the Service interface
type service struct {
	repo   Repository
	logger *logging.Logger
	queue  chan *AuditEntry
	done   chan struct{}
}

// NewService creates a new audit service with a background worker for async logging
func NewService(repo Repository) *service {
	s := &service{
		repo:   repo,
		logger: logging.New("audit"),
		queue:  make(chan *AuditEntry, 1000),
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
	select {
	case s.queue <- entry:
		// queued
	default:
		s.logger.Warn(ctx, "audit queue full, dropping entry",
			"action", entry.Action,
			"userID", entry.UserID,
		)
	}
}

// worker processes audit entries from the queue until the channel is closed.
func (s *service) worker() {
	defer close(s.done)
	for entry := range s.queue {
		bgCtx := context.Background()
		if err := s.Log(bgCtx, entry); err != nil {
			s.logger.Error(bgCtx, "failed to create audit log", err,
				"action", entry.Action,
				"userID", entry.UserID,
			)
		}
	}
}

// Shutdown drains the audit queue and waits for the worker to finish.
func (s *service) Shutdown(ctx context.Context) error {
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
		return nil, err
	}

	return common.NewPaginatedResult(logs, total, pagination), nil
}

// ListFiltered retrieves audit logs with dynamic filtering
func (s *service) ListFiltered(ctx context.Context, params *filter.Params) (*common.PaginatedResult[AuditLog], error) {
	logs, total, err := s.repo.FindAllFiltered(ctx, params)
	if err != nil {
		return nil, err
	}

	return common.NewPaginatedResultFromFilter(logs, total, params), nil
}

// ListByUserID retrieves audit logs for a specific user
func (s *service) ListByUserID(ctx context.Context, userID uint, pagination *common.Pagination) (*common.PaginatedResult[AuditLog], error) {
	logs, total, err := s.repo.FindByUserID(ctx, userID, pagination)
	if err != nil {
		return nil, err
	}

	return common.NewPaginatedResult(logs, total, pagination), nil
}
