# V-Insight — Multi-tenant Monitoring SaaS

V-Insight is a Docker-based multi-tenant monitoring platform with a focus on reliability and observability. It includes an API backend, background worker jobs, and a SvelteKit frontend with a CORS-free API proxy.

- Backend: Go, Gin (port 8080)
- Worker: Go (port 8081)
- Frontend: SvelteKit (port 3000)
- Database: PostgreSQL 15

---

## Quick overview

V-Insight performs automated health checks (HTTP/HTTPS) and SSL expiry monitoring, evaluates alert rules, creates incidents, and notifies configured channels (webhook, Discord, email-ready).

It supports multi-tenant isolation via `tenant_id` on all main tables and enforces tenant-scoped queries consistently in handlers and repositories.

---

## Key features

- HTTP/HTTPS health checks, uptime and response time tracking
- SSL certificate expiry checks
- Flexible alert rules (down, slow_response, ssl_expiry)
- Incident lifecycle with automatic resolution
- Notification channels: webhook, Discord, (email-ready)
- Docker-first developer workflow and automatic DB migrations

---

## Quick start (development)

```bash
git clone https://github.com/eovipmak/v-insight.git
cd v-insight
cp .env.example .env
make up
```

Or without `make`:

```bash
docker compose up -d
```

- Backend: http://localhost:8080
- Frontend: http://localhost:3000
- Swagger API: http://localhost:8080/swagger/

Migrations run automatically on startup.

---

## For developers and contributors

- Backend tests: `cd backend && go test ./...`
- Worker tests: `cd worker && go test ./...`
- Frontend TypeScript checks: `cd frontend && npm run check`
- E2E: `cd frontend && npx playwright test`

Tips:
- Do not commit `.env` files to the repository – use environment configurations for production.
- Use `make` convenience commands (`make up`, `make logs`, `make rebuild`, `make migrate-up`, etc.)

---

## For AI agents and automations

To make it easier for LLMs, bots, or automation agents to work in this repo, see `docs/ai_agents.md`. Key notes:

- API proxy: `frontend/src/routes/api/[...path]/+server.ts` — **do NOT** add CORS middleware.
- Multi-tenant: Always include and verify `tenant_id` context when querying or modifying tenant-scoped resources.
- Migrations: Located in `backend/migrations/` — use `make migrate-create` then edit up/down SQL files; run `make migrate-up`.
- Tests: Backend/Worker use Go tests. Ensure tests pass before opening PRs.

---

## Docs

Detailed documentation: `docs/` includes guides for architecture, configuration, installation, usage, troubleshooting, contributing, and AI agent guidelines.

---

## License

This project is licensed under the MIT License — see `LICENSE` for details.
