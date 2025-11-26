# Testing Observability Features

This guide provides step-by-step instructions for testing the logging and monitoring features.

## Prerequisites

1. Docker and Docker Compose installed
2. `.env` file created from `.env.example`:
   ```bash
   cp .env.example .env
   ```

## Quick Start Testing

### 1. Start Services

```bash
# Start all services with Docker Compose
make up

# Wait for services to be healthy (~30 seconds)
# Check service status
make ps
```

### 2. Verify Dependencies Installation

The Docker images will automatically run `go mod tidy` during build to fetch:
- `go.uber.org/zap` - Structured logging library
- `github.com/prometheus/client_golang` - Prometheus metrics client

Check build logs:
```bash
docker compose logs backend | grep "go mod"
docker compose logs worker | grep "go mod"
```

### 3. Test Metrics Endpoints

**Backend Metrics:**
```bash
# Check metrics endpoint is accessible
curl http://localhost:8080/metrics

# Should return Prometheus metrics format, including:
# - http_request_total
# - http_request_duration_seconds
# - monitor_count
# - open_incidents_count
```

**Worker Metrics:**
```bash
# Check worker metrics endpoint
curl http://localhost:8081/metrics

# Should return metrics including:
# - worker_job_execution_total
# - worker_monitor_check_total
# - worker_incident_created_total
# - worker_notification_sent_total
```

### 4. Test Structured Logging

**Backend Logs:**
```bash
# View backend logs in real-time
docker compose logs -f backend

# Generate some requests
curl http://localhost:8080/health
curl http://localhost:8080/api/v1/

# Logs should show JSON format with fields:
# - timestamp
# - level
# - msg
# - method
# - path
# - status
# - duration
# - request_id
```

Example expected log output:
```json
{"timestamp":"2024-01-15T10:30:45.123Z","level":"info","msg":"HTTP request","method":"GET","path":"/health","status":200,"duration":0.001234,"client_ip":"172.20.0.1"}
```

**Worker Logs:**
```bash
# View worker logs in real-time
docker compose logs -f worker

# Logs should show JSON format with job execution details
# Example log entries:
# - "Starting health check run"
# - "Monitor check successful" (with monitor details)
# - "Alert evaluation completed" (with metrics)
```

### 5. Test Authentication Logging

```bash
# Register a new user (generates auth logs)
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "password123",
    "tenant_name": "Test Tenant"
  }'

# Check backend logs for:
# - "User registration attempt" (info level)
# - "User registration successful" (info level)

# Try logging in
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "password123"
  }'

# Check backend logs for:
# - "User login attempt" (info level)
# - "User login successful" (info level)
```

### 6. Test Monitor Check Logging

```bash
# First, get an auth token
TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password123"}' | \
  jq -r '.token')

# Create a monitor
curl -X POST http://localhost:8080/api/v1/monitors \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test Monitor",
    "url": "https://httpbin.org/status/200",
    "check_interval": 60,
    "timeout": 5
  }'

# Wait for the health check job to run (every 30 seconds)
# Check worker logs for:
docker compose logs -f worker | grep "Monitor check"

# Should show:
# - "Checking monitor" (debug level)
# - "Monitor check successful" (info level) with status_code and response_time_ms
```

### 7. Test Alert Evaluation Logging

If you have alert rules configured:

```bash
# View worker logs for alert evaluation
docker compose logs -f worker | grep -E "(Alert evaluation|Incident)"

# Should show:
# - "Starting alert evaluation run" (info level)
# - "Alert evaluation completed" (info level with metrics)
# - "Incident created" (warn level) when alert triggers
# - "Incident resolved" (info level) when alert clears
```

### 8. Test Notification Logging

If you have alert channels configured:

```bash
# View worker logs for notifications
docker compose logs -f worker | grep "Notification"

# Should show:
# - "Starting notification processing run" (info level)
# - "Notification sent" (info level) with channel details
# - "Notification processing completed" (info level with metrics)
```

## Metrics Validation

### Check Specific Metrics

```bash
# Monitor check metrics
curl -s http://localhost:8081/metrics | grep worker_monitor_check_total

# Job execution metrics
curl -s http://localhost:8081/metrics | grep worker_job_execution

# HTTP request metrics
curl -s http://localhost:8080/metrics | grep http_request_total

# Incident metrics
curl -s http://localhost:8081/metrics | grep incident
```

### Visualize Metrics with Prometheus (Optional)

1. Add Prometheus to docker-compose.yml or run standalone:
   ```yaml
   prometheus:
     image: prom/prometheus:latest
     ports:
       - "9090:9090"
     volumes:
       - ./prometheus.yml:/etc/prometheus/prometheus.yml
     command:
       - '--config.file=/etc/prometheus/prometheus.yml'
   ```

2. Create `prometheus.yml`:
   ```yaml
   global:
     scrape_interval: 15s
   
   scrape_configs:
     - job_name: 'v-insight-backend'
       static_configs:
         - targets: ['backend:8080']
     
     - job_name: 'v-insight-worker'
       static_configs:
         - targets: ['worker:8081']
   ```

3. Access Prometheus UI at http://localhost:9090

## Running Unit Tests

### Backend Tests

```bash
# Run all backend tests
cd backend && go test ./...

# Run middleware tests specifically
cd backend && go test ./internal/api/middleware/... -v

# Run with coverage
cd backend && go test ./... -cover
```

### Worker Tests

```bash
# Run all worker tests
cd worker && go test ./...

# Run with coverage
cd worker && go test ./... -cover
```

## Troubleshooting

### Logs Not in JSON Format

- Check that logger is initialized: look for "Starting V-Insight" log
- Verify environment variable: `ENV=development` or `ENV=production`
- Check for initialization errors in startup logs

### Metrics Endpoint Returns 404

- Verify service is running: `docker compose ps`
- Check health endpoint first: `curl http://localhost:8080/health`
- Review service logs for startup errors: `docker compose logs backend`

### No Logs Appearing

- Check if services are running: `make ps`
- Restart services: `make restart`
- Check Docker logs: `docker compose logs -f`

### Tests Failing

- Ensure dependencies are installed: `cd backend && go mod tidy`
- Check Go version: requires Go 1.21+
- Run tests individually to isolate issues

## Expected Behavior Summary

### ✅ Successful Implementation Indicators

1. **Logs:**
   - All logs are in JSON format
   - Logs include timestamp, level, msg fields
   - Request logs include request_id, method, path, status, duration
   - Auth logs include email field
   - Monitor logs include monitor_name, response_time_ms
   - Incident logs include monitor_id, rule_name

2. **Metrics:**
   - `/metrics` endpoints respond with Prometheus format
   - Counters increment with activity
   - Histograms show duration distributions
   - Gauges reflect current state

3. **Tests:**
   - All unit tests pass
   - Request logger middleware tests pass
   - No test failures in CI/CD

## Performance Validation

Monitor the impact of logging/metrics:

```bash
# Check service resource usage
docker stats v-insight-backend v-insight-worker

# Typical overhead should be:
# - CPU: < 5% increase
# - Memory: < 50MB increase
# - No noticeable latency increase
```

## Next Steps

After validation:
1. Configure log aggregation (ELK, Loki, CloudWatch)
2. Set up Prometheus scraping
3. Create Grafana dashboards
4. Configure alerting rules
5. Set up log retention policies

## Support

If issues persist:
- Check OBSERVABILITY.md for detailed documentation
- Review Docker logs: `make logs`
- Check service health: `make ps`
- Verify .env configuration
