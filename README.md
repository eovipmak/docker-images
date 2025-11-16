# V-Insight - Multi-tenant Monitoring SaaS

A Docker-based multi-tenant monitoring SaaS platform built with Go and SvelteKit.

## Architecture

- **Backend API** (Go, Fiber): REST API service on port 8080
- **Worker** (Go): Background job processing service on port 8081
- **Frontend** (SvelteKit): Web interface on port 3000
- **PostgreSQL 15**: Database on port 5432

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

## Development

### Hot Reload

All services support hot-reload in development mode:

- **Backend & Worker**: Using Air for automatic Go code reloading
- **Frontend**: Using Vite's built-in hot module replacement (HMR)

### CORS Configuration in Development

By default, the development environment allows all origins (`*`) for CORS, making it easy to develop and debug from any frontend location.

**Environment Variables:**
- `ENV=development` - Sets the environment mode
- `CORS_ALLOWED_ORIGINS=*` - Allows all origins (default in development)
- `CORS_ALLOWED_ORIGINS=` - Empty value also allows all origins in development

**Alternative Configurations:**
- Specific origins: `CORS_ALLOWED_ORIGINS=http://localhost:3000,http://127.0.0.1:3000`
- Custom domain: `CORS_ALLOWED_ORIGINS=http://192.168.1.100:3000`

**Important:** The wildcard (`*`) CORS setting is automatically disabled in production mode for security.

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

When deploying to a VPS with a public IP address, you need to configure the environment variables properly:

1. **Copy and edit the environment file**:
```bash
cp .env.example .env
```

2. **Update the following variables in `.env`**:

```bash
# Replace 1.2.3.4 with your actual VPS IP address

# Backend API - accessible from browser
PUBLIC_API_URL=http://1.2.3.4:8080

# Environment - set to production for deployment
ENV=production

# CORS - must include the frontend URL that users will access
# Include both IP and localhost for flexibility
CORS_ALLOWED_ORIGINS=http://1.2.3.4:3000,http://localhost:3000,http://127.0.0.1:3000
```

3. **Start the services**:
```bash
make up
```

4. **Access the application**:
- Frontend: `http://YOUR_VPS_IP:3000`
- Backend API: `http://YOUR_VPS_IP:8080`

**Important Notes**:
- The `PUBLIC_API_URL` is used by the frontend (browser) to make API requests
- The `CORS_ALLOWED_ORIGINS` must match the URL where users access the frontend
- For production with a domain name, update both variables to use `https://` and your domain
- The wildcard (`*`) for CORS is only allowed in development mode (`ENV=development`)
- In production mode, wildcard is automatically replaced with safe default origins

## License

MIT
