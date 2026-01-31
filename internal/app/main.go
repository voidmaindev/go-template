package app

import (
	"github.com/voidmaindev/go-template/internal/container"
	"github.com/voidmaindev/go-template/internal/domain/audit"
	"github.com/voidmaindev/go-template/internal/domain/auth"
	"github.com/voidmaindev/go-template/internal/domain/city"
	"github.com/voidmaindev/go-template/internal/domain/country"
	"github.com/voidmaindev/go-template/internal/domain/document"
	"github.com/voidmaindev/go-template/internal/domain/email"
	"github.com/voidmaindev/go-template/internal/domain/item"
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
		item.NewDomain(),
		country.NewDomain(),

		// Domains with dependencies
		city.NewDomain(),     // depends on: country
		document.NewDomain(), // depends on: city, item
	}
}
