package rbac

import (
	"context"

	"github.com/voidmaindev/go-template/internal/common"
	"gorm.io/gorm"
)

// Repository defines the interface for role data access
type Repository interface {
	common.Repository[Role]

	// FindByCode finds a role by its code
	FindByCode(ctx context.Context, code string) (*Role, error)

	// ExistsByCode checks if a role with the given code exists
	ExistsByCode(ctx context.Context, code string) (bool, error)

	// FindSystemRoles returns all system roles
	FindSystemRoles(ctx context.Context) ([]Role, error)
}

type repository struct {
	*common.BaseRepository[Role]
}

// NewRepository creates a new RBAC repository
func NewRepository(db *gorm.DB) Repository {
	return &repository{
		BaseRepository: common.NewBaseRepository[Role](db, "role"),
	}
}

// FindByCode finds a role by its code
func (r *repository) FindByCode(ctx context.Context, code string) (*Role, error) {
	return r.FindOne(ctx, map[string]any{"code": code})
}

// ExistsByCode checks if a role with the given code exists
func (r *repository) ExistsByCode(ctx context.Context, code string) (bool, error) {
	return r.Exists(ctx, map[string]any{"code": code})
}

// FindSystemRoles returns all system roles
func (r *repository) FindSystemRoles(ctx context.Context) ([]Role, error) {
	var roles []Role
	err := r.DB().WithContext(ctx).Where("is_system = ?", true).Find(&roles).Error
	if err != nil {
		return nil, err
	}
	return roles, nil
}
