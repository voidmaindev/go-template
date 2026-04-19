package example_document

import "github.com/voidmaindev/go-template/internal/common/errors"

const domainName = "document"

// Domain-specific errors for document operations.
// Builder methods (WithOperation, WithContext, etc.) are clone-on-write safe.
var (
	// ErrDocumentNotFound is returned when a document cannot be found
	ErrDocumentNotFound = errors.NotFound(domainName, "document")

	// ErrDocumentCodeExists is returned when trying to create document with existing code
	ErrDocumentCodeExists = errors.AlreadyExists(domainName, "document", "code")

	// ErrInvalidCityRef is returned when a request payload references a city ID
	// that does not exist. Input validation, so maps to HTTP 400.
	ErrInvalidCityRef = errors.Validation(domainName, "referenced city not found")

	// ErrInvalidItemRef is returned when a request payload references an item ID
	// that does not exist. Input validation, so maps to HTTP 400.
	ErrInvalidItemRef = errors.Validation(domainName, "referenced item not found")

	// ErrDocumentItemNotFound is returned when a document item cannot be found
	ErrDocumentItemNotFound = errors.NotFound(domainName, "document item")
)
