# cmd/servers/

**Purpose:** HTTP API server entry points — can contain multiple server binaries.

```
cmd/servers/
├── main.go        # public API server
├── admin.go       # admin-only server (separate port/auth)
└── internal.go    # internal service-to-service API
```

## main.go example

```go
func main() {
    cfg := configs.Load()
    logger.Init(cfg.Logger)
    db := database.Init(cfg.Database)

    r := setupRouter()
    registerRoutes(r, db)
    r.Run(":" + cfg.Server.Port)
}
```

**Recommendations:**
- Framework: Gin, Echo, Chi, Fiber
- Rely on internal packages for all logic — entry points are wiring only
- Split across multiple files in the same `package main`
