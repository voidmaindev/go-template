package migrations

import (
	"github.com/voidmaindev/go-template/internal/domain/country"
	"gorm.io/gorm"
)

// CreateCountriesTable creates the countries table.
type CreateCountriesTable struct{}

func (m *CreateCountriesTable) Version() string { return "000003" }
func (m *CreateCountriesTable) Name() string    { return "create_countries_table" }

func (m *CreateCountriesTable) Up(tx *gorm.DB) error {
	return tx.Migrator().CreateTable(&country.Country{})
}

func (m *CreateCountriesTable) Down(tx *gorm.DB) error {
	return tx.Migrator().DropTable(&country.Country{})
}
