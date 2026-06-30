# internal/auth/

**Purpose:** Authentication and authorization — token verification, role gates, ownership checks, access decisions.

## Authentication

**Pattern:** An `AuthProvider` interface with `VerifyToken(ctx, token) (UserClaims, error)`. Implement once per provider (Firebase, Auth0, Supabase Auth, custom JWT) and swap in `Init()`.

Auth middleware sets `userID` and `role` in `context.Context` — consumed by policy functions and services.

## Authorization

### Actor pattern (recommended)

The `Actor` struct bundles user identity (ID + role). Services extract it from context only when authorization is needed:

```go
type Actor struct {
    UserID string
    Role   enums.UserRole
}

func ActorFromContext(ctx context.Context) (*Actor, error) {
    userID := utils.GetCtx[string](ctx, enums.ContextKeyUserID)
    if userID == "" {
        return nil, ErrUnauthorized
    }
    role := utils.GetCtx[enums.UserRole](ctx, enums.ContextKeyUserRole)
    return &Actor{UserID: userID, Role: role}, nil
}
```

### Policy functions

Pure functions — actor + resource owner => allow/deny:

```go
func CanMutate(actor *Actor, resourceOwnerID string) error {
    if actor.Role == enums.UserRoleAdmin {
        return nil
    }
    if actor.UserID == resourceOwnerID {
        return nil
    }
    return ErrForbidden
}

func CanRead(actor *Actor, resourceOwnerID string) error {
    return CanMutate(actor, resourceOwnerID)
}
```

**Usage in services:**

```go
// Auth check — service uses ActorFromContext + CanMutate
func (s *exampleService) Delete(ctx context.Context, id uint) *errors.AppError {
    actor, err := auth.ActorFromContext(ctx)
    if err != nil {
        return errors.Unauthorized()
    }
    item, err := s.repo.FindByID(ctx, id)
    if err != nil {
        return errors.NotFound()
    }
    if err := auth.CanMutate(actor, item.OwnerID); err != nil {
        return errors.Forbidden()
    }
    return s.repo.Delete(ctx, id)
}

// Domain data — controller passes userID, no auth check needed
func (s *exampleService) Create(ctx context.Context, userID string, body req.CreateReq) (*models.Example, *errors.AppError) {
    item := &models.Example{Name: body.Name, OwnerID: userID}
    return item, s.repo.Create(ctx, item)
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
- Simple checks (role + ownership): keep inline with `Actor` + `CanMutate`/`CanRead`
- Complex RBAC/ABAC: `github.com/casbin/casbin/v2` — flexible model-based authorization
- Alternative: `github.com/ory/keto` — Ory's access control server (gRPC)
- For multi-tenant: add tenant scoping to all policy checks

Extend this package with domain-specific policy functions as your authorization model grows.
