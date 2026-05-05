# internal/middleware/

**Purpose:** HTTP middleware — request pipeline filters executed before handlers.

**Common middleware:**
- **Auth** — extracts Bearer token, calls `IAuthProvider.VerifyToken`, stores claims in context
- **Role** — checks user role from context against required roles
- Logging, rate limiting, CORS, recovery

**Recommendation:** Use your framework's middleware system (Gin, Echo, Chi, etc). Keep middleware composable — each does one thing and passes to `next`.
