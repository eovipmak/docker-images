# V-Insight Development Guide

## Project Overview

V-Insight is a Docker-based multi-tenant monitoring SaaS platform. **Repository:** ~50 files. **Stack:** Go 1.23 (Gin backend on :8080, Fiber worker on :8081), SvelteKit/TypeScript frontend (:3000), PostgreSQL 15 (:5432).

**CRITICAL:** Uses CORS-free proxy architecture. Frontend (`src/routes/api/[...path]/+server.ts`) proxies all `/api/*` requests to backend. NEVER add CORS middleware.

## Quick Start (ALWAYS Follow This Order)

**REQUIRED:** Copy `.env.example` to `.env` before starting: `cp .env.example .env`

**Start services (PRIMARY METHOD):**
```bash
make up          # Start all services
# OR: docker compose up -d
```

**Wait ~30s for startup.** Services health-checked via `docker compose ps`.

**Hot reload enabled:** Backend/worker use Air (`.air.toml`), frontend uses Vite HMR. Code changes auto-reload.

## Commands Reference

**Docker (primary):** `make up|down|logs|logs-backend|logs-worker|logs-frontend|rebuild|clean`
**Backend tests:** `cd backend && go test ./...` (takes ~10s, must pass before committing)
**Frontend check:** `cd frontend && npm run check` (validates TypeScript/Svelte)
**Permission fix:** `sudo chown -R $USER:$USER frontend/node_modules frontend/.svelte-kit` (if EACCES errors)

## Testing

**Backend:** Tests in `backend/internal/config/config_test.go` and `backend/cmd/api/main_test.go`. Run: `cd backend && go test ./...`
**Frontend:** No test suite. Use `npm run check` for type validation.
**Expected:** All backend tests pass (~5-10s with dependency downloads).

## Project Structure

```
/
├── .github/workflows/           # Manual workflows (ssl-checker.yml, ssl-monitor.yml)
├── backend/
│   ├── cmd/api/main.go          # Entry point, Gin setup, routes
│   ├── internal/config/         # Config loader (config.go, config_test.go)
│   ├── .air.toml                # Hot-reload config
│   └── go.mod                   # Go 1.23.0, gin-gonic/gin, lib/pq, godotenv
├── worker/
│   ├── cmd/worker/main.go       # Background job processor
│   ├── .air.toml                # Hot-reload config  
│   └── go.mod                   # Go 1.21, gofiber/fiber/v2
├── frontend/
│   ├── src/routes/
│   │   ├── api/[...path]/+server.ts   # API proxy (critical!)
│   │   ├── +layout.svelte, +page.svelte
│   │   ├── login/, dashboard/, domains/, alerts/, settings/
│   │   └── lib/api/client.ts, lib/components/Nav.svelte
│   ├── package.json             # Node 20, SvelteKit v2, Vite v5, Tailwind v3
│   ├── svelte.config.js         # Uses adapter-node
│   └── vite.config.js           # Port 3000, usePolling: true (for Docker)
├── docker/                      # Dockerfile.{backend,worker,frontend}
├── docker-compose.yml           # All services with health checks
├── Makefile                     # Convenience commands
└── .env.example                 # Template (copy to .env)

## Key Configuration Details

**Backend (.air.toml):** Watches `cmd/`, `internal/`; excludes `tmp/`, `_test.go`; builds to `./tmp/main`
**Worker (.air.toml):** Same as backend, builds `cmd/worker` to `./tmp/main`
**Frontend (vite.config.js):** `usePolling: true` required for Docker volume watching
**Docker Compose:** PostgreSQL health check takes ~10s; backend health check has 30s start_period

## Environment Variables

Default `.env.example` values work for local dev. Service environment mappings:
- **Backend:** `POSTGRES_HOST` (default: localhost, Docker: postgres), `POSTGRES_PORT` (5432), `PORT` (8080), `ENV` (development/production)
- **Worker:** `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`, `PORT` (8081)
- **Frontend:** `NODE_ENV` (development), `BACKEND_API_URL` (http://backend:8080)

## Common Issues & Solutions

**Services fail to start:** Ensure `.env` exists: `cp .env.example .env`
**Permission denied (npm):** Run `sudo chown -R $USER:$USER frontend/node_modules frontend/.svelte-kit`
**Hot-reload not working:** Check Air logs (`docker compose logs backend/worker`), look for syntax errors in `build-errors.log`
**Frontend changes not appearing:** Vite polling enabled in `vite.config.js`; if needed: `docker compose restart frontend`
**Database connection errors:** Wait ~10s for PostgreSQL init; verify with `docker compose ps`
**Port conflicts:** Run `docker compose down` or change ports in `.env`

## Making Code Changes

**Backend:** Edit in `backend/cmd/` or `backend/internal/` → Air auto-rebuilds → Test: `curl http://localhost:8080/health`
**Worker:** Edit in `worker/cmd/` → Air auto-rebuilds → Test: `curl http://localhost:8081/health`
**Frontend:** Edit in `frontend/src/` → Vite HMR updates browser → Type check: `cd frontend && npm run check`

## Validation Before Committing

```bash
make rebuild && docker compose ps              # Verify all healthy
curl http://localhost:8080/health              # Test backend
curl http://localhost:8081/health              # Test worker
cd backend && go test ./...                    # Run backend tests (REQUIRED)
cd frontend && npm run check                   # Validate TypeScript (REQUIRED)
git status                                     # Ensure no bin/, tmp/, node_modules/ committed
```

## Critical Notes

1. **NEVER add CORS middleware** - Proxy architecture eliminates CORS
2. **API proxy location:** `frontend/src/routes/api/[...path]/+server.ts`
3. **Docker is primary dev method** - Use local builds only as fallback
4. **PostgreSQL startup:** ~10s health check delay
5. **`.env` is REQUIRED** - Copy from `.env.example` before starting
6. **GitHub workflows are manual-only** - ssl-checker.yml and ssl-monitor.yml not triggered on push/PR
