# services/

**Purpose:** Business logic layer — orchestrates repositories, enforces policy, maps DTOs ↔ models.

**Rules:**
- Accept DTOs from controllers (for reusability and clear contracts)
- Extract user claims from context via `utils.GetCtx[T any](ctx, key)` (keeps function signatures clean)
- Map repo errors (`errors.Is(err, coreError.ErrNotFound)`) to `*response.AppError`
- Map models ↔ DTOs with `utils.MapToDTO` / `utils.MapToDTOs`
- No HTTP concerns — no `*gin.Context`, no status codes in logic

## Example

```go
type IExampleService interface {
    List(ctx context.Context, query req.ListExampleQuery) (*response.PaginatedResponse[res.ExampleRes], *response.AppError)
    GetByID(ctx context.Context, body req.GetExampleReq) (*res.ExampleRes, *response.AppError)
    Create(ctx context.Context, body req.CreateExampleReq) (*res.ExampleRes, *response.AppError)
    Update(ctx context.Context, id uint, body req.UpdateExampleReq) (*res.ExampleRes, *response.AppError)
    Delete(ctx context.Context, id uint) *response.AppError
}

type exampleService struct {
    repo repositories.IExampleRepo
}

func NewExampleService(repo repositories.IExampleRepo) IExampleService {
    return &exampleService{repo: repo}
}

func (s *exampleService) List(ctx context.Context, query req.ListExampleQuery) (*response.PaginatedResponse[res.ExampleRes], *response.AppError) {
    result, err := s.repo.FindAll(ctx, query.ToQuery())
    if err != nil {
        return nil, response.Internal("failed to list examples")
    }

    var examples []res.ExampleRes
    if err := utils.MapToDTOs(result.Data, &examples); err != nil {
        return nil, response.Internal("failed to map examples")
    }

    return &response.PaginatedResponse[res.ExampleRes]{Data: examples, Meta: result.Meta}, nil
}

func (s *exampleService) GetByID(ctx context.Context, body req.GetExampleReq) (*res.ExampleRes, *response.AppError) {
    example, err := s.repo.FindByID(ctx, body.ID)
    if err != nil {
        if errors.Is(err, coreError.ErrNotFound) {
            return nil, response.NotFound("example not found")
        }
        return nil, response.Internal("failed to get example")
    }

    var result res.ExampleRes
    if err := utils.MapToDTO(example, &result); err != nil {
        return nil, response.Internal("failed to map example")
    }

    return &result, nil
}

func (s *exampleService) Create(ctx context.Context, body req.CreateExampleReq) (*res.ExampleRes, *response.AppError) {
    userID := utils.GetCtx[string](ctx, enums.ContextKeyUserID)

    m := &models.Example{
        Name:        body.Name,
        Description: body.Description,
        OwnerID:     userID,
    }

    if err := s.repo.Create(ctx, m); err != nil {
        return nil, response.Internal("failed to create example")
    }

    var result res.ExampleRes
    if err := utils.MapToDTO(m, &result); err != nil {
        return nil, response.Internal("failed to map example")
    }

    return &result, nil
}

func (s *exampleService) Update(ctx context.Context, id uint, body req.UpdateExampleReq) (*res.ExampleRes, *response.AppError) {
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

    var result res.ExampleRes
    if err := utils.MapToDTO(example, &result); err != nil {
        return nil, response.Internal("failed to map example")
    }

    return &result, nil
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
- Extract `userID := utils.GetCtx[string](ctx, enums.ContextKeyUserID)` for ownership — keeps function signatures concise
- In `Create`, set ownership fields from the extracted userID
- In `Update`, fetch-then-mutate: find model, set fields, save
- Keep manual field mapping (`example.Name = body.Name`) rather than dumping all fields
