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

### Security Features
- **Rate Limiting**: 
  - Per-IP: 100 requests/minute (configurable via `RATE_LIMIT_PER_IP`)
  - Per-User: 1000 requests/hour (configurable via `RATE_LIMIT_PER_USER`)
  - Applied to public endpoints (login, register)
- **Security Headers**:
  - `X-Content-Type-Options: nosniff` - Prevents MIME type sniffing
  - `X-Frame-Options: DENY` - Prevents clickjacking attacks
  - `X-XSS-Protection: 1; mode=block` - Enables XSS protection
  - `Strict-Transport-Security` - Enforces HTTPS (configurable via `HSTS_MAX_AGE`)
- **Request ID Tracking**: Unique ID per request for logging and debugging
- **Request Size Limits**: 10MB default limit on request body size
- **Input Sanitization**: HTML escaping and validation for user inputs

## API Documentation

V-Insight provides comprehensive API documentation using Swagger/OpenAPI:

### Access Documentation

- **Swagger UI** (Development): `http://localhost:8080/swagger/` - Interactive API documentation with "Try it out" functionality
- **Frontend Docs Page**: `http://localhost:3000/docs` - Embedded Swagger UI with quick start guide
- **OpenAPI Spec**: 
  - JSON: `backend/docs/swagger.json`
  - YAML: `backend/docs/swagger.yaml`

### Quick Start with API

1. **Register/Login**: Use `/api/v1/auth/register` or `/api/v1/auth/login` to get a JWT token
2. **Authorize**: Click "Authorize" button in Swagger UI and enter `Bearer <your-token>`
3. **Try Endpoints**: Use the "Try it out" button to test endpoints directly

### Available API Groups

- **Authentication**: Register, login, user info
- **Monitors**: Create and manage monitoring endpoints
- **Alert Rules**: Configure alerting conditions
- **Alert Channels**: Setup notification channels (Webhook, Discord, Email)
- **Incidents**: View and manage triggered alerts
- **Dashboard**: Get overview statistics and data
- **Metrics**: Monitor performance metrics

See [Backend API Documentation](backend/README.md#api-endpoints) for complete endpoint reference.

## Documentation

Comprehensive documentation is available for each component:

- **[Backend README](backend/README.md)**: API architecture, endpoints, development guide, testing, and deployment
- **[Frontend README](frontend/README.md)**: SvelteKit setup, component development, routing, styling, and testing
- **[Worker README](worker/README.md)**: Job scheduler, background tasks, monitoring, and performance tuning

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

**Service Management:**
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

**Testing:**
- `make test-all` - Run all tests (backend, worker, frontend)
- `make test-backend` - Run backend unit tests
- `make test-worker` - Run worker unit tests
- `make test-frontend` - Run frontend unit tests

**Database Migrations:**
- `make migrate-up` - Run database migrations
- `make migrate-down` - Rollback migrations
- `make migrate-create name=<name>` - Create new migration
- `make migrate-version` - Show current migration version

**Help:**
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

## Configuration

The application is configured using environment variables. Copy `.env.example` to `.env` and customize as needed:

### Core Configuration
- `ENV` - Environment mode (`development`, `production`) - Default: `development`
- `PORT` - Backend API port - Default: `8080`
- `POSTGRES_HOST` - PostgreSQL host - Default: `localhost`
- `POSTGRES_PORT` - PostgreSQL port - Default: `5432`
- `POSTGRES_USER` - Database user - Default: `postgres`
- `POSTGRES_PASSWORD` - Database password - Default: `postgres`
- `POSTGRES_DB` - Database name - Default: `v_insight`
- `JWT_SECRET` - Secret key for JWT tokens - **Change in production!**

### Security Configuration
- `HSTS_MAX_AGE` - HSTS header max-age in seconds - Default: `31536000` (1 year)
- `HSTS_INCLUDE_SUBDOMAINS` - Include subdomains in HSTS policy - Default: `true`

### Rate Limiting Configuration
- `RATE_LIMIT_PER_IP` - Requests per minute per IP address - Default: `100`
- `RATE_LIMIT_PER_USER` - Requests per hour per authenticated user - Default: `1000`

**Note:** Rate limiting is applied only to public endpoints (login, register) to prevent abuse while allowing normal API usage for authenticated users.

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

V-Insight has comprehensive test coverage across all layers. See [TESTING.md](TESTING.md) for detailed testing documentation.

**Quick Test Commands:**
```bash
make test-all          # Run all tests (backend, worker, frontend)
make test-backend      # Run backend tests only
make test-worker       # Run worker tests only  
make test-frontend     # Run frontend tests only
```

**Test Coverage:**
- Backend: Auth services, repositories, alert evaluation, utilities
- Worker: HTTP/SSL checkers, task executor, concurrent processing
- Frontend: Stores, API client, components

**Backend Tests (Go + Testify):**
```bash
cd backend
go test ./... -v -cover
```

**Worker Tests (Go + Testify):**
```bash
cd worker
go test ./... -v -cover
```

**Frontend Tests (Vitest + Testing Library):**
```bash
cd frontend
npm test                # Run once
npm run test:watch      # Watch mode
npm run test:coverage   # With coverage
```

**E2E Tests (Playwright):**
```bash
cd frontend
npm run test:e2e
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

### CI/CD Pipeline for Auto-Deploy

This project includes a GitHub Actions CI/CD pipeline for automatic deployment to your VPS.

#### Setup Steps

1. **Configure GitHub Secrets**:
   - Go to your repository Settings > Secrets and variables > Actions
   - Add the following secrets:
     - `VPS_HOST`: Your VPS IP address or domain
     - `VPS_USER`: SSH username for your VPS
     - `SSH_PRIVATE_KEY`: Private SSH key for authentication (generate with `ssh-keygen -t rsa -b 4096`)
     - `SSH_PORT`: SSH port (default 22, optional)

2. **On your VPS**:
   - Ensure Docker and Docker Compose are installed
   - Clone the repository: `git clone https://github.com/eovipmak/v-insight.git`
   - Copy `.env.example` to `.env` and configure production settings
   - Ensure the SSH key is added to `~/.ssh/authorized_keys` on the VPS
   - For production, use `docker-compose.prod.yml`: `docker compose -f docker-compose.prod.yml up -d`

3. **Staging Environment for PR Testing**:
   - Create a separate directory for staging: `mkdir /path/to/v-insight-staging && cd /path/to/v-insight-staging`
   - Clone the repo again: `git clone https://github.com/eovipmak/v-insight.git .`
   - Copy `.env.example` to `.env` and configure staging settings (different ports if needed, e.g., 3001 for frontend)
   - The `deploy-pr-staging.yml` workflow will automatically deploy PRs to this staging environment on ports 3001 (frontend) and 8081 (backend) when you comment `/test` on the PR
   - Access staging at: `http://YOUR_VPS_IP:3001`
   - The workflow will comment the staging URL directly on the PR
   - To cleanup staging, comment `/cleanup` on the PR

4. **Deployment Flow**:
   - On push to `main` branch, the pipeline runs tests
   - If tests pass, it builds Docker images and pushes to GitHub Container Registry
   - Then SSHs into your VPS, pulls the latest images, and restarts services
   - For PRs, staging is deployed automatically for UI testing

#### Manual Deployment Alternative

If you prefer manual deployment:

```bash
# On your VPS
git pull origin main
make up
```

Or build images locally and push:

```bash
# Build and push images
docker build -t your-registry/v-insight-backend:latest -f docker/Dockerfile.backend .
docker build -t your-registry/v-insight-worker:latest -f docker/Dockerfile.worker ./worker
docker build -t your-registry/v-insight-frontend:latest -f docker/Dockerfile.frontend ./frontend
docker push your-registry/v-insight-backend:latest
# ... push others

# On VPS
docker compose pull
docker compose up -d
```

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

For detailed troubleshooting of Docker health checks and service startup issues, see:
- **[Docker Health Check Troubleshooting Guide](DOCKER_HEALTHCHECK_GUIDE.md)** - Comprehensive guide for diagnosing and fixing service health issues
- **[E2E Test Fix Summary](E2E_TEST_FIX_SUMMARY.md)** - Details on recent fixes to E2E test reliability

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

## Testing

### Demo User

A demo user is automatically created via database migrations for testing purposes:

- **Email**: `test@gmail.com`
- **Password**: `Password!`
- **Tenant**: Demo Tenant

This user is available immediately after starting the services and can be used for:
- Manual testing
- Automated E2E tests
- Development and debugging

### Running Tests

**Backend Tests:**
```bash
cd backend
go test ./...
```

**Frontend Type Check:**
```bash
cd frontend
npm run check
```

**E2E Tests:**
```bash
# Install Playwright browsers first
npx playwright install --with-deps

# Run all E2E tests
npx playwright test

# Run specific test file
npx playwright test tests/e2e-workflow.spec.ts

# Run tests with UI
npx playwright test --ui
```

### Automated Testing (GitHub Actions)

The repository includes automated E2E testing via GitHub Actions that:
1. Builds and starts all services (backend, worker, frontend, PostgreSQL)
2. Waits for services to be healthy
3. Runs comprehensive E2E tests covering:
   - User login with demo account
   - Monitor (domain) creation and editing
   - Alert rule creation and editing

Tests run automatically on:
- Push to main/master branches
- Pull requests to main/master
- Manual workflow dispatch
- Comment with `/test` on pull requests

View test results and screenshots in the GitHub Actions artifacts.

### Test Coverage

The E2E test suite (`tests/e2e-workflow.spec.ts`) validates:
- ✅ Login with demo user credentials
- ✅ Add monitor for google.com with custom interval/timeout
- ✅ Edit monitor properties (name, interval, timeout)
- ✅ Add alert rule with trigger conditions
- ✅ Edit alert rule properties
- ✅ Complete workflow from login to alert configuration

All test runs generate screenshots at each step for visual verification.

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Run tests: `cd backend && go test ./...`
5. Validate frontend: `cd frontend && npm run check`
6. Submit a pull request

## License

This project is licensed under the MIT License - see the LICENSE file for details.
