# Observability Features

This document describes the logging and monitoring (observability) features implemented in V-Insight.

## Overview

V-Insight now includes comprehensive observability features:
- **Structured JSON logging** using Uber's Zap library
- **Prometheus metrics** for monitoring and alerting
- **Request tracing** with request IDs
- **Multi-tenant context** in logs

## Structured Logging

### Backend Logging

The backend API uses structured JSON logging for all requests and important events.

**Log Fields:**
- `timestamp`: ISO8601 formatted timestamp
- `level`: Log level (debug, info, warn, error)
- `msg`: Log message
- `method`: HTTP method
- `path`: Request path
- `status`: HTTP status code
- `duration`: Request duration
- `request_id`: Unique request identifier
- `tenant_id`: Tenant context (when available)
- `user_id`: User context (when available)
- `client_ip`: Client IP address

**Example Log Output:**
```json
{
  "timestamp": "2024-01-15T10:30:45.123Z",
  "level": "info",
  "msg": "HTTP request",
  "method": "GET",
  "path": "/api/v1/monitors",
  "status": 200,
  "duration": 0.045,
  "request_id": "550e8400-e29b-41d4-a716-446655440000",
  "tenant_id": 1,
  "user_id": 5,
  "client_ip": "192.168.1.100"
}
```

### Worker Logging

The worker service logs all job executions and important events.

**Logged Events:**
- Health check execution and results
- SSL certificate checks
- Alert rule evaluations
- Incident creation and resolution
- Notification sending (success/failure)

**Example Log Output:**
```json
{
  "timestamp": "2024-01-15T10:30:45.123Z",
  "level": "info",
  "msg": "Monitor check successful",
  "monitor_name": "Production API",
  "status_code": 200,
  "response_time_ms": 145
}
```

```json
{
  "timestamp": "2024-01-15T10:31:00.456Z",
  "level": "warn",
  "msg": "Incident created",
  "monitor_id": "mon_123",
  "rule_name": "API Down Alert",
  "trigger_value": "Monitor is down: Connection refused"
}
```

## Prometheus Metrics

### Backend Metrics

Available at `http://localhost:8080/metrics`

**HTTP Request Metrics:**
- `http_request_total`: Total HTTP requests (labels: method, path, status)
- `http_request_duration_seconds`: Request duration histogram (labels: method, path, status)

**Monitor Metrics:**
- `monitor_count`: Current number of active monitors (gauge)
- `open_incidents_count`: Current number of open incidents (gauge)
- `monitor_check_total`: Total monitor checks performed (labels: status)
- `monitor_check_failed_total`: Total failed monitor checks (counter)
- `monitor_check_duration_seconds`: Monitor check duration (labels: monitor_id)
- `monitor_response_time_ms`: HTTP response time (labels: monitor_id)

### Worker Metrics

Available at `http://localhost:8081/metrics`

**Job Execution Metrics:**
- `worker_job_execution_total`: Total job executions (labels: job_name, status)
- `worker_job_execution_duration_seconds`: Job execution duration (labels: job_name)

**Monitor Check Metrics:**
- `worker_monitor_check_total`: Total monitor checks (labels: status)

**Alert Metrics:**
- `worker_alert_evaluation_total`: Total alert evaluations (labels: status)
- `worker_incident_created_total`: Total incidents created (counter)
- `worker_incident_resolved_total`: Total incidents resolved (counter)

**Notification Metrics:**
- `worker_notification_sent_total`: Total notifications sent (labels: channel_type, status)

### Example Prometheus Queries

Monitor success rate:
```promql
rate(worker_monitor_check_total{status="success"}[5m]) / rate(worker_monitor_check_total[5m])
```

Average response time:
```promql
rate(monitor_response_time_ms_sum[5m]) / rate(monitor_response_time_ms_count[5m])
```

Open incidents over time:
```promql
open_incidents_count
```

Notification failure rate:
```promql
rate(worker_notification_sent_total{status="failure"}[5m]) / rate(worker_notification_sent_total[5m])
```

## Environment Configuration

Logging behavior can be configured via environment variables:

- `SERVER_ENV=production`: Enables production logging mode (no color, optimized for parsing)
- `SERVER_ENV=development`: Enables development logging mode (includes caller information)

## Integration with Monitoring Tools

### Prometheus

Add these scrape configs to your `prometheus.yml`:

```yaml
scrape_configs:
  - job_name: 'v-insight-backend'
    static_configs:
      - targets: ['localhost:8080']
    metrics_path: '/metrics'
    scrape_interval: 15s

  - job_name: 'v-insight-worker'
    static_configs:
      - targets: ['localhost:8081']
    metrics_path: '/metrics'
    scrape_interval: 15s
```

### Grafana Dashboards

Example dashboard panels:

1. **Request Rate**: `rate(http_request_total[5m])`
2. **Error Rate**: `rate(http_request_total{status=~"5.."}[5m])`
3. **Response Time (p95)**: `histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))`
4. **Active Monitors**: `monitor_count`
5. **Open Incidents**: `open_incidents_count`

### Log Aggregation (ELK, Loki, etc.)

JSON logs can be easily ingested by log aggregation tools:

- **Elasticsearch/Logstash/Kibana (ELK)**: Parse JSON logs with Logstash
- **Grafana Loki**: Use JSON parser in Loki configuration
- **CloudWatch Logs**: Use JSON log format for structured querying

Example Loki query:
```logql
{service="v-insight-backend"} | json | level="error" | tenant_id="1"
```

## Testing

To verify logging and metrics:

```bash
# Start services
make up

# Check backend metrics endpoint
curl http://localhost:8080/metrics

# Check worker metrics endpoint  
curl http://localhost:8081/metrics

# Trigger some activity and check logs
docker compose logs -f backend
docker compose logs -f worker

# Run tests
cd backend && go test ./...
cd worker && go test ./...
```

## Best Practices

1. **Always include request_id** in logs for request tracing
2. **Add tenant_id and user_id** when available for multi-tenant debugging
3. **Use appropriate log levels**:
   - `debug`: Detailed debugging information
   - `info`: Normal operation events (requests, job runs)
   - `warn`: Warning conditions (auth failures, slow responses)
   - `error`: Error conditions (failures, exceptions)
4. **Monitor key metrics**:
   - Request rate and latency
   - Error rates
   - Monitor check success rate
   - Incident creation rate
   - Notification delivery rate

## Future Enhancements

Potential improvements for observability:

- [ ] Distributed tracing with OpenTelemetry
- [ ] Custom metrics for business KPIs
- [ ] Log sampling for high-volume endpoints
- [ ] Anomaly detection on metrics
- [ ] Alerting rules for critical metrics
