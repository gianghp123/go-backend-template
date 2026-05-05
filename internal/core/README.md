# internal/core/

**Purpose:** Shared, reusable primitives used across all modules. No business logic.

## What it contains

| Package | Responsibility |
|---------|---------------|
| `enums/` | Shared enum types (user role, context keys) |
| `errors/` | `AppError` with predefined `BadRequest`, `NotFound`, etc. (accepts optional messages) |
| `response/` | Standardized API response envelope (`BaseResponse[T]`) and pagination helpers |
| `logger/` | Logger singleton pattern (README only — implement with your logger) |

## Guidelines

- Core packages must NOT import from `internal/configs/`, `internal/database/`, or `internal/modules/`
- Adding new primitives here is rare — most code lives in modules
- Concrete implementations (database, logger, policy) are described in READMEs — choose your own libraries
