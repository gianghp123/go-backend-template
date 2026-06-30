# controllers/

**Purpose:** HTTP request/response handlers. Thin layer — binds input, extracts userID from context, calls service, maps model→DTO, writes output.

**Rules:**
- No business logic — delegate to services
- Bind request params with your framework (Gin, Echo, Chi)
- Extract `userID` from context via `utils.GetCtx` and pass to service as param (for Create, List, etc.)
- Convert framework errors to `response.AppError`
- Map models from services to response DTOs using `utils.MapToDTO` / `utils.MapToDTOs`
- Write responses via `response.Success()` / `response.Fail()`

## Example

```go
type ExampleController struct {
    svc ExampleService
}

func NewController(svc ExampleService) *ExampleController {
    return &ExampleController{svc: svc}
}

// List — controller extracts userID, passes to service
func (ctrl *ExampleController) List(c *gin.Context) {
    var q req.ListExampleQuery
    if err := c.ShouldBindQuery(&q); err != nil {
        c.JSON(http.StatusBadRequest, response.Fail(response.BadRequest(err.Error())))
        return
    }

    userID := utils.GetCtx[string](c.Request.Context(), enums.ContextKeyUserID)

    result, appErr := ctrl.svc.List(c.Request.Context(), userID, q)
    if appErr != nil {
        c.JSON(appErr.Code, response.Fail(appErr))
        return
    }

    var examples []res.ExampleRes
    if err := utils.MapToDTOs(result.Data, &examples); err != nil {
        c.JSON(http.StatusInternalServerError, response.Fail(response.Internal("failed to map examples")))
        return
    }

    c.JSON(http.StatusOK, response.Success(response.PaginatedResponse[res.ExampleRes]{Data: examples, Meta: result.Meta}))
}

// GetByID — auth handled inside service, controller just passes the ID
func (ctrl *ExampleController) GetByID(c *gin.Context) {
    var body req.GetExampleReq
    if err := c.ShouldBindUri(&body); err != nil {
        c.JSON(http.StatusBadRequest, response.Fail(response.BadRequest(err.Error())))
        return
    }

    example, appErr := ctrl.svc.GetByID(c.Request.Context(), body.ID)
    if appErr != nil {
        c.JSON(appErr.Code, response.Fail(appErr))
        return
    }

    var result res.ExampleRes
    if err := utils.MapToDTO(example, &result); err != nil {
        c.JSON(http.StatusInternalServerError, response.Fail(response.Internal("failed to map example")))
        return
    }

    c.JSON(http.StatusOK, response.Success(result))
}

// Create — controller extracts userID, passes to service
func (ctrl *ExampleController) Create(c *gin.Context) {
    var body req.CreateExampleReq
    if err := c.ShouldBindJSON(&body); err != nil {
        c.JSON(http.StatusBadRequest, response.Fail(response.BadRequest(err.Error())))
        return
    }

    userID := utils.GetCtx[string](c.Request.Context(), enums.ContextKeyUserID)

    example, appErr := ctrl.svc.Create(c.Request.Context(), userID, body)
    if appErr != nil {
        c.JSON(appErr.Code, response.Fail(appErr))
        return
    }

    var result res.ExampleRes
    if err := utils.MapToDTO(example, &result); err != nil {
        c.JSON(http.StatusInternalServerError, response.Fail(response.Internal("failed to map example")))
        return
    }

    c.JSON(http.StatusCreated, response.Success(result))
}
```

## Admin controller

A separate admin controller shares the same service but routes are guarded by role middleware:

```go
type ExampleAdminController struct {
    svc ExampleService
}

func NewAdminController(svc ExampleService) *ExampleAdminController {
    return &ExampleAdminController{svc: svc}
}
```

**Recommendations:**
- Bind query: `c.ShouldBindQuery` | body: `c.ShouldBindJSON` | path: `c.ShouldBindUri`
- Always use `c.Request.Context()` for context propagation
- Swagger docs via `@Summary`, `@Tags`, `@Success`, `@Router` annotations
- Map models to response DTOs here, not in the service layer
- Extract `userID` from context here — service doesn't do this
- Accept the service **interface** in the struct field, not the concrete type
