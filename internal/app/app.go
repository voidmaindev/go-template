package app

import (
	"sort"
	"sync"

	"github.com/voidmaindev/go-template/internal/container"
)

// App represents a runnable application with specific domains
type App struct {
	Name        string
	Description string
	Domains     func() []container.Domain
}

// Registry of all available apps (protected by registryMu)
var (
	registry   = make(map[string]*App)
	registryMu sync.RWMutex
)

// Register adds an app to the registry (thread-safe)
func Register(a *App) {
	registryMu.Lock()
	defer registryMu.Unlock()
	registry[a.Name] = a
}

// Get returns an app by name (thread-safe)
func Get(name string) *App {
	registryMu.RLock()
	defer registryMu.RUnlock()
	return registry[name]
}

// List returns all registered app names sorted alphabetically (thread-safe)
func List() []string {
	registryMu.RLock()
	defer registryMu.RUnlock()
	names := make([]string, 0, len(registry))
	for name := range registry {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
