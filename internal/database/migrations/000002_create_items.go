package migrations

import (
	"github.com/voidmaindev/go-template/internal/domain/example_item"
	"gorm.io/gorm"
)

// CreateItemsTable creates the items table.
type CreateItemsTable struct{}

func (m *CreateItemsTable) Version() string { return "000002" }
func (m *CreateItemsTable) Name() string    { return "create_items_table" }

func (m *CreateItemsTable) Up(tx *gorm.DB) error {
	return tx.Migrator().CreateTable(&example_item.Item{})
}

func (m *CreateItemsTable) Down(tx *gorm.DB) error {
	return tx.Migrator().DropTable(&example_item.Item{})
}
