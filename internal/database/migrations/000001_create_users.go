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
	return tx.Migrator().CreateTable(&user.User{})
}

func (m *CreateUsersTable) Down(tx *gorm.DB) error {
	return tx.Migrator().DropTable(&user.User{})
}
