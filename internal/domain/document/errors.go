package document

import "github.com/voidmaindev/go-template/internal/common/errors"

const domainName = "document"

// These errors are package-level singletons. NEVER chain builder methods
// (WithOperation, WithContext, etc.) on them at runtime — doing so would
// mutate the shared instance. Return them directly or create new errors
// with errors.New()/errors.Internal() for context-enriched variants.
//
// Domain-specific errors for document operations
var (
	// ErrDocumentNotFound is returned when a document cannot be found
	ErrDocumentNotFound = errors.NotFound(domainName, "document")

	// ErrDocumentCodeExists is returned when trying to create document with existing code
	ErrDocumentCodeExists = errors.AlreadyExists(domainName, "document", "code")

	// ErrCityNotFound is returned when the referenced city cannot be found
	ErrCityNotFound = errors.NotFound(domainName, "city")

	// ErrItemNotFound is returned when the referenced item cannot be found
	ErrItemNotFound = errors.NotFound(domainName, "item")

	// ErrDocumentItemNotFound is returned when a document item cannot be found
	ErrDocumentItemNotFound = errors.NotFound(domainName, "document item")
)
