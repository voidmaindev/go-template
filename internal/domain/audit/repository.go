package audit

import (
	"context"
	"database/sql"

	"github.com/voidmaindev/go-template/internal/common"
	"github.com/voidmaindev/go-template/internal/common/filter"
	"gorm.io/gorm"
)

// Repository defines the audit log repository interface
type Repository interface {
	Create(ctx context.Context, log *AuditLog) error
	FindAll(ctx context.Context, pagination *common.Pagination) ([]AuditLog, int64, error)
	FindAllFiltered(ctx context.Context, params *filter.Params) ([]AuditLog, int64, error)
	FindByUserID(ctx context.Context, userID uint, pagination *common.Pagination) ([]AuditLog, int64, error)
}

// repository implements the Repository interface
type repository struct {
	db *gorm.DB
}

// NewRepository creates a new audit repository
func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

// Create inserts a new audit log entry
func (r *repository) Create(ctx context.Context, log *AuditLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

// FindAll retrieves audit logs with pagination
func (r *repository) FindAll(ctx context.Context, pagination *common.Pagination) ([]AuditLog, int64, error) {
	var logs []AuditLog
	var total int64

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		query := tx.Model(&AuditLog{})

		if err := query.Count(&total).Error; err != nil {
			return err
		}

		return query.
			Order("timestamp DESC").
			Offset(pagination.GetOffset()).
			Limit(pagination.GetLimit()).
			Find(&logs).Error
	}, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

// FindAllFiltered retrieves audit logs with dynamic filtering
func (r *repository) FindAllFiltered(ctx context.Context, params *filter.Params) ([]AuditLog, int64, error) {
	var logs []AuditLog
	var total int64

	config := AuditLog{}.FilterConfig()

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Count with filters only (no pagination)
		countQuery := tx.Model(&AuditLog{})
		countQuery = filter.ApplyFiltersOnly(countQuery, config, params)
		if err := countQuery.Count(&total).Error; err != nil {
			return err
		}

		// Query with filters, sorting, and pagination
		query := tx.Model(&AuditLog{})
		query = filter.Apply(query, config, params)
		return query.Find(&logs).Error
	}, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

// FindByUserID retrieves audit logs for a specific user
func (r *repository) FindByUserID(ctx context.Context, userID uint, pagination *common.Pagination) ([]AuditLog, int64, error) {
	var logs []AuditLog
	var total int64

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		query := tx.Model(&AuditLog{}).Where("user_id = ?", userID)

		if err := query.Count(&total).Error; err != nil {
			return err
		}

		return query.
			Order("timestamp DESC").
			Offset(pagination.GetOffset()).
			Limit(pagination.GetLimit()).
			Find(&logs).Error
	}, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}
