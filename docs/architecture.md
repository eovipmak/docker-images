# Architecture

v-insight is composed of three main components:

- Backend API (Go + Gin) — REST API handling authentication, monitors, alert rules, incidents.
- Worker (Go) — background jobs for health checks, SSL checks, alert evaluation, and notifications.
- Frontend (SvelteKit) — web UI and an API proxy that forwards browser requests to the backend (CORS-free).

## CORS-free proxy

The frontend proxies `/api/*` requests to the backend so the browser never performs cross-origin requests. This eliminates the need for CORS middleware.

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
├── docker/     # Dockerfiles
└── docs/       # Documentation (this folder)
```

## Multi-tenant model

All main tables include `tenant_id` for strong data isolation. Handlers and repositories filter by tenant context.
