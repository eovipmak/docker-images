# Troubleshooting

Common problems and quick fixes.

## Services not starting

1. Ensure `.env` exists:

```bash
cp .env.example .env
```

2. Check container status:

```bash
docker compose ps
```

3. View logs:

```bash
make logs
```

## Frontend permission issues

If you see permission errors for `node_modules`:

```bash
sudo chown -R $USER:$USER frontend/node_modules frontend/.svelte-kit
```

## Hot-reload not working

- Backend/Worker use Air — inspect logs: `docker compose logs backend` or `docker compose logs worker`.
- Restart the service if needed: `docker compose restart backend`.

## Database connection issues

- Wait a short time for Postgres to initialize.
- Verify Postgres container status: `docker compose ps postgres`.
- Check Postgres logs: `make logs-postgres`.

## Deployment issues

- Ensure `.env.production` is copied to `.env` and production secrets are set.
- Check backups in `./backups/` created by `deploy.sh`.

## Health check troubleshooting

- Use the health endpoints to verify services:
  - Backend: `/health`, `/health/live`, `/health/ready`
  - Worker: `/health`, `/health/live`, `/health/ready`

## Where to get help

- Check component READMEs: `backend/README.md`, `frontend/README.md`, `worker/README.md`.
- Open an issue on the project repository with logs and reproduction steps.
