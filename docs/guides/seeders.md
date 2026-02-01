# Seeders Guide

This guide covers creating and managing database seeders for initial and test data.

## Overview

Seeders populate the database with initial data. Each seeder is:
- **Idempotent**: Safe to run multiple times
- **Tracked**: Records which seeders have run
- **Configurable**: Can use environment variables
- **Transactional**: Runs in a transaction

## Seeder Files

Location: `internal/database/seeders/`

```
internal/database/seeders/
├── seeders.go                    # SeederManager and interface
├── admin_user.go                 # Creates default admin user
├── rbac_seeder.go                # Creates system roles and policies
└── self_registered_role.go       # Creates self-registration role
```

## Creating a Seeder

### 1. Create the Seeder File

Create `internal/database/seeders/<name>_seeder.go`:

```go
package seeders

import (
    "github.com/voidmaindev/go-template/internal/config"
    "gorm.io/gorm"
)

type ProductCategoriesSeeder struct{}

func (s *ProductCategoriesSeeder) Name() string {
    return "product_categories"
}

func (s *ProductCategoriesSeeder) Run(db *gorm.DB, cfg *config.Config) error {
    categories := []map[string]any{
        {"name": "Electronics", "slug": "electronics"},
        {"name": "Clothing", "slug": "clothing"},
        {"name": "Books", "slug": "books"},
        {"name": "Home & Garden", "slug": "home-garden"},
    }

    for _, cat := range categories {
        // Use FirstOrCreate for idempotency
        result := db.Table("categories").
            Where("slug = ?", cat["slug"]).
            FirstOrCreate(&cat)
        if result.Error != nil {
            return result.Error
        }
    }

    return nil
}
```

### 2. Register the Seeder

Add to `internal/database/seeders/seeders.go`:

```go
func init() {
    defaultManager.Register(
        &AdminUserSeeder{},
        &RBACSeeder{},
        &SelfRegisteredRoleSeeder{},
        &ProductCategoriesSeeder{},  // Add new seeder
    )
}
```

## Seeder Interface

```go
type Seeder interface {
    Name() string                              // Unique seeder identifier
    Run(db *gorm.DB, cfg *config.Config) error // Execute seeder logic
}
```

## CLI Commands

```bash
# Run all pending seeders
go run ./cmd/api seed

# Reset and re-run all seeders
go run ./cmd/api seed --fresh

# Show seeder status
go run ./cmd/api seed status

# Reset seeder records (allows re-run)
go run ./cmd/api seed reset
```

## Seeder Status

The `status` command shows a table:

```
┌─────────────────────────┬─────────┐
│ Seeder                  │ Status  │
├─────────────────────────┼─────────┤
│ admin_user              │ Applied │
│ rbac_seeder             │ Applied │
│ self_registered_role    │ Applied │
│ product_categories      │ Pending │
└─────────────────────────┴─────────┘
```

## Seeders Tracking Table

Applied seeders are tracked in the `seeders` table:

```sql
CREATE TABLE seeders (
    name VARCHAR(255) PRIMARY KEY,
    seeded_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
```

## Idempotency Patterns

### FirstOrCreate Pattern

The most common pattern - creates only if not exists:

```go
func (s *CountriesSeeder) Run(db *gorm.DB, cfg *config.Config) error {
    countries := []Country{
        {Code: "US", Name: "United States"},
        {Code: "GB", Name: "United Kingdom"},
        {Code: "CA", Name: "Canada"},
    }

    for _, country := range countries {
        // Won't duplicate if code already exists
        db.Where("code = ?", country.Code).FirstOrCreate(&country)
    }

    return nil
}
```

### Upsert Pattern

Update if exists, create if not:

```go
func (s *ConfigSeeder) Run(db *gorm.DB, cfg *config.Config) error {
    configs := []AppConfig{
        {Key: "app.name", Value: "My App"},
        {Key: "app.version", Value: "1.0.0"},
    }

    for _, c := range configs {
        db.Clauses(clause.OnConflict{
            Columns:   []clause.Column{{Name: "key"}},
            DoUpdates: clause.AssignmentColumns([]string{"value"}),
        }).Create(&c)
    }

    return nil
}
```

### Check Before Insert

Explicit existence check:

```go
func (s *DefaultRolesSeeder) Run(db *gorm.DB, cfg *config.Config) error {
    var count int64
    db.Table("roles").Where("name = ?", "admin").Count(&count)

    if count == 0 {
        db.Create(&Role{Name: "admin", Description: "Administrator"})
    }

    return nil
}
```

## Environment Configuration

Use config for dynamic values:

```go
func (s *AdminUserSeeder) Run(db *gorm.DB, cfg *config.Config) error {
    // Get values from environment via config
    email := cfg.Seed.AdminEmail
    if email == "" {
        email = "admin@example.com"  // Default fallback
    }

    password := cfg.Seed.AdminPassword
    if password == "" {
        password = "Ab123456"  // Development fallback
    }

    // Hash password
    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    if err != nil {
        return err
    }

    admin := User{
        Email:    email,
        Password: string(hashedPassword),
        Name:     cfg.Seed.AdminName,
    }

    return db.Where("email = ?", email).FirstOrCreate(&admin).Error
}
```

Environment variables:
```bash
SEED_ADMIN_EMAIL=admin@company.com
SEED_ADMIN_PASSWORD=secure-password-here
SEED_ADMIN_NAME="System Administrator"
```

## Dependent Seeders

When seeders depend on data from other seeders:

```go
func (s *RBACSeeder) Run(db *gorm.DB, cfg *config.Config) error {
    // Create roles first
    roles := []Role{
        {Name: "admin"},
        {Name: "user"},
    }
    for _, r := range roles {
        db.Where("name = ?", r.Name).FirstOrCreate(&r)
    }

    // Then create policies
    var adminRole Role
    if err := db.Where("name = ?", "admin").First(&adminRole).Error; err != nil {
        return err  // Admin role must exist
    }

    // Assign admin wildcard policy
    policy := RBACPolicy{
        RoleID:   adminRole.ID,
        Domain:   "*",
        Action:   "*",
    }
    return db.Where("role_id = ? AND domain = ?", adminRole.ID, "*").
        FirstOrCreate(&policy).Error
}
```

**Registration Order Matters**: Register seeders in dependency order.

## Handling Related Data

For seeding with relationships:

```go
func (s *TestDataSeeder) Run(db *gorm.DB, cfg *config.Config) error {
    // Create parent
    country := Country{Code: "US", Name: "United States"}
    db.Where("code = ?", country.Code).FirstOrCreate(&country)

    // Create children referencing parent
    cities := []City{
        {Name: "New York", CountryID: country.ID},
        {Name: "Los Angeles", CountryID: country.ID},
        {Name: "Chicago", CountryID: country.ID},
    }

    for _, city := range cities {
        db.Where("name = ? AND country_id = ?", city.Name, city.CountryID).
            FirstOrCreate(&city)
    }

    return nil
}
```

## Development vs Production

### Development Seeders

Add test data for development:

```go
type DevelopmentDataSeeder struct{}

func (s *DevelopmentDataSeeder) Name() string {
    return "development_data"
}

func (s *DevelopmentDataSeeder) Run(db *gorm.DB, cfg *config.Config) error {
    // Only run in development
    if cfg.App.Environment != "development" {
        return nil  // Skip in production
    }

    // Create test users, sample products, etc.
    // ...

    return nil
}
```

### Production Seeders

Keep production seeders minimal:
- System roles and permissions
- Default admin user
- Required lookup data (countries, categories)

Avoid:
- Test users
- Sample content
- Debug data

## Error Handling

Seeders run in transactions. On error:
1. Transaction is rolled back
2. Seeder is NOT marked as applied
3. Error is logged and returned

```go
func (s *MySeeder) Run(db *gorm.DB, cfg *config.Config) error {
    // If any operation fails, entire seeder is rolled back
    if err := db.Create(&item1).Error; err != nil {
        return fmt.Errorf("failed to create item1: %w", err)
    }

    if err := db.Create(&item2).Error; err != nil {
        return fmt.Errorf("failed to create item2: %w", err)
    }

    return nil  // Success - seeder marked as applied
}
```

## Best Practices

### Always Be Idempotent

```go
// Good - safe to run multiple times
db.Where("code = ?", "US").FirstOrCreate(&country)

// Bad - will create duplicates
db.Create(&country)
```

### Use Meaningful Names

```go
// Good
func (s *InitialRBACRolesSeeder) Name() string {
    return "initial_rbac_roles"
}

// Bad
func (s *Seeder1) Name() string {
    return "seeder1"
}
```

### Keep Seeders Focused

- One logical group of data per seeder
- Split large seeders into multiple smaller ones
- Order by dependencies

### Log Important Actions

```go
func (s *AdminUserSeeder) Run(db *gorm.DB, cfg *config.Config) error {
    email := cfg.Seed.AdminEmail

    var existing User
    result := db.Where("email = ?", email).First(&existing)
    if result.Error == nil {
        slog.Info("admin user already exists", "email", email)
        return nil
    }

    // Create admin...
    slog.Info("created admin user", "email", email)
    return nil
}
```

## Checklist

- [ ] Seeder implements `Seeder` interface
- [ ] `Name()` returns unique identifier
- [ ] `Run()` is idempotent (safe to run multiple times)
- [ ] Registered in correct order (dependencies first)
- [ ] Uses config for environment-specific values
- [ ] Appropriate for target environment (dev vs prod)
- [ ] Error handling with wrapped errors
