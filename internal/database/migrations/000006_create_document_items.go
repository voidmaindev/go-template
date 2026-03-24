package migrations

import (
	"github.com/voidmaindev/go-template/internal/domain/example_document"
	"gorm.io/gorm"
)

// CreateDocumentItemsTable creates the document_items table.
type CreateDocumentItemsTable struct{}

func (m *CreateDocumentItemsTable) Version() string { return "000006" }
func (m *CreateDocumentItemsTable) Name() string    { return "create_document_items_table" }

func (m *CreateDocumentItemsTable) Up(tx *gorm.DB) error {
	if err := tx.Migrator().CreateTable(&example_document.DocumentItem{}); err != nil {
		return err
	}

	// Create indexes for foreign keys
	if !tx.Migrator().HasIndex(&example_document.DocumentItem{}, "idx_document_items_document_id") {
		if err := tx.Exec("CREATE INDEX IF NOT EXISTS idx_document_items_document_id ON document_items(document_id)").Error; err != nil {
			return err
		}
	}

	if !tx.Migrator().HasIndex(&example_document.DocumentItem{}, "idx_document_items_item_id") {
		if err := tx.Exec("CREATE INDEX IF NOT EXISTS idx_document_items_item_id ON document_items(item_id)").Error; err != nil {
			return err
		}
	}

	return nil
}

func (m *CreateDocumentItemsTable) Down(tx *gorm.DB) error {
	return tx.Migrator().DropTable(&example_document.DocumentItem{})
}
