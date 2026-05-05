# dtos/req/

**Purpose:** Request DTOs — shape and validate incoming data from the API layer.

**Rules:**
- Use `json:"..."` tags for body binding
- Use `form:"..."` tags for query params
- Use `uri:"..."` tags for path params
- `ListQuery` DTOs should have a `ToQuery()` method that builds a `*database.Query`

## Example

```go
type ListExampleQuery struct {
    OwnerID string `form:"ownerId"`
    Search  string `form:"search"`
    Page    int    `form:"page"`
    Limit   int    `form:"limit"`
}

func (q ListExampleQuery) ToQuery() *database.Query {
    query := database.NewQuery().SetPage(q.Page).SetLimit(q.Limit)
    if q.OwnerID != "" {
        query.SetFilter("owner_id", q.OwnerID)
    }
    if q.Search != "" {
        query.SetFilter("search", q.Search)
    }
    return query
}

type CreateExampleReq struct {
    Name        string `json:"name" binding:"required"`
    Description string `json:"description"`
}

type UpdateExampleReq struct {
    Name        string `json:"name"`
    Description string `json:"description"`
}

type GetExampleReq struct {
    ID uint `uri:"exampleId"`
}
```

**Recommendations:**
- Validate with `binding:"required"`, `binding:"min=1"`, etc.
- Use pointer types (`*uint`, `*string`) for nullable filter fields
- `ToQuery()` keeps filter logic self-contained in the DTO — repo doesn't need to know about filters
