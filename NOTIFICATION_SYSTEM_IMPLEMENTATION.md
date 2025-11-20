# Notification Delivery System Implementation

## Overview

This document describes the notification delivery system implemented for V-Insight to send alerts when incidents occur or are resolved.

## Architecture

### Database Changes

**Migration 000007_add_notified_at**
- Adds `notified_at` TIMESTAMP column to the `incidents` table
- Creates index `idx_incidents_notified_at` for efficient queries
- Allows the system to track which incidents have been notified

### Notification Job

The `NotificationJob` runs every 30 seconds as a scheduled worker job and:

1. **Queries unnotified incidents** - Fetches incidents where `notified_at IS NULL`
2. **Retrieves alert channels** - Gets associated channels via the `alert_rule_channels` junction table
3. **Sends notifications** - Dispatches notifications based on channel type
4. **Marks as notified** - Updates `notified_at` timestamp to prevent duplicate notifications

### Supported Channel Types

#### 1. Webhook Notifier

Sends POST request with JSON payload:

```json
{
  "incident_id": "uuid-here",
  "monitor_name": "My Website",
  "monitor_url": "https://example.com",
  "status": "open",
  "message": "Monitor is down: Connection timeout",
  "timestamp": "2024-01-15T10:30:00Z"
}
```

**Configuration:**
```json
{
  "url": "https://your-webhook-endpoint.com/notify"
}
```

#### 2. Discord Notifier

Sends rich embed notifications using Discord webhook API:

**For Open Incidents:**
- 🚨 Red color (0xFF0000)
- Title: "🚨 New Incident Detected"

**For Resolved Incidents:**
- ✅ Green color (0x00FF00)
- Title: "✅ Incident Resolved"

**Configuration:**
```json
{
  "url": "https://discord.com/api/webhooks/YOUR_WEBHOOK_ID/YOUR_WEBHOOK_TOKEN"
}
```

## How It Works

### Flow Diagram

```
Incident Created (by AlertEvaluatorJob)
    ↓
Incident saved with notified_at = NULL
    ↓
NotificationJob runs (every 30s)
    ↓
Query: SELECT * FROM incidents WHERE notified_at IS NULL
    ↓
For each incident:
    ↓
    Get alert_channels via alert_rule_channels
    ↓
    For each enabled channel:
        ↓
        Send notification (webhook/Discord)
        ↓
        If successful: UPDATE incidents SET notified_at = NOW()
```

### Key Features

1. **Idempotency** - Each incident is notified exactly once
2. **Multi-channel** - Supports multiple notification channels per alert rule
3. **Extensible** - Easy to add new notifier types (email, Slack, etc.)
4. **Resilient** - Failed notifications don't block other incidents
5. **Efficient** - Index on `notified_at` for fast queries

## Testing

### Database Setup

1. Ensure migration 000007 is applied:
```bash
make migrate-up
```

2. Verify the column exists:
```sql
\d incidents
-- Should show notified_at column
```

### Manual Testing

#### 1. Create Alert Channel (Webhook)

```sql
INSERT INTO alert_channels (id, tenant_id, type, name, config, enabled)
VALUES (
  uuid_generate_v4(),
  1,
  'webhook',
  'My Webhook',
  '{"url": "https://webhook.site/your-unique-url"}'::jsonb,
  true
);
```

#### 2. Create Alert Channel (Discord)

```sql
INSERT INTO alert_channels (id, tenant_id, type, name, config, enabled)
VALUES (
  uuid_generate_v4(),
  1,
  'discord',
  'Discord Alerts',
  '{"url": "https://discord.com/api/webhooks/YOUR_WEBHOOK_ID/YOUR_TOKEN"}'::jsonb,
  true
);
```

#### 3. Link Channel to Alert Rule

```sql
INSERT INTO alert_rule_channels (alert_rule_id, alert_channel_id)
VALUES (
  'your-alert-rule-id',
  'your-alert-channel-id'
);
```

#### 4. Trigger an Incident

Wait for the AlertEvaluatorJob to detect a monitor issue, or manually create:

```sql
INSERT INTO incidents (id, monitor_id, alert_rule_id, started_at, status, trigger_value)
VALUES (
  uuid_generate_v4(),
  'your-monitor-id',
  'your-alert-rule-id',
  NOW(),
  'open',
  'Monitor is down: Connection timeout'
);
```

#### 5. Verify Notification

Check worker logs:
```bash
make logs-worker
```

Expected output:
```
[NotificationJob] Starting notification processing run
[NotificationJob] Found 1 unnotified incidents
[NotificationJob] ✓ Sent webhook notification for incident xxx via channel 'My Webhook'
[NotificationJob] Notification processing completed in 234ms - Sent: 1, Failed: 0
```

#### 6. Verify Database Update

```sql
SELECT id, status, notified_at FROM incidents WHERE id = 'your-incident-id';
-- notified_at should now have a timestamp
```

## Configuration

### Worker Schedule

The notification job is scheduled in `worker/cmd/worker/main.go`:

```go
// Schedule notification job to run every 30 seconds
if err := sched.AddJob("*/30 * * * * *", notificationJob); err != nil {
    log.Fatalf("Failed to schedule notification job: %v", err)
}
```

**Cron pattern:** `*/30 * * * * *` (every 30 seconds)

To change the frequency, modify the cron pattern:
- Every 15 seconds: `*/15 * * * * *`
- Every minute: `0 * * * * *`
- Every 2 minutes: `0 */2 * * * *`

### HTTP Client Timeout

Default timeout for webhook/Discord requests: **10 seconds**

Configured in `NotificationJob`:
```go
httpClient: &http.Client{
    Timeout: 10 * time.Second,
}
```

## Troubleshooting

### Notifications Not Sending

1. **Check job is running:**
```bash
curl http://localhost:8081/health
# Should list NotificationJob in the jobs array
```

2. **Check for unnotified incidents:**
```sql
SELECT COUNT(*) FROM incidents WHERE notified_at IS NULL;
```

3. **Check alert channels are enabled:**
```sql
SELECT id, name, type, enabled FROM alert_channels WHERE enabled = true;
```

4. **Check channel associations:**
```sql
SELECT ar.name, ac.name, ac.type
FROM alert_rule_channels arc
JOIN alert_rules ar ON arc.alert_rule_id = ar.id
JOIN alert_channels ac ON arc.alert_channel_id = ac.id;
```

### Webhook Failures

1. **Verify webhook URL is accessible:**
```bash
curl -X POST https://your-webhook-url.com \
  -H "Content-Type: application/json" \
  -d '{"test": "message"}'
```

2. **Check worker logs for errors:**
```bash
make logs-worker | grep NotificationJob
```

Common errors:
- `webhook URL not configured` - Missing or invalid URL in config
- `webhook returned non-success status: XXX` - Endpoint returned error
- `failed to send webhook request` - Network/connectivity issue

### Discord Failures

1. **Verify Discord webhook URL format:**
   - Should match: `https://discord.com/api/webhooks/{id}/{token}`

2. **Test webhook directly:**
```bash
curl -X POST "https://discord.com/api/webhooks/YOUR_ID/YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"content": "Test message"}'
```

3. **Check rate limits:**
   - Discord webhooks: 30 requests per minute
   - Worker processes max 100 incidents per run

## Extending the System

### Adding a New Notifier Type

To add support for email, Slack, or other channels:

1. **Add the channel type to database:**
```sql
-- Update the CHECK constraint in alert_channels table
ALTER TABLE alert_channels 
DROP CONSTRAINT alert_channels_type_check;

ALTER TABLE alert_channels 
ADD CONSTRAINT alert_channels_type_check 
CHECK (type IN ('webhook', 'discord', 'email', 'slack'));
```

2. **Implement the notifier in notification_job.go:**
```go
func (j *NotificationJob) sendEmailNotification(incident *IncidentNotificationData, channel *AlertChannelConfig) error {
    // Get email config
    to, ok := channel.Config["to"].(string)
    if !ok || to == "" {
        return fmt.Errorf("email 'to' address not configured")
    }
    
    // Send email logic here
    // ...
    
    return nil
}
```

3. **Add case to processIncidentNotifications:**
```go
case "email":
    err = j.sendEmailNotification(incident, channel)
```

## Security Considerations

1. **Webhook URLs** - Stored in database config JSONB, should use HTTPS
2. **Discord webhooks** - Tokens in URLs are sensitive, protect database access
3. **Timeout** - 10 second timeout prevents hanging on slow endpoints
4. **Validation** - Channel config URLs are validated before sending
5. **Error handling** - Failed notifications don't crash the worker

## Performance

### Metrics

- **Query efficiency** - Index on `notified_at` for fast filtering
- **Batch size** - Processes up to 100 incidents per run (configurable)
- **Frequency** - Runs every 30 seconds
- **Timeout** - 10 seconds per HTTP request
- **Concurrency** - Sequential processing (could be parallelized if needed)

### Scalability

For high-volume scenarios:
1. Increase batch size (change LIMIT in query)
2. Add worker parallelization
3. Consider message queue (RabbitMQ, Redis) for reliability
4. Implement retry logic with exponential backoff

## Maintenance

### Monitoring

Key metrics to track:
- Notifications sent per run
- Failed notification count
- Average processing time
- Unnotified incidents count

### Cleanup

Old incidents with notifications can be archived:
```sql
-- Archive incidents older than 90 days
DELETE FROM incidents 
WHERE notified_at < NOW() - INTERVAL '90 days'
  AND status = 'resolved';
```

## References

- Alert Evaluator Job: `worker/internal/jobs/alert_evaluator_job.go`
- Database Schema: `backend/migrations/000006_create_alerts.up.sql`
- Worker Main: `worker/cmd/worker/main.go`
- Discord Webhook API: https://discord.com/developers/docs/resources/webhook
