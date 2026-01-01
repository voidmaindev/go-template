package user

import (
	"context"

	"github.com/voidmaindev/GoTemplate/internal/common"
	"gorm.io/gorm"
)

// Repository defines the user repository interface
type Repository interface {
	common.Repository[User]

	// FindByEmail finds a user by email
	FindByEmail(ctx context.Context, email string) (*User, error)

	// ExistsByEmail checks if a user with the given email exists
	ExistsByEmail(ctx context.Context, email string) (bool, error)
}

// repository implements the Repository interface
type repository struct {
	*common.BaseRepository[User]
}

// NewRepository creates a new user repository
func NewRepository(db *gorm.DB) Repository {
	return &repository{
		BaseRepository: common.NewBaseRepository[User](db),
	}
}

// FindByEmail finds a user by email
func (r *repository) FindByEmail(ctx context.Context, email string) (*User, error) {
	return r.FindOne(ctx, map[string]any{"email": email})
}

// ExistsByEmail checks if a user with the given email exists
func (r *repository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	return r.Exists(ctx, map[string]any{"email": email})
}
