package city

import "github.com/voidmaindev/go-template/internal/common/errors"

const domainName = "city"

// Domain-specific errors for city operations.
// Builder methods (WithOperation, WithContext, etc.) are clone-on-write safe.
var (
	// ErrCityNotFound is returned when a city cannot be found
	ErrCityNotFound = errors.NotFound(domainName, "city")

	// ErrCountryNotFound is returned when the referenced country cannot be found
	ErrCountryNotFound = errors.NotFound(domainName, "country")

	// ErrCityNameExists is returned when trying to create city with existing name in same country
	ErrCityNameExists = errors.AlreadyExists(domainName, "city", "name")
)
