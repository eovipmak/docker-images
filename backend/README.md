# V-Insight Backend API

Go-based REST API server for V-Insight monitoring platform, providing endpoints for monitor management, alert configuration, and incident tracking.

## Notes for AI Agents & Automation

- The backend enforces multi-tenant behavior: every persisted object includes a `tenant_id` and requests must be validated via `TenantRequired` middleware.
- Avoid adding CORS handling in the backend; the frontend proxy handles cross-origin behavior.
- Migrations are in `backend/migrations/` and run automatically at startup — add `.up.sql` and corresponding `.down.sql` files for schema changes.
- Refer to `docs/ai_agents.md` for safe automation patterns and validation scripts.

## Architecture

### Tech Stack

- **Language**: Go 1.23+
- **Framework**: Gin (HTTP router)
- **Database**: PostgreSQL 15 with sqlx
- **Authentication**: JWT (golang-jwt/jwt/v5)
- **API Documentation**: Swagger/OpenAPI (swaggo/gin-swagger)
- **Migrations**: golang-migrate
- **Hot Reload**: Air (development)

### Project Structure

```
backend/
├── cmd/api/                    # Application entry point
│   └── main.go                # Server initialization, routing
├── internal/
│   ├── api/
│   │   ├── handlers/          # HTTP request handlers
│   │   │   ├── auth_handler.go          # Authentication endpoints
│   │   │   ├── monitor_handler.go       # Monitor CRUD
│   │   │   ├── alert_rule_handler.go    # Alert rule CRUD
│   │   │   ├── alert_channel_handler.go # Notification channels
│   │   │   ├── incident_handler.go      # Incident management
│   │   │   ├── dashboard_handler.go     # Dashboard statistics
│   │   │   ├── metrics_handler.go       # Metrics endpoints
│   │   │   └── stream_handler.go        # SSE real-time events
│   │   └── middleware/        # HTTP middleware
│   │       ├── auth.go               # JWT authentication
│   │       ├── tenant.go             # Multi-tenant context
│   │       ├── rate_limit.go         # Rate limiting
│   │       ├── security_headers.go   # Security headers
│   │       ├── request_id.go         # Request tracking
│   │       └── performance.go        # Gzip, caching
│   ├── domain/
│   │   ├── entities/          # Domain models (Monitor, AlertRule, etc.)
│   │   ├── repository/        # Repository interfaces
│   │   └── service/           # Business logic services
│   ├── repository/postgres/   # PostgreSQL implementations
│   ├── database/              # Database connection management
│   ├── config/                # Configuration management
│   ├── auth/                  # JWT and password utilities
│   └── utils/                 # Helper functions (sanitization, validation)
├── migrations/                # Database migrations (auto-run on startup)
├── docs/                      # Generated Swagger documentation
├── go.mod                     # Go dependencies
├── go.sum                     # Dependency checksums
└── .air.toml                  # Hot-reload configuration
```

## Features

### Multi-Tenant Architecture

- **Tenant Isolation**: All database queries scoped to tenant_id
- **Middleware Enforcement**: TenantRequired middleware validates access
- **Row-Level Security**: Database-level tenant filtering

### Security

- **JWT Authentication**: Secure token-based authentication
- **Password Hashing**: bcrypt with cost factor 12
- **Rate Limiting**: 
  - Per-IP: 100 req/min (configurable via `RATE_LIMIT_PER_IP`)
  - Per-User: 1000 req/hour (configurable via `RATE_LIMIT_PER_USER`)
- **Security Headers**: HSTS, X-Content-Type-Options, X-Frame-Options, X-XSS-Protection
- **Input Sanitization**: HTML escaping for user inputs
- **Request Size Limits**: 10MB default

### API Documentation

- **Swagger UI**: Available at `/swagger/` (development only)
- **OpenAPI 3.0**: Full API specification
- **Interactive Testing**: Try endpoints directly in browser
- **Authentication**: Bearer token support in Swagger UI

## Getting Started

### Prerequisites

- Go 1.23 or higher
- PostgreSQL 15
- Make (optional)

### Environment Variables

Copy `.env.example` to `.env` in the project root and configure:

```bash
# Database
POSTGRES_USER=postgres
POSTGRES_PASSWORD=postgres
POSTGRES_DB=v_insight

# Backend API
BACKEND_PORT=8080

# Environment
ENV=development  # production | staging | development

# Security
HSTS_MAX_AGE=31536000
HSTS_INCLUDE_SUBDOMAINS=true
RATE_LIMIT_PER_IP=100
RATE_LIMIT_PER_USER=1000
```

### Local Development

#### Option 1: Docker (Recommended)

```bash
# From project root
make up        # Start all services
make logs-backend  # View backend logs
```

The backend will be available at `http://localhost:8080`

#### Option 2: Local Build

```bash
cd backend

# Install dependencies
go mod download

# Run migrations (requires PostgreSQL running)
# Migrations run automatically when starting the server

# Run with hot-reload (development)
air

# Or run directly
go run cmd/api/main.go
```

### Building for Production

```bash
cd backend

# Build binary
go build -o bin/api cmd/api/main.go

# Run
./bin/api
```

## API Endpoints

### Authentication (Public)

- `POST /api/v1/auth/register` - Register new user and tenant
- `POST /api/v1/auth/login` - Login and get JWT token
- `GET /api/v1/auth/me` - Get current user info (protected)

### Monitors (Protected)

- `POST /api/v1/monitors` - Create monitor
- `GET /api/v1/monitors` - List all monitors
- `GET /api/v1/monitors/:id` - Get monitor details
- `PUT /api/v1/monitors/:id` - Update monitor
- `DELETE /api/v1/monitors/:id` - Delete monitor
- `GET /api/v1/monitors/:id/checks` - Get check history
- `GET /api/v1/monitors/:id/ssl-status` - Get SSL certificate status
- `GET /api/v1/monitors/:id/metrics` - Get monitor metrics

### Alert Rules (Protected)

- `POST /api/v1/alert-rules` - Create alert rule
- `GET /api/v1/alert-rules` - List all alert rules
- `GET /api/v1/alert-rules/:id` - Get alert rule details
- `PUT /api/v1/alert-rules/:id` - Update alert rule
- `DELETE /api/v1/alert-rules/:id` - Delete alert rule
- `POST /api/v1/alert-rules/:id/test` - Test alert rule configuration

### Alert Channels (Protected)

- `POST /api/v1/alert-channels` - Create notification channel
- `GET /api/v1/alert-channels` - List all channels
- `GET /api/v1/alert-channels/:id` - Get channel details
- `PUT /api/v1/alert-channels/:id` - Update channel
- `DELETE /api/v1/alert-channels/:id` - Delete channel
- `POST /api/v1/alert-channels/:id/test` - Test channel configuration

### Incidents (Protected)

- `GET /api/v1/incidents` - List incidents (filterable by status, monitor)
- `GET /api/v1/incidents/:id` - Get incident details
- `POST /api/v1/incidents/:id/resolve` - Manually resolve incident

### Dashboard (Protected)

- `GET /api/v1/dashboard/stats` - Get dashboard statistics
- `GET /api/v1/dashboard` - Get full dashboard data

### Real-Time Events (Protected)

- `GET /api/v1/stream/events` - Server-Sent Events stream

### Health Check (Public)

- `GET /health` - Health status and database connectivity

### API Documentation (Development Only)

- `GET /swagger/` - Swagger UI interface

## Testing

```bash
cd backend

# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests in verbose mode
go test -v ./...

# Run specific package tests
go test ./internal/repository/postgres/...
```

## Database Migrations

Migrations are located in `backend/migrations/` and run automatically on server startup.

### Manual Migration Operations

```bash
# From project root
make migrate-up          # Apply pending migrations
make migrate-down        # Rollback last migration
make migrate-version     # Show current migration version
make migrate-create name=description  # Create new migration files
```

### Creating New Migrations

1. Create migration files:
   ```bash
   make migrate-create name=add_new_feature
   ```

2. Edit the generated files:
   - `migrations/YYYYMMDDHHMMSS_add_new_feature.up.sql` - Schema changes
   - `migrations/YYYYMMDDHHMMSS_add_new_feature.down.sql` - Rollback changes

3. Migrations run automatically on next backend restart

## Configuration

### Server Configuration

Configured via environment variables (see `.env.example`):

- `BACKEND_PORT`: HTTP server port (default: 8080)
- `ENV`: Environment mode (production|staging|development)
- Database connection settings
- Security settings (HSTS, rate limits)
- JWT secret key

### Gin Mode

- `production`: Minimal logging, optimized performance
- Other environments: Verbose logging, debug mode

## Common Tasks

### Regenerate Swagger Documentation

```bash
cd backend

# Install swag CLI (if not installed)
go install github.com/swaggo/swag/cmd/swag@latest

# Generate docs
swag init -g cmd/api/main.go -o docs
```

### Add New Endpoint

1. Create handler method in appropriate handler file
2. Add Swagger annotations (godoc format)
3. Register route in `cmd/api/main.go`
4. Regenerate Swagger docs
5. Add tests

### Add Middleware

1. Create middleware function in `internal/api/middleware/`
2. Apply globally in `main.go` or to specific route groups
3. Document in this README

## Troubleshooting

### Port Already in Use

```bash
# Find process using port 8080
lsof -i :8080

# Kill the process
kill -9 <PID>
```

### Database Connection Issues

- Verify PostgreSQL is running: `docker compose ps`
- Check credentials in `.env`
- Ensure database exists: `psql -U postgres -c "CREATE DATABASE v_insight;"`

### Migration Errors

- Check migration files for syntax errors
- Verify migration version: `make migrate-version`
- Manually rollback if needed: `make migrate-down`

### Hot Reload Not Working

- Check `.air.toml` configuration
- Look for syntax errors in `build-errors.log`
- Restart Air manually

## Production Deployment

### Build Optimizations

```bash
# Build with optimizations
CGO_ENABLED=0 go build -ldflags="-w -s" -o bin/api cmd/api/main.go
```

### Environment Settings

- Set `ENV=production`
- Use strong `JWT_SECRET`
- Enable HSTS with long max age
- Configure appropriate rate limits
- Use SSL/TLS termination proxy (nginx, Caddy)

### Health Checks

Configure health check endpoint (`/health`) in:
- Load balancer
- Container orchestration (Kubernetes, Docker Swarm)
- Monitoring systems

## Contributing

1. Run tests before committing: `go test ./...`
2. Format code: `go fmt ./...`
3. Update Swagger docs when modifying endpoints
4. Follow existing code patterns and naming conventions
5. Add tests for new functionality

## Resources

- [Gin Framework Documentation](https://gin-gonic.com/docs/)
- [sqlx Documentation](http://jmoiron.github.io/sqlx/)
- [Swagger/OpenAPI Specification](https://swagger.io/specification/)
- [golang-migrate Documentation](https://github.com/golang-migrate/migrate)
