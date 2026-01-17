package app

import (
	"github.com/voidmaindev/go-template/internal/container"
	"github.com/voidmaindev/go-template/internal/domain/city"
	"github.com/voidmaindev/go-template/internal/domain/country"
	"github.com/voidmaindev/go-template/internal/domain/document"
	"github.com/voidmaindev/go-template/internal/domain/item"
	"github.com/voidmaindev/go-template/internal/domain/rbac"
	"github.com/voidmaindev/go-template/internal/domain/user"
)

func init() {
	Register(&App{
		Name:        "main",
		Description: "Full application with all domains",
		Domains: func() []container.Domain {
			return []container.Domain{
				// Core domains (no dependencies)
				rbac.NewDomain(), // must be registered first (user depends on rbac.Service)
				user.NewDomain(), // depends on: rbac
				item.NewDomain(),
				country.NewDomain(),

				// Domains with dependencies
				city.NewDomain(),     // depends on: country
				document.NewDomain(), // depends on: city, item
			}
		},
	})
}
