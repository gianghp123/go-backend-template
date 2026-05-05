# cmd/

**Purpose:** Application entry points — one folder per binary type, each containing multiple executables.

**What it can contain:**

| Folder | Purpose |
|--------|---------|
| `servers/` | HTTP API servers (multiple: main.go, admin.go, internal.go, etc.) |
| `webhooks/` | Webhook handlers (Stripe events, GitHub hooks) |
| `workers/` | Background job processors (queue consumers, schedulers) |
| `lambdas/` | AWS Lambda function handlers (multiple lambdas) |
| `jobs/` | CLI tools (migrations, seeding, maintenance) |

**Conventions:**
- Each folder produces multiple binaries: `go build ./cmd/servers/main` or via Makefile targets
- Keep entry points thin — call setup functions from internal packages
- One `package main` per binary; one folder can hold many `main.go` variants
