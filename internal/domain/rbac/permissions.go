package rbac

// Action constants for RBAC (standard CRUD terminology)
const (
	ActionRead   = "read"
	ActionCreate = "create"
	ActionUpdate = "update"
	ActionDelete = "delete"
)

// AllActions returns all available actions
func AllActions() []string {
	return []string{ActionRead, ActionCreate, ActionUpdate, ActionDelete}
}

// IsValidAction checks if the given action is valid
func IsValidAction(action string) bool {
	switch action {
	case ActionRead, ActionCreate, ActionUpdate, ActionDelete:
		return true
	default:
		return false
	}
}

// ProtectedDomains returns domains that should have restricted access
// (only read access for full_writer)
func ProtectedDomains() []string {
	return []string{"user", "rbac"}
}

// IsProtectedDomain checks if a domain is protected
func IsProtectedDomain(domain string) bool {
	for _, d := range ProtectedDomains() {
		if d == domain {
			return true
		}
	}
	return false
}

// SystemRoles returns the codes of system-defined roles that cannot be deleted
func SystemRoles() []string {
	return []string{RoleCodeAdmin, RoleCodeFullReader, RoleCodeFullWriter, RoleCodeUser}
}

// IsSystemRole checks if a role code is a system role
func IsSystemRole(code string) bool {
	for _, r := range SystemRoles() {
		if r == code {
			return true
		}
	}
	return false
}

// System role codes
const (
	RoleCodeAdmin      = "admin"
	RoleCodeFullReader = "full_reader"
	RoleCodeFullWriter = "full_writer"
	RoleCodeUser       = "user"
)
