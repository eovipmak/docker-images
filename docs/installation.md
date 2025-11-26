# Installation

This guide covers installing and running v-insight in development and production.

## Prerequisites

- Docker
- Docker Compose
- Make (optional, for convenience)

## Quick Start (Development)

```bash
git clone https://github.com/eovipmak/v-insight.git
cd v-insight
cp .env.example .env
make up
```

Or without Make:

```bash
docker-compose up -d
```

The backend automatically applies database migrations on startup.

## Make Commands (Common)

- `make up` - Start all services
- `make down` - Stop all services
- `make logs` - View logs from all services
- `make logs-backend` - View backend logs
- `make logs-worker` - View worker logs
- `make logs-frontend` - View frontend logs
- `make rebuild` - Rebuild and restart all services
- `make clean` - Remove containers, volumes, images
- `make ps` - Show service status

## Database Migrations

Migration files are in `backend/migrations/` and v-insight uses `golang-migrate`.

Common commands:

```bash
make migrate-up        # run migrations
make migrate-down      # rollback
make migrate-create name=your_migration_name
make migrate-version   # show current version
```

## Production Deployment

1. Copy production env:

```bash
cp .env.production .env
# Edit .env with production values (POSTGRES password, JWT_SECRET, PUBLIC_API_URL, etc.)
```

2. Deploy using the script:

```bash
chmod +x deploy.sh
./deploy.sh
```

The script performs checks, backups, builds images, runs migrations, and restarts services.

### Manual production (alternative)

```bash
docker build -t your-registry/v-insight-backend:latest -f docker/Dockerfile.backend .
docker build -t your-registry/v-insight-worker:latest -f docker/Dockerfile.worker ./worker
docker build -t your-registry/v-insight-frontend:latest -f docker/Dockerfile.frontend ./frontend
docker push your-registry/v-insight-backend:latest
# Pull and restart on server
docker compose -f docker-compose.prod.yml pull
docker compose -f docker-compose.prod.yml up -d
```

## Backups

Backups are SQL dumps in `./backups/` (deployment script creates them). Manual backup example:

```bash
docker compose -f docker-compose.prod.yml exec -T postgres \
  pg_dump -U $POSTGRES_USER $POSTGRES_DB > backups/manual_$(date +%Y%m%d_%H%M%S).sql
```

Restore example (outline):

1. Stop or scale down services as appropriate.
2. Start the database service only.
3. Import the SQL dump into the database with `psql`.
4. Restart all services.
