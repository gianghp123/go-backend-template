# internal/modules/

**Purpose:** Domain modules — one folder per business feature. All business logic lives here.

**Structure per module:**

```
<module>/
├── controllers/
│   ├── <name>.controller.go        # User-facing handlers
│   └── <name>-admin.controller.go  # Admin handlers (same service)
├── services/
│   └── <name>.service.go           # Business logic
├── repositories/
│   └── <name>.repo.go              # Data access
├── dtos/
│   ├── req/                        # Request DTOs
│   └── res/                        # Response DTOs
└── <name>.module.go                # Route registration
```

**Rules:**
- Modules are independent — no cross-module imports
- Controllers call services, services call repositories
- Both user and admin controllers share the same service
- See `MODULE_GUIDE.md` for step-by-step creation
