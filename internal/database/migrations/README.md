# migrations/

**Purpose:** SQL migration files for managing database schema changes.

**What it can contain:**
- Timestamp-prefixed `.sql` files with `UP` and `DOWN` sections
- Seed data scripts
- Migration config (if using goose, golang-migrate, etc.)

**Recommendations:**
- `pressly/goose` — SQL-first migration tool, supports Go migrations too
- `golang-migrate/migrate` — popular, supports many DB drivers
- `atlasgo.io` — declarative schema migrations

**Conventions:**
- File format: `YYYYMMDDHHMMSS_description.sql`
- Each file has:
  ```sql
  -- +goose Up
  CREATE TABLE examples (...);

  -- +goose Down
  DROP TABLE IF EXISTS examples;
  ```
- One conceptual change per migration (avoid mixing table creation + data seed)

**Commands (goose):**

```bash
make migrate-create name=create_users_table   # scaffold new migration
make migrate-up                                # apply pending
make migrate-down                              # rollback last batch
```

Add more migration tools or seed scripts here as your project grows.
