package rbac

import "github.com/voidmaindev/go-template/internal/common/errors"

const domainName = "rbac"

// These errors are package-level singletons. NEVER chain builder methods
// (WithOperation, WithContext, etc.) on them at runtime — doing so would
// mutate the shared instance. Return them directly or create new errors
// with errors.New()/errors.Internal() for context-enriched variants.
//
// Domain-specific errors for RBAC operations
var (
	// ErrRoleNotFound is returned when a role cannot be found
	ErrRoleNotFound = errors.NotFound(domainName, "role")

	// ErrRoleCodeExists is returned when trying to create a role with an existing code
	ErrRoleCodeExists = errors.AlreadyExists(domainName, "role", "code")

	// ErrSystemRoleCannotBeDeleted is returned when trying to delete a system role
	ErrSystemRoleCannotBeDeleted = errors.New(domainName, errors.CodeForbidden).
					WithMessage("system roles cannot be deleted")

	// ErrSystemRoleCannotBeModified is returned when trying to modify a system role's permissions
	ErrSystemRoleCannotBeModified = errors.New(domainName, errors.CodeForbidden).
					WithMessage("system role permissions cannot be modified")

	// ErrCannotRemoveLastAdmin is returned when trying to remove the last admin
	ErrCannotRemoveLastAdmin = errors.New(domainName, errors.CodeForbidden).
					WithMessage("cannot remove the last admin user")

	// ErrRoleAlreadyAssigned is returned when a role is already assigned to a user
	ErrRoleAlreadyAssigned = errors.New(domainName, errors.CodeConflict).
				WithMessage("role is already assigned to this user")

	// ErrRoleNotAssigned is returned when trying to remove a role that is not assigned
	ErrRoleNotAssigned = errors.New(domainName, errors.CodeNotFound).
				WithMessage("role is not assigned to this user")

	// ErrInvalidAction is returned when an invalid action is provided
	ErrInvalidAction = errors.Validation(domainName, "invalid action")

	// ErrInvalidDomain is returned when an invalid domain is provided
	ErrInvalidDomain = errors.Validation(domainName, "invalid domain")

	// ErrUserNotFound is returned when a user cannot be found
	ErrUserNotFound = errors.NotFound(domainName, "user")

	// ErrEnforcerNotInitialized is returned when the enforcer is not initialized
	ErrEnforcerNotInitialized = errors.New(domainName, errors.CodeInternal).
					WithStack().
					WithMessage("RBAC enforcer not initialized")
)
