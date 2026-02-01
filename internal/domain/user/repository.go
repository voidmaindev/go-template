package user

import (
	"context"

	"github.com/voidmaindev/go-template/internal/common"
	"gorm.io/gorm"
)

// Repository defines the user repository interface.
// It extends common.Repository[User] with domain-specific queries
// for authentication and user management.
type Repository interface {
	common.Repository[User]

	// FindByEmail retrieves a user by their unique email address.
	// Returns a NotFound error if no user with the given email exists.
	FindByEmail(ctx context.Context, email string) (*User, error)

	// ExistsByEmail checks if a user with the given email address exists.
	// Useful for registration validation without loading the full user entity.
	ExistsByEmail(ctx context.Context, email string) (bool, error)

	// DeleteExternalIdentitiesByUserID soft-deletes all external OAuth identities
	// associated with a user. Called during user deletion for cascade cleanup.
	DeleteExternalIdentitiesByUserID(ctx context.Context, userID uint) error
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

// DeleteExternalIdentitiesByUserID soft-deletes all external identities for a user
func (r *repository) DeleteExternalIdentitiesByUserID(ctx context.Context, userID uint) error {
	return r.DB().WithContext(ctx).Where("user_id = ?", userID).Delete(&ExternalIdentity{}).Error
}
