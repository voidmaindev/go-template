package user

import (
	"context"
	"errors"
	"time"

	"github.com/voidmaindev/go-template/internal/common"
	commonerrors "github.com/voidmaindev/go-template/internal/common/errors"
	"gorm.io/gorm"
)

// repoDomain is used for wrapping repository errors with a consistent domain tag.
const repoDomain = "user"

// wrapRepoErr wraps a raw DB error with domain + operation context. Returns nil on nil.
func wrapRepoErr(op string, err error) error {
	if err == nil {
		return nil
	}
	return commonerrors.Internal(repoDomain, err).WithOperation(op)
}

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

	// UpdateEmailVerifiedAt sets the email_verified_at timestamp for a user.
	UpdateEmailVerifiedAt(ctx context.Context, id uint, verifiedAt time.Time) error

	// UpdatePassword updates a user's password hash.
	UpdatePassword(ctx context.Context, id uint, hashedPassword string) error

	// FindExternalIdentityByProvider finds an external identity by provider and provider ID.
	FindExternalIdentityByProvider(ctx context.Context, provider, providerID string) (*ExternalIdentity, error)

	// FindExternalIdentitiesByUserID returns all external identities for a user.
	FindExternalIdentitiesByUserID(ctx context.Context, userID uint) ([]ExternalIdentity, error)

	// CountExternalIdentitiesByUserID counts external identities for a user.
	CountExternalIdentitiesByUserID(ctx context.Context, userID uint) (int64, error)

	// CreateExternalIdentity creates a new external identity record.
	CreateExternalIdentity(ctx context.Context, identity *ExternalIdentity) error

	// DeleteExternalIdentityByProvider deletes an identity for a specific user and provider.
	// Returns the number of rows affected.
	DeleteExternalIdentityByProvider(ctx context.Context, userID uint, provider string) (int64, error)

	// BeginTx starts a GORM transaction and returns a repository scoped to it.
	// The caller is responsible for committing or rolling back via the returned *gorm.DB.
	BeginTx(ctx context.Context) (Repository, *gorm.DB, error)
}

// repository implements the Repository interface
type repository struct {
	*common.BaseRepository[User]
}

// NewRepository creates a new user repository
func NewRepository(db *gorm.DB) Repository {
	return &repository{
		BaseRepository: common.NewBaseRepository[User](db, "user"),
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
	return wrapRepoErr("DeleteExternalIdentitiesByUserID",
		r.DB().WithContext(ctx).Where("user_id = ?", userID).Delete(&ExternalIdentity{}).Error)
}

// UpdateEmailVerifiedAt sets the email_verified_at timestamp for a user
func (r *repository) UpdateEmailVerifiedAt(ctx context.Context, id uint, verifiedAt time.Time) error {
	return wrapRepoErr("UpdateEmailVerifiedAt",
		r.DB().WithContext(ctx).Model(&User{}).Where("id = ?", id).Update("email_verified_at", verifiedAt).Error)
}

// UpdatePassword updates a user's password hash
func (r *repository) UpdatePassword(ctx context.Context, id uint, hashedPassword string) error {
	return wrapRepoErr("UpdatePassword",
		r.DB().WithContext(ctx).Model(&User{}).Where("id = ?", id).Update("password", hashedPassword).Error)
}

// FindExternalIdentityByProvider finds an external identity by provider and provider ID
func (r *repository) FindExternalIdentityByProvider(ctx context.Context, provider, providerID string) (*ExternalIdentity, error) {
	var identity ExternalIdentity
	err := r.DB().WithContext(ctx).
		Where("provider = ? AND provider_id = ?", provider, providerID).
		First(&identity).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, commonerrors.NotFound(repoDomain, "external_identity")
		}
		return nil, wrapRepoErr("FindExternalIdentityByProvider", err)
	}
	return &identity, nil
}

// FindExternalIdentitiesByUserID returns all external identities for a user
func (r *repository) FindExternalIdentitiesByUserID(ctx context.Context, userID uint) ([]ExternalIdentity, error) {
	var identities []ExternalIdentity
	if err := r.DB().WithContext(ctx).Where("user_id = ?", userID).Find(&identities).Error; err != nil {
		return nil, wrapRepoErr("FindExternalIdentitiesByUserID", err)
	}
	return identities, nil
}

// CountExternalIdentitiesByUserID counts external identities for a user
func (r *repository) CountExternalIdentitiesByUserID(ctx context.Context, userID uint) (int64, error) {
	var count int64
	err := r.DB().WithContext(ctx).Model(&ExternalIdentity{}).
		Where("user_id = ?", userID).Count(&count).Error
	return count, wrapRepoErr("CountExternalIdentitiesByUserID", err)
}

// CreateExternalIdentity creates a new external identity record
func (r *repository) CreateExternalIdentity(ctx context.Context, identity *ExternalIdentity) error {
	return wrapRepoErr("CreateExternalIdentity",
		r.DB().WithContext(ctx).Create(identity).Error)
}

// DeleteExternalIdentityByProvider deletes an identity for a specific user and provider
func (r *repository) DeleteExternalIdentityByProvider(ctx context.Context, userID uint, provider string) (int64, error) {
	result := r.DB().WithContext(ctx).
		Where("user_id = ? AND provider = ?", userID, provider).
		Delete(&ExternalIdentity{})
	return result.RowsAffected, wrapRepoErr("DeleteExternalIdentityByProvider", result.Error)
}

// BeginTx starts a GORM transaction and returns a repository scoped to it
func (r *repository) BeginTx(ctx context.Context) (Repository, *gorm.DB, error) {
	tx := r.DB().WithContext(ctx).Begin()
	if tx.Error != nil {
		return nil, nil, wrapRepoErr("BeginTx", tx.Error)
	}
	txRepo := &repository{
		BaseRepository: common.NewBaseRepository[User](tx, "user"),
	}
	return txRepo, tx, nil
}
