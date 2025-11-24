# SSE Real-Time Updates - Testing Guide

## Overview
This document provides testing instructions for the Server-Sent Events (SSE) implementation in V-Insight.

## Architecture Summary

### Flow
1. **Worker** performs health checks/alert evaluations
2. **Worker** sends event to Backend via `/internal/broadcast`
3. **Backend** broadcasts event to all connected SSE clients for that tenant
4. **Frontend** receives event via EventSource and updates UI stores
5. **UI** reactively updates based on store changes

### Event Types
- `monitor_check`: Emitted after each monitor health check
- `incident_created`: Emitted when a new incident is created
- `incident_resolved`: Emitted when an incident is resolved

## Testing Instructions

### Prerequisites
1. Start all services: `make up`
2. Wait for services to be healthy (~30 seconds)
3. Access frontend at http://localhost:3000

### Manual Testing Scenario

#### 1. User Registration & Login
```bash
# Register a new user
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "password123",
    "full_name": "Test User"
  }'

# Login and get token
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "password123"
  }'

# Save the returned token as TOKEN
export TOKEN="<your_token_here>"
```

#### 2. Create a Monitor
```bash
# Create a monitor
curl -X POST http://localhost:8080/api/v1/monitors \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "name": "Google",
    "url": "https://www.google.com",
    "check_interval": 60,
    "timeout": 10,
    "enabled": true,
    "check_ssl": true,
    "ssl_alert_days": 30
  }'
```

#### 3. Test SSE Connection

**Option A: Using curl (simple test)**
```bash
# Test SSE endpoint (should see connection event)
curl -N -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8080/api/v1/stream/events?token=$TOKEN"

# Expected output:
# event: connected
# data: {"message": "Connected to event stream"}
#
# : heartbeat
# (repeats every 30 seconds)
```

**Option B: Using browser console**
1. Open frontend at http://localhost:3000
2. Login with test@example.com
3. Open browser DevTools Console (F12)
4. Navigate to Dashboard
5. Look for SSE connection logs:
   ```
   [SSE] Connecting to event stream...
   [SSE] Connected to event stream
   [SSE] Connection established: ...
   ```

#### 4. Trigger Monitor Check Events

**Option A: Wait for automatic check** (recommended)
- Worker runs health checks every 30 seconds
- Once a monitor exists, wait up to 60 seconds for first check
- Watch browser console for: `[SSE] Monitor check event: ...`
- Check dashboard - stats should update automatically

**Option B: Manually trigger via broadcast** (for immediate testing)
```bash
# Simulate a monitor check event
curl -X POST http://localhost:8080/internal/broadcast \
  -H "Content-Type: application/json" \
  -d '{
    "type": "monitor_check",
    "data": {
      "monitor_id": "<monitor_id_from_step_2>",
      "monitor_name": "Google",
      "success": true,
      "status_code": 200,
      "response_time_ms": 150,
      "checked_at": "'$(date -u +"%Y-%m-%dT%H:%M:%SZ")'"
    },
    "tenant_id": 1
  }'
```

#### 5. Test Incident Events

Create an alert rule that will trigger:
```bash
# Create a "down" alert rule
curl -X POST http://localhost:8080/api/v1/alert-rules \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "name": "Monitor Down Alert",
    "trigger_type": "down",
    "threshold_value": 1,
    "enabled": true
  }'

# Manually trigger incident_created event
curl -X POST http://localhost:8080/internal/broadcast \
  -H "Content-Type: application/json" \
  -d '{
    "type": "incident_created",
    "data": {
      "monitor_id": "<monitor_id>",
      "alert_rule_id": "<alert_rule_id>",
      "rule_name": "Monitor Down Alert",
      "trigger_value": "Monitor is down",
      "status": "open"
    },
    "tenant_id": 1
  }'

# Watch browser console for: [SSE] Incident created event: ...
# Dashboard should show updated incident count
```

### Expected Behaviors

#### Connection Management
- ✅ SSE connection establishes when dashboard is loaded
- ✅ Heartbeat messages every 30 seconds keep connection alive
- ✅ Connection auto-reconnects if dropped (exponential backoff)
- ✅ Connection closes when navigating away from dashboard

#### Event Handling
- ✅ Monitor check events update monitor status in real-time
- ✅ Dashboard stats update without page refresh
- ✅ Incident events update open incidents count
- ✅ Events are tenant-isolated (only see your tenant's events)

#### UI Updates
- ✅ Domains page shows real-time monitor status
- ✅ Dashboard stats update in real-time
- ✅ No polling required - updates are push-based
- ✅ Multiple tabs/windows all receive events

### Debugging

#### Check SSE Connection Status
```javascript
// In browser console
console.log(window.eventSource?.readyState);
// 0 = CONNECTING, 1 = OPEN, 2 = CLOSED
```

#### View SSE Logs
```bash
# Backend logs
docker compose logs -f backend | grep SSE

# Worker logs  
docker compose logs -f worker | grep Event

# Frontend logs (browser console)
# Filter by "SSE" prefix
```

#### Common Issues

1. **SSE connection fails with 401**
   - Check that auth token is being passed in URL
   - Verify token hasn't expired
   - Check browser console for token value

2. **No events received**
   - Ensure monitors exist and are enabled
   - Wait for next worker cycle (30-60 seconds)
   - Check worker logs for job execution
   - Verify backend is receiving broadcast requests

3. **Events received but UI not updating**
   - Check browser console for errors
   - Verify Svelte stores are subscribed
   - Check component lifecycle (onMount/onDestroy)

## Performance Considerations

- Each SSE connection uses one persistent HTTP connection
- Event channels are buffered (capacity: 10 events)
- Dropped events if client channel is full (logged)
- Heartbeats prevent connection timeouts
- Reconnection uses exponential backoff (max 30s)

## Security

- ✅ Token-based authentication (query parameter for SSE)
- ✅ Tenant isolation enforced
- ✅ /internal/broadcast should be internal network only (no auth)
- ⚠️ Consider rate limiting for SSE connections
- ⚠️ Token in URL is visible in logs (acceptable for SSE)

## Next Steps

For production deployment:
1. Add rate limiting for SSE connections
2. Consider using Redis pub/sub for multi-instance scaling
3. Add monitoring for SSE connection count
4. Implement connection limit per tenant
5. Add SSE connection metrics to dashboard
