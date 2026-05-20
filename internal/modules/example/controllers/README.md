# controllers/

**Purpose:** HTTP request/response handlers. Thin layer — binds input, calls service, maps model→DTO, writes output.

**Rules:**
- No business logic — delegate to services
- Bind request params with your framework (Gin, Echo, Chi)
- Convert framework errors to `response.AppError`
- Map models from services to response DTOs using `utils.MapToDTO` / `utils.MapToDTOs`
- Write responses via `response.Success()` / `response.Fail()`

## Example

```go
type ExampleController struct {
    svc services.IExampleService
}

func NewExampleController(svc services.IExampleService) *ExampleController {
    return &ExampleController{svc: svc}
}

// List godoc
// @Summary List examples
// @Tags examples
// @Accept json
// @Produce json
// @Param page query int false "Page" default(1)
// @Param limit query int false "Limit" default(10)
// @Success 200 {object} response.BaseResponse[response.PaginatedResponse[res.ExampleRes]]
// @Router /examples [get]
func (ctrl *ExampleController) List(c *gin.Context) {
    var q req.ListExampleQuery
    if err := c.ShouldBindQuery(&q); err != nil {
        c.JSON(http.StatusBadRequest, response.Fail(response.BadRequest(err.Error())))
        return
    }

    result, appErr := ctrl.svc.List(c.Request.Context(), q)
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

// GetByID godoc
// @Summary Get example by ID
// @Tags examples
// @Accept json
// @Produce json
// @Param id path int true "Example ID"
// @Success 200 {object} response.BaseResponse[res.ExampleRes]
// @Router /examples/{id} [get]
func (ctrl *ExampleController) GetByID(c *gin.Context) {
    var body req.GetExampleReq
    if err := c.ShouldBindUri(&body); err != nil {
        c.JSON(http.StatusBadRequest, response.Fail(response.BadRequest(err.Error())))
        return
    }

    example, appErr := ctrl.svc.GetByID(c.Request.Context(), body)
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

// Create godoc
// @Summary Create a new example
// @Tags examples
// @Accept json
// @Produce json
// @Param body body req.CreateExampleReq true "Example body"
// @Success 201 {object} response.BaseResponse[res.ExampleRes]
// @Router /examples [post]
func (ctrl *ExampleController) Create(c *gin.Context) {
    var body req.CreateExampleReq
    if err := c.ShouldBindJSON(&body); err != nil {
        c.JSON(http.StatusBadRequest, response.Fail(response.BadRequest(err.Error())))
        return
    }

    example, appErr := ctrl.svc.Create(c.Request.Context(), body)
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

## Ownership enforcement via DTO injection

Controllers can extract user claims from context and inject owner filters into DTOs before calling the service:

```go
func (ctrl *ExampleController) ListMyExamples(c *gin.Context) {
    var q req.ListExampleQuery
    if err := c.ShouldBindQuery(&q); err != nil {
        c.JSON(http.StatusBadRequest, response.Fail(response.BadRequest(err.Error())))
        return
    }
    // Force ownership: only list examples belonging to the current user
    userID := utils.GetCtx[string](c.Request.Context(), enums.ContextKeyUserID)
    q.OwnerID = userID

    result, appErr := ctrl.svc.List(c.Request.Context(), q)
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
```

## Admin controller

A separate admin controller shares the same service but routes are guarded by role middleware:

```go
type ExampleAdminController struct {
    svc services.IExampleService
}

func (ctrl *ExampleAdminController) Create(c *gin.Context) {
    var body req.CreateExampleReq
    if err := c.ShouldBindJSON(&body); err != nil {
        c.JSON(http.StatusBadRequest, response.Fail(response.BadRequest(err.Error())))
        return
    }

    example, appErr := ctrl.svc.Create(c.Request.Context(), body)
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

**Recommendations:**
- Bind query: `c.ShouldBindQuery` | body: `c.ShouldBindJSON` | path: `c.ShouldBindUri`
- Always use `c.Request.Context()` for context propagation
- Swagger docs via `@Summary`, `@Tags`, `@Success`, `@Router` annotations
- Map models to response DTOs here, not in the service layer
- For ownership enforcement, inject `utils.GetCtx[string](ctx, enums.ContextKeyUserID)` into request DTOs at the controller level
