# unit-of-work/

**Purpose:** GORM-specific UnitOfWork implementation — database transaction coordination.

The `IUnitOfWork` and `IProvider` interfaces are defined in `internal/database/ports/`.

## Implementation

**`internal/database/unit-of-work/unit_of_work.go`:**
```go
package unitofwork

import (
    "context"
    "gorm.io/gorm"
    "github.com/your-org/your-project/internal/database/ports"
    "github.com/your-org/your-project/internal/database/provider"
)

var _ ports.IUnitOfWork = (*UnitOfWork)(nil)

type UnitOfWork struct {
    db *gorm.DB
}

func New(db *gorm.DB) *UnitOfWork {
    return &UnitOfWork{db: db}
}

func (u *UnitOfWork) Do(ctx context.Context, fn func(ctx context.Context, p ports.IProvider) error) error {
    return u.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
        return fn(ctx, provider.New(tx))
    })
}
```

Usage:
```go
uow := unitofwork.New(db)

err := uow.Do(ctx, func(txCtx context.Context, p ports.IProvider) error {
    example, err := p.Example().FindByID(txCtx, id)
    if err != nil {
        return err // rollback
    }
    example.Name = "updated"
    return p.Example().Update(txCtx, example) // commit
})
```

Any error returned by `fn` triggers a rollback; nil commits.
