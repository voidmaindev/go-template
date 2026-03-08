package audit

import (
	"context"
	"log/slog"

	"github.com/voidmaindev/go-template/internal/common"
	"github.com/voidmaindev/go-template/internal/common/filter"
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
	ListFiltered(ctx context.Context, params *filter.Params) (*common.FilteredResult[AuditLog], error)
	// ListByUserID retrieves audit logs for a specific user
	ListByUserID(ctx context.Context, userID uint, pagination *common.Pagination) (*common.PaginatedResult[AuditLog], error)
}

// service implements the Service interface
type service struct {
	repo Repository
}

// NewService creates a new audit service
func NewService(repo Repository) Service {
	return &service{repo: repo}
}

// Log creates an audit log entry synchronously
func (s *service) Log(ctx context.Context, entry *AuditEntry) error {
	log := entry.ToAuditLog()
	return s.repo.Create(ctx, log)
}

// LogAsync creates an audit log entry asynchronously (fire-and-forget)
// Errors are logged but do not affect the caller
func (s *service) LogAsync(ctx context.Context, entry *AuditEntry) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				slog.Error("panic in async audit log", "recovered", r, "action", entry.Action)
			}
		}()
		// Use a background context to prevent cancellation from parent
		bgCtx := context.Background()
		if err := s.Log(bgCtx, entry); err != nil {
			slog.Error("failed to create audit log",
				"action", entry.Action,
				"userID", entry.UserID,
				"error", err,
			)
		}
	}()
}

// List retrieves audit logs with pagination
func (s *service) List(ctx context.Context, pagination *common.Pagination) (*common.PaginatedResult[AuditLog], error) {
	logs, total, err := s.repo.FindAll(ctx, pagination)
	if err != nil {
		return nil, err
	}

	// Convert pointers to values
	values := make([]AuditLog, len(logs))
	for i, log := range logs {
		values[i] = *log
	}

	return common.NewPaginatedResult(values, total, pagination), nil
}

// ListFiltered retrieves audit logs with dynamic filtering
func (s *service) ListFiltered(ctx context.Context, params *filter.Params) (*common.FilteredResult[AuditLog], error) {
	logs, total, err := s.repo.FindAllFiltered(ctx, params)
	if err != nil {
		return nil, err
	}

	// Convert pointers to values
	values := make([]AuditLog, len(logs))
	for i, log := range logs {
		values[i] = *log
	}

	return common.NewFilteredResult(values, total, params), nil
}

// ListByUserID retrieves audit logs for a specific user
func (s *service) ListByUserID(ctx context.Context, userID uint, pagination *common.Pagination) (*common.PaginatedResult[AuditLog], error) {
	logs, total, err := s.repo.FindByUserID(ctx, userID, pagination)
	if err != nil {
		return nil, err
	}

	// Convert pointers to values
	values := make([]AuditLog, len(logs))
	for i, log := range logs {
		values[i] = *log
	}

	return common.NewPaginatedResult(values, total, pagination), nil
}
