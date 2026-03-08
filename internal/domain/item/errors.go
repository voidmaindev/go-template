package item

import "github.com/voidmaindev/go-template/internal/common/errors"

const domainName = "item"

// These errors are package-level singletons. NEVER chain builder methods
// (WithOperation, WithContext, etc.) on them at runtime — doing so would
// mutate the shared instance. Return them directly or create new errors
// with errors.New()/errors.Internal() for context-enriched variants.
//
// Domain-specific errors for item operations
var (
	// ErrItemNotFound is returned when an item cannot be found
	ErrItemNotFound = errors.NotFound(domainName, "item")

	// ErrItemNameExists is returned when trying to create item with existing name
	ErrItemNameExists = errors.AlreadyExists(domainName, "item", "name")
)
