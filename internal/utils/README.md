# internal/utils/

**Purpose:** General-purpose helper functions used across the project.

**Files:**

| File | Contents |
|------|----------|
| `env.util.go` | `GetEnv(key, fallback)` / `GetEnvInt(key, fallback)` — safe env var readers |
| `dto.util.go` | `MapToDTO[T]` / `MapToDTOs[T]` — generic struct mapping helpers |
| `context.util.go` | `GetCtx[T any](ctx, key)` — generic context value extractor |

**Guidelines:**
- Functions here must NOT import from `internal/modules/` or `internal/database/`
- Keep them pure, stateless, and testable
- Only add utilities that are genuinely shared across multiple packages

**Recommendations:**
- DTO mapping: `github.com/jinzhu/copier` — deep copy between structs
- Validation: `github.com/go-playground/validator` — struct tag validation
- Pagination: manual slice operations or `copier` + offset math
