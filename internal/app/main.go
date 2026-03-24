package app

import (
	"github.com/voidmaindev/go-template/internal/container"
	"github.com/voidmaindev/go-template/internal/domain/audit"
	"github.com/voidmaindev/go-template/internal/domain/auth"
	"github.com/voidmaindev/go-template/internal/domain/example_city"
	"github.com/voidmaindev/go-template/internal/domain/example_country"
	"github.com/voidmaindev/go-template/internal/domain/example_document"
	"github.com/voidmaindev/go-template/internal/domain/email"
	"github.com/voidmaindev/go-template/internal/domain/example_item"
	"github.com/voidmaindev/go-template/internal/domain/rbac"
	"github.com/voidmaindev/go-template/internal/domain/user"
)

// MainApp returns the main application configuration.
// Includes all domains for a full-featured deployment.
func MainApp() *App {
	return &App{
		Name:        "main",
		Description: "Full application with all domains",
		Domains:     mainDomains,
	}
}

func mainDomains() []container.Domain {
	return []container.Domain{
		// Core domains (order matters for dependencies)
		rbac.NewDomain(),  // must be registered first (user depends on rbac.Service)
		user.NewDomain(),  // depends on: rbac
		email.NewDomain(), // standalone, no dependencies
		audit.NewDomain(), // depends on: user (for tokenStore)
		auth.NewDomain(),  // depends on: user, email, rbac, audit

		// Example domains — delete when building your own app
		example_item.NewDomain(),
		example_country.NewDomain(),
		example_city.NewDomain(),     // depends on: example_country
		example_document.NewDomain(), // depends on: example_city, example_item
	}
}
