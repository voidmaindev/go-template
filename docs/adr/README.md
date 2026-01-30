# Architecture Decision Records

This directory contains Architecture Decision Records (ADRs) documenting significant architectural decisions made in this project.

## Format

Each ADR follows this structure:

- **Title**: Short descriptive name
- **Status**: Proposed | Accepted | Deprecated | Superseded
- **Context**: The situation and problem being addressed
- **Decision**: The chosen approach
- **Consequences**: Trade-offs and implications

## Index

| ADR | Title | Status |
|-----|-------|--------|
| [001](001-domain-driven-architecture.md) | Domain-Driven Architecture | Accepted |
| [002](002-generic-repository-pattern.md) | Generic Repository Pattern | Accepted |
| [003](003-rbac-with-casbin.md) | RBAC with Casbin | Accepted |
| [004](004-openapi-first-design.md) | OpenAPI-First Design | Accepted |
| [005](005-type-safe-dependency-injection.md) | Type-Safe Dependency Injection | Accepted |
| [006a](006-self-registration-oauth.md) | Self-Registration and OAuth | Accepted |
| [006b](006-pluggable-email-provider.md) | SendGrid Email Provider | Accepted |

## Creating New ADRs

1. Copy the template structure from an existing ADR
2. Use the next sequential number
3. Update this README index
