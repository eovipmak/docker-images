# Configuration

Configuration in v-insight is primarily via environment variables. Copy `.env.example` to `.env` and update values.

## Core variables

- `ENV` – environment: `development` or `production` (default: `development`)
- `PORT` – backend port (default: `8080`)
- `POSTGRES_HOST`, `POSTGRES_PORT`, `POSTGRES_USER`, `POSTGRES_PASSWORD`, `POSTGRES_DB`
- `JWT_SECRET` – change for production

## Security

- `HSTS_MAX_AGE` – HSTS header max-age (seconds). Default `31536000`.
- `HSTS_INCLUDE_SUBDOMAINS` – include subdomains in HSTS (true/false)

## Rate limiting

- `RATE_LIMIT_PER_IP` – requests per minute per IP (default: `100`)
- `RATE_LIMIT_PER_USER` – requests per hour per authenticated user (default: `1000`)

## Production tips

- Never commit `.env` to source control. Use `.env.production` as a template.
- Generate secure secrets: `openssl rand -base64 32` for `JWT_SECRET` and DB passwords.

## Resource limits (production)

Adjust `docker-compose.prod.yml` resource constraints as needed:

- PostgreSQL: recommended 2 CPU / 2GB RAM
- Backend/Worker: recommended 1 CPU / 1GB RAM each
- Frontend: recommended 0.5 CPU / 512MB
