# Migrations Guide

This guide covers creating and managing database migrations.

## Overview

Migrations are versioned database changes that can be applied and rolled back. Each migration is:
- **Versioned**: Sequential number ensures correct order
- **Named**: Descriptive name explains the change
- **Reversible**: Has both `Up()` and `Down()` methods
- **Transactional**: Runs in a transaction for atomicity

## Migration Files

Location: `internal/database/migrations/`

```
internal/database/migrations/
├── migrations.go           # Migrator and interface
├── 000001_create_users.go
├── 000002_create_items.go
├── 000003_create_rbac_tables.go
└── ...
```

## Creating a Migration

### 1. Create the Migration File

Create `internal/database/migrations/NNNNNN_description.go`:

```go
package migrations

import "gorm.io/gorm"

type CreateProducts struct{}

func (m *CreateProducts) Version() string {
    return "000009"  // Next sequential number
}

func (m *CreateProducts) Name() string {
    return "create_products"
}

func (m *CreateProducts) Up(tx *gorm.DB) error {
    return tx.Exec(`
        CREATE TABLE products (
            id SERIAL PRIMARY KEY,
            name VARCHAR(200) NOT NULL,
            description TEXT,
            price BIGINT NOT NULL DEFAULT 0,
            sku VARCHAR(50) NOT NULL UNIQUE,
            created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
            updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
            deleted_at TIMESTAMP WITH TIME ZONE
        );

        CREATE INDEX idx_products_sku ON products(sku);
        CREATE INDEX idx_products_deleted_at ON products(deleted_at);
    `).Error
}

func (m *CreateProducts) Down(tx *gorm.DB) error {
    return tx.Exec(`DROP TABLE IF EXISTS products`).Error
}
```

### 2. Register the Migration

Add to `internal/database/migrations/migrations.go`:

```go
func init() {
    defaultMigrator.Register(
        &CreateUsers{},
        &CreateItems{},
        &CreateRBACTables{},
        // ... existing migrations
        &CreateProducts{},  // Add new migration
    )
}
```

## Migration Interface

```go
type Migration interface {
    Version() string           // "000001", "000002", etc.
    Name() string              // "create_users", "add_email_index"
    Up(tx *gorm.DB) error      // Apply migration
    Down(tx *gorm.DB) error    // Revert migration
}
```

## CLI Commands

```bash
# Apply all pending migrations
go run ./cmd/api migrate up

# Apply up to a specific version
go run ./cmd/api migrate up --to 000005

# Revert the last migration
go run ./cmd/api migrate down

# Revert last N migrations
go run ./cmd/api migrate down 3

# Show migration status
go run ./cmd/api migrate status

# Revert all migrations
go run ./cmd/api migrate reset

# Revert all and re-apply (reset + up)
go run ./cmd/api migrate refresh
```

## Migration Status

The `status` command shows a table:

```
┌─────────┬────────────────────────┬─────────┐
│ Version │ Name                   │ Status  │
├─────────┼────────────────────────┼─────────┤
│ 000001  │ create_users           │ Applied │
│ 000002  │ create_items           │ Applied │
│ 000003  │ create_rbac_tables     │ Applied │
│ 000009  │ create_products        │ Pending │
└─────────┴────────────────────────┴─────────┘
```

## Schema Migrations Table

Applied migrations are tracked in `schema_migrations`:

```sql
CREATE TABLE schema_migrations (
    version VARCHAR(255) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    applied_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
```

## Common Migration Patterns

### Add Column

```go
func (m *AddEmailVerifiedAt) Up(tx *gorm.DB) error {
    return tx.Exec(`
        ALTER TABLE users
        ADD COLUMN email_verified_at TIMESTAMP WITH TIME ZONE
    `).Error
}

func (m *AddEmailVerifiedAt) Down(tx *gorm.DB) error {
    return tx.Exec(`
        ALTER TABLE users
        DROP COLUMN email_verified_at
    `).Error
}
```

### Add Index

```go
func (m *AddProductsNameIndex) Up(tx *gorm.DB) error {
    return tx.Exec(`
        CREATE INDEX CONCURRENTLY idx_products_name
        ON products(name)
    `).Error
}

func (m *AddProductsNameIndex) Down(tx *gorm.DB) error {
    return tx.Exec(`DROP INDEX IF EXISTS idx_products_name`).Error
}
```

### Add Foreign Key

```go
func (m *AddProductCategoryFK) Up(tx *gorm.DB) error {
    return tx.Exec(`
        ALTER TABLE products
        ADD COLUMN category_id INTEGER REFERENCES categories(id);

        CREATE INDEX idx_products_category_id ON products(category_id);
    `).Error
}

func (m *AddProductCategoryFK) Down(tx *gorm.DB) error {
    return tx.Exec(`
        ALTER TABLE products
        DROP COLUMN category_id
    `).Error
}
```

### Rename Column

```go
func (m *RenameUserNameToFullName) Up(tx *gorm.DB) error {
    return tx.Exec(`
        ALTER TABLE users
        RENAME COLUMN name TO full_name
    `).Error
}

func (m *RenameUserNameToFullName) Down(tx *gorm.DB) error {
    return tx.Exec(`
        ALTER TABLE users
        RENAME COLUMN full_name TO name
    `).Error
}
```

### Data Migration

```go
func (m *MigrateUserStatuses) Up(tx *gorm.DB) error {
    // Add new column
    if err := tx.Exec(`
        ALTER TABLE users ADD COLUMN status VARCHAR(20) DEFAULT 'active'
    `).Error; err != nil {
        return err
    }

    // Migrate data
    return tx.Exec(`
        UPDATE users
        SET status = CASE
            WHEN is_active = true THEN 'active'
            ELSE 'inactive'
        END
    `).Error
}

func (m *MigrateUserStatuses) Down(tx *gorm.DB) error {
    return tx.Exec(`ALTER TABLE users DROP COLUMN status`).Error
}
```

## Best Practices

### Version Numbering

Use 6-digit sequential numbers:
- `000001`, `000002`, `000003`...
- Check existing migrations before assigning a number
- Numbers sort lexicographically

### Naming Conventions

Use descriptive snake_case names:
- `create_<table>` - New table
- `add_<column>_to_<table>` - New column
- `add_<table>_<column>_index` - New index
- `rename_<old>_to_<new>` - Rename
- `drop_<table>` - Remove table

### Keep Migrations Small

- One logical change per migration
- Easier to debug failures
- Easier to roll back specific changes

### Always Test Down()

```bash
# Apply migration
go run ./cmd/api migrate up

# Verify down works
go run ./cmd/api migrate down
go run ./cmd/api migrate up
```

### Avoid Destructive Changes in Production

- Don't drop columns/tables in the same release as code changes
- Use multi-step migrations for breaking changes:
  1. Add new column
  2. Deploy code that writes to both
  3. Migrate data
  4. Deploy code that reads from new column
  5. Drop old column (separate release)

### Use Raw SQL

Prefer raw SQL over GORM AutoMigrate:
- Explicit control over schema
- No surprises from ORM inference
- Easier to review and audit

## Troubleshooting

### Migration Failed Mid-Way

If a migration fails after partial execution:
1. Check `schema_migrations` table
2. Manually fix the schema if needed
3. Re-run the migration

### Version Conflict

If two developers create the same version:
1. Coordinate on version numbers
2. Renumber one migration
3. Update registration order

### Down Migration Doesn't Work

Ensure `Down()` is the exact inverse of `Up()`:
- If `Up()` adds a column, `Down()` drops it
- If `Up()` creates a table, `Down()` drops it
- Test both directions

## Checklist

- [ ] Use next sequential version number
- [ ] Descriptive migration name
- [ ] `Up()` applies the change
- [ ] `Down()` reverses the change exactly
- [ ] Register in `migrations.go`
- [ ] Test `up`, `down`, `up` sequence
- [ ] No destructive changes without multi-step plan
