# internal/core/enums/

**Purpose:** Shared enum type definitions used across the project.

## Files

| File | Contents |
|------|----------|
| `context-key.enum.go` | `ContextKey` type and keys for storing values in `context.Context` |
| `user-role.enum.go` | `UserRole` type and role constants (`UserRoleAdmin`, `UserRoleUser`) |

## Usage

```go
import "your-project/internal/core/enums"

ctx = context.WithValue(ctx, enums.ContextKeyUserID, "user-1")
ctx = context.WithValue(ctx, enums.ContextKeyUserRole, enums.UserRoleAdmin)

userID := utils.GetCtx[string](ctx, enums.ContextKeyUserID)
role := utils.GetCtx[enums.UserRole](ctx, enums.ContextKeyUserRole)
```

## Conventions
- One file per enum domain
- Type-safe string constants (`type MyEnum string`)
- Add new context keys or role types here as the project grows
