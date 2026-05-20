# services/

**Purpose:** Business logic layer — orchestrates repositories, enforces policy.

**Rules:**
- Accept DTOs from controllers (for reusability and clear contracts)
- Extract user claims from context via `utils.GetCtx[T any](ctx, key)` (keeps function signatures clean)
- Return domain models directly — no DTO mapping here
- Map repo errors (`errors.Is(err, coreError.ErrNotFound)`) to `*response.AppError`
- No HTTP concerns — no `*gin.Context`, no status codes in logic

## Example

```go
type IExampleService interface {
    List(ctx context.Context, query req.ListExampleQuery) (*response.PaginatedResult[*models.Example], *response.AppError)
    GetByID(ctx context.Context, body req.GetExampleReq) (*models.Example, *response.AppError)
    Create(ctx context.Context, body req.CreateExampleReq) (*models.Example, *response.AppError)
    Update(ctx context.Context, id uint, body req.UpdateExampleReq) (*models.Example, *response.AppError)
    Delete(ctx context.Context, id uint) *response.AppError
}

type exampleService struct {
    repo ports.IExampleRepo
}

func NewExampleService(repo ports.IExampleRepo) IExampleService {
    return &exampleService{repo: repo}
}

func (s *exampleService) List(ctx context.Context, query req.ListExampleQuery) (*response.PaginatedResult[*models.Example], *response.AppError) {
    result, err := s.repo.FindAll(ctx, query.ToQuery())
    if err != nil {
        return nil, response.Internal("failed to list examples")
    }
    return result, nil
}

func (s *exampleService) GetByID(ctx context.Context, body req.GetExampleReq) (*models.Example, *response.AppError) {
    userID := utils.GetCtx[string](ctx, enums.ContextKeyUserID)
    example, err := s.repo.FindByID(ctx, body.ID)
    if err != nil {
        if errors.Is(err, coreError.ErrNotFound) {
            return nil, response.NotFound("example not found")
        }
        return nil, response.Internal("failed to get example")
    }

    if example.OwnerID != userID {
        return nil, response.Forbidden("not authorized to access this resource")
    }

    return example, nil
}

func (s *exampleService) Create(ctx context.Context, body req.CreateExampleReq) (*models.Example, *response.AppError) {
    userID := utils.GetCtx[string](ctx, enums.ContextKeyUserID)

    m := &models.Example{
        Name:        body.Name,
        Description: body.Description,
        OwnerID:     userID,
    }

    if err := s.repo.Create(ctx, m); err != nil {
        return nil, response.Internal("failed to create example")
    }

    return m, nil
}

func (s *exampleService) Update(ctx context.Context, id uint, body req.UpdateExampleReq) (*models.Example, *response.AppError) {
    example, err := s.repo.FindByID(ctx, id)
    if err != nil {
        if errors.Is(err, coreError.ErrNotFound) {
            return nil, response.NotFound("example not found")
        }
        return nil, response.Internal("failed to get example for update")
    }

    example.Name = body.Name
    example.Description = body.Description

    if err := s.repo.Update(ctx, example); err != nil {
        return nil, response.Internal("failed to update example")
    }

    return example, nil
}

func (s *exampleService) Delete(ctx context.Context, id uint) *response.AppError {
    if err := s.repo.Delete(ctx, id); err != nil {
        return response.Internal("failed to delete example")
    }
    return nil
}
```

**Recommendations:**
- Always return `*response.AppError` (not raw `error`) — controllers rely on `.Code` for HTTP status
- Use `errors.Is(err, coreError.ErrNotFound)` to detect 404 from repo
- Extract `userID := utils.GetCtx[string](ctx, enums.ContextKeyUserID)` from context for ownership checks — keeps function signatures concise
- In `GetByID` / `Update` / `Delete`, compare `item.OwnerID != userID` and return `Forbidden` to enforce ownership
- In `Create`, set ownership fields from the extracted userID
- In `Update`, fetch-then-mutate: find model, set fields, save
- Keep manual field mapping (`example.Name = body.Name`) rather than dumping all fields
- Model-to-DTO mapping belongs in the controller layer, not here
