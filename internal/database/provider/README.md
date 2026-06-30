# provider/

**Purpose:** GORM-specific Provider implementation — singleton access to repository instances.

The `Provider` interface is defined in `internal/database/ports/`.

## Implementation

**`internal/database/provider/provider.go`:**
```go
package provider

import (
    "gorm.io/gorm"
    "github.com/your-org/your-project/internal/database/ports"
    "github.com/your-org/your-project/internal/modules/example/repositories"
)

var _ ports.Provider = (*Provider)(nil)

type Provider struct {
    db          *gorm.DB
    exampleRepo ports.ExampleRepository
}

func New(db *gorm.DB) *Provider {
    return &Provider{db: db}
}

func (p *Provider) Example() ports.ExampleRepository {
    if p.exampleRepo == nil {
        p.exampleRepo = repositories.NewRepo(p.db)
    }
    return p.exampleRepo
}
```

Usage:
```go
p := provider.New(db)
svc := services.NewService(p.Example())
```

## Adding a new repository

1. Define the interface in `internal/database/ports/`
2. Implement it in `internal/modules/<name>/repositories/`
3. Add a field + nil-check accessor on `Provider`
4. Add the accessor to `ports.Provider` interface
