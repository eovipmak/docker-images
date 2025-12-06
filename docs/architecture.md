# Architecture

v-insight is composed of three main components:

- Backend API (Go + Gin) — REST API handling authentication, monitors, alert rules, incidents.
- Worker (Go) — background jobs for health checks, SSL checks, alert evaluation, and notifications.
- Frontend (SvelteKit) — web UI and an API proxy that forwards browser requests to the backend (CORS-free).

## CORS-free proxy

The frontend proxies `/api/*` requests to the backend so the browser never performs cross-origin requests. This eliminates the need for CORS middleware.

## Notes for AI Agents & Automation

- The frontend API proxy is implemented at `frontend/src/routes/api/[...path]/+server.ts`. Do not add CORS middleware — the proxy pattern is intentional.
- User enforcement is critical: always filter queries and commands by `user_id` and validate user access in handlers and services.
- The backend and worker both rely on shared domain models in `shared/domain` and `shared/repository`.
- For schema changes, add migrations under `backend/migrations/`; ensure both `.up.sql` and `.down.sql` are present and tested.

## Worker jobs (schedules)

- Health checks: every 30s
- SSL checks: every 5m
- Alert evaluator: every 1m
- Notification job: every 30s

## Project structure (high level)

```
.
├── backend/    # Go backend (API, migrations)
├── worker/     # Background jobs
├── frontend/   # SvelteKit app + API proxy
├── shared/     # Shared domain logic and repositories
├── docker/     # Dockerfiles
└── docs/       # Documentation (this folder)
```

## User Isolation model

All main tables include `user_id` for strong data isolation. Handlers and repositories filter by user context.
