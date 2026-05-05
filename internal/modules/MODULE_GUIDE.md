# Module Creation Guide

## Folder Structure

```
internal/modules/<name>/
├── controllers/
│   ├── <name>.controller.go       # User-facing handlers
│   └── <name>-admin.controller.go # Admin handlers (same service)
├── services/
│   ├── <name>.service.go          # Business logic + DTO ↔ model mapping
│   └── <name>.service_test.go     # Table-driven tests with mock interfaces
├── repositories/
│   ├── interfaces.go              # Repository interface
│   └── <name>.repo.go             # Repository implementation
├── dtos/
│   ├── req/                       # Request DTOs (API layer)
│   └── res/                       # Response DTOs (API layer)
└── <name>.module.go               # Route registration
```

Model/entity structs live in `internal/database/models/`, separate from DTOs.

## Data Flow

```
Controller  →  DTOs (req/res)
    ↓
Service     →  maps DTO ↔ Model, applies policy
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

```go
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

```go
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

## Step 5: Write the Service

Services accept DTOs, build queries, map DTOs ↔ models, and enforce policy:

```go
func (s *Service) List(ctx context.Context, q req.ListQuery) (*response.PaginatedResponse[res.ExampleRes], *coreError.AppError) {
    result, err := s.repo.FindAll(ctx, q.ToQuery())
    if err != nil {
        return nil, coreError.Internal("failed to list examples")
    }

    var examples []res.ExampleRes
    if err := utils.MapToDTOs(result.Data, &examples); err != nil {
        return nil, coreError.Internal("failed to map examples")
    }

    return &response.PaginatedResponse[res.ExampleRes]{Data: examples, Meta: result.Meta}, nil
}

func (s *Service) GetByID(ctx context.Context, body req.GetExampleReq) (*res.ExampleRes, *coreError.AppError) {
    userID := utils.GetCtx[string](ctx, enums.ContextKeyUserID)
    item, err := s.repo.FindByID(ctx, body.ID)
    if err != nil {
        if errors.Is(err, coreError.ErrNotFound) { return nil, coreError.NotFound() }
        return nil, coreError.Internal(err.Error())
    }
    // Optional: if err := policy.EnforceOwnerOrAdmin(userID, role, item.OwnerID); err != nil { ... }

    var result res.ExampleRes
    if err := utils.MapToDTO(item, &result); err != nil {
        return nil, coreError.Internal("failed to map example")
    }
    return &result, nil
}
```

Key patterns:
- Services return `(*data, *coreError.AppError)` — controllers use `.Code` for HTTP status
- Service maps repo errors to `coreError.*` (AppError) via `errors.Is(err, coreError.ErrNotFound)` → `coreError.NotFound()`
- Service maps models ↔ DTOs with `utils.MapToDTO` / `utils.MapToDTOs`
- `utils.GetCtx[T any](ctx, key)` extracts user claims from context (set by middleware)
- `policy.Enforce*` functions handle authorization (see `auth/README.md`)
- Service functions accept DTOs (not separate params), including `GetByID` via `req.GetExampleReq`

## Step 6: Write Controllers

Controllers pass DTOs from HTTP layer to service:

```go
type Controller struct { Service services.IExampleService }

func (h *Controller) List(ctx context.Context, q req.ListQuery) (*response.PaginatedResponse[res.ExampleRes], *coreError.AppError) {
    return h.Service.List(ctx, q)
}
```

Admin controllers share the same service:

```go
type AdminController struct { Service *services.Service }
```

## Step 7: Register Routes

```go
func RegisterRoutes(rg *gin.RouterGroup, svc *services.Service, authMw, adminMw gin.HandlerFunc) {
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
| Mapper | `mapToRes` / `mapToModel` | lowercase, in service |
