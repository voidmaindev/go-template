package document

import "github.com/voidmaindev/go-template/internal/common/errors"

const domainName = "document"

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
