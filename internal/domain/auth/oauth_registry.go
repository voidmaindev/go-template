package auth

import (
	"github.com/voidmaindev/go-template/internal/config"
)

// OAuthRegistry manages OAuth providers
type OAuthRegistry struct {
	providers map[string]OAuthProvider
}

// NewOAuthRegistry creates a new OAuth registry with configured providers
func NewOAuthRegistry(cfg *config.OAuthConfig) *OAuthRegistry {
	registry := &OAuthRegistry{
		providers: make(map[string]OAuthProvider),
	}

	// Register enabled providers
	if cfg.Google.Enabled {
		registry.providers["google"] = NewGoogleProvider(&cfg.Google)
	}

	if cfg.Facebook.Enabled {
		registry.providers["facebook"] = NewFacebookProvider(&cfg.Facebook)
	}

	if cfg.Apple.Enabled {
		registry.providers["apple"] = NewAppleProvider(&cfg.Apple)
	}

	return registry
}

// Get returns a provider by name, or nil if not found/disabled
func (r *OAuthRegistry) Get(name string) OAuthProvider {
	return r.providers[name]
}

// List returns all registered provider names
func (r *OAuthRegistry) List() []string {
	names := make([]string, 0, len(r.providers))
	for name := range r.providers {
		names = append(names, name)
	}
	return names
}
