package rbac

import (
	"context"

	"github.com/casbin/casbin/v3"
	"github.com/voidmaindev/go-template/internal/common"
	"github.com/voidmaindev/go-template/internal/common/errors"
	"github.com/voidmaindev/go-template/internal/common/filter"
	"github.com/voidmaindev/go-template/internal/common/logging"
)

// DomainProvider is an interface for getting registered domain names
type DomainProvider interface {
	GetDomainNames() []string
}

// Service defines the interface for RBAC operations
type Service interface {
	// Role management
	CreateRole(ctx context.Context, req *CreateRoleRequest) (*Role, error)
	GetRoleByCode(ctx context.Context, code string) (*RoleWithPermissions, error)
	ListRoles(ctx context.Context, params *filter.Params) (*common.FilteredResult[Role], error)
	UpdateRolePermissions(ctx context.Context, code string, req *UpdateRolePermissionsRequest) (*RoleWithPermissions, error)
	DeleteRole(ctx context.Context, code string) error

	// User-role management
	GetUserRoles(ctx context.Context, userID uint) ([]UserRoleResponse, error)
	AssignRole(ctx context.Context, userID uint, roleCode string) error
	AssignRoleInTx(tx *casbin.Transaction, ctx context.Context, userID uint, roleCode string) error
	RemoveRole(ctx context.Context, userID uint, roleCode string) error

	// Permission checking
	CheckPermission(ctx context.Context, userID uint, domain, action string) (bool, error)

	// Discovery
	GetDomains(ctx context.Context) []DomainResponse
	GetActions(ctx context.Context) []string

	// Sync system roles policies with registered domains
	SyncGlobalRoles(ctx context.Context) error

	// Admin check
	CountAdminUsers(ctx context.Context) (int, error)

	// Transaction support
	GetTransactionalEnforcer() *casbin.TransactionalEnforcer
}

type service struct {
	repo           Repository
	enforcer       *casbin.TransactionalEnforcer
	domainProvider DomainProvider
	logger         *logging.Logger
}

// NewService creates a new RBAC service
func NewService(repo Repository, enforcer *casbin.TransactionalEnforcer, domainProvider DomainProvider) Service {
	return &service{
		repo:           repo,
		enforcer:       enforcer,
		domainProvider: domainProvider,
		logger:         logging.New(domainName),
	}
}

// CreateRole creates a new role with permissions
func (s *service) CreateRole(ctx context.Context, req *CreateRoleRequest) (*Role, error) {
	// Check if role code already exists
	exists, err := s.repo.ExistsByCode(ctx, req.Code)
	if err != nil {
		return nil, errors.Internal(domainName, err).WithOperation("CreateRole")
	}
	if exists {
		return nil, ErrRoleCodeExists
	}

	// Validate domains
	validDomains := s.getValidDomains()
	for _, perm := range req.Permissions {
		if !s.isValidDomain(perm.Domain, validDomains) {
			return nil, ErrInvalidDomain
		}
		for _, action := range perm.Actions {
			if !IsValidAction(action) {
				return nil, ErrInvalidAction
			}
		}
	}

	// Create role metadata
	role := &Role{
		Code:        req.Code,
		Name:        req.Name,
		Description: req.Description,
		IsSystem:    false,
	}

	// Track added policies for rollback if needed
	var addedPolicies [][]string

	// Use transaction for atomicity
	err = s.repo.Transaction(ctx, func(txRepo common.Repository[Role]) error {
		// Create role within transaction
		if err := txRepo.Create(ctx, role); err != nil {
			return err
		}

		// Add policies to Casbin (in-memory)
		for _, perm := range req.Permissions {
			for _, action := range perm.Actions {
				added, err := s.enforcer.AddPolicy(req.Code, perm.Domain, action)
				if err != nil {
					s.logger.Error(ctx, "failed to add policy", err,
						"role", req.Code,
						"domain", perm.Domain,
						"action", action,
					)
					return err
				}
				if added {
					addedPolicies = append(addedPolicies, []string{req.Code, perm.Domain, action})
				}
			}
		}

		// Save policies - if this fails, transaction rolls back
		if err := s.enforcer.SavePolicy(); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		// Rollback: remove any policies we added to Casbin's in-memory state
		for _, policy := range addedPolicies {
			s.enforcer.RemovePolicy(policy[0], policy[1], policy[2])
		}
		return nil, errors.Internal(domainName, err).WithOperation("CreateRole")
	}

	return role, nil
}

// GetRoleByCode gets a role by code with its permissions
func (s *service) GetRoleByCode(ctx context.Context, code string) (*RoleWithPermissions, error) {
	role, err := s.repo.FindByCode(ctx, code)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, ErrRoleNotFound
		}
		return nil, errors.Internal(domainName, err).WithOperation("GetRoleByCode")
	}

	permissions := s.getRolePermissions(code)

	return &RoleWithPermissions{
		Role:        role,
		Permissions: permissions,
	}, nil
}

// ListRoles lists all roles with filtering
func (s *service) ListRoles(ctx context.Context, params *filter.Params) (*common.FilteredResult[Role], error) {
	roles, total, err := s.repo.FindAllFiltered(ctx, params)
	if err != nil {
		return nil, errors.Internal(domainName, err).WithOperation("ListRoles")
	}
	return common.NewFilteredResult(roles, total, params), nil
}

// UpdateRolePermissions updates the permissions of a role atomically
func (s *service) UpdateRolePermissions(ctx context.Context, code string, req *UpdateRolePermissionsRequest) (*RoleWithPermissions, error) {
	role, err := s.repo.FindByCode(ctx, code)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, ErrRoleNotFound
		}
		return nil, errors.Internal(domainName, err).WithOperation("UpdateRolePermissions")
	}

	// Cannot modify system roles
	if role.IsSystem {
		return nil, ErrSystemRoleCannotBeModified
	}

	// Validate domains and actions before any modifications
	validDomains := s.getValidDomains()
	for _, perm := range req.Permissions {
		if !s.isValidDomain(perm.Domain, validDomains) {
			return nil, ErrInvalidDomain
		}
		for _, action := range perm.Actions {
			if !IsValidAction(action) {
				return nil, ErrInvalidAction
			}
		}
	}

	// Track existing policies for rollback
	existingPolicies, _ := s.enforcer.GetFilteredPolicy(0, code)

	// Track added policies for rollback if needed
	var addedPolicies [][]string

	// Use transaction wrapper for atomicity
	err = s.repo.Transaction(ctx, func(txRepo common.Repository[Role]) error {
		// Remove all existing policies for this role
		if _, err := s.enforcer.RemoveFilteredPolicy(0, code); err != nil {
			return err
		}

		// Add new policies
		for _, perm := range req.Permissions {
			for _, action := range perm.Actions {
				added, err := s.enforcer.AddPolicy(code, perm.Domain, action)
				if err != nil {
					s.logger.Error(ctx, "failed to add policy", err,
						"role", code,
						"domain", perm.Domain,
						"action", action,
					)
					return err
				}
				if added {
					addedPolicies = append(addedPolicies, []string{code, perm.Domain, action})
				}
			}
		}

		// Save policies - if this fails, we need to rollback
		if err := s.enforcer.SavePolicy(); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		// Rollback: remove any policies we added and restore the original policies
		for _, policy := range addedPolicies {
			s.enforcer.RemovePolicy(policy[0], policy[1], policy[2])
		}
		// Restore original policies
		for _, policy := range existingPolicies {
			if len(policy) >= 3 {
				s.enforcer.AddPolicy(policy[0], policy[1], policy[2])
			}
		}
		// Try to save the restored state (best effort)
		if saveErr := s.enforcer.SavePolicy(); saveErr != nil {
			s.logger.Error(ctx, "failed to save restored policies during rollback", saveErr)
		}
		return nil, errors.Internal(domainName, err).WithOperation("UpdateRolePermissions")
	}

	permissions := s.getRolePermissions(code)

	return &RoleWithPermissions{
		Role:        role,
		Permissions: permissions,
	}, nil
}

// DeleteRole deletes a role atomically with rollback support
func (s *service) DeleteRole(ctx context.Context, code string) error {
	role, err := s.repo.FindByCode(ctx, code)
	if err != nil {
		if errors.IsNotFound(err) {
			return ErrRoleNotFound
		}
		return errors.Internal(domainName, err).WithOperation("DeleteRole")
	}

	// Cannot delete system roles
	if role.IsSystem {
		return ErrSystemRoleCannotBeDeleted
	}

	// Capture existing policies for rollback
	existingPolicies, _ := s.enforcer.GetFilteredPolicy(0, code)
	existingGroupings, _ := s.enforcer.GetFilteredGroupingPolicy(1, code)

	// Use transaction for atomicity
	err = s.repo.Transaction(ctx, func(txRepo common.Repository[Role]) error {
		// Remove all policies for this role
		if _, err := s.enforcer.RemoveFilteredPolicy(0, code); err != nil {
			return err
		}

		// Remove all user-role assignments
		if _, err := s.enforcer.RemoveFilteredGroupingPolicy(1, code); err != nil {
			return err
		}

		// Save policy changes
		if err := s.enforcer.SavePolicy(); err != nil {
			return err
		}

		// Delete role metadata within transaction
		if err := txRepo.Delete(ctx, role.ID); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		// Rollback: restore policies and groupings
		for _, policy := range existingPolicies {
			if len(policy) >= 3 {
				s.enforcer.AddPolicy(policy[0], policy[1], policy[2])
			}
		}
		for _, grouping := range existingGroupings {
			if len(grouping) >= 2 {
				s.enforcer.AddGroupingPolicy(grouping[0], grouping[1])
			}
		}
		// Best effort save of restored state
		if saveErr := s.enforcer.SavePolicy(); saveErr != nil {
			s.logger.Error(ctx, "failed to save restored policies during rollback", saveErr)
		}
		return errors.Internal(domainName, err).WithOperation("DeleteRole")
	}

	return nil
}

// GetUserRoles gets all roles assigned to a user
func (s *service) GetUserRoles(ctx context.Context, userID uint) ([]UserRoleResponse, error) {
	subject := FormatUserSubject(userID)
	roleCodes, err := s.enforcer.GetRolesForUser(subject)
	if err != nil {
		return nil, errors.Internal(domainName, err).WithOperation("GetUserRoles")
	}

	var roles []UserRoleResponse
	for _, code := range roleCodes {
		role, err := s.repo.FindByCode(ctx, code)
		if err != nil {
			// Role might not exist in metadata, skip
			continue
		}
		roles = append(roles, UserRoleResponse{
			Code: role.Code,
			Name: role.Name,
		})
	}

	return roles, nil
}

// AssignRole assigns a role to a user
func (s *service) AssignRole(ctx context.Context, userID uint, roleCode string) error {
	// Check if role exists
	_, err := s.repo.FindByCode(ctx, roleCode)
	if err != nil {
		if errors.IsNotFound(err) {
			return ErrRoleNotFound
		}
		return errors.Internal(domainName, err).WithOperation("AssignRole")
	}

	subject := FormatUserSubject(userID)

	// Check if already assigned
	hasRole, err := s.enforcer.HasRoleForUser(subject, roleCode)
	if err != nil {
		return errors.Internal(domainName, err).WithOperation("AssignRole.HasRole")
	}
	if hasRole {
		return ErrRoleAlreadyAssigned
	}

	// Assign role
	if _, err := s.enforcer.AddRoleForUser(subject, roleCode); err != nil {
		return errors.Internal(domainName, err).WithOperation("AssignRole.AddRole")
	}

	if err := s.enforcer.SavePolicy(); err != nil {
		return errors.Internal(domainName, err).WithOperation("AssignRole.SavePolicy")
	}

	return nil
}

// RemoveRole removes a role from a user
func (s *service) RemoveRole(ctx context.Context, userID uint, roleCode string) error {
	subject := FormatUserSubject(userID)

	// Check if assigned
	hasRole, err := s.enforcer.HasRoleForUser(subject, roleCode)
	if err != nil {
		return errors.Internal(domainName, err).WithOperation("RemoveRole.HasRole")
	}
	if !hasRole {
		return ErrRoleNotAssigned
	}

	// If removing admin role, check if this is the last admin
	if roleCode == RoleCodeAdmin {
		count, err := s.CountAdminUsers(ctx)
		if err != nil {
			return err
		}
		if count <= 1 {
			return ErrCannotRemoveLastAdmin
		}
	}

	// Remove role
	if _, err := s.enforcer.DeleteRoleForUser(subject, roleCode); err != nil {
		return errors.Internal(domainName, err).WithOperation("RemoveRole.DeleteRole")
	}

	if err := s.enforcer.SavePolicy(); err != nil {
		return errors.Internal(domainName, err).WithOperation("RemoveRole.SavePolicy")
	}

	return nil
}

// CheckPermission checks if a user has permission for a domain and action
func (s *service) CheckPermission(ctx context.Context, userID uint, domain, action string) (bool, error) {
	subject := FormatUserSubject(userID)
	allowed, err := s.enforcer.Enforce(subject, domain, action)
	if err != nil {
		return false, errors.Internal(domainName, err).WithOperation("CheckPermission")
	}
	return allowed, nil
}

// GetDomains returns all registered domains
func (s *service) GetDomains(ctx context.Context) []DomainResponse {
	domains := s.domainProvider.GetDomainNames()
	result := make([]DomainResponse, len(domains))
	for i, d := range domains {
		result[i] = DomainResponse{
			Name:        d,
			IsProtected: IsProtectedDomain(d),
		}
	}
	return result
}

// GetActions returns all available actions
func (s *service) GetActions(ctx context.Context) []string {
	return AllActions()
}

// CountAdminUsers counts users with admin role
func (s *service) CountAdminUsers(ctx context.Context) (int, error) {
	users, err := s.enforcer.GetUsersForRole(RoleCodeAdmin)
	if err != nil {
		return 0, errors.Internal(domainName, err).WithOperation("CountAdminUsers")
	}
	return len(users), nil
}

// SyncGlobalRoles syncs system role policies with all registered domains
func (s *service) SyncGlobalRoles(ctx context.Context) error {
	allDomains := s.domainProvider.GetDomainNames()
	actions := AllActions()

	s.logger.Info(ctx, "syncing global RBAC roles", "domains", allDomains)

	// 1. Admin = wildcard (full access to everything)
	s.enforcer.RemoveFilteredPolicy(0, RoleCodeAdmin)
	if _, err := s.enforcer.AddPolicy(RoleCodeAdmin, "*", "*"); err != nil {
		s.logger.Error(ctx, "failed to add admin wildcard policy", err)
	}

	// 2. full_reader = read NON-PROTECTED domains only
	s.enforcer.RemoveFilteredPolicy(0, RoleCodeFullReader)
	for _, dom := range allDomains {
		if !IsProtectedDomain(dom) {
			if _, err := s.enforcer.AddPolicy(RoleCodeFullReader, dom, ActionRead); err != nil {
				s.logger.Error(ctx, "failed to add full_reader policy", err, "domain", dom)
			}
		}
	}

	// 3. full_writer = CRUD on non-protected, read on protected
	s.enforcer.RemoveFilteredPolicy(0, RoleCodeFullWriter)
	for _, dom := range allDomains {
		if IsProtectedDomain(dom) {
			// Read-only on protected domains
			if _, err := s.enforcer.AddPolicy(RoleCodeFullWriter, dom, ActionRead); err != nil {
				s.logger.Error(ctx, "failed to add full_writer read policy", err, "domain", dom)
			}
		} else {
			// Full CRUD on non-protected domains
			for _, act := range actions {
				if _, err := s.enforcer.AddPolicy(RoleCodeFullWriter, dom, act); err != nil {
					s.logger.Error(ctx, "failed to add full_writer policy", err, "domain", dom, "action", act)
				}
			}
		}
	}

	if err := s.enforcer.SavePolicy(); err != nil {
		return errors.Internal(domainName, err).WithOperation("SyncGlobalRoles.SavePolicy")
	}

	s.logger.Info(ctx, "global RBAC roles synced successfully")
	return nil
}

// getRolePermissions extracts permissions from Casbin policies for a role
func (s *service) getRolePermissions(roleCode string) []Permission {
	policies, _ := s.enforcer.GetFilteredPolicy(0, roleCode)

	// Group by domain
	domainActions := make(map[string][]string)
	for _, p := range policies {
		if len(p) >= 3 {
			domain := p[1]
			action := p[2]
			domainActions[domain] = append(domainActions[domain], action)
		}
	}

	var permissions []Permission
	for domain, actions := range domainActions {
		permissions = append(permissions, Permission{
			Domain:  domain,
			Actions: actions,
		})
	}

	return permissions
}

// getValidDomains returns valid domain names
func (s *service) getValidDomains() map[string]bool {
	domains := s.domainProvider.GetDomainNames()
	result := make(map[string]bool)
	for _, d := range domains {
		result[d] = true
	}
	// Also allow wildcard
	result["*"] = true
	return result
}

// isValidDomain checks if a domain is valid
func (s *service) isValidDomain(domain string, validDomains map[string]bool) bool {
	return validDomains[domain]
}

// GetEnforcer returns the Casbin enforcer (for middleware use)
func (s *service) GetEnforcer() *casbin.TransactionalEnforcer {
	return s.enforcer
}

// GetTransactionalEnforcer returns the enforcer for transaction coordination
func (s *service) GetTransactionalEnforcer() *casbin.TransactionalEnforcer {
	return s.enforcer
}

// AssignRoleInTx assigns a role without calling SavePolicy.
// Used when the caller manages the transaction via TransactionalEnforcer.
func (s *service) AssignRoleInTx(tx *casbin.Transaction, ctx context.Context, userID uint, roleCode string) error {
	// Check if role exists
	_, err := s.repo.FindByCode(ctx, roleCode)
	if err != nil {
		if errors.IsNotFound(err) {
			return ErrRoleNotFound
		}
		return errors.Internal(domainName, err).WithOperation("AssignRoleInTx")
	}

	subject := FormatUserSubject(userID)

	// Check if already assigned (using main enforcer for read)
	hasRole, err := s.enforcer.HasRoleForUser(subject, roleCode)
	if err != nil {
		return errors.Internal(domainName, err).WithOperation("AssignRoleInTx.HasRole")
	}
	if hasRole {
		return ErrRoleAlreadyAssigned
	}

	// Add role using transaction (no SavePolicy - transaction handles commit)
	if _, err := tx.AddGroupingPolicy(subject, roleCode); err != nil {
		return errors.Internal(domainName, err).WithOperation("AssignRoleInTx.AddRole")
	}

	return nil
}
