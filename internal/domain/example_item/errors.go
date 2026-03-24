package example_item

import "github.com/voidmaindev/go-template/internal/common/errors"

const domainName = "item"

// Domain-specific errors for item operations.
// Builder methods (WithOperation, WithContext, etc.) are clone-on-write safe.
var (
	// ErrItemNotFound is returned when an item cannot be found
	ErrItemNotFound = errors.NotFound(domainName, "item")

	// ErrItemNameExists is returned when trying to create item with existing name
	ErrItemNameExists = errors.AlreadyExists(domainName, "item", "name")
)
