# internal/core/response/

**Purpose:** Standardized API response envelope.

## Types

- `BaseResponse[T]` — generic response wrapper with `Success`, `Data`, `Error`, `Meta`
- `PaginatedResult[T]` — paginated list with `Items` and `Meta`
- `Meta` — `Page`, `Limit`, `Total`, `TotalPages`

## Functions

```go
response.Success(data)
response.SuccessWithMeta(data, meta)
response.Fail[T](errorMessage)
response.PaginatedResponse(items, total, page, limit)
```

## Recommendations

- Always use `BaseResponse` for consistency
- Controllers return `response.Success()` or `response.Fail()` in their HTTP responses
