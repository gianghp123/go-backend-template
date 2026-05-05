# internal/auth/

**Purpose:** Authentication and authorization — token verification, role gates, ownership checks, access decisions.

## Authentication

**Pattern:** An `IAuthProvider` interface with `VerifyToken(ctx, token) (UserClaims, error)`. Implement once per provider (Firebase, Auth0, Supabase Auth, custom JWT) and swap in `Init()`.

Auth middleware reads from `context.Context` — set by `middleware/auth`, consumed by policy functions and services.

## Authorization

### Simple enforcement (built-in)

For most cases, inline checks are sufficient:

```go
func EnforceRole(role enums.UserRole, allowedRoles ...enums.UserRole) *errors.AppError {
    for _, r := range allowedRoles {
        if role == r {
            return nil
        }
    }
    return errors.Forbidden("insufficient permissions")
}

func EnforceOwnerOrAdmin(userID string, role enums.UserRole, resourceOwnerID string) *errors.AppError {
    if role == enums.UserRoleAdmin {
        return nil
    }
    if userID == resourceOwnerID {
        return nil
    }
    return errors.Forbidden("access denied")
}
```

**Usage in services:**

```go
func (s *exampleService) GetByID(ctx context.Context, body req.GetExampleReq) (*res.ExampleRes, *errors.AppError) {
    userID := utils.GetCtx[string](ctx, enums.ContextKeyUserID)
    role := utils.GetCtx[enums.UserRole](ctx, enums.ContextKeyUserRole)

    item, err := s.repo.FindByID(ctx, body.ID)
    if err != nil {
        return nil, mapError(err)
    }
    if err := policy.EnforceOwnerOrAdmin(userID, role, item.OwnerID); err != nil {
        return nil, err
    }
    ...
}
```

### Complex authorization (RBAC / ABAC)

For role hierarchies, permission matrices, or attribute-based rules, integrate with a dedicated library:

```go
import "github.com/casbin/casbin/v2"

e, _ := casbin.NewEnforcer("model.conf", "policy.csv")
ok, _ := e.Enforce(userID, resource, action)
```

**Recommendations:**
- Simple checks (role + ownership): keep inline — no external dependency needed
- Complex RBAC/ABAC: `github.com/casbin/casbin/v2` — flexible model-based authorization
- Alternative: `github.com/ory/keto` — Ory's access control server (gRPC)
- For multi-tenant: add tenant scoping to all policy checks

Extend this package with domain-specific policy functions as your authorization model grows.
