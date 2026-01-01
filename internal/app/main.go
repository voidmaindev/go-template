package app

import (
	"github.com/voidmaindev/GoTemplate/internal/container"
	"github.com/voidmaindev/GoTemplate/internal/domain/city"
	"github.com/voidmaindev/GoTemplate/internal/domain/country"
	"github.com/voidmaindev/GoTemplate/internal/domain/document"
	"github.com/voidmaindev/GoTemplate/internal/domain/item"
	"github.com/voidmaindev/GoTemplate/internal/domain/user"
)

func init() {
	Register(&App{
		Name:        "main",
		Description: "Full application with all domains",
		Domains: func() []container.Domain {
			return []container.Domain{
				// Core domains (no dependencies)
				user.NewDomain(),
				item.NewDomain(),
				country.NewDomain(),

				// Domains with dependencies
				city.NewDomain(),     // depends on: country
				document.NewDomain(), // depends on: city, item
			}
		},
	})
}
