# Alert System - Complete Implementation Guide

## Overview

This implementation adds a comprehensive alert system to V-Insight, allowing users to:
- Create alert rules that trigger based on monitor conditions
- Configure multiple notification channels (webhook, discord, email)
- Track incidents when alerts are triggered
- Manage alerts with full CRUD operations

## Files Created/Modified

### Database Migrations
- `backend/migrations/000006_create_alerts.up.sql` - Creates alert_rules, alert_channels, alert_rule_channels, and incidents tables
- `backend/migrations/000006_create_alerts.down.sql` - Rollback migration

### Domain Entities
- `backend/internal/domain/entities/alert_rule.go` - AlertRule entity with AlertRuleWithChannels
- `backend/internal/domain/entities/alert_channel.go` - AlertChannel entity with custom JSONB config type
- `backend/internal/domain/entities/incident.go` - Incident entity for tracking triggered alerts

### Repository Interfaces
- `backend/internal/domain/repository/alert_rule_repository.go` - Interface for alert rule operations
- `backend/internal/domain/repository/alert_channel_repository.go` - Interface for alert channel operations

### Repository Implementations
- `backend/internal/repository/postgres/alert_rule_repository.go` - PostgreSQL implementation with channel management
- `backend/internal/repository/postgres/alert_channel_repository.go` - PostgreSQL implementation
- `backend/internal/repository/postgres/alert_rule_repository_test.go` - Unit tests (9 test cases)
- `backend/internal/repository/postgres/alert_channel_repository_test.go` - Unit tests (5 test cases)

### API Handlers
- `backend/internal/api/handlers/alert_rule_handler.go` - CRUD handlers for alert rules
- `backend/internal/api/handlers/alert_channel_handler.go` - CRUD handlers for alert channels

### Main Application
- `backend/cmd/api/main.go` - Updated to wire up new repositories, handlers, and routes

## API Endpoints

All endpoints require authentication and tenant context.

### Alert Rules
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/alert-rules` | Create alert rule |
| GET | `/api/v1/alert-rules` | List all alert rules |
| GET | `/api/v1/alert-rules/:id` | Get alert rule with channels |
| PUT | `/api/v1/alert-rules/:id` | Update alert rule |
| DELETE | `/api/v1/alert-rules/:id` | Delete alert rule |

### Alert Channels
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/alert-channels` | Create alert channel |
| GET | `/api/v1/alert-channels` | List all alert channels |
| GET | `/api/v1/alert-channels/:id` | Get alert channel |
| PUT | `/api/v1/alert-channels/:id` | Update alert channel |
| DELETE | `/api/v1/alert-channels/:id` | Delete alert channel |

## Usage Examples

### 1. Create a Webhook Channel
```bash
curl -X POST http://localhost:8080/api/v1/alert-channels \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "type": "webhook",
    "name": "Slack Notifications",
    "config": {
      "url": "https://hooks.slack.com/services/xxx/yyy/zzz",
      "method": "POST"
    },
    "enabled": true
  }'
```

### 2. Create a Discord Channel
```bash
curl -X POST http://localhost:8080/api/v1/alert-channels \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "type": "discord",
    "name": "Discord Alerts",
    "config": {
      "webhook_url": "https://discord.com/api/webhooks/xxx/yyy"
    },
    "enabled": true
  }'
```

### 3. Create an Alert Rule for Monitor Downtime
```bash
curl -X POST http://localhost:8080/api/v1/alert-rules \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "monitor_id": "uuid-of-your-monitor",
    "name": "Website Down Alert",
    "trigger_type": "down",
    "threshold_value": 3,
    "enabled": true,
    "channel_ids": ["channel-uuid-1", "channel-uuid-2"]
  }'
```

### 4. Create an SSL Expiry Alert
```bash
curl -X POST http://localhost:8080/api/v1/alert-rules \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "monitor_id": "uuid-of-your-monitor",
    "name": "SSL Expiring Soon",
    "trigger_type": "ssl_expiry",
    "threshold_value": 7,
    "enabled": true,
    "channel_ids": ["channel-uuid-1"]
  }'
```

### 5. List All Alert Rules
```bash
curl http://localhost:8080/api/v1/alert-rules \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

Response:
```json
[
  {
    "id": "rule-uuid-1",
    "tenant_id": 1,
    "monitor_id": "monitor-uuid",
    "name": "Website Down Alert",
    "trigger_type": "down",
    "threshold_value": 3,
    "enabled": true,
    "channel_ids": ["channel-uuid-1", "channel-uuid-2"],
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  }
]
```

## Trigger Types

### down
Triggers when a monitor fails consecutive health checks.
- **threshold_value**: Number of consecutive failed checks (e.g., 3 means alert after 3 failures)
- Use case: Detect website/service downtime

### ssl_expiry
Triggers when SSL certificate is expiring soon.
- **threshold_value**: Number of days before expiry (e.g., 7 means alert 7 days before expiration)
- Use case: Prevent SSL certificate expiration

### slow_response
Triggers when response time exceeds threshold.
- **threshold_value**: Response time in milliseconds (e.g., 5000 means alert when response > 5 seconds)
- Use case: Detect performance degradation

## Channel Types

### webhook
Generic webhook integration for custom services.
```json
{
  "type": "webhook",
  "config": {
    "url": "https://your-service.com/webhook",
    "method": "POST",
    "headers": {
      "Authorization": "Bearer token",
      "Content-Type": "application/json"
    }
  }
}
```

### discord
Discord webhook integration.
```json
{
  "type": "discord",
  "config": {
    "webhook_url": "https://discord.com/api/webhooks/xxx/yyy",
    "username": "V-Insight Bot",
    "avatar_url": "https://example.com/avatar.png"
  }
}
```

### email
Email notification integration.
```json
{
  "type": "email",
  "config": {
    "to": ["admin@example.com", "team@example.com"],
    "subject_prefix": "[V-Insight]",
    "from": "alerts@v-insight.com"
  }
}
```

## Database Schema Details

### alert_rules
- **Constraints**: Check constraint on trigger_type, foreign keys with CASCADE delete
- **Indexes**: tenant_id, monitor_id, enabled
- **Triggers**: Auto-update updated_at timestamp

### alert_channels
- **Constraints**: Check constraint on type, foreign key with CASCADE delete
- **Indexes**: tenant_id, enabled
- **Triggers**: Auto-update updated_at timestamp
- **JSONB Config**: Flexible configuration storage for different channel types

### alert_rule_channels
- **Composite PK**: (alert_rule_id, alert_channel_id)
- **Indexes**: alert_channel_id for reverse lookups
- **ON DELETE CASCADE**: Auto-cleanup when rules or channels are deleted

### incidents
- **Status Values**: 'open', 'resolved'
- **Indexes**: monitor_id, alert_rule_id, status, composite (monitor_id, started_at)
- **Purpose**: Track all triggered alerts for history and analytics

## Security Features

1. **Multi-tenant Isolation**: All queries filtered by tenant_id from JWT token
2. **Authorization Checks**: Verify resource ownership before access
3. **Channel Validation**: Ensure channels belong to tenant when attaching to rules
4. **Input Validation**: Gin binding tags validate request payloads
5. **Error Handling**: Generic error messages to prevent information leakage

## Testing

### Run Unit Tests
```bash
cd backend
go test ./internal/repository/postgres/...
```

### Test Coverage
- AlertRuleRepository: 9 test cases covering CRUD and channel management
- AlertChannelRepository: 5 test cases covering CRUD operations
- All tests use sqlmock for database mocking

## Future Enhancements

Ready for implementation:

1. **Incident Management API**
   - GET /api/v1/incidents - List incidents
   - GET /api/v1/incidents/:id - Get incident details
   - POST /api/v1/incidents/:id/resolve - Mark as resolved

2. **Alert Delivery Service** (Worker)
   - Poll for triggered conditions
   - Send notifications via configured channels
   - Track delivery status
   - Implement retry logic

3. **Alert Templates**
   - Customizable message templates per channel type
   - Variable substitution (monitor name, status, etc.)

4. **Escalation Policies**
   - Progressive alert escalation
   - Time-based escalation rules
   - Multiple notification tiers

5. **Alert History**
   - Track all sent notifications
   - Delivery status and timestamps
   - Failure reasons

## Migration

The database migration will run automatically on backend startup. To manually run:

```bash
# Using golang-migrate CLI
migrate -path ./backend/migrations -database "postgres://user:pass@localhost:5432/dbname?sslmode=disable" up
```

## Notes

- All UUIDs are auto-generated using PostgreSQL's uuid_generate_v4()
- Timestamps are managed by PostgreSQL (created_at, updated_at)
- Empty arrays returned for list endpoints (not null)
- Soft deletes are NOT implemented - resources are permanently deleted
- Channel configs are stored as JSONB for flexibility
- Monitor_id in alert_rules is nullable for global rules (future feature)

## Support

For issues or questions:
1. Check migration logs for database errors
2. Review API handler logs for request validation errors
3. Verify JWT token is valid and contains tenant_id
4. Ensure monitor_id exists before creating rules
5. Validate channel ownership before attaching to rules
