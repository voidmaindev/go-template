package example_city

import "github.com/voidmaindev/go-template/internal/common/errors"

const domainName = "city"

// Domain-specific errors for city operations.
// Builder methods (WithOperation, WithContext, etc.) are clone-on-write safe.
var (
	// ErrCityNotFound is returned when a city cannot be found.
	ErrCityNotFound = errors.NotFound(domainName, "city")

	// ErrCountryNotFound is returned when the country resource itself cannot be
	// resolved — e.g. a country ID came from the URL (ListByCountry) and no
	// such country exists. Maps to HTTP 404.
	ErrCountryNotFound = errors.NotFound(domainName, "country")

	// ErrInvalidCountryRef is returned when a request payload references a
	// country ID that does not exist. This is input validation, not a missing
	// resource, so it maps to HTTP 400.
	ErrInvalidCountryRef = errors.Validation(domainName, "referenced country not found")

	// ErrCityNameExists is returned when trying to create city with existing name in same country
	ErrCityNameExists = errors.AlreadyExists(domainName, "city", "name")
)
