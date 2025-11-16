# V-Insight Development Guide

## Project Overview

V-Insight is a Docker-based multi-tenant monitoring SaaS platform built with Go (backend & worker) and SvelteKit (frontend). It uses a CORS-free proxy architecture where the SvelteKit frontend proxies all API requests to the backend service.

**Repository Size:** Small (~50 files)
**Languages:** Go 1.23+, TypeScript/JavaScript, Svelte
**Frameworks:** Gin (backend), Fiber (worker), SvelteKit (frontend)
**Database:** PostgreSQL 15
**Container Orchestration:** Docker Compose

## Architecture

The application consists of four services:
1. **Backend API** (Go/Gin) - REST API on port 8080
2. **Worker** (Go/Fiber) - Background job processing on port 8081
3. **Frontend** (SvelteKit) - Web interface on port 3000 with built-in API proxy
4. **PostgreSQL 15** - Database on port 5432

**CRITICAL ARCHITECTURAL NOTE:** This application uses a proxy architecture that eliminates CORS. The frontend serves the UI and proxies all `/api/*` requests to the backend. NO CORS headers or middleware should ever be added to the backend.

## Quick Start - ALWAYS Follow This Order

### Environment Setup

**ALWAYS run these steps before starting development:**

```bash
# 1. Copy environment file (REQUIRED)
cp .env.example .env

# 2. Start services with Docker Compose (PRIMARY METHOD)
make up
# OR without Make:
docker compose up -d

# Services will be available after ~30 seconds startup time:
# - Backend: http://localhost:8080
# - Worker: http://localhost:8081
# - Frontend: http://localhost:3000
# - PostgreSQL: localhost:5432
```

**Important:** The `.env` file MUST exist before starting services. Default values in `.env.example` work for local development.

### Building and Testing

#### Docker-Based Development (PRIMARY METHOD - ALWAYS USE THIS)

Docker Compose is the primary development method. All services support hot-reload:

```bash
# Build all images (run when Dockerfiles change)
docker compose build

# Start services (ALWAYS run this after build)
docker compose up -d

# View logs
make logs                # All services
make logs-backend        # Backend only
make logs-worker         # Worker only
make logs-frontend       # Frontend only
make logs-postgres       # Database only

# Rebuild everything from scratch
make rebuild

# Stop services
make down

# Complete cleanup (removes volumes and images)
make clean
```

**Hot Reload Details:**
- Backend & Worker: Use Air for automatic Go code reloading (see `.air.toml` files)
- Frontend: Vite HMR handles hot module replacement
- Code changes appear automatically; NO restart needed

#### Local Development (FALLBACK ONLY)

Use local builds ONLY when Docker is unavailable or for quick testing:

**Backend (Go):**
```bash
cd backend
go test ./...           # Run tests (ALWAYS run before committing)
go build -o bin/api ./cmd/api
```

**Worker (Go):**
```bash
cd worker
go build -o bin/worker ./cmd/worker
```

**Frontend (SvelteKit):**
```bash
cd frontend
npm install             # MUST run first if node_modules is empty/missing
npm run check           # Type checking (run before committing)
npm run build           # Production build
npm run dev             # Development server
```

**CRITICAL PERMISSION ISSUE:** If you encounter `EACCES` errors with `npm install` or `npm run check`, the `node_modules/` or `.svelte-kit/` directories may have been created by Docker with root ownership. Fix with:
```bash
sudo chown -R $USER:$USER frontend/node_modules frontend/.svelte-kit
```

## Testing

### Backend Tests
```bash
cd backend
go test ./...           # Run all tests
go test ./... -v        # Verbose output
go test -run TestName   # Run specific test
```

**Test Coverage:**
- `backend/internal/config/config_test.go` - Configuration loading tests
- `backend/cmd/api/main_test.go` - API tests (placeholder)

**Expected Results:** All tests should pass. Test run takes ~5-10 seconds including dependency downloads.

### Frontend Tests

No test suite is currently configured for the frontend. The `npm run check` command validates TypeScript types and Svelte syntax.

## Project Structure

```
/
├── .github/
│   └── workflows/
│       ├── ssl-checker.yml      # Manual workflow for SSL checker Docker build
│       └── ssl-monitor.yml      # Manual workflow for SSL monitor Docker build
├── backend/
│   ├── cmd/api/                 # Main application entry point
│   │   └── main.go              # Server setup and routes
│   ├── internal/config/         # Configuration management
│   │   ├── config.go            # Config loader with environment variables
│   │   └── config_test.go       # Config tests
│   ├── .air.toml                # Hot-reload config for Air
│   ├── go.mod                   # Go module definition (Go 1.23.0)
│   └── go.sum                   # Dependency checksums
├── worker/
│   ├── cmd/worker/              # Worker entry point
│   │   └── main.go              # Background job processor
│   ├── .air.toml                # Hot-reload config for Air
│   ├── go.mod                   # Go module definition (Go 1.21)
│   └── go.sum                   # Dependency checksums
├── frontend/
│   ├── src/
│   │   ├── routes/
│   │   │   ├── api/[...path]/+server.ts  # API proxy (handles all /api/* requests)
│   │   │   ├── +layout.svelte            # Root layout
│   │   │   ├── +page.svelte              # Home page
│   │   │   ├── login/+page.svelte        # Login page
│   │   │   ├── dashboard/                # Dashboard section
│   │   │   ├── domains/+page.svelte      # Domains page
│   │   │   ├── alerts/+page.svelte       # Alerts page
│   │   │   └── settings/+page.svelte     # Settings page
│   │   ├── lib/
│   │   │   ├── api/client.ts             # API client utilities
│   │   │   └── components/Nav.svelte     # Navigation component
│   │   ├── app.css                       # Global styles (Tailwind)
│   │   └── app.html                      # HTML template
│   ├── package.json                      # Dependencies (Node 20)
│   ├── svelte.config.js                  # SvelteKit config (uses adapter-node)
│   ├── vite.config.js                    # Vite config (server on port 3000)
│   ├── tsconfig.json                     # TypeScript config
│   ├── tailwind.config.js                # Tailwind CSS config
│   └── postcss.config.js                 # PostCSS config
├── docker/
│   ├── Dockerfile.backend                # Multi-stage: builder, development, production
│   ├── Dockerfile.worker                 # Multi-stage: builder, development, production
│   └── Dockerfile.frontend               # Single-stage Node.js development image
├── docker-compose.yml                    # Service orchestration
├── Makefile                              # Convenience commands (up, down, logs, etc.)
├── .env.example                          # Environment template
├── .gitignore                            # Git ignore patterns
└── README.md                             # Project documentation
```

## Configuration Files

### Backend Configuration
- **go.mod**: Uses Go 1.23.0, depends on gin-gonic/gin, lib/pq, joho/godotenv
- **.air.toml**: Hot-reload watches `cmd/` and `internal/`, excludes `tmp/` and `_test.go`

### Worker Configuration
- **go.mod**: Uses Go 1.21, depends on gofiber/fiber/v2
- **.air.toml**: Similar to backend, watches all Go files except tests

### Frontend Configuration
- **package.json**: Node 20, uses SvelteKit v2, Vite v5, Tailwind CSS v3
- **svelte.config.js**: Uses adapter-node for production builds
- **vite.config.js**: Server on port 3000, polling enabled for Docker volume watching
- **tsconfig.json**: Strict TypeScript with bundler module resolution

### Docker Configuration
- **docker-compose.yml**: Defines all four services with health checks, volumes, and networking
- **Dockerfiles**: Multi-stage builds for backend/worker (development uses Air), single-stage for frontend

## GitHub Workflows

**NOTE:** Both workflows are manual-only (`workflow_dispatch`) and are NOT triggered automatically on push/PR.

1. **ssl-checker.yml** - Builds and pushes SSL checker Docker image to GHCR
2. **ssl-monitor.yml** - Builds and pushes SSL monitor Docker image to GHCR

These workflows are for additional services not part of the main v-insight application.

## Common Issues and Solutions

### Issue: Services fail to start
**Solution:** Ensure `.env` file exists: `cp .env.example .env`

### Issue: Permission denied in frontend
**Problem:** Docker creates `node_modules/` and `.svelte-kit/` with root ownership
**Solution:**
```bash
sudo chown -R $USER:$USER frontend/node_modules frontend/.svelte-kit
```

### Issue: Backend/Worker not hot-reloading
**Check:** Air logs in `docker compose logs backend` or `docker compose logs worker`
**Common Cause:** Syntax errors in Go code prevent rebuild. Check `build-errors.log`

### Issue: Frontend changes not appearing
**Check:** Vite uses polling for Docker volumes. Verify `usePolling: true` in `vite.config.js`
**Fallback:** Restart frontend container: `docker compose restart frontend`

### Issue: Database connection errors
**Solution:** PostgreSQL takes ~10 seconds to initialize. Check health: `docker compose ps`

### Issue: Port already in use
**Solution:** Stop existing services: `docker compose down` or change ports in `.env`

## Environment Variables

All services use environment variables from `.env` file:

```bash
# Database (used by backend)
POSTGRES_USER=postgres
POSTGRES_PASSWORD=postgres
POSTGRES_DB=v_insight

# Service ports
BACKEND_PORT=8080
WORKER_PORT=8081
FRONTEND_PORT=3000

# Environment mode
ENV=development  # or production
```

**Backend environment mapping:**
- `POSTGRES_HOST` - Database host (defaults to "localhost", Docker uses "postgres")
- `POSTGRES_PORT` - Database port (defaults to "5432")
- `PORT` - Server port (defaults to "8080")
- `ENV` - Sets Gin mode (defaults to "development")

## Making Code Changes

### Backend Changes
1. Edit Go files in `backend/cmd/` or `backend/internal/`
2. Air automatically rebuilds and restarts (check logs)
3. Test with: `curl http://localhost:8080/health`
4. Run tests: `cd backend && go test ./...`

### Worker Changes
1. Edit Go files in `worker/cmd/`
2. Air automatically rebuilds and restarts
3. Test with: `curl http://localhost:8081/health`

### Frontend Changes
1. Edit files in `frontend/src/`
2. Vite HMR updates browser automatically
3. Type check: `cd frontend && npm run check`
4. Access at: http://localhost:3000

## Critical Implementation Notes

1. **NEVER add CORS middleware to backend** - The proxy architecture eliminates CORS completely
2. **API proxy is at** `frontend/src/routes/api/[...path]/+server.ts` - All API routes go through here
3. **Hot reload depends on Air** - Check `.air.toml` if builds fail
4. **Docker is the primary development environment** - Local builds are secondary
5. **PostgreSQL startup takes ~10 seconds** - Wait for health check before connecting
6. **Environment file is REQUIRED** - Copy `.env.example` before starting

## Validation Before Committing

Always run these checks before committing:

```bash
# 1. Test with Docker (PRIMARY)
make rebuild              # Clean build
docker compose ps         # Verify all healthy
curl http://localhost:8080/health   # Test backend
curl http://localhost:8081/health   # Test worker

# 2. Run backend tests
cd backend && go test ./...

# 3. Check frontend types
cd frontend && npm run check

# 4. Verify no build artifacts committed
git status                # Check for bin/, tmp/, node_modules/, etc.
```

## Additional Resources

- **Backend API Framework:** [Gin Documentation](https://gin-gonic.com/docs/)
- **Worker Framework:** [Fiber Documentation](https://docs.gofiber.io/)
- **Frontend Framework:** [SvelteKit Documentation](https://kit.svelte.dev/docs)
- **Hot Reload Tool:** [Air Documentation](https://github.com/cosmtrek/air)
