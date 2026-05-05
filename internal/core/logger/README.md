# internal/core/logger/

**Purpose:** Logging abstraction — a singleton accessible across the application.

**Pattern:** Initialize with `Init(cfg)` at startup and use via `logger.L`.

**Recommendation:** Use `go.uber.org/zap` for structured logging or `log/slog` (Go 1.21+).
