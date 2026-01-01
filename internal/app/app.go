package app

import (
	"sort"

	"github.com/voidmaindev/GoTemplate/internal/container"
)

// App represents a runnable application with specific domains
type App struct {
	Name        string
	Description string
	Domains     func() []container.Domain
}

// Registry of all available apps
var Registry = make(map[string]*App)

// Register adds an app to the registry
func Register(a *App) {
	Registry[a.Name] = a
}

// Get returns an app by name
func Get(name string) *App {
	return Registry[name]
}

// List returns all registered app names sorted alphabetically
func List() []string {
	names := make([]string, 0, len(Registry))
	for name := range Registry {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
