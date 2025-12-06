# V-Insight Development Guide

## Project Overview

V-Insight is a Docker-based multi-tenant monitoring SaaS platform with automated alerting. **Stack:** Go 1.23 (Gin backend :8080, Fiber worker :8081), SvelteKit/TypeScript frontend (:3000), PostgreSQL 15 (:5432).

**CRITICAL:** Uses CORS-free proxy architecture. Frontend (`src/routes/api/[...path]/+server.ts`) proxies all `/api/*` requests to backend. Do not add CORS middleware to backend.

**Architecture Refactoring:** The project now uses a **Go Workspace** (`go.work`) with a `shared` module to centralize domain logic and repositories, reducing duplication between Backend and Worker services.

## Features

**Monitoring:** HTTP/HTTPS health checks, SSL certificate monitoring, custom intervals
**Alerting:** Alert rules (down, slow_response, ssl_expiry), automatic incident creation/resolution
**Notifications:** Webhook, Discord integrations; email-ready
**Worker Jobs:** Health checks (30s), SSL checks (5m), alert evaluation (1m), notifications (30s)

## Quick Start

**REQUIRED:** `cp .env.example .env` before starting

```bash
make up          # Start all services (wait ~30s for PostgreSQL init)
```

**Hot reload enabled:** Backend/worker use Air (`.air.toml`), frontend uses Vite HMR.

## Commands

**Docker:** `make up|down|logs|logs-backend|logs-worker|logs-frontend|rebuild|clean|ps|restart`
**Tests:** `cd backend && go test ./...` (required before commit)
**Frontend:** `cd frontend && npm run check` (TypeScript validation)
**E2E:** `cd frontend && npm run test:e2e`
**Migrations:** `make migrate-up|migrate-down|migrate-create|migrate-version`

## Project Structure

```
/
├── backend/
│   ├── cmd/api/main.go          # Entry point, Gin routes, middleware
│   ├── internal/
│   │   ├── api/handlers/        # HTTP handlers (monitors, alerts, channels)
│   │   ├── domain/
│   │   │   ├── entities/        # Core entities (Monitor, AlertRule, Incident, etc.)
│   │   │   ├── repository/      # Repository interfaces
│   │   │   └── service/         # Business logic (alert evaluation)
│   │   └── repository/postgres/ # PostgreSQL implementations + tests
│   ├── migrations/              # Database migrations (auto-run on startup)
│   └── .air.toml                # Hot-reload config
├── worker/
│   ├── cmd/worker/main.go       # Job scheduler, cron jobs
│   ├── internal/jobs/           # HealthCheckJob, SSLCheckJob, AlertEvaluatorJob, NotificationJob
│   └── .air.toml                # Hot-reload config
├── frontend/
│   ├── src/routes/
│   │   ├── api/[...path]/+server.ts   # API proxy (CRITICAL!)
│   │   ├── login/, dashboard/, monitors/, alerts/, settings/
│   │   └── lib/                 # Shared components, API client
│   ├── svelte.config.js         # Uses adapter-node
│   ├── vite.config.js           # Port 3000, usePolling: true
│   └── tests/                   # Playwright E2E tests
├── docker/                      # Dockerfiles for backend, worker, frontend
├── docker-compose.yml           # All services with health checks
├── Makefile                     # Convenience commands
└── .env.example                 # Template (MUST copy to .env)
```

## Database Schema

**User-scoped tables (multi-tenant ready):**
- `tenants`, `users`, `tenant_users` - Multi-tenant organization and accounts (partially implemented)
- `monitors`, `monitor_checks` - Website/service monitoring and check results (scoped by user_id)
- `alert_rules`, `alert_channels`, `alert_rule_channels` - Alert configuration (scoped by user_id)
- `incidents` - Triggered alerts with resolution tracking (scoped by user_id)

**Migrations:** Auto-run on backend startup. Files in `backend/migrations/`. Use `golang-migrate` for manual operations.

## API Endpoints

**Monitors:** `/api/v1/monitors` (POST, GET, GET/:id, PUT/:id, DELETE/:id)
**Alert Rules:** `/api/v1/alert-rules` (POST, GET, GET/:id, PUT/:id, DELETE/:id)
**Alert Channels:** `/api/v1/alert-channels` (POST, GET, GET/:id, PUT/:id, DELETE/:id)
**Health:** `/health` on backend:8080 and worker:8081

## Alert System

**Trigger Types:**
- `down`: Monitor fails N consecutive checks (threshold = count)
- `slow_response`: Response time > threshold (ms)
- `ssl_expiry`: Certificate expires within threshold (days)

**Channel Types:**
- `webhook`: Generic POST with JSON payload
- `discord`: Rich embed with color-coded status
- `email`: Ready for implementation

**Flow:** Monitor check → Alert Evaluator (1m) → Create/resolve incident → Notification Job (30s) → Send to channels

## Common Issues & Solutions

**Services fail:** Ensure `.env` exists: `cp .env.example .env`
**Permission denied (npm):** `sudo chown -R $USER:$USER frontend/node_modules frontend/.svelte-kit`
**Hot-reload not working:** Check Air logs, look for syntax errors in `build-errors.log`
**Frontend changes not appearing:** Vite polling enabled; try `docker compose restart frontend`
**Database connection errors:** Wait ~10s for PostgreSQL init; verify with `docker compose ps`
**Port conflicts:** `docker compose down` or change ports in `.env`

## Making Code Changes

**Backend:** Edit `backend/cmd/` or `backend/internal/` → Air rebuilds → Test: `curl localhost:8080/health`
**Worker:** Edit `worker/cmd/` or `worker/internal/jobs/` → Air rebuilds → Test: `curl localhost:8081/health`
**Frontend:** Edit `frontend/src/` → Vite HMR → Type check: `npm run check`

**Add migration:** `make migrate-create name=description` → Edit `.up.sql` and `.down.sql` → Restart backend

## Validation Before Committing

```bash
make rebuild && docker compose ps      # All services healthy
cd backend && go test ./...            # Backend tests pass (REQUIRED)
cd frontend && npm run check           # TypeScript validates (REQUIRED)
git status                             # No tmp/, node_modules/ committed
```

## Code Patterns and Conventions

**User-based isolation (multi-tenant ready):**
- Entities include `user_id int` field for ownership (tenant tables exist but not fully integrated)
- Handlers retrieve user from context: `userIDValue, _ := c.Get("user_id")`
- Always verify ownership: `if monitor.UserID != userID { return forbidden }`
- Repositories have `GetByUserID(userID int)` methods

**API proxy implementation:**
- Frontend `src/routes/api/[...path]/+server.ts` forwards all methods to backend
- Strips problematic headers ('host', 'connection') and adds CORS headers for browser compatibility
- Uses `BACKEND_API_URL` env var (default: 'http://backend:8080')

**Worker job patterns:**
- Jobs query monitors directly without user filtering (process all users)
- Use executor for concurrent processing
- Store results in `monitor_checks` table with success/error details

**Middleware chain:**
- AuthRequired validates JWT and sets user_id/role in context
- AdminRequired checks role == "admin" for admin routes

**Database patterns:**
- Auto-migrations run on backend startup
- Use `sql.Null*` types for optional fields in structs
- Timestamps: `created_at`, `updated_at` managed by triggers

**Go Workspace:**
- Project uses `go.work` with `shared/` module for domain entities and repositories
- Run `go work sync` to update workspace

## Critical Architecture Notes

1. **CORS handled by proxy** - Frontend proxy adds CORS headers; do not add CORS middleware to backend
2. **API proxy:** `frontend/src/routes/api/[...path]/+server.ts` handles all API forwarding
3. **User isolation:** Every query MUST filter by `user_id` from JWT context
4. **Docker primary:** Use Docker for development; local builds are fallback only
5. **Migrations auto-run:** Backend applies migrations on startup automatically
6. **Worker jobs schedule:**
   - Health checks: every 30 seconds
   - SSL checks: every 5 minutes  
   - Alert evaluation: every minute
   - Notifications: every 30 seconds

## Debugging Tips

- When debugging, check logs periodically using `make logs-backend`, `make logs-worker`, or `make logs-frontend` to identify errors, but avoid continuous monitoring to prevent getting stuck in log viewing loops.
- Use specific log filters (e.g., `grep` or `tail -f` with filters) to find relevant information quickly without overwhelming output.
- Prefer using breakpoints in code editors, unit tests, or curl commands for step-by-step debugging instead of relying solely on logs.
- If logs are not showing expected output, ensure services are running with `docker compose ps` and check for syntax errors in code.

## Development Workflow

1. Start services: `make up` (wait ~30s)
2. Make changes (hot-reload enabled)
3. Test manually via browser/curl
4. Run tests: `go test ./...` and `npm run check`
5. Check logs: `make logs-backend` / `make logs-worker` / `make logs-frontend`
6. Commit only when tests pass

## Key Technologies

**Backend:** Go 1.23, Gin (HTTP), GORM (ORM), golang-migrate, Air (hot-reload), testify (testing)
**Worker:** Go 1.23, Fiber (HTTP server), robfig/cron (scheduler), same database/ORM as backend
**Frontend:** SvelteKit v2, TypeScript, Vite v5, Tailwind CSS, Playwright (E2E)
**Database:** PostgreSQL 15 with JSONB, UUIDs, auto-migrations
**Infrastructure:** Docker Compose, health checks, volume mounts for hot-reload