# Adding a New App

Apps are configurations that define which domains run together. Use apps to create different deployments from the same codebase (e.g., a full app vs. a lightweight microservice).

## Steps

### 1. Create the App File

Create `internal/app/<appname>.go`:

```go
package app

import (
    "github.com/voidmaindev/go-template/internal/container"
    "github.com/voidmaindev/go-template/internal/domain/rbac"
    "github.com/voidmaindev/go-template/internal/domain/user"
    // Import your domains
)

// MyApp returns the application configuration.
func MyApp() *App {
    return &App{
        Name:        "myapp",
        Description: "Description of what this app does",
        Domains:     myAppDomains,
    }
}

func myAppDomains() []container.Domain {
    return []container.Domain{
        // Core domains (required for auth)
        rbac.NewDomain(), // Must be first (user depends on it)
        user.NewDomain(), // Required for authentication

        // Add your domains here
        // product.NewDomain(),
        // order.NewDomain(),
    }
}
```

### 2. Register in the App Registry

Edit `internal/app/app.go` and add your app to the `All()` function:

```go
func All() map[string]*App {
    main := MainApp()
    geo := GeographyApp()
    my := MyApp()  // Add this line
    return map[string]*App{
        main.Name: main,
        geo.Name:  geo,
        my.Name:   my,  // Add this line
    }
}
```

### 3. Run Your App

```bash
go run ./cmd/api serve myapp
```

## Domain Ordering

Domain order matters when there are dependencies:

```go
func myAppDomains() []container.Domain {
    return []container.Domain{
        // 1. Core domains first
        rbac.NewDomain(),    // No dependencies
        user.NewDomain(),    // Depends on: rbac

        // 2. Independent domains
        item.NewDomain(),    // No dependencies
        country.NewDomain(), // No dependencies

        // 3. Dependent domains last
        city.NewDomain(),     // Depends on: country
        document.NewDomain(), // Depends on: city, item
    }
}
```

## Example: Minimal Auth-Only App

```go
package app

import (
    "github.com/voidmaindev/go-template/internal/container"
    "github.com/voidmaindev/go-template/internal/domain/rbac"
    "github.com/voidmaindev/go-template/internal/domain/user"
)

func AuthOnlyApp() *App {
    return &App{
        Name:        "auth",
        Description: "Authentication service only",
        Domains:     authDomains,
    }
}

func authDomains() []container.Domain {
    return []container.Domain{
        rbac.NewDomain(),
        user.NewDomain(),
    }
}
```

## Checklist

- [ ] Created `internal/app/<appname>.go` with factory function
- [ ] Added app to `All()` in `internal/app/app.go`
- [ ] Verified domain order respects dependencies
- [ ] Tested with `go run ./cmd/api serve <appname>`
