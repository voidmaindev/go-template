# Developer Guides

Step-by-step guides for common development tasks.

## Index

| Guide | Description |
|-------|-------------|
| [Adding a New App](adding-new-app.md) | Create a new application configuration |
| [Adding a New Domain](adding-new-domain.md) | Create a new domain with full CRUD |
| [Testing](testing.md) | Unit and integration testing patterns |
| [Migrations](migrations.md) | Creating and managing database migrations |
| [Seeders](seeders.md) | Creating idempotent database seeders |
| [Removing Domains](removing-domains.md) | Cleanup checklist for example domains |

## Architecture Decision Records

For detailed architectural decisions, see the [ADR documentation](../adr/README.md):

### Core Architecture
| ADR | Description |
|-----|-------------|
| [ADR-001](../adr/001-domain-driven-architecture.md) | Domain-Driven Architecture |
| [ADR-002](../adr/002-generic-repository-pattern.md) | Generic Repository Pattern |
| [ADR-005](../adr/005-type-safe-dependency-injection.md) | Type-Safe Dependency Injection |
| [ADR-010](../adr/010-domain-lifecycle.md) | Domain Lifecycle Management (Shutdowner) |

### Authentication & Security
| ADR | Description |
|-----|-------------|
| [ADR-003](../adr/003-rbac-with-casbin.md) | RBAC with Casbin |
| [ADR-006](../adr/006-self-registration-oauth.md) | Self-Registration and OAuth |
| [ADR-007](../adr/007-rate-limiting.md) | Rate Limiting Strategy |
| [ADR-009](../adr/009-oauth-security-hardening.md) | OAuth Security Hardening |
| [ADR-011](../adr/011-token-invalidation.md) | Token Invalidation Strategy |

### API & Features
| ADR | Description |
|-----|-------------|
| [ADR-004](../adr/004-openapi-first-design.md) | OpenAPI-First Design |
| [ADR-008](../adr/008-pluggable-email-provider.md) | SendGrid Email Provider |
| [ADR-012](../adr/012-filtering-system.md) | Django-Style Filtering |
