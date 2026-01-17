package app

import (
	"github.com/voidmaindev/go-template/internal/container"
	"github.com/voidmaindev/go-template/internal/domain/city"
	"github.com/voidmaindev/go-template/internal/domain/country"
	"github.com/voidmaindev/go-template/internal/domain/rbac"
	"github.com/voidmaindev/go-template/internal/domain/user"
)

func init() {
	Register(&App{
		Name:        "geography",
		Description: "Geography service (countries and cities)",
		Domains: func() []container.Domain {
			return []container.Domain{
				// Core domains (user depends on rbac)
				rbac.NewDomain(), // must be registered first (user depends on rbac.Service)
				user.NewDomain(), // depends on: rbac

				// Geography domains
				country.NewDomain(),
				city.NewDomain(), // depends on: country
			}
		},
	})
}
