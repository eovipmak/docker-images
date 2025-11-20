# V-Insight - Multi-tenant Monitoring SaaS

A Docker-based multi-tenant monitoring SaaS platform built with Go and SvelteKit. Monitor your websites, APIs, and services with automated health checks, SSL monitoring, and intelligent alerting.

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

## Features

### Monitoring
- **HTTP/HTTPS Health Checks**: Monitor website uptime and response times
- **SSL Certificate Monitoring**: Track SSL expiration dates
- **Custom Check Intervals**: Configure monitoring frequency per monitor
- **Multi-tenant Support**: Complete isolation between tenant data

### Alerting System
- **Alert Rules**: Create rules for monitor downtime, slow response, and SSL expiry
- **Multiple Trigger Types**:
  - `down`: Consecutive failed health checks
  - `slow_response`: Response time exceeds threshold (ms)
  - `ssl_expiry`: Certificate expires within threshold days
- **Incident Management**: Automatic incident creation and resolution tracking

### Notification Channels
- **Webhook**: Generic webhook integration for custom services
- **Discord**: Rich embed notifications with color-coded status
- **Email**: (Ready for implementation)
- **Multi-channel Support**: Send alerts to multiple channels simultaneously

### Worker Jobs
- **Health Check Job**: Runs every 30 seconds
- **SSL Check Job**: Runs every 5 minutes
- **Alert Evaluator Job**: Runs every minute to evaluate alert conditions
- **Notification Job**: Runs every 30 seconds to send notifications

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

## API Endpoints

### Monitors
- `POST /api/v1/monitors` - Create a new monitor
- `GET /api/v1/monitors` - List all monitors
- `GET /api/v1/monitors/:id` - Get monitor details
- `PUT /api/v1/monitors/:id` - Update monitor
- `DELETE /api/v1/monitors/:id` - Delete monitor

### Alert Rules
- `POST /api/v1/alert-rules` - Create alert rule
- `GET /api/v1/alert-rules` - List all alert rules
- `GET /api/v1/alert-rules/:id` - Get alert rule with channels
- `PUT /api/v1/alert-rules/:id` - Update alert rule
- `DELETE /api/v1/alert-rules/:id` - Delete alert rule

### Alert Channels
- `POST /api/v1/alert-channels` - Create notification channel
- `GET /api/v1/alert-channels` - List all channels
- `GET /api/v1/alert-channels/:id` - Get channel details
- `PUT /api/v1/alert-channels/:id` - Update channel
- `DELETE /api/v1/alert-channels/:id` - Delete channel

### Incidents
- Incidents are automatically created by the Alert Evaluator job
- Notifications are sent automatically via the Notification job

## Development

### Hot Reload

All services support hot-reload in development mode:

- **Backend & Worker**: Using Air for automatic Go code reloading
- **Frontend**: Using Vite's built-in hot module replacement (HMR)

### Testing

**Backend Tests:**
```bash
cd backend
go test ./...
```

**Frontend Type Checking:**
```bash
cd frontend
npm run check
```

**E2E Tests:**
```bash
cd frontend
npm run test:e2e        # Run tests
npm run test:e2e:ui     # Run with UI
```

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

## Alert System Configuration

### Creating an Alert Rule

Example: Alert when website is down for 3 consecutive checks
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

### Trigger Types

1. **down** - Monitor fails consecutive health checks
   - `threshold_value`: Number of consecutive failures (e.g., 3)

2. **slow_response** - Response time exceeds threshold
   - `threshold_value`: Time in milliseconds (e.g., 5000 = 5 seconds)

3. **ssl_expiry** - SSL certificate expiring soon
   - `threshold_value`: Days before expiry (e.g., 7)

### Notification Channels

**Webhook Example:**
```json
{
  "type": "webhook",
  "name": "My Webhook",
  "config": {
    "url": "https://your-service.com/webhook"
  },
  "enabled": true
}
```

**Discord Example:**
```json
{
  "type": "discord",
  "name": "Discord Alerts",
  "config": {
    "url": "https://discord.com/api/webhooks/YOUR_WEBHOOK_ID/YOUR_TOKEN"
  },
  "enabled": true
}
```

### How Alerts Work

1. **Monitor Check**: Health check or SSL check runs on schedule
2. **Evaluation**: Alert Evaluator job checks results against alert rules (every minute)
3. **Incident Creation**: If rule triggers and no open incident exists, creates new incident
4. **Notification**: Notification job sends alerts to configured channels (every 30 seconds)
5. **Resolution**: When monitor recovers, incident is automatically resolved and notification sent

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

## Database Schema

### Core Tables

- **tenants**: Multi-tenant organization data
- **users**: User accounts with tenant association
- **monitors**: Website/service monitoring configuration
- **monitor_checks**: Health check and SSL check results
- **alert_rules**: Alert trigger conditions
- **alert_channels**: Notification delivery channels (webhook, discord, email)
- **alert_rule_channels**: Many-to-many relationship between rules and channels
- **incidents**: Triggered alert instances with resolution tracking

### Multi-Tenant Design

- Every table includes `tenant_id` for data isolation
- All queries automatically filter by tenant context
- Foreign keys enforce referential integrity
- Indexes on `tenant_id` for query performance

## Troubleshooting

### Services Not Starting
```bash
# Ensure .env exists
cp .env.example .env

# Check service status
docker compose ps

# View logs
make logs
```

### Permission Issues (Frontend)
```bash
sudo chown -R $USER:$USER frontend/node_modules frontend/.svelte-kit
```

### Hot Reload Not Working
- Check Air logs: `docker compose logs backend` or `docker compose logs worker`
- Look for syntax errors in `build-errors.log`
- Restart service: `docker compose restart backend`

### Database Connection Issues
- Wait ~10 seconds for PostgreSQL initialization
- Verify: `docker compose ps postgres`
- Check logs: `make logs-postgres`

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Run tests: `cd backend && go test ./...`
5. Validate frontend: `cd frontend && npm run check`
6. Submit a pull request

## License

This project is licensed under the MIT License - see the LICENSE file for details.
