# V-Insight - Multi-tenant Monitoring SaaS

A Docker-based multi-tenant monitoring SaaS platform built with Go and SvelteKit.

## Architecture

- **Backend API** (Go, Gin): REST API service on port 8080
- **Worker** (Go): Background job processing service on port 8081
- **Frontend** (SvelteKit): Web interface on port 3000 with built-in API proxy
- **PostgreSQL 15**: Database on port 5432

### CORS-Free Architecture

This application uses a proxy architecture that completely eliminates CORS:
- The frontend (SvelteKit) runs on port 3000 and serves the web interface
- All API requests from the browser go to the same origin (localhost:3000/api/*)
- SvelteKit's server-side proxy forwards these requests to the backend (port 8080)
- No cross-origin requests occur, so CORS is never triggered
- No CORS headers or middleware are needed anywhere in the codebase

## Quick Start

### Prerequisites

- Docker
- Docker Compose
- Make (optional, for convenience)

### Setup

1. Clone the repository:
```bash
git clone https://github.com/eovipmak/v-insight.git
cd v-insight
```

2. Copy the environment file:
```bash
cp .env.example .env
```

3. Start all services:
```bash
make up
```

Or without Make:
```bash
docker-compose up -d
```

**Note:** Database migrations run automatically on startup, creating all necessary tables (users, tenants, etc.).

### Available Make Commands

- `make up` - Start all services
- `make down` - Stop all services
- `make logs` - View logs from all services
- `make logs-backend` - View backend logs
- `make logs-worker` - View worker logs
- `make logs-frontend` - View frontend logs
- `make logs-postgres` - View PostgreSQL logs
- `make rebuild` - Rebuild and restart all services
- `make clean` - Remove all containers, volumes, and images
- `make ps` - Show status of all services
- `make restart` - Restart all services
- `make help` - Show all available commands

## Services

### Backend API (http://localhost:8080)

- Health check: `GET /health`
- API v1: `GET /api/v1`

### Worker (http://localhost:8081)

- Health check: `GET /health`

### Frontend (http://localhost:3000)

- Main application interface

### PostgreSQL (localhost:5432)

- Database: `v_insight`
- User: `postgres`
- Password: `postgres`

**Migrations:** The backend service automatically runs database migrations on startup using the `migrate` tool. Migration files are located in `backend/migrations/`.

## Development

### Hot Reload

All services support hot-reload in development mode:

- **Backend & Worker**: Using Air for automatic Go code reloading
- **Frontend**: Using Vite's built-in hot module replacement (HMR)

### Database Migrations

Database migrations are automatically applied when the backend service starts. For manual migration management, use these commands:

```bash
# Check current migration version
make migrate-version

# Manually run migrations (usually not needed)
make migrate-up

# Rollback migrations
make migrate-down

# Create a new migration
make migrate-create name=your_migration_name

# Force migration to a specific version
make migrate-force version=<version_number>
```

Migration files are located in `backend/migrations/` and use the [golang-migrate](https://github.com/golang-migrate/migrate) tool.

### Project Structure

```
.
├── backend/              # Go backend API
│   ├── cmd/api/         # Main application entry point
│   ├── internal/        # Private application code
│   └── pkg/             # Public libraries
├── frontend/            # SvelteKit frontend
│   └── src/             # Source code
│       └── routes/      # SvelteKit routes
├── worker/              # Go worker service
│   ├── cmd/worker/      # Main worker entry point
│   └── internal/        # Private worker code
├── docker/              # Dockerfiles
│   ├── Dockerfile.backend
│   ├── Dockerfile.frontend
│   └── Dockerfile.worker
├── docker-compose.yml   # Docker Compose configuration
├── Makefile            # Make commands
└── .env.example        # Environment variables template
```

## Deployment

### Deploying to a VPS

When deploying to a VPS with a public IP address:

1. **Copy and edit the environment file**:
```bash
cp .env.example .env
```

2. **Update the environment in `.env` if needed**:

```bash
# Set to production for deployment
ENV=production
```

3. **Start the services**:
```bash
make up
```

The backend service will automatically run database migrations on startup.

4. **Access the application**:
- Frontend: `http://YOUR_VPS_IP:3000`
- Backend API (internal): `http://YOUR_VPS_IP:8080`

**Important Notes**:
- All API requests from the browser are automatically proxied through the frontend server
- The backend does not need to be directly accessible from the browser
- No CORS configuration is needed since all requests appear to come from the same origin
