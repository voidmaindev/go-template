package migrations

import (
	"github.com/voidmaindev/go-template/internal/domain/user"
	"gorm.io/gorm"
)

// CreateUsersTable creates the users table.
type CreateUsersTable struct{}

func (m *CreateUsersTable) Version() string { return "000001" }
func (m *CreateUsersTable) Name() string    { return "create_users_table" }

func (m *CreateUsersTable) Up(tx *gorm.DB) error {
	if err := tx.Migrator().CreateTable(&user.User{}); err != nil {
		return err
	}
	// Create partial unique index on email (only for non-deleted records)
	// This allows soft-deleted users' emails to be reused
	return tx.Exec("CREATE UNIQUE INDEX idx_users_email ON users (email) WHERE deleted_at IS NULL").Error
}

func (m *CreateUsersTable) Down(tx *gorm.DB) error {
	return tx.Migrator().DropTable(&user.User{})
}
