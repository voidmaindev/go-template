package app

import (
	"github.com/voidmaindev/go-template/internal/container"
	"github.com/voidmaindev/go-template/internal/domain/city"
	"github.com/voidmaindev/go-template/internal/domain/country"
	"github.com/voidmaindev/go-template/internal/domain/document"
	"github.com/voidmaindev/go-template/internal/domain/item"
	"github.com/voidmaindev/go-template/internal/domain/user"
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
