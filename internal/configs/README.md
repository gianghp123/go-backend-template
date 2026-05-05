# internal/configs/

**Purpose:** Loads environment variables and exposes typed configuration structs.

**Structure:**
- `index.go` — aggregates all configs, calls `godotenv.Load`
- `*.config.go` — one file per config domain (server, database, auth, logger)

**Conventions:**
- Use `utils.GetEnv` / `utils.GetEnvInt` for reading env vars
- Each sub-config has its own `load*Config()` function
- Provide sensible defaults

**Recommendation:** Use `os.Getenv` for system env or `github.com/joho/godotenv` for `.env` file loading.
