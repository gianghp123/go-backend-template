# repositories/

**Purpose:** Data access layer — raw database operations. No business logic.

**Rules:**
- No DTOs — accept and return only `models.*` and `*database.Query`
- No AppErrors — return stdlib errors only (`errors.New`, `gorm.ErrRecordNotFound`, etc.)
- No context extraction — the requester is a business concern, stay in services
- All methods accept `context.Context` for cancellation and tracing

## Interface

```go
type IExampleRepo interface {
    FindAll(ctx context.Context, query *database.Query) (*response.PaginatedResult[*models.Example], error)
    FindByID(ctx context.Context, id uint) (*models.Example, error)
    Create(ctx context.Context, example *models.Example) error
    Update(ctx context.Context, example *models.Example) error
    Delete(ctx context.Context, id uint) error
}
```

## Implementation (GORM example)

```go
type exampleRepo struct {
    db *gorm.DB
}

func NewExampleRepo(db *gorm.DB) IExampleRepo {
    return &exampleRepo{db: db}
}

func (r *exampleRepo) FindAll(ctx context.Context, query *database.Query) (*response.PaginatedResult[*models.Example], error) {
    var examples []*models.Example

    tx := r.db.WithContext(ctx).Model(&models.Example{})
    var total int64
    query.Count(tx).Count(&total)

    err := query.Apply(tx).Find(&examples).Error
    if err != nil {
        return nil, errors.MapRepoError(err)
    }

    return &response.PaginatedResult[*models.Example]{
        Data: examples,
        Meta: response.NewMeta(query.Page, query.Limit, total),
    }, nil
}

func (r *exampleRepo) FindByID(ctx context.Context, id uint) (*models.Example, error) {
    var example models.Example
    err := r.db.WithContext(ctx).First(&example, id).Error
    if err != nil {
        return nil, errors.MapRepoError(err)
    }
    return &example, nil
}

func (r *exampleRepo) Create(ctx context.Context, example *models.Example) error {
    return r.db.WithContext(ctx).Create(example).Error
}

func (r *exampleRepo) Update(ctx context.Context, example *models.Example) error {
    return r.db.WithContext(ctx).Save(example).Error
}

func (r *exampleRepo) Delete(ctx context.Context, id uint) error {
    return r.db.WithContext(ctx).Delete(&models.Example{}, id).Error
}
```

**Recommendations:**
- Return pointers (`*models.Example`) from FindByID — nil means not found
- Use `errors.MapRepoError(err)` to normalize ORM errors into sentinel errors
- `FindAll` returns `*response.PaginatedResult` with both data and meta
- Keep the concrete struct unexported (`exampleRepo`) — only the interface is public
- Constructor returns the interface type (`NewExampleRepo(...) IExampleRepo`)
