# HTTP Health Check Worker Implementation

## Overview

This implementation provides automated HTTP health checking for monitors in the V-Insight platform. The worker runs every 30 seconds to check monitors that are due for their scheduled health checks.

## Key Features

### 1. Concurrent Processing
- Worker pool with maximum 10 concurrent monitor checks
- Uses Go semaphores for concurrency control
- Non-blocking concurrent execution with proper synchronization

### 2. Smart Scheduling
- Monitors are checked based on their `check_interval` setting
- Only enabled monitors are checked
- Monitors with `last_checked_at` = NULL (never checked) are prioritized
- SQL query efficiently identifies monitors needing checks

### 3. HTTP Health Checking
- HTTP GET requests to monitor URLs
- Configurable timeout per monitor
- Follows up to 5 redirects
- Tracks status code and response time
- Custom User-Agent: "V-Insight-Monitor/1.0"

### 4. Result Storage
- Each check result saved to `monitor_checks` table
- Records: status code, response time (ms), success/failure, error messages
- Updates `last_checked_at` timestamp on monitors table

### 5. Context-Aware
- Respects context cancellation
- Monitor-specific timeouts
- Graceful shutdown support

## Architecture

### Components

#### 1. `http_checker.go`
- `HTTPChecker` struct with HTTP client
- `CheckURL()` method performs actual HTTP requests
- Returns `HTTPCheckResult` with status, timing, and error info
- Configurable redirect limit (max 5)

#### 2. `health_check_job.go`
- `HealthCheckJob` implements scheduler.Job interface
- `Run()` - main entry point called by scheduler
- `getMonitorsNeedingCheck()` - queries database for monitors to check
- `checkMonitorsConcurrently()` - manages worker pool
- `checkMonitor()` - performs individual monitor check
- `saveCheck()` - persists check result
- `updateLastCheckedAt()` - updates monitor timestamp

#### 3. Database Schema
- Migration `000004_add_last_checked_at_to_monitors.up.sql`
- Adds `last_checked_at TIMESTAMP` to monitors table
- Creates composite index for efficient queries

#### 4. Repository Methods (Backend)
- `GetMonitorsNeedingCheck(now)` - finds monitors due for checking
- `SaveCheck(check)` - saves check result
- `UpdateLastCheckedAt(id, time)` - updates timestamp

## Database Queries

### Get Monitors Needing Check
```sql
SELECT id, tenant_id, name, url, check_interval, timeout, enabled, 
       last_checked_at, created_at, updated_at
FROM monitors
WHERE enabled = true
  AND (
      last_checked_at IS NULL
      OR last_checked_at + (check_interval || ' seconds')::INTERVAL <= $1
  )
ORDER BY last_checked_at ASC NULLS FIRST
```

### Save Check Result
```sql
INSERT INTO monitor_checks (
    monitor_id, checked_at, status_code, response_time_ms, 
    ssl_valid, ssl_expires_at, error_message, success
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING id
```

### Update Last Checked
```sql
UPDATE monitors
SET last_checked_at = $1
WHERE id = $2
```

## Scheduling

The health check job runs every 30 seconds using cron syntax:
```
*/30 * * * * *
```

This schedule is configured in `worker/cmd/worker/main.go`.

## Logging

Detailed logging for debugging:
- Job start/completion with duration
- Number of monitors found
- Individual monitor check results
- Success: "✓ Monitor {name} is UP - Status: {code}, Response: {ms}ms"
- Failure: "✗ Monitor {name} is DOWN - Error: {error}"
- Database operation failures

## Error Handling

- Database errors are logged but don't stop other checks
- HTTP errors are captured in check results
- Context cancellation is handled gracefully
- Individual monitor failures don't affect others

## Testing

### Unit Tests

#### HTTP Checker Tests (`http_checker_test.go`)
- Success cases with 200 OK
- Server errors (500)
- Redirects (up to 5)
- Timeouts
- Invalid URLs
- Context cancellation
- User-Agent verification

#### Job Tests (`health_check_job_test.go`)
- Struct field validation
- Job creation
- Empty monitor list handling
- Context cancellation

### Integration Tests
- Requires database connection
- Skipped in CI without database
- Can be enabled with test database configuration

## Performance Considerations

1. **Concurrency Limit**: Max 10 concurrent checks prevents overwhelming the system
2. **Per-Monitor Timeout**: Each monitor has its own timeout setting
3. **Efficient Queries**: Indexes on `enabled` and `last_checked_at` for fast lookups
4. **Batch Processing**: All due monitors checked in single job run

## Configuration

Monitor settings (per monitor):
- `check_interval`: How often to check (seconds)
- `timeout`: Maximum time for HTTP request (seconds)
- `enabled`: Whether monitoring is active

Worker settings:
- Runs every 30 seconds
- 10 concurrent worker limit (hardcoded, can be made configurable)

## Future Enhancements

Potential improvements:
1. Configurable worker pool size
2. Retry logic for failed checks
3. Alert triggering on status changes
4. Multiple check methods (TCP, ICMP)
5. Custom HTTP headers/methods
6. Response body validation
7. Metrics collection (Prometheus)
