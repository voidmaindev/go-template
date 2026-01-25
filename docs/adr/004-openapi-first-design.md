# ADR-004: OpenAPI-First Design

## Status

Accepted

## Context

We need an API design approach that:
- Provides clear API contracts
- Generates type-safe server code
- Enables automatic documentation
- Supports client code generation

## Decision

Use OpenAPI 3.0 specification as the source of truth with `oapi-codegen` for code generation.

**Workflow**:
1. Define API in `api/openapi.yaml`
2. Generate types and interfaces with `oapi-codegen`
3. Implement generated interfaces in handlers
4. Serve spec for Swagger UI

**Code Generation Configuration** (`api/oapi-codegen.yaml`):
```yaml
package: api
generate:
  models: true
  chi-server: true
  strict-server: true
output: api/api.gen.go
```

**Generated Artifacts**:
- Request/response DTOs
- Chi router interface
- Strict server interface with typed handlers

## Consequences

### Positive
- **Contract-first**: API design before implementation
- **Type safety**: Generated types match spec exactly
- **Documentation**: Spec serves as living documentation
- **Client generation**: Consumers can generate clients

### Negative
- **Build step**: Spec changes require regeneration
- **Strict typing**: Less flexibility in handlers
- **Learning curve**: OpenAPI spec syntax

### Mitigations
- Makefile target for code generation (`make generate`)
- Clear examples in existing spec
- Strict server pattern enforces correct response types
