package migrations

import (
	"github.com/voidmaindev/go-template/internal/domain/example_city"
	"gorm.io/gorm"
)

// CreateCitiesTable creates the cities table with foreign key to countries.
type CreateCitiesTable struct{}

func (m *CreateCitiesTable) Version() string { return "000004" }
func (m *CreateCitiesTable) Name() string    { return "create_cities_table" }

func (m *CreateCitiesTable) Up(tx *gorm.DB) error {
	if err := tx.Migrator().CreateTable(&example_city.City{}); err != nil {
		return err
	}

	// Create index for country_id if not already created by GORM
	if !tx.Migrator().HasIndex(&example_city.City{}, "idx_cities_country_id") {
		return tx.Exec("CREATE INDEX IF NOT EXISTS idx_cities_country_id ON cities(country_id)").Error
	}
	return nil
}

func (m *CreateCitiesTable) Down(tx *gorm.DB) error {
	return tx.Migrator().DropTable(&example_city.City{})
}
