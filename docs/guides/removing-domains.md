# Removing Example Domains

This guide provides a checklist for removing the example domains (item, country, city, document) when starting a new project.

## Example Domains

The template includes these example domains to demonstrate patterns:

| Domain | Location | Dependencies |
|--------|----------|--------------|
| `item` | `internal/domain/item/` | None |
| `country` | `internal/domain/country/` | None |
| `city` | `internal/domain/city/` | country |
| `document` | `internal/domain/document/` | city, item |

## Removal Order

Remove domains in reverse dependency order to avoid compilation errors:

1. **document** (depends on city, item)
2. **city** (depends on country)
3. **country** (standalone)
4. **item** (standalone)

## Step-by-Step Removal

### 1. Remove Domain from App Registration

Edit `internal/app/main.go`:

```go
func mainDomains() []container.Domain {
    return []container.Domain{
        rbac.NewDomain(),
        user.NewDomain(),
        email.NewDomain(),
        audit.NewDomain(),
        auth.NewDomain(),
        // Remove these lines:
        // item.NewDomain(),
        // country.NewDomain(),
        // city.NewDomain(),
        // document.NewDomain(),
    }
}
```

Also update any other app files (e.g., `internal/app/geography.go`) that reference these domains.

### 2. Update Seeders

Edit `internal/database/seeders/self_registered_role.go`:

Remove permissions for deleted domains:

```go
// Before
domains := []string{"item", "city", "country", "document"}

// After - remove references to deleted domains
domains := []string{} // Or your new domains
```

### 3. Remove Domain Directories

```bash
# Remove domain directories
rm -rf internal/domain/item
rm -rf internal/domain/country
rm -rf internal/domain/city
rm -rf internal/domain/document
```

### 4. Remove Migration Files (Optional)

If starting fresh, remove example migrations:

```bash
# Remove example domain migrations
rm internal/database/migrations/000002_create_items.go
rm internal/database/migrations/000004_create_countries.go
rm internal/database/migrations/000005_create_cities.go
rm internal/database/migrations/000006_create_documents.go
```

Update `internal/database/migrations/migrations.go` to remove their registration.

### 5. Update Integration Tests

Remove or update tests that reference deleted domains:

```bash
# Check for references
grep -r "item\." internal/integration/
grep -r "country\." internal/integration/
grep -r "city\." internal/integration/
grep -r "document\." internal/integration/
```

Update `internal/integration/setup_test.go` to remove helpers for deleted domains.

### 6. Remove Test Fixtures

Edit `internal/testutil/fixtures.go`:

Remove fixture builders for deleted domains:
- `ItemFixture()`
- `CountryFixture()`
- `CityFixture()`
- `DocumentFixture()`

### 7. Update Imports

Search for and remove unused imports:

```bash
# Find files importing deleted domains
grep -r "github.com/voidmaindev/go-template/internal/domain/item" .
grep -r "github.com/voidmaindev/go-template/internal/domain/country" .
grep -r "github.com/voidmaindev/go-template/internal/domain/city" .
grep -r "github.com/voidmaindev/go-template/internal/domain/document" .
```

### 8. Verify Compilation

```bash
# Ensure the project still compiles
go build ./...

# Run tests
go test ./...
```

### 9. Database Cleanup (Development)

If you have an existing development database:

```bash
# Reset migrations (drops all tables)
go run ./cmd/api migrate reset

# Re-run migrations
go run ./cmd/api migrate up

# Re-run seeders
go run ./cmd/api seed
```

Or manually drop tables:

```sql
DROP TABLE IF EXISTS documents CASCADE;
DROP TABLE IF EXISTS cities CASCADE;
DROP TABLE IF EXISTS countries CASCADE;
DROP TABLE IF EXISTS items CASCADE;
```

## Keeping Some Domains

If you want to keep some example domains:

### Keep only geography (country, city)

1. Remove `item` and `document` from app registration
2. Update `document` dependencies in code (or remove `document`)
3. Remove `internal/domain/item/` and `internal/domain/document/`

### Keep only item

1. Remove `country`, `city`, `document` from app registration
2. Remove their directories

## After Removal

### Update Documentation

- Update `README.md` if it references example domains
- Update `docs/ARCHITECTURE.md` domain list
- Remove example API endpoints from OpenAPI spec

### Create Your Domains

Follow the [Adding a New Domain](adding-new-domain.md) guide to create your own domains.

## Checklist

### Per Domain Removal

- [ ] Remove from app registration (`internal/app/`)
- [ ] Update seeder permissions if needed
- [ ] Remove domain directory (`internal/domain/<name>/`)
- [ ] Remove migration file (if starting fresh)
- [ ] Update migration registration
- [ ] Remove integration test references
- [ ] Remove test fixtures
- [ ] Remove unused imports

### After All Removals

- [ ] Project compiles: `go build ./...`
- [ ] Tests pass: `go test ./...`
- [ ] Database reset or tables dropped
- [ ] Documentation updated
- [ ] OpenAPI spec updated (if applicable)

## Quick Commands

```bash
# Full cleanup of all example domains
# Step 1: Edit internal/app/main.go to remove domains
# Step 2: Edit internal/database/seeders/self_registered_role.go

# Step 3: Remove directories
rm -rf internal/domain/item
rm -rf internal/domain/country
rm -rf internal/domain/city
rm -rf internal/domain/document

# Step 4: Verify
go build ./...
go test -short ./...

# Step 5: Reset database (development only)
go run ./cmd/api migrate reset
go run ./cmd/api migrate up
go run ./cmd/api seed
```
