# services/

**Purpose:** Business logic layer — orchestrates repositories, enforces policy.

**Rules:**
- Accept only the params you need — `userID` is passed from the controller when it's domain data
- Call `policy.ActorFromContext(ctx)` **only** when authorization is needed (GetByID, Update, Delete)
- Return domain models directly — no DTO mapping here
- Map repo errors (`errors.Is(err, coreError.ErrNotFound)`) to `*response.AppError`
- No HTTP concerns — no `*gin.Context`, no status codes in logic
- Never call `utils.GetCtx` or `utils.MapToDTO` — those are controller concerns

## Example

```go
// Interface — exported for controllers to depend on
type ExampleService interface {
    List(ctx context.Context, userID string, query req.ListExampleQuery) (*response.PaginatedResult[*models.Example], *response.AppError)
    GetByID(ctx context.Context, id uint) (*models.Example, *response.AppError)
    Create(ctx context.Context, userID string, body req.CreateExampleReq) (*models.Example, *response.AppError)
    Update(ctx context.Context, id uint, body req.UpdateExampleReq) (*models.Example, *response.AppError)
    Delete(ctx context.Context, id uint) *response.AppError
}

// Struct — unexported
type exampleService struct {
    repo ports.ExampleRepository
}

// Constructor — returns concrete struct
func NewService(repo ports.ExampleRepository) *exampleService {
    return &exampleService{repo: repo}
}

// List — userID is domain data, controller passes it
func (s *exampleService) List(ctx context.Context, userID string, query req.ListExampleQuery) (*response.PaginatedResult[*models.Example], *response.AppError) {
    dbQuery := query.ToQuery()
    dbQuery.SetFilter("owner_id", userID)
    result, err := s.repo.FindAll(ctx, dbQuery)
    if err != nil {
        return nil, response.Internal("failed to list examples")
    }
    return result, nil
}

// GetByID — needs auth check, uses ActorFromContext + CanRead
func (s *exampleService) GetByID(ctx context.Context, id uint) (*models.Example, *response.AppError) {
    actor, err := policy.ActorFromContext(ctx)
    if err != nil {
        return nil, response.Unauthorized()
    }
    example, err := s.repo.FindByID(ctx, id)
    if err != nil {
        if errors.Is(err, coreError.ErrNotFound) {
            return nil, response.NotFound("example not found")
        }
        return nil, response.Internal("failed to get example")
    }
    if err := policy.CanRead(actor, example.OwnerID); err != nil {
        return nil, response.Forbidden()
    }
    return example, nil
}

// Create — userID is domain data, controller passes it
func (s *exampleService) Create(ctx context.Context, userID string, body req.CreateExampleReq) (*models.Example, *response.AppError) {
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

// Update — needs auth check, uses ActorFromContext + CanMutate
func (s *exampleService) Update(ctx context.Context, id uint, body req.UpdateExampleReq) (*models.Example, *response.AppError) {
    actor, err := policy.ActorFromContext(ctx)
    if err != nil {
        return nil, response.Unauthorized()
    }
    example, err := s.repo.FindByID(ctx, id)
    if err != nil {
        if errors.Is(err, coreError.ErrNotFound) {
            return nil, response.NotFound("example not found")
        }
        return nil, response.Internal("failed to get example for update")
    }
    if err := policy.CanMutate(actor, example.OwnerID); err != nil {
        return nil, response.Forbidden()
    }
    example.Name = body.Name
    example.Description = body.Description
    if err := s.repo.Update(ctx, example); err != nil {
        return nil, response.Internal("failed to update example")
    }
    return example, nil
}

// Delete — needs auth check, uses ActorFromContext + CanMutate
func (s *exampleService) Delete(ctx context.Context, id uint) *response.AppError {
    actor, err := policy.ActorFromContext(ctx)
    if err != nil {
        return response.Unauthorized()
    }
    example, err := s.repo.FindByID(ctx, id)
    if err != nil {
        if errors.Is(err, coreError.ErrNotFound) {
            return response.NotFound("example not found")
        }
        return response.Internal("failed to get example")
    }
    if err := policy.CanMutate(actor, example.OwnerID); err != nil {
        return response.Forbidden()
    }
    if err := s.repo.Delete(ctx, id); err != nil {
        return response.Internal("failed to delete example")
    }
    return nil
}
```

**Recommendations:**
- Always return `*response.AppError` (not raw `error`) — controllers rely on `.Code` for HTTP status
- Use `errors.Is(err, coreError.ErrNotFound)` to detect 404 from repo
- **Create/List** — `userID` is domain data, controller passes it as a param. Service never extracts from context.
- **GetByID/Update/Delete** — needs authorization, service calls `policy.ActorFromContext(ctx)` + `policy.CanRead`/`policy.CanMutate`
- In `Update`, fetch-then-mutate: find model, set fields, save
- Keep manual field mapping (`example.Name = body.Name`) rather than dumping all fields
- Model-to-DTO mapping belongs in the controller layer, not here
- Constructor (`NewService`) returns concrete struct, not the interface — callers accept the interface
