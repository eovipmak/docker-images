# V-Insight Worker

Background job processor for V-Insight monitoring platform, responsible for executing health checks, SSL monitoring, alert evaluation, and notification delivery.

## Architecture

### Tech Stack

- **Language**: Go 1.23+
- **Job Scheduler**: robfig/cron v3
- **Database**: PostgreSQL 15 with sqlx (shared with backend)
- **HTTP Client**: net/http (health checks)
- **TLS Client**: crypto/tls (SSL certificate checks)
- **Hot Reload**: Air (development)

### Project Structure

```
worker/
├── cmd/worker/               # Application entry point
│   └── main.go              # Worker initialization, job scheduling
├── internal/
│   ├── jobs/                # Job implementations
│   │   ├── health_check_job.go      # HTTP/HTTPS health checks
│   │   ├── ssl_check_job.go         # SSL certificate monitoring
│   │   ├── alert_evaluator_job.go   # Alert condition evaluation
│   │   ├── notification_job.go      # Notification delivery
│   │   └── event_broadcaster.go     # SSE event broadcasting
│   ├── executor/            # Job execution engines
│   │   ├── executor.go             # Concurrent job executor
│   │   ├── http_checker.go         # HTTP health check logic
│   │   └── ssl_checker.go          # SSL verification logic
│   ├── scheduler/           # Job scheduling
│   │   └── scheduler.go            # Cron job manager
│   ├── database/            # Database connection management
│   └── config/              # Configuration management
├── go.mod                   # Go dependencies
├── go.sum                   # Dependency checksums
└── .air.toml               # Hot-reload configuration
```

## Notes for AI Agents & Automation

- The worker runs scheduled jobs that process all tenant monitors — any code changes must preserve multi-tenant behavior.
- Avoid hardcoding tenant IDs or modifying the way tenant context is loaded.
- Tests: `cd worker && go test ./...` ; add tests for concurrency and job behavior changes.
- See `docs/ai_agents.md` for automation and safe PR patterns.

## Features

### Job Types

#### 1. Health Check Job
**Schedule**: Every 30 seconds

Performs HTTP/HTTPS health checks for all enabled monitors:
- Sends HTTP request to monitor URL
- Records response time
- Captures HTTP status code
- Detects connection errors
- Stores results in `monitor_checks` table

**Execution flow:**
1. Query all enabled monitors from database
2. Execute checks concurrently (configurable parallelism)
3. For each monitor:
   - Send HTTP request with configured timeout
   - Measure response time
   - Record success/failure
   - Store check result with timestamp
4. Update monitor status based on check result

#### 2. SSL Check Job
**Schedule**: Every 5 minutes

Monitors SSL/TLS certificates for HTTPS endpoints:
- Establishes TLS connection
- Extracts certificate information
- Calculates days until expiration
- Verifies certificate validity
- Stores SSL status in `monitor_checks`

**Execution flow:**
1. Query all enabled monitors with SSL checking enabled
2. Execute SSL checks concurrently
3. For each monitor:
   - Connect to host via TLS
   - Parse certificate chain
   - Extract expiration date
   - Validate certificate
   - Calculate days until expiry
4. Store SSL check results

#### 3. Alert Evaluator Job
**Schedule**: Every 60 seconds

Evaluates alert rules against recent monitoring data:
- Checks for monitor downtime
- Detects slow response times
- Identifies expiring SSL certificates
- Creates/resolves incidents
- Triggers notifications

**Alert Types:**
- **down**: Monitor failed N consecutive checks
- **slow_response**: Response time exceeds threshold (ms)
- **ssl_expiry**: Certificate expires within threshold (days)

**Execution flow:**
1. Query all enabled alert rules
2. For each rule:
   - Fetch recent monitor checks
   - Evaluate trigger condition
   - Create incident if condition met
   - Resolve incident if condition cleared
3. Queue notifications for new/resolved incidents

#### 4. Notification Job
**Schedule**: Every 30 seconds

Delivers notifications through configured channels:
- Webhook (generic HTTP POST)
- Discord (rich embed)
- Email (ready for implementation)

**Execution flow:**
1. Query pending notifications (not yet sent)
2. For each notification:
   - Fetch incident details
   - Fetch associated channels
   - Send notification to each channel
   - Mark notification as sent
   - Handle delivery failures

#### 5. Event Broadcaster
**On-Demand** (triggered by notifications)

Broadcasts real-time events via Server-Sent Events (SSE):
- Sends events to backend `/internal/broadcast` endpoint
- Backend forwards to connected SSE clients
- Enables real-time UI updates

### Concurrent Execution

The worker uses a concurrent executor pattern:

```go
type Executor struct {
    maxConcurrency int
    semaphore      chan struct{}
}

// Execute up to N jobs concurrently
func (e *Executor) Execute(monitors []*Monitor, fn func(*Monitor)) {
    var wg sync.WaitGroup
    for _, monitor := range monitors {
        e.semaphore <- struct{}{} // Acquire
        wg.Add(1)
        go func(m *Monitor) {
            defer wg.Done()
            defer func() { <-e.semaphore }() // Release
            fn(m)
        }(monitor)
    }
    wg.Wait()
}
```

**Benefits:**
- Prevents resource exhaustion
- Limits concurrent database connections
- Configurable parallelism
- Graceful handling of large monitor counts

## Getting Started

### Prerequisites

- Go 1.23 or higher
- PostgreSQL 15 (shared with backend)
- Backend API running (for event broadcasting)

### Environment Variables

Uses the same `.env` file as backend (in project root):

```bash
# Database (shared with backend)
POSTGRES_USER=postgres
POSTGRES_PASSWORD=postgres
POSTGRES_DB=v_insight

# Worker Configuration
WORKER_PORT=8081

# Backend API URL (for event broadcasting)
BACKEND_API_URL=http://backend:8080

# Environment
ENV=development  # production | staging | development
```

### Local Development

#### Option 1: Docker (Recommended)

```bash
# From project root
make up            # Start all services
make logs-worker   # View worker logs
```

The worker health check is available at `http://localhost:8081/health`

#### Option 2: Local Build

```bash
cd worker

# Install dependencies
go mod download

# Run with hot-reload (development)
air

# Or run directly
go run cmd/worker/main.go
```

### Building for Production

```bash
cd worker

# Build binary
go build -o bin/worker cmd/worker/main.go

# Run
./bin/worker
```

## Job Scheduling

### Cron Schedule Format

```
┌─────────────── minute (0 - 59)
│ ┌───────────── hour (0 - 23)
│ │ ┌─────────── day of month (1 - 31)
│ │ │ ┌───────── month (1 - 12)
│ │ │ │ ┌─────── day of week (0 - 6) (Sunday to Saturday)
│ │ │ │ │
* * * * *
```

### Current Schedules

```go
// Health checks every 30 seconds
scheduler.AddJob("*/30 * * * * *", healthCheckJob)

// SSL checks every 5 minutes
scheduler.AddJob("*/5 * * * *", sslCheckJob)

// Alert evaluation every 60 seconds
scheduler.AddJob("*/60 * * * * *", alertEvaluatorJob)

// Notifications every 30 seconds
scheduler.AddJob("*/30 * * * * *", notificationJob)
```

### Modifying Schedules

Edit `cmd/worker/main.go`:

```go
// Change health check to every 60 seconds
scheduler.AddJob("*/60 * * * * *", healthCheckJob)

// Change SSL check to every 10 minutes
scheduler.AddJob("*/10 * * * *", sslCheckJob)
```

## Database Schema

### Tables Used by Worker

**Read from:**
- `monitors` - Active monitors to check
- `alert_rules` - Alert conditions to evaluate
- `alert_channels` - Notification destinations
- `alert_rule_channels` - Rule-to-channel associations
- `incidents` - Open incidents (for evaluation)

**Write to:**
- `monitor_checks` - Health and SSL check results
- `incidents` - Created/resolved incidents

### Monitor Checks Schema

```sql
CREATE TABLE monitor_checks (
    id UUID PRIMARY KEY,
    monitor_id UUID REFERENCES monitors(id),
    checked_at TIMESTAMP NOT NULL,
    status_code INTEGER,
    response_time_ms INTEGER,
    ssl_valid BOOLEAN,
    ssl_expires_at TIMESTAMP,
    error_message TEXT,
    success BOOLEAN NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);
```

### Incident Schema

```sql
CREATE TABLE incidents (
    id UUID PRIMARY KEY,
    monitor_id UUID REFERENCES monitors(id),
    alert_rule_id UUID REFERENCES alert_rules(id),
    started_at TIMESTAMP NOT NULL,
    resolved_at TIMESTAMP,
    status VARCHAR(20) NOT NULL,
    trigger_value TEXT,
    notified_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW()
);
```

## Testing

```bash
cd worker

# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific package tests
go test ./internal/jobs/...
go test ./internal/executor/...

# Run tests in verbose mode
go test -v ./...
```

### Test Coverage

Current test files:
- `internal/jobs/health_check_job_test.go`
- `internal/jobs/ssl_check_job_test.go`
- `internal/jobs/alert_evaluator_job_test.go`
- `internal/jobs/notification_job_test.go`
- `internal/executor/executor_test.go`
- `internal/executor/http_checker_test.go`
- `internal/executor/ssl_checker_test.go`
- `internal/scheduler/scheduler_test.go`

## Monitoring and Observability

### Health Check Endpoint

```bash
curl http://localhost:8081/health
```

Response:
```json
{
  "status": "ok",
  "database": "connected",
  "jobs": {
    "health_check": "running",
    "ssl_check": "running",
    "alert_evaluator": "running",
    "notification": "running"
  }
}
```

### Logs

Worker logs include:
- Job start/completion times
- Number of monitors processed
- Errors and warnings
- Database query performance
- Notification delivery status

**Example log output:**
```
[Worker] Starting health check job
[HealthCheck] Processing 45 monitors
[HealthCheck] Completed in 2.3s - Success: 43, Failed: 2
[AlertEvaluator] Evaluated 12 alert rules
[AlertEvaluator] Created 1 new incident(s)
[Notification] Sent 3 notifications
```

### Performance Metrics

Monitor these metrics in production:
- Job execution duration
- Monitor check success rate
- Alert evaluation latency
- Notification delivery success rate
- Database connection pool usage

## Configuration

### Executor Concurrency

Default: 10 concurrent operations

Modify in job initialization:

```go
executor := executor.NewExecutor(20) // 20 concurrent checks
```

**Guidelines:**
- Lower values: Reduced load, slower execution
- Higher values: Faster execution, more database connections
- Consider: Database connection limits, CPU cores

### HTTP Client Timeouts

Default timeouts in `http_checker.go`:

```go
client := &http.Client{
    Timeout: time.Duration(monitor.Timeout) * time.Second,
    Transport: &http.Transport{
        TLSHandshakeTimeout:   10 * time.Second,
        ResponseHeaderTimeout: 10 * time.Second,
        IdleConnTimeout:       90 * time.Second,
    },
}
```

### SSL Verification

Configure in `ssl_checker.go`:

```go
config := &tls.Config{
    InsecureSkipVerify: false, // Set true to accept self-signed certs
    MinVersion:        tls.VersionTLS12,
}
```

## Common Tasks

### Adding a New Job

1. Create job file in `internal/jobs/`:

```go
// internal/jobs/my_new_job.go
package jobs

type MyNewJob struct {
    db *database.DB
}

func NewMyNewJob(db *database.DB) *MyNewJob {
    return &MyNewJob{db: db}
}

func (j *MyNewJob) Run() {
    log.Println("[MyNewJob] Starting...")
    // Job logic here
    log.Println("[MyNewJob] Completed")
}
```

2. Register in `cmd/worker/main.go`:

```go
myNewJob := jobs.NewMyNewJob(db)
scheduler.AddJob("*/5 * * * *", myNewJob) // Every 5 minutes
```

### Debugging Job Execution

Enable verbose logging:

```go
// In job file
log.Printf("[Debug] Processing item: %+v", item)
log.Printf("[Debug] Query result: %v rows", len(results))
```

View logs:
```bash
make logs-worker
```

### Testing Notification Channels

Use the backend API to test channels:

```bash
# Test webhook channel
curl -X POST http://localhost:8080/api/v1/alert-channels/{id}/test \
  -H "Authorization: Bearer $TOKEN"

# Test Discord channel
curl -X POST http://localhost:8080/api/v1/alert-channels/{id}/test \
  -H "Authorization: Bearer $TOKEN"
```

## Troubleshooting

### Job Not Running

1. Check scheduler logs:
   ```bash
   make logs-worker | grep -i "scheduler"
   ```

2. Verify cron expression:
   ```bash
   # Test expression at https://crontab.guru/
   ```

3. Check for panics:
   ```bash
   make logs-worker | grep -i "panic"
   ```

### Database Connection Issues

```bash
# Check database connectivity
docker compose exec worker psql -U postgres -d v_insight -c "SELECT 1;"

# Check connection pool
make logs-worker | grep -i "database"
```

### High Memory Usage

Reduce executor concurrency:

```go
// In cmd/worker/main.go
executor := executor.NewExecutor(5) // Lower from 10 to 5
```

### Notification Delivery Failures

Check logs for specific errors:

```bash
make logs-worker | grep -i "notification"
```

Common issues:
- Invalid webhook URL
- Discord webhook expired
- Network connectivity problems

## Production Deployment

### Build Optimizations

```bash
# Build with optimizations
CGO_ENABLED=0 go build -ldflags="-w -s" -o bin/worker cmd/worker/main.go
```

### Environment Settings

- Set `ENV=production`
- Configure appropriate job schedules
- Adjust executor concurrency based on load
- Set reasonable HTTP timeouts

### Resource Requirements

**Minimum:**
- CPU: 1 core
- Memory: 256MB
- Storage: 100MB (binary + logs)

**Recommended (1000 monitors):**
- CPU: 2 cores
- Memory: 512MB
- Storage: 1GB

### Health Checks

Configure health check in orchestration:

```yaml
# Kubernetes
livenessProbe:
  httpGet:
    path: /health
    port: 8081
  initialDelaySeconds: 10
  periodSeconds: 30

# Docker Compose
healthcheck:
  test: ["CMD", "curl", "-f", "http://localhost:8081/health"]
  interval: 30s
  timeout: 10s
  retries: 3
```

### Scaling

Worker is designed to run as a single instance per environment:
- Jobs use database locking where needed
- Concurrent execution handled internally
- Running multiple instances may cause duplicate work

For high-volume scenarios:
- Increase executor concurrency
- Optimize database queries
- Consider sharding monitors across multiple workers

## Performance Tuning

### Database Query Optimization

1. Add indexes for frequently queried fields
2. Use batch operations where possible
3. Implement query result caching
4. Monitor slow queries

### Job Optimization

1. **Health Checks**: Adjust timeout based on monitor requirements
2. **SSL Checks**: Cache certificate info for repeated checks
3. **Alert Evaluation**: Optimize lookback window
4. **Notifications**: Batch notifications by channel

### Monitoring Job Performance

```go
import "time"

func (j *Job) Run() {
    start := time.Now()
    defer func() {
        duration := time.Since(start)
        log.Printf("[Job] Completed in %v", duration)
    }()
    
    // Job logic
}
```

## Contributing

1. Run tests before committing: `go test ./...`
2. Format code: `go fmt ./...`
3. Add tests for new jobs
4. Document new configuration options
5. Update this README for significant changes

## Resources

- [robfig/cron Documentation](https://pkg.go.dev/github.com/robfig/cron/v3)
- [Go Concurrency Patterns](https://go.dev/blog/pipelines)
- [PostgreSQL Performance Tuning](https://wiki.postgresql.org/wiki/Performance_Optimization)
- [Crontab Guru](https://crontab.guru/) - Cron expression tester
