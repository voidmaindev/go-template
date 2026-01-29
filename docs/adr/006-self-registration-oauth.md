# ADR-006: Self-Registration and OAuth Authentication

## Status

Accepted

## Context

The application needs to support user self-registration in addition to admin-created accounts. This includes:

1. Email-based self-registration with verification
2. OAuth authentication via third-party providers (Google, Facebook, Apple)
3. Password reset functionality
4. External identity management (linking/unlinking OAuth accounts)

## Decision

### Single Toggle for Self-Registration

Self-registration (both email and OAuth) is controlled by a single `SELF_REGISTRATION_ENABLED` toggle. When disabled:
- All self-registration endpoints return 403 Forbidden
- OAuth providers cannot create new users
- Existing OAuth-linked users can still authenticate

**Rationale**: Simplifies configuration and provides a single kill-switch for public registration while maintaining existing user access.

### Redis for Token Storage

Verification and password reset tokens are stored in Redis with TTL:
- Email verification tokens: 24h default (`SELF_REGISTRATION_VERIFICATION_TOKEN_EXPIRY`)
- Password reset tokens: 1h default (`SELF_REGISTRATION_PASSWORD_RESET_TOKEN_EXPIRY`)

**Rationale**:
- Automatic expiration via Redis TTL
- Distributed access across multiple instances
- Consistent with existing JWT blacklist storage
- No database migrations required

### OAuth Provider Abstraction

OAuth providers implement a common interface:

```go
type OAuthProvider interface {
    GetAuthURL(state string) string
    Exchange(ctx context.Context, code string) (*oauth2.Token, error)
    GetUserInfo(ctx context.Context, token *oauth2.Token) (*OAuthUserInfo, error)
}
```

**Rationale**: Allows easy addition of new providers without modifying core authentication logic.

### Auto-Link Behavior for Matching Emails

When a user authenticates via OAuth:
1. If the OAuth email matches an existing user, the identity is automatically linked
2. If no match exists and self-registration is enabled, a new user is created
3. If no match exists and self-registration is disabled, authentication fails

**Rationale**: Reduces friction for existing users trying OAuth login while maintaining security.

### `self_registered` Role

Self-registered users receive a limited-permission `self_registered` role:
- Read access to non-protected domains only
- No write, update, or delete permissions by default
- Administrators can manually upgrade users to other roles

**Rationale**: Provides a security boundary between public registrations and admin-created users. Allows administrators to vet self-registered users before granting elevated access.

### External Identity Model

External identities are stored separately from users:

```go
type ExternalIdentity struct {
    ID         uint
    UserID     uint
    Provider   string // google, facebook, apple
    ProviderID string // Unique ID from provider
    Email      string // Email from provider (may differ from user's primary)
    CreatedAt  time.Time
}
```

**Rationale**:
- Users can have multiple OAuth providers linked
- Provider-specific data is isolated
- Easy to add/remove providers without affecting user records

### Email Verification Flow

1. User registers with email/password
2. System sends verification email with token
3. User clicks link, token is validated
4. User's `email_verified_at` is set
5. If verification required, user can now log in

**Rationale**: Standard security practice to verify email ownership before granting access.

## Consequences

### Positive

- Users can self-register without admin intervention
- Multiple authentication methods (email + OAuth providers)
- Flexible permission model for self-registered users
- Standard OAuth2 flow familiar to users

### Negative

- Additional configuration complexity (SMTP, OAuth credentials)
- Need to manage token storage in Redis
- Potential for spam registrations (mitigated by email verification)

### Neutral

- New dependencies: SMTP library, OAuth2 clients
- Additional database tables for external identities
- New seeder for `self_registered` role

## Related

- ADR-003: RBAC with Casbin (role management)
- ADR-005: Type-Safe Dependency Injection (service registration)
