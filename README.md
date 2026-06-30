# Go Backend Template

A reusable Go backend project skeleton — no framework lock-in, plug in your own stack.

```
├── cmd/
│   ├── servers/         # HTTP server binaries (main, admin, internal, etc.)
│   ├── workers/         # Background job processors (add as needed)
│   ├── lambdas/         # Lambda function handlers (add as needed)
│   └── webhooks/        # Webhook handlers (add as needed)
├── docker/              # Docker Compose for local services
├── docs/                # Specifications and plans
├── internal/
│   ├── auth/            # Auth provider + authorization policy (README)
│   ├── configs/         # Typed config loading from env
│   ├── core/
│   │   ├── enums/       # Shared enum types + context keys
│   │   ├── errors/      # AppError + sentinel errors + error mapping
│   │   ├── response/    # API response envelope + pagination
│   │   └── logger/      # Logger pattern (README)
│   ├── database/        # Query builder, models, migrations
│   ├── middleware/       # Auth & role middleware pattern (README)
│   ├── modules/         # Domain features (README patterns)
│   │   ├── example/     # Pattern docs per layer (controllers, services, repos, DTOs)
│   │   └── MODULE_GUIDE.md
│   ├── storage/         # Object storage abstraction (interface + noop)
│   └── utils/           # GetCtx, MapToDTO, GetEnv, GetEnvInt
├── .env.example
├── .gitignore
├── Makefile
└── README.md
```

## Get started

```bash
cp .env.example .env
make up
```

## Add a feature

1. Create models in `internal/database/models/`
2. Add migrations to `internal/database/migrations/`
3. Follow `internal/modules/MODULE_GUIDE.md` — it walks through controllers, services, repos, DTOs
4. Register routes in `cmd/servers/`

## Swap providers

| To swap | Edit |
|---------|------|
| Auth | Implement `AuthProvider` interface (see `internal/auth/README.md`) |
| Database | Replace GORM in `internal/database/` with SQLx, Bun, etc. |
| Storage | Implement `StorageProvider` interface (see `internal/storage/README.md`) |
| Logger | Replace Zap in `core/logger/` with slog, logrus, etc. |

## Key patterns

- **DTOs** — request/response structs in `dtos/req/` and `dtos/res/`, models in `database/models/`
- **Repos** — accept models + `database.Query`, return models + stdlib errors, never DTOs
- **Services** — accept only the params they need (controller passes `userID` as param), use `policy.ActorFromContext` for auth checks, return `*errors.AppError`
- **Controllers** — bind input, call service, write response via `response.Success/Fail`
