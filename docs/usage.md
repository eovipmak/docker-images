# Usage

This document explains how to use v-insight: API usage, common workflows, and examples.

## API Quick Start

1. Register or login to get a JWT:

```bash
POST /api/v1/auth/register
POST /api/v1/auth/login
```

2. Authorize in Swagger UI (or set `Authorization: Bearer <token>` header).

3. Use API endpoints to manage monitors, alert rules, channels, and incidents.

## Common Endpoints

- `POST /api/v1/monitors` – create a monitor
- `GET /api/v1/monitors` – list monitors
- `POST /api/v1/alert-rules` – create alert rules
- `POST /api/v1/alert-channels` – create notification channels

See `backend/README.md` for full API reference and the interactive Swagger UI at `http://localhost:8080/swagger/`.

## Example: Create an Alert Rule (curl)

```bash
curl -X POST http://localhost:8080/api/v1/alert-rules \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "monitor_id": "monitor-uuid",
    "name": "Website Down Alert",
    "trigger_type": "down",
    "threshold_value": 3,
    "enabled": true,
    "channel_ids": ["channel-uuid"]
  }'
```

## Notification Channel Examples

- Webhook:

```json
{
  "type": "webhook",
  "name": "My Webhook",
  "config": {"url": "https://your-service.com/webhook"},
  "enabled": true
}
```

- Discord (webhook): similar JSON payload containing `url`.

## How Alerts Work (summary)

1. Monitor check runs on schedule (HTTP/SSL checks).
2. Alert Evaluator (every minute) compares checks vs rules.
3. New incident is created if a rule triggers.
4. Notification job sends alerts to configured channels (every 30s).
5. Incident resolved automatically when conditions clear.

## Testing and E2E

- Backend: `cd backend && go test ./...`
- Worker: `cd worker && go test ./...`
- Frontend: `cd frontend && npm run check`
- E2E (Playwright): `cd frontend && npx playwright test`
