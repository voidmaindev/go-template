package domain

import (
	"github.com/voidmaindev/GoTemplate/internal/container"
	"github.com/voidmaindev/GoTemplate/internal/domain/city"
	"github.com/voidmaindev/GoTemplate/internal/domain/country"
	"github.com/voidmaindev/GoTemplate/internal/domain/document"
	"github.com/voidmaindev/GoTemplate/internal/domain/item"
	"github.com/voidmaindev/GoTemplate/internal/domain/user"
)

// All returns all domains in registration order.
// Order matters: domains with dependencies must come after their dependencies.
// For example: city depends on country, so country comes first.
func All() []container.Domain {
	return []container.Domain{
		// Core domains (no dependencies)
		user.NewDomain(),
		item.NewDomain(),
		country.NewDomain(),

		// Domains with dependencies
		city.NewDomain(),    // depends on: country
		document.NewDomain(), // depends on: city, item
	}
}
