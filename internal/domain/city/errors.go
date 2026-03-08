package city

import "github.com/voidmaindev/go-template/internal/common/errors"

const domainName = "city"

// These errors are package-level singletons. NEVER chain builder methods
// (WithOperation, WithContext, etc.) on them at runtime — doing so would
// mutate the shared instance. Return them directly or create new errors
// with errors.New()/errors.Internal() for context-enriched variants.
//
// Domain-specific errors for city operations
var (
	// ErrCityNotFound is returned when a city cannot be found
	ErrCityNotFound = errors.NotFound(domainName, "city")

	// ErrCountryNotFound is returned when the referenced country cannot be found
	ErrCountryNotFound = errors.NotFound(domainName, "country")

	// ErrCityNameExists is returned when trying to create city with existing name in same country
	ErrCityNameExists = errors.AlreadyExists(domainName, "city", "name")
)
