package country

import "github.com/voidmaindev/go-template/internal/common/errors"

const domainName = "country"

// These errors are package-level singletons. NEVER chain builder methods
// (WithOperation, WithContext, etc.) on them at runtime — doing so would
// mutate the shared instance. Return them directly or create new errors
// with errors.New()/errors.Internal() for context-enriched variants.
//
// Domain-specific errors for country operations
var (
	// ErrCountryNotFound is returned when a country cannot be found
	ErrCountryNotFound = errors.NotFound(domainName, "country")

	// ErrCountryCodeExists is returned when trying to create country with existing code
	ErrCountryCodeExists = errors.AlreadyExists(domainName, "country", "code")

	// ErrCountryNameExists is returned when trying to create country with existing name
	ErrCountryNameExists = errors.AlreadyExists(domainName, "country", "name")
)
