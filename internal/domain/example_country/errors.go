package example_country

import "github.com/voidmaindev/go-template/internal/common/errors"

const domainName = "country"

// Domain-specific errors for country operations.
// Builder methods (WithOperation, WithContext, etc.) are clone-on-write safe.
var (
	// ErrCountryNotFound is returned when a country cannot be found
	ErrCountryNotFound = errors.NotFound(domainName, "country")

	// ErrCountryCodeExists is returned when trying to create country with existing code
	ErrCountryCodeExists = errors.AlreadyExists(domainName, "country", "code")

	// ErrCountryNameExists is returned when trying to create country with existing name
	ErrCountryNameExists = errors.AlreadyExists(domainName, "country", "name")
)
