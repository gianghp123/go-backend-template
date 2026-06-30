# Module Creation Guide

## Folder Structure

```
internal/modules/<name>/
├── controllers/
│   ├── controller.go              # User-facing handlers
│   └── admin_controller.go        # Admin handlers (same service)
├── services/
│   ├── service.go                 # Business logic — returns models, no DTO mapping
│   └── service_test.go            # Table-driven tests with mock implementations
├── repositories/
│   └── repository.go              # Repository implementation (interface in database/ports/)
├── dtos/
│   ├── req/                       # Request DTOs (API layer)
│   └── res/                       # Response DTOs (API layer)
└── module.go                      # Route registration
```

Model/entity structs live in `internal/database/models/`, separate from DTOs.

Port **interfaces** live in `internal/database/ports/` (e.g., `ExampleRepository`, `Provider`).
Repository **implementations** live in `internal/modules/<name>/repositories/`.
The **Provider** lives in `internal/database/provider/` — singleton repo accessor.
The **UnitOfWork** lives in `internal/database/unit-of-work/` — transaction coordinator.

## Data Flow

```
Controller  →  binds req DTO, extracts userID from context via utils.GetCtx, calls service, maps model → res DTO
    ↓
Service     →  enforces policy when needed (policy.ActorFromContext + policy.CanMutate/CanRead), returns Model + AppError
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

Repository interfaces live in `internal/database/ports/`, one file per module.
Go convention: no `I` prefix on interfaces. Name by what it does.

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
var _ ports.ExampleRepository = (*Repo)(nil)

type Repo struct { db *gorm.DB }

func NewRepo(db *gorm.DB) *Repo {
    return &Repo{db: db}
}

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

Services are **pure** — they accept only the params they need for pure business logic.
The controller extracts `userID` from context and passes it as a param when the service needs domain data.
The service calls `policy.ActorFromContext(ctx)` **only** when it needs to enforce authorization.
The repo field uses the interface from `database/ports/`.
Constructor returns the concrete struct, not the interface.

```go
// Interface — exported for controllers to depend on
type ExampleService interface {
    Create(ctx context.Context, userID string, body req.CreateReq) (*models.Example, *coreError.AppError)
    List(ctx context.Context, userID string, q req.ListQuery) (*response.PaginatedResult[*models.Example], *coreError.AppError)
    GetByID(ctx context.Context, id uint) (*models.Example, *coreError.AppError)
    Delete(ctx context.Context, id uint) *coreError.AppError
}

// Struct — unexported, returned as concrete type
type exampleService struct {
    repo ports.ExampleRepository
}

func NewService(repo ports.ExampleRepository) *exampleService {
    return &exampleService{repo: repo}
}

// Create — userID is domain data, controller passes it
func (s *exampleService) Create(ctx context.Context, userID string, body req.CreateReq) (*models.Example, *coreError.AppError) {
    item := &models.Example{
        Name:    body.Name,
        OwnerID: userID,
    }
    if err := s.repo.Create(ctx, item); err != nil {
        return nil, coreError.Internal("failed to create example")
    }
    return item, nil
}

// List — userID scopes the query, controller passes it
func (s *exampleService) List(ctx context.Context, userID string, q req.ListQuery) (*response.PaginatedResult[*models.Example], *coreError.AppError) {
    dbQuery := q.ToQuery()
    dbQuery.SetFilter("owner_id", userID)
    result, err := s.repo.FindAll(ctx, dbQuery)
    if err != nil {
        return nil, coreError.Internal("failed to list examples")
    }
    return result, nil
}

// GetByID — needs auth check, so uses ActorFromContext + CanRead
func (s *exampleService) GetByID(ctx context.Context, id uint) (*models.Example, *coreError.AppError) {
    actor, err := policy.ActorFromContext(ctx)
    if err != nil {
        return nil, coreError.Unauthorized()
    }
    item, err := s.repo.FindByID(ctx, id)
    if err != nil {
        if errors.Is(err, coreError.ErrNotFound) { return nil, coreError.NotFound() }
        return nil, coreError.Internal(err.Error())
    }
    if err := policy.CanRead(actor, item.OwnerID); err != nil {
        return nil, coreError.Forbidden()
    }
    return item, nil
}

// Delete — needs auth check, so uses ActorFromContext + CanMutate
func (s *exampleService) Delete(ctx context.Context, id uint) *coreError.AppError {
    actor, err := policy.ActorFromContext(ctx)
    if err != nil {
        return coreError.Unauthorized()
    }
    item, err := s.repo.FindByID(ctx, id)
    if err != nil {
        if errors.Is(err, coreError.ErrNotFound) { return coreError.NotFound() }
        return coreError.Internal(err.Error())
    }
    if err := policy.CanMutate(actor, item.OwnerID); err != nil {
        return coreError.Forbidden()
    }
    if err := s.repo.Delete(ctx, id); err != nil {
        return coreError.Internal("failed to delete example")
    }
    return nil
}
```

Key rules:
- **Create/List** — `userID` is domain data, controller passes it as a param. Service never extracts from context.
- **GetByID/Update/Delete** — needs authorization, service calls `policy.ActorFromContext(ctx)` + `policy.CanRead`/`policy.CanMutate`
- Services return `(*model, *coreError.AppError)` — model types, not DTOs
- Controllers are responsible for mapping models to response DTOs
- Service maps repo errors to `coreError.*` (AppError) via `errors.Is(err, coreError.ErrNotFound)` → `coreError.NotFound()`
- Services never call `utils.GetCtx` or `utils.MapToDTO` — those are controller concerns
- Service functions accept the minimum params needed — this makes them reusable across different controllers
- Constructor (`NewService`) returns concrete struct `*exampleService`, not the interface — callers accept the interface

## Step 6: Write Controllers

Controllers bind request DTOs, extract `userID` from context, call services, map models to response DTOs, and write HTTP responses.
Controllers accept the service **interface** (not concrete type) for testability:

```go
type Controller struct { svc ExampleService }

func (h *Controller) Create(c *gin.Context) {
    var body req.CreateReq
    if err := c.ShouldBindJSON(&body); err != nil {
        c.JSON(http.StatusBadRequest, response.Fail(coreError.BadRequest(err.Error())))
        return
    }

    // Controller extracts userID from context — service doesn't
    userID := utils.GetCtx[string](c.Request.Context(), enums.ContextKeyUserID)

    item, appErr := h.svc.Create(c.Request.Context(), userID, body)
    if appErr != nil {
        c.JSON(appErr.Code, response.Fail(appErr))
        return
    }

    var result res.ExampleRes
    if err := utils.MapToDTO(item, &result); err != nil {
        c.JSON(http.StatusInternalServerError, response.Fail(coreError.Internal("failed to map")))
        return
    }
    c.JSON(http.StatusOK, response.Success(result))
}

func (h *Controller) List(c *gin.Context) {
    var q req.ListQuery
    if err := c.ShouldBindQuery(&q); err != nil {
        c.JSON(http.StatusBadRequest, response.Fail(coreError.BadRequest(err.Error())))
        return
    }

    // Controller extracts userID — service receives it as param
    userID := utils.GetCtx[string](c.Request.Context(), enums.ContextKeyUserID)

    result, appErr := h.svc.List(c.Request.Context(), userID, q)
    if appErr != nil {
        c.JSON(appErr.Code, response.Fail(appErr))
        return
    }

    var examples []res.ExampleRes
    if err := utils.MapToDTOs(result.Data, &examples); err != nil {
        c.JSON(http.StatusInternalServerError, response.Fail(coreError.Internal("failed to map")))
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

    // Auth is handled inside the service via policy.ActorFromContext — controller just passes the ID
    item, appErr := h.svc.GetByID(c.Request.Context(), body.ID)
    if appErr != nil {
        c.JSON(appErr.Code, response.Fail(appErr))
        return
    }

    var result res.ExampleRes
    if err := utils.MapToDTO(item, &result); err != nil {
        c.JSON(http.StatusInternalServerError, response.Fail(coreError.Internal("failed to map")))
        return
    }
    c.JSON(http.StatusOK, response.Success(result))
}
```

Model-to-DTO mapping happens in controllers — services return plain models.

Admin controllers share the same service:

```go
type AdminController struct { svc ExampleService }
```

## Step 7: Register Routes

At the composition root, use the Provider to wire services:

```go
import (
    "github.com/your-org/your-project/internal/database/provider"
)

func RegisterRoutes(rg *gin.RouterGroup, p *provider.Provider, authMw, adminMw gin.HandlerFunc) {
    svc := services.NewService(p.Example())
    h := controllers.NewController(svc)
    admin := controllers.NewAdminController(svc)
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

The `Provider` is for simple repo access. For transactions, use `UnitOfWork` from `internal/database/unit-of-work/`:

Usage:

```go
p := provider.New(db)
uow := unitofwork.New(db)

// Non-transactional
svc := services.NewService(p.Example())

// Transactional
uow.Do(ctx, func(txCtx context.Context, p ports.Provider) error {
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
| Service interface | PascalCase (no `I` prefix) | `ExampleService` |
| Service struct | unexported | `exampleService` |
| Service constructor | returns concrete struct | `NewService(...) *exampleService` |
| Controller | PascalCase | `ExampleController` |
| Admin controller | PascalCase | `ExampleAdminController` |
| Repo interface | PascalCase (no `I` prefix) | `ExampleRepository` |
| Repo struct | unexported | `Repo` |
| Repo constructor | returns concrete struct | `NewRepo(...) *Repo` |
| Mapper | `utils.MapToDTO` / `utils.MapToDTOs` | generic, in controller (or utils) |
