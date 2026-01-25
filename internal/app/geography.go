package app

import (
	"github.com/voidmaindev/go-template/internal/container"
	"github.com/voidmaindev/go-template/internal/domain/city"
	"github.com/voidmaindev/go-template/internal/domain/country"
	"github.com/voidmaindev/go-template/internal/domain/rbac"
	"github.com/voidmaindev/go-template/internal/domain/user"
)

// GeographyApp returns the geography service application configuration.
// A lightweight service focused on country and city management.
func GeographyApp() *App {
	return &App{
		Name:        "geography",
		Description: "Geography service (countries and cities)",
		Domains:     geographyDomains,
	}
}

func geographyDomains() []container.Domain {
	return []container.Domain{
		// Core domains (user depends on rbac)
		rbac.NewDomain(), // must be registered first (user depends on rbac.Service)
		user.NewDomain(), // depends on: rbac

		// Geography domains
		country.NewDomain(),
		city.NewDomain(), // depends on: country
	}
}
