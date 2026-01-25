# ADR-003: RBAC with Casbin

## Status

Accepted

## Context

We need an authorization system that:
- Supports role-based access control (RBAC)
- Allows flexible policy definitions
- Supports multiple roles per user
- Can be extended to resource-level permissions

## Decision

Use Casbin for RBAC with a domain/action model:

**Model Definition (RBAC with domains)**:
```
[request_definition]
r = sub, dom, obj, act

[policy_definition]
p = sub, dom, obj, act

[role_definition]
g = _, _, _

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = g(r.sub, p.sub, r.dom) && r.dom == p.dom && r.obj == p.obj && r.act == p.act
```

**Policy Storage**: Policies stored in database via GORM adapter.

**Integration**: Casbin enforcer injected via dependency container and used in middleware.

## Consequences

### Positive
- **Flexibility**: Policy changes without code changes
- **Well-tested**: Casbin is mature and widely used
- **Multi-role**: Users can have multiple roles per domain
- **Extensible**: Can add ABAC (attribute-based) rules later

### Negative
- **Complexity**: Casbin model syntax has learning curve
- **Storage**: Policy changes require database operations
- **Performance**: Enforcement adds overhead (mitigated by caching)

### Mitigations
- Document policy model and common patterns
- Use Casbin's built-in caching
- Provide helper functions for common permission checks
