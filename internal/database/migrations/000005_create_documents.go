package migrations

import (
	"github.com/voidmaindev/go-template/internal/domain/example_document"
	"gorm.io/gorm"
)

// CreateDocumentsTable creates the documents table.
type CreateDocumentsTable struct{}

func (m *CreateDocumentsTable) Version() string { return "000005" }
func (m *CreateDocumentsTable) Name() string    { return "create_documents_table" }

func (m *CreateDocumentsTable) Up(tx *gorm.DB) error {
	// Create table without Items relation (DocumentItem table will be created separately)
	type DocumentWithoutItems struct {
		example_document.Document
	}

	// Use raw Document model - GORM will handle the table structure
	if err := tx.Migrator().CreateTable(&example_document.Document{}); err != nil {
		return err
	}

	// Create index for city_id if not already created
	if !tx.Migrator().HasIndex(&example_document.Document{}, "idx_documents_city_id") {
		return tx.Exec("CREATE INDEX IF NOT EXISTS idx_documents_city_id ON documents(city_id)").Error
	}
	return nil
}

func (m *CreateDocumentsTable) Down(tx *gorm.DB) error {
	return tx.Migrator().DropTable(&example_document.Document{})
}
