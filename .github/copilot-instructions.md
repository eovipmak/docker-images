# Copilot Instructions for v-insight Repository

## Repository Overview

**v-insight** is a dual-service SSL/TLS certificate monitoring and checking application. It provides both real-time SSL certificate checking and historical monitoring capabilities with alerting features.

**Repository Type:** Microservices architecture with Docker containerization  
**Primary Languages:** Python 3.12 (backend), TypeScript/React (frontend), Vanilla JavaScript (simple UI)  
**Size:** ~4,500+ lines of Python code, React TypeScript frontend  
**Main Components:**
- `ssl-checker`: Standalone SSL certificate checker with REST API and web UI
- `ssl-monitor`: Full-featured SSL monitoring platform with React frontend, authentication, and database

---

## Project Structure

### Root Level
```
.
├── .github/
│   └── workflows/           # CI/CD workflows for Docker builds
├── docker-compose.yml       # Multi-service orchestration
├── ssl-checker/            # SSL certificate checker service (port 8000)
└── ssl-monitor/            # SSL monitoring platform (port 8001)
```

### ssl-checker Service
```
ssl-checker/
├── api/                    # FastAPI backend
│   ├── main.py            # FastAPI app with API endpoints
│   ├── ssl_checker.py     # SSL checking and validation logic
│   ├── cert_utils.py      # Certificate parsing utilities
│   ├── network_utils.py   # DNS, HTTP, socket operations
│   ├── constants.py       # App constants and configuration
│   └── requirements.txt   # Python dependencies
├── ui/                     # Vanilla JavaScript frontend
│   ├── index.html         # Main HTML page
│   ├── styles.css         # Styles
│   └── app.js             # JavaScript logic
├── Dockerfile             # Python 3.12-slim based image
└── README.md
```

### ssl-monitor Service
```
ssl-monitor/
├── api/                    # FastAPI backend with database
│   ├── main.py            # Main FastAPI application (40,583 lines)
│   ├── database.py        # SQLAlchemy database models and sessions
│   ├── schemas.py         # Pydantic schemas for API
│   ├── auth.py            # JWT authentication logic
│   ├── alert_service.py   # Alert notification service
│   ├── alembic/           # Database migrations
│   │   ├── alembic.ini    # Alembic configuration
│   │   └── versions/      # Migration files
│   ├── test_*.py          # Python test files
│   └── requirements.txt   # Python dependencies
├── frontend/              # React TypeScript application
│   ├── src/
│   │   ├── pages/         # Dashboard, Login, AddDomain, AlertSettings
│   │   ├── components/    # React components
│   │   ├── services/      # API client, authService
│   │   ├── tests/         # Vitest unit tests
│   │   └── types/         # TypeScript type definitions
│   ├── e2e/               # Playwright E2E tests
│   ├── package.json       # Node dependencies
│   ├── vite.config.ts     # Vite build configuration
│   ├── playwright.config.ts # E2E test configuration
│   ├── eslint.config.js   # ESLint configuration
│   └── TESTING.md         # Comprehensive testing guide
├── Dockerfile             # Multi-stage: Node 20 + Python 3.12
└── entrypoint.sh          # Runs Alembic migrations then starts app
```

---

## Build Instructions

### Prerequisites
- **Docker:** Version 20.x+ with Docker Compose v2.38.2+
- **Python:** 3.12.x (for local development)
- **Node.js:** 20.x (for frontend development)
- **npm:** 10.8.2+

### Docker Build (Recommended for Production/Testing)

**ALWAYS use Docker builds to verify changes in a clean environment.**

#### Build Single Service
```bash
# SSL Checker (takes ~60 seconds)
docker build -t ssl-checker -f ssl-checker/Dockerfile ssl-checker

# SSL Monitor (takes ~120 seconds due to frontend build)
docker build -t ssl-monitor -f ssl-monitor/Dockerfile ssl-monitor
```

#### Build with Docker Compose
```bash
# Build both services
docker compose build

# Build specific service
docker compose build ssl-checker
docker compose build ssl-monitor

# Run both services
docker compose up -d

# View logs
docker compose logs -f ssl-checker
docker compose logs -f ssl-monitor
```

**Service URLs after `docker compose up`:**
- SSL Checker: http://localhost:8000
- SSL Checker API docs: http://localhost:8000/docs
- SSL Monitor: http://localhost:8001
- SSL Monitor API docs: http://localhost:8001/docs

### ssl-checker Local Development

```bash
cd ssl-checker/api

# Install dependencies (first time)
pip install -r requirements.txt

# Run development server
uvicorn main:app --reload --host 0.0.0.0 --port 8000
```

**Dependencies:**
- fastapi, uvicorn[standard], requests, dnspython, aiofiles

**No tests** exist for ssl-checker Python code. Only manual testing via UI/API.

### ssl-monitor Backend Local Development

```bash
cd ssl-monitor/api

# Install dependencies (first time)
pip install -r requirements.txt

# Run database migrations (REQUIRED before first run)
alembic upgrade head

# Run development server
uvicorn main:app --reload --host 0.0.0.0 --port 8001
```

**Environment Variables Required:**
```bash
export SSL_CHECKER_URL=http://localhost:8000
export DATABASE_URL=sqlite:///./ssl_monitor.db
export ASYNC_DATABASE_URL=sqlite+aiosqlite:///./ssl_monitor.db
```

**Database Migrations:**
- **ALWAYS run** `alembic upgrade head` after pulling changes that may include new migrations
- New migrations in `api/alembic/versions/` indicate database schema changes
- The `entrypoint.sh` automatically runs migrations in Docker

**Python Testing:**
Test files exist (`test_*.py`) but there's **no pytest in requirements.txt**. To run tests:
```bash
# Install pytest first
pip install pytest httpx

# Run specific test file
python -m pytest test_405_error.py -v
```

### ssl-monitor Frontend Development

```bash
cd ssl-monitor/frontend

# Install dependencies (first time, takes ~7 seconds)
npm ci

# Development server with hot reload (starts on port 3000)
npm run dev

# Build for production (takes ~8 seconds)
npm run build

# Output: dist/ directory
```

**Linting:**
```bash
# ALWAYS run before committing frontend changes
npm run lint

# Expected output: May show 10 errors, 1 warning (existing issues)
# Focus on not introducing NEW lint errors
```

**Unit Tests:**
```bash
# Run all unit tests (Vitest, takes ~60 seconds)
npm test

# Run in watch mode (development)
npm test -- --watch

# Run tests with UI
npm run test:ui

# Expected: 22 tests across 4 test files should pass
```

**End-to-End Tests:**
```bash
# Install Playwright browsers (first time only)
npx playwright install chromium

# Requires app running on http://localhost:3000
# Create .env.test.local with test credentials:
# TEST_USER_EMAIL=test@example.com
# TEST_USER_PASSWORD=testpass123

# Run E2E tests
npm run test:e2e

# Run in UI mode (interactive debugging)
npm run test:e2e:ui

# Run specific test
npm run test:e2e -- domain-deletion.spec.ts
```

**IMPORTANT:** E2E tests auto-start dev server via `webServer` config in `playwright.config.ts`.

---

## Validation & CI/CD

### GitHub Workflows

Located in `.github/workflows/`:

1. **ssl-checker.yml** - Builds and pushes Docker image to GHCR
   - Trigger: Manual (`workflow_dispatch`)
   - Builds from `ssl-checker/Dockerfile`
   - Tags: `build_at-{timestamp}` and `latest`

2. **ssl-monitor.yml** - Builds and pushes Docker image to GHCR
   - Trigger: Manual (`workflow_dispatch`)
   - Multi-stage build: Node 20 → Python 3.12
   - Tags: `build_at-{timestamp}` and `latest`

**No automated CI tests run on push/PR.** Validation is manual via Docker builds.

### Pre-Checkin Validation Steps

**ALWAYS perform these steps before finalizing changes:**

1. **For ssl-checker changes:**
   ```bash
   docker build -t ssl-checker-test -f ssl-checker/Dockerfile ssl-checker
   ```

2. **For ssl-monitor backend changes:**
   ```bash
   cd ssl-monitor/api
   pip install -r requirements.txt
   # If migrations exist: alembic upgrade head
   uvicorn main:app --reload  # Verify no startup errors
   ```

3. **For ssl-monitor frontend changes:**
   ```bash
   cd ssl-monitor/frontend
   npm ci
   npm run lint        # Check for NEW errors only
   npm run build       # Must succeed
   npm test -- --run   # All 22 tests must pass
   ```

4. **Full integration test:**
   ```bash
   docker compose build
   docker compose up -d
   # Verify both services start successfully
   docker compose logs ssl-checker
   docker compose logs ssl-monitor
   docker compose down
   ```

---

## Common Patterns & Conventions

### Code Style
- **Python:** FastAPI async/await patterns, Pydantic models for validation
- **TypeScript/React:** Functional components with hooks, Material-UI components
- **JavaScript:** ES6+ for ssl-checker UI

### API Patterns
- **ssl-checker:** REST endpoints at `/api/*` (e.g., `/api/check`, `/api/batch_check`)
- **ssl-monitor:** REST + WebSocket endpoints, authentication via JWT in localStorage

### Database
- **ORM:** SQLAlchemy 2.x with async support
- **Migrations:** Alembic (run via `entrypoint.sh` in Docker)
- **Default:** SQLite (`ssl_monitor.db`) for development

### Testing
- **Frontend Unit:** Vitest + React Testing Library
- **Frontend E2E:** Playwright (Chromium only)
- **Backend:** Manual testing or pytest (not configured by default)

### Known Issues & Workarounds

1. **Frontend Linting:** There are existing lint errors (10 errors, 1 warning). Don't introduce new ones.

2. **Database Migrations:** If you get "table already exists" errors:
   ```bash
   # Delete database and recreate
   rm ssl_monitor.db
   alembic upgrade head
   ```

3. **Frontend Proxy:** Vite dev server proxies `/api` to `http://localhost:8000`. Ensure ssl-checker backend runs on 8000.

4. **Docker Build Times:**
   - ssl-checker: ~60 seconds
   - ssl-monitor: ~120 seconds (includes npm build)
   - Use `docker compose build` for parallel builds

5. **Port Conflicts:** Default ports are 8000 (ssl-checker) and 8001 (ssl-monitor). If unavailable, modify `docker-compose.yml` ports mapping.

---

## Quick Reference

### File Locations
- **Workflow Configs:** `.github/workflows/ssl-checker.yml`, `ssl-monitor.yml`
- **Docker Config:** `docker-compose.yml`, `*/Dockerfile`
- **Python Deps:** `ssl-checker/api/requirements.txt`, `ssl-monitor/api/requirements.txt`
- **Frontend Deps:** `ssl-monitor/frontend/package.json`
- **DB Migrations:** `ssl-monitor/api/alembic/versions/`
- **Lint Config:** `ssl-monitor/frontend/eslint.config.js`
- **Test Config:** `ssl-monitor/frontend/vite.config.ts` (unit), `playwright.config.ts` (E2E)
- **Test Docs:** `ssl-monitor/frontend/TESTING.md`

### Key Commands Summary
```bash
# Build everything
docker compose build

# Run everything
docker compose up -d

# Frontend dev
cd ssl-monitor/frontend && npm ci && npm run dev

# Backend dev
cd ssl-monitor/api && pip install -r requirements.txt && alembic upgrade head && uvicorn main:app --reload

# Test frontend
cd ssl-monitor/frontend && npm test

# E2E tests
cd ssl-monitor/frontend && npm run test:e2e

# Build checks
npm run build  # Frontend
docker build -t test -f ssl-checker/Dockerfile ssl-checker  # Backend
```

### Important Notes
- **TRUST THESE INSTRUCTIONS:** They have been validated by building, testing, and running all services
- Only search/explore if these instructions are incomplete or incorrect
- When in doubt, use Docker builds—they are the source of truth
- Database migrations run automatically in Docker via `entrypoint.sh`
- Frontend E2E tests auto-start the dev server—no manual startup needed
