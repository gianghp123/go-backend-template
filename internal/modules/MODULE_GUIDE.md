# Module Creation Guide

## Folder Structure

```
internal/modules/<name>/
├── controllers/
│   ├── <name>.controller.go       # User-facing handlers
│   └── <name>-admin.controller.go # Admin handlers (same service)
├── services/
│   ├── <name>.service.go          # Business logic — returns models, no DTO mapping
│   └── <name>.service_test.go     # Table-driven tests with mock interfaces
├── repositories/
│   └── <name>.repo.go             # Repository implementation (interface in database/ports/)
├── dtos/
│   ├── req/                       # Request DTOs (API layer)
│   └── res/                       # Response DTOs (API layer)
└── <name>.module.go               # Route registration
```

Model/entity structs live in `internal/database/models/`, separate from DTOs.

Port **interfaces** live in `internal/database/ports/`.
Repository **implementations** live in `internal/modules/<name>/repositories/`.
The **Provider** lives in `internal/database/provider/` — singleton repo accessor.
The **UnitOfWork** lives in `internal/database/unit-of-work/` — transaction coordinator.

## Data Flow

```
Controller  →  binds req DTO, calls service, maps model → res DTO
    ↓
Service     →  applies policy, returns Model + AppError
    ↓
Provider    →  provides Repository instances (singleton)
 UnitOfWork →  wraps work in a DB transaction (optional)
    ↓
Repository  →  accepts Model + Query, returns Model + error
    ↓
Database    →  entities in internal/database/models/
```

## Step 1: Define the Model

Put entity structs in `internal/database/models/<name>.model.go`:

```go
type Example struct {
    ID          uint   `gorm:"primaryKey"`
    Name        string `gorm:"type:varchar(255)"`
    OwnerID     string `gorm:"not null"`
}
func (Example) TableName() string { return "examples" }
```

Models are pure data structs — no business logic.

## Step 2: Define DTOs

**Request DTOs** (`dtos/req/`):

```go
type ListQuery struct {
    Page    int    `json:"page"`
    Limit   int    `json:"limit"`
    OwnerID string `json:"ownerId"`
}

type CreateReq struct {
    Name string `json:"name"`
}
```

**Response DTOs** (`dtos/res/`):

```go
type ExampleRes struct {
    ID      uint   `json:"id"`
    Name    string `json:"name"`
    OwnerID string `json:"ownerId"`
}
```

## Step 3: Define the Repository Interface

Repository interfaces live in `internal/database/ports/`, one file per module:

**`internal/database/ports/example.repo.interface.go`:**
```go
package ports

import (
    "context"
    "github.com/your-org/your-project/internal/core/response"
    "github.com/your-org/your-project/internal/database"
    "github.com/your-org/your-project/internal/database/models"
)

type IExampleRepo interface {
    FindAll(ctx context.Context, q *database.Query) (*response.PaginatedResult[*models.Example], error)
    FindByID(ctx context.Context, id uint) (*models.Example, error)
    Create(ctx context.Context, m *models.Example) error
    Update(ctx context.Context, m *models.Example) error
    Delete(ctx context.Context, id uint) error
}
```

Repositories accept models and `database.Query`, never DTOs. Return `[]*models.Example` (slice of pointers) and `*models.Example` (pointer) from FindAll/FindByID. Errors are standard library errors (e.g. `errors.New("not found")`).

## Step 4: Implement the Repository

Repository implementations stay in `internal/modules/<name>/repositories/`.
They import the interface from `database/ports/`:

```go
package repositories

import (
    "context"
    "gorm.io/gorm"
    "github.com/your-org/your-project/internal/database"
    "github.com/your-org/your-project/internal/database/models"
    "github.com/your-org/your-project/internal/database/ports"
)

// Ensure compile-time interface compliance
var _ ports.IExampleRepo = (*Repo)(nil)

type Repo struct { db *gorm.DB }

func (r *Repo) FindAll(ctx context.Context, q *database.Query) (*response.PaginatedResult[*models.Example], error) {
    var examples []*models.Example
    // ... apply q.Count / q.Apply with your ORM
    return &response.PaginatedResult[*models.Example]{Data: examples, Meta: meta}, nil
}

func (r *Repo) FindByID(ctx context.Context, id uint) (*models.Example, error) {
    var m models.Example
    if err := r.db.WithContext(ctx).First(&m, id).Error; err != nil {
        return nil, err
    }
    return &m, nil
}
```

After implementing, register the repository in the Provider (see Step 9).

## Step 5: Write the Service

Services accept DTOs, build queries, enforce policy, and return models directly.
The repo field uses the interface from `database/ports/`:

```go
type Service struct {
    repo ports.IExampleRepo
}

func NewService(repo ports.IExampleRepo) *Service {
    return &Service{repo: repo}
}

func (s *Service) List(ctx context.Context, q req.ListQuery) (*response.PaginatedResult[*models.Example], *coreError.AppError) {
    result, err := s.repo.FindAll(ctx, q.ToQuery())
    if err != nil {
        return nil, coreError.Internal("failed to list examples")
    }
    return result, nil
}

func (s *Service) GetByID(ctx context.Context, body req.GetExampleReq) (*models.Example, *coreError.AppError) {
    userID := utils.GetCtx[string](ctx, enums.ContextKeyUserID)
    item, err := s.repo.FindByID(ctx, body.ID)
    if err != nil {
        if errors.Is(err, coreError.ErrNotFound) { return nil, coreError.NotFound() }
        return nil, coreError.Internal(err.Error())
    }

    if item.OwnerID != userID {
        return nil, coreError.Forbidden("not authorized to access this resource")
    }

    return item, nil
}
```

Key patterns:
- Services return `(*model, *coreError.AppError)` — model types, not DTOs
- Controllers are responsible for mapping models to response DTOs
- Service maps repo errors to `coreError.*` (AppError) via `errors.Is(err, coreError.ErrNotFound)` → `coreError.NotFound()`
- Services no longer call `utils.MapToDTO` / `utils.MapToDTOs` — that happens in controllers
- `utils.GetCtx[T any](ctx, key)` extracts user claims from context (set by middleware)
- Use the extracted `userID` from context to enforce ownership directly in the service
- Service functions accept DTOs (not separate params), including `GetByID` via `req.GetExampleReq`

## Step 6: Write Controllers

Controllers bind request DTOs, call services, map models to response DTOs, and write HTTP responses:

```go
type Controller struct { Service services.IExampleService }

func (h *Controller) List(c *gin.Context) {
    var q req.ListQuery
    if err := c.ShouldBindQuery(&q); err != nil {
        c.JSON(http.StatusBadRequest, response.Fail(coreError.BadRequest(err.Error())))
        return
    }

    result, appErr := h.Service.List(c.Request.Context(), q)
    if appErr != nil {
        c.JSON(appErr.Code, response.Fail(appErr))
        return
    }

    var examples []res.ExampleRes
    if err := utils.MapToDTOs(result.Data, &examples); err != nil {
        c.JSON(http.StatusInternalServerError, response.Fail(coreError.Internal("failed to map examples")))
        return
    }

    c.JSON(http.StatusOK, response.Success(response.PaginatedResponse[res.ExampleRes]{Data: examples, Meta: result.Meta}))
}

func (h *Controller) GetByID(c *gin.Context) {
    var body req.GetExampleReq
    if err := c.ShouldBindUri(&body); err != nil {
        c.JSON(http.StatusBadRequest, response.Fail(coreError.BadRequest(err.Error())))
        return
    }

    item, appErr := h.Service.GetByID(c.Request.Context(), body)
    if appErr != nil {
        c.JSON(appErr.Code, response.Fail(appErr))
        return
    }

    var result res.ExampleRes
    if err := utils.MapToDTO(item, &result); err != nil {
        c.JSON(http.StatusInternalServerError, response.Fail(coreError.Internal("failed to map example")))
        return
    }

    c.JSON(http.StatusOK, response.Success(result))
}
```

Model-to-DTO mapping happens in controllers — services return plain models.

Admin controllers share the same service:

```go
type AdminController struct { Service *services.Service }
```

## Step 7: Register Routes

At the composition root, use the Provider to wire services:

```go
import (
    "github.com/your-org/your-project/internal/database/provider"
)

func RegisterRoutes(rg *gin.RouterGroup, p *provider.Provider, authMw, adminMw gin.HandlerFunc) {
    svc := services.NewExampleService(p.Example())
    h := &controllers.Controller{Service: svc}
    admin := &controllers.AdminController{Service: svc}
    rg.GET("/examples", authMw, h.List)
    rg.DELETE("/examples/:id", authMw, adminMw, admin.Delete)
}
```

## Step 8: Write Tests

Use mock implementations of interfaces. Tests are table-driven:

```go
type mockRepo struct { findByIDFn func(ctx context.Context, id uint) (*models.Example, error) }

func TestService_GetByID(t *testing.T) {
    tests := []struct {
        name    string
        id      uint
        wantErr bool
    }{
        {name: "found", id: 1, wantErr: false},
        {name: "not found", id: 999, wantErr: true},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) { ... })
    }
}
```

## Step 9: Register Repository in the Provider

After implementing a repository, add it to `internal/database/provider/provider.go`:

```go
package provider

import (
    "gorm.io/gorm"
    "github.com/your-org/your-project/internal/database/ports"
    "github.com/your-org/your-project/internal/modules/example/repositories"
)

var _ ports.IProvider = (*Provider)(nil)

type Provider struct {
    db          *gorm.DB
    exampleRepo ports.IExampleRepo
}

func New(db *gorm.DB) *Provider {
    return &Provider{db: db}
}

func (p *Provider) Example() ports.IExampleRepo {
    if p.exampleRepo == nil {
        p.exampleRepo = repositories.NewExampleRepo(p.db)
    }
    return p.exampleRepo
}
```

The `Provider` is for simple repo access. For transactions, use `UnitOfWork` from `internal/database/unit-of-work/`:

Usage:

```go
p := provider.New(db)
uow := unitofwork.New(db)

// Non-transactional
svc := example_service.NewExampleService(p.Example())

// Transactional
uow.Do(ctx, func(txCtx context.Context, p ports.IProvider) error {
    example, err := p.Example().FindByID(txCtx, id)
    // ...
    return p.Example().Update(txCtx, example)
})
```

## Naming Conventions

| Artifact | Convention | Example |
|----------|-----------|---------|
| Model | PascalCase | `Example` |
| DTO req | PascalCase | `CreateReq`, `ListQuery` |
| DTO res | PascalCase | `ExampleRes` |
| Service | `XxxService` | `ExampleService` |
| Controller | `XxxController` | `ExampleController` |
| Admin controller | `XxxAdminController` | `ExampleAdminController` |
| Repo interface | `IXxxRepo` | `IExampleRepo` |
| Mapper | `utils.MapToDTO` / `utils.MapToDTOs` | generic, in controller (or utils) |
