# ports/

**Purpose:** Port definitions — interfaces that decouple core logic from framework-specific implementations.

**Rules:**
- Interfaces only, no implementations
- One file per concern (e.g., `example.repository.go`, `provider.go`, `unit_of_work.go`)
- Never reference DTOs, HTTP, or framework types (e.g., `*gorm.DB`)
- Consumers (services, handlers) depend on these ports, not on implementations
- Go convention: no `I` prefix on interfaces — name by what it does

## Repository interface

**`internal/database/ports/example.repository.go`:**
```go
package ports

import (
    "context"
    "github.com/your-org/your-project/internal/core/response"
    "github.com/your-org/your-project/internal/database"
    "github.com/your-org/your-project/internal/database/models"
)

type ExampleRepository interface {
    FindAll(ctx context.Context, q *database.Query) (*response.PaginatedResult[*models.Example], error)
    FindByID(ctx context.Context, id uint) (*models.Example, error)
    Create(ctx context.Context, m *models.Example) error
    Update(ctx context.Context, m *models.Example) error
    Delete(ctx context.Context, id uint) error
}
```

## Provider interface

**`internal/database/ports/provider.go`:**
```go
package ports

type Provider interface {
    Example() ExampleRepository
}
```

## UnitOfWork interface

**`internal/database/ports/unit_of_work.go`:**
```go
package ports

import "context"

type UnitOfWork interface {
    Do(ctx context.Context, fn func(ctx context.Context, provider Provider) error) error
}
```

## Why centralise ports here?

- Prevents circular imports between modules
- Framework-specific implementations (`provider/`, `unit-of-work/`) depend on these ports
- Repository implementations in `modules/<name>/repositories/` also depend on these ports
- Services and controllers depend only on ports, never on implementations
