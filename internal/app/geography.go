package app

import (
	"github.com/voidmaindev/GoTemplate/internal/container"
	"github.com/voidmaindev/GoTemplate/internal/domain/city"
	"github.com/voidmaindev/GoTemplate/internal/domain/country"
	"github.com/voidmaindev/GoTemplate/internal/domain/user"
)

func init() {
	Register(&App{
		Name:        "geography",
		Description: "Geography service (countries and cities)",
		Domains: func() []container.Domain {
			return []container.Domain{
				// User domain required for JWT auth
				user.NewDomain(),

				// Geography domains
				country.NewDomain(),
				city.NewDomain(), // depends on: country
			}
		},
	})
}
