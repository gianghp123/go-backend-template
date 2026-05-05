# internal/database/

**Purpose:** Database connection, query builder, entity definitions, and SQL migrations.

**Structure:**

```
database/
├── README.md
├── query.go          # Generic query builder (filters, pagination, sorting)
├── models/           # Entity structs (one file per domain entity)
└── migrations/       # SQL migration files (timestamped)
```

## query.go

Generic `Query` struct with chainable methods:

```go
q := database.NewQuery().
    SetPage(1).
    SetLimit(10).
    SetFilter("owner_id", "user-1").
    SetOrderBy("created_at DESC")
```

Add `Count(tx)` and `Apply(tx)` methods adapted to your ORM (GORM, SQLx, etc.).

**Imports:**

```go
import "your-project/internal/database"
```

## Recommendations
- ORMs: GORM, SQLx, Bun, Ent
- Migrations: `pressly/goose` (SQL-first), `golang-migrate/migrate`, `atlasgo.io`
- Connection pooling: tune `MaxOpenConns`, `MaxIdleConns`, `ConnMaxLifetime`

## Conventions
- Models stay in `models/`, repositories stay in `modules/<name>/repositories/`
- Migration files: `YYYYMMDDHHMMSS_description.sql`
- Each migration is a single conceptual change

This folder is not limited to these files — add seed scripts, connection init, or DB adapters as needed.
