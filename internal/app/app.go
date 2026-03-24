package app

import (
	"sort"

	"github.com/voidmaindev/go-template/internal/container"
)

// App represents a runnable application with specific domains
type App struct {
	Name        string
	Description string
	Domains     func() []container.Domain
}

// All returns all available app configurations.
// This is the single source of truth for available apps.
func All() map[string]*App {
	main := MainApp()
	geo := ExampleGeographyApp()
	return map[string]*App{
		main.Name: main,
		geo.Name:  geo,
	}
}

// Get returns an app by name, or nil if not found.
func Get(name string) *App {
	return All()[name]
}

// List returns all registered app names sorted alphabetically.
func List() []string {
	apps := All()
	names := make([]string, 0, len(apps))
	for name := range apps {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
