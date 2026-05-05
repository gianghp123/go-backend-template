# internal/core/errors/

**Purpose:** Standardized application errors.

## Types

- `AppError{Code int, Message string}` — implements `error` interface
- `New(code, defaultMsg, ...msg)` — create with optional custom message
- Predefined functions: `BadRequest()`, `Unauthorized()`, `Forbidden()`, `NotFound()`, `Conflict()`, `Internal()` — all accept optional `msg ...string`

## Usage

```go
if err != nil {
    return errors.NotFound("user not found")
}

// or with custom message
return errors.BadRequest("invalid email format")
```
