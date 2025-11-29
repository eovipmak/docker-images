# Project: V-Insight

## General Instructions
- You are an AI coding agent working on V-Insight, a Docker-based multi-tenant monitoring SaaS with Go backend, worker, and SvelteKit frontend.
- Prioritize multi-tenant security: every query/modification must filter by `tenant_id` from JWT context.
- Use CORS-free proxy architecture; never add CORS middleware—proxy handles all `/api/*` requests.
- Follow existing patterns: handlers retrieve `tenantID := c.Get("tenant_id")`; verify ownership; repos use `GetByTenantID(tenantID int)`.
- Worker jobs process all tenants without filtering; store results in `monitor_checks` with success/error details.
- Auto-migrations run on backend startup; use `sql.Null*` for optional fields; timestamps managed by triggers.

## Coding Style
- Go: Use tabs for indentation; PascalCase for exported types/functions; camelCase for unexported; Gin for HTTP, Fiber for worker.
- TypeScript/SvelteKit: Use 2 spaces for indentation; camelCase for variables/functions; PascalCase for types.
- Middleware chain: AuthRequired → TenantRequired; validate access via `tenantUserRepo.HasAccess()`.
- API proxy in `frontend/src/routes/api/[...path]/+server.ts`: strips 'host', 'connection'; uses `BACKEND_API_URL`.
- Database: PostgreSQL 15 with JSONB; migrations in `backend/migrations/`; create with `make migrate-create`.
- Tests: Go tests in backend/worker; TypeScript checks in frontend; Playwright E2E.

## Specific Components
- **Backend Handlers** (`backend/internal/api/handlers/`): Enforce tenant scoping; example: `if monitor.TenantID != tenantID { return forbidden }`.
- **Worker Jobs** (`worker/internal/jobs/`): Use executor for concurrent processing; query all monitors.
- **Frontend Proxy** (`frontend/src/routes/api/[...path]/+server.ts`): Forwards methods to backend without CORS.
- **Migrations** (`backend/migrations/`): Up/down SQL files; auto-run on startup.
- **Alert Flow**: Check → Eval (1m) → Incident → Notify (30s) to webhook/Discord channels.

## Dependencies and Integration
- Avoid new deps unless necessary; state reason if adding.
- External integrations: Webhook POST JSON; Discord embeds; email ready.
- Worker schedules: Health checks 30s, SSL 5m, alert eval 1m, notifications 30s.

## Workflows
- Start: `cp .env.example .env; make up` (wait 30s).
- Test: `cd backend && go test ./...`; `cd frontend && npm run check`; `cd frontend && npm run test:e2e`.
- Debug: Use `make logs-*`; prefer breakpoints; check `docker compose ps`; limit log viewing to max 1 minute to avoid loops.