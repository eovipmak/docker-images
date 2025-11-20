# Alert Evaluation Implementation Summary

## Overview
Implemented automatic alert evaluation in the worker service to create and resolve incidents based on monitor check results.

## Components Implemented

### 1. Incident Repository (`backend/internal/domain/repository/incident_repository.go`)
Interface for incident data operations:
- `Create(incident)` - Create a new incident
- `GetByID(id)` - Retrieve an incident by ID
- `GetOpenIncident(monitorID, alertRuleID)` - Get open incident for monitor+rule
- `GetByMonitorID(monitorID)` - Get all incidents for a monitor
- `Update(incident)` - Update an existing incident
- `Resolve(id)` - Mark incident as resolved

### 2. PostgreSQL Incident Repository (`backend/internal/repository/postgres/incident_repository.go`)
PostgreSQL implementation with:
- Proper error handling
- Prevention of duplicate incidents
- Support for both open and resolved incidents
- Tenant isolation through monitor relationships

### 3. Alert Service (`backend/internal/domain/service/alert_service.go`)
Service layer for alert evaluation logic:
- `EvaluateCheck(check, rules)` - Evaluate a check against multiple rules
- `CreateIncident(monitorID, ruleID, value)` - Create incident (no duplicates)
- `ResolveIncident(incidentID)` - Resolve a specific incident
- `ResolveMonitorIncidents(monitorID, ruleID)` - Resolve monitor incidents
- `GetAllEnabledRules()` - Get all enabled alert rules

Supports three trigger types:
1. **down** - Monitor is unreachable/failed
2. **slow_response** - Response time exceeds threshold (in milliseconds)
3. **ssl_expiry** - SSL certificate expires within threshold days

### 4. Alert Evaluator Job (`worker/internal/jobs/alert_evaluator_job.go`)
Worker job that runs every minute:
- Fetches all enabled alert rules
- Gets latest monitor checks (last 5 minutes)
- Evaluates each check against applicable rules
- Creates incidents when alerts trigger
- Resolves incidents when monitors recover
- Prevents duplicate incidents for same monitor+rule combination
- Comprehensive logging for debugging

### 5. Worker Integration (`worker/cmd/worker/main.go`)
Registered alert evaluator job to run every minute using cron syntax: `* * * * *`

### 6. Extended Alert Rule Repository
Added `GetAllEnabled()` method to retrieve all enabled alert rules across all tenants.

## Incident Lifecycle

### Creating Incidents
1. Monitor check is performed
2. Alert evaluator job retrieves the check
3. Check is evaluated against all enabled rules
4. If rule triggers and no open incident exists → create new incident
5. If open incident already exists → skip (no duplicate)

### Resolving Incidents
1. Monitor check is performed successfully
2. Alert evaluator job evaluates the check
3. If alert no longer triggers and open incident exists → resolve incident
4. Incident status changed to 'resolved' with resolved_at timestamp

## Database Schema
The incidents table already exists with:
- `id` - UUID primary key
- `monitor_id` - Reference to monitor
- `alert_rule_id` - Reference to alert rule
- `started_at` - When incident started
- `resolved_at` - When incident was resolved (nullable)
- `status` - 'open' or 'resolved'
- `trigger_value` - Description of what triggered the alert
- `created_at` - Record creation timestamp

## Testing
Created comprehensive tests:
- `incident_repository_test.go` - Repository operations
- `alert_service_test.go` - Alert evaluation logic with mocks
- `alert_evaluator_job_test.go` - Job evaluation logic

## Key Features
✅ Automatic incident creation when alerts trigger
✅ Automatic incident resolution when monitors recover
✅ No duplicate incidents for same monitor+rule combination
✅ Support for monitor-specific and global alert rules
✅ Detailed logging for debugging
✅ Efficient queries using DISTINCT ON for latest checks
✅ Proper error handling and validation
✅ Multi-tenant support through existing relationships

## Execution Schedule
- Health Check Job: Every 30 seconds
- SSL Check Job: Every 5 minutes
- **Alert Evaluator Job: Every 1 minute** (new)

## Example Log Output
```
[AlertEvaluatorJob] Starting alert evaluation run
[AlertEvaluatorJob] Found 3 enabled alert rules
[AlertEvaluatorJob] Found 5 recent monitor checks
[AlertEvaluatorJob] ⚠ Incident created: Monitor abc123 triggered rule 'API Down Alert' - Monitor is down: Connection refused
[AlertEvaluatorJob] ✓ Incident resolved: Monitor def456 recovered from rule 'Slow Response Alert'
[AlertEvaluatorJob] Evaluation completed in 125ms - Created: 1 incidents, Resolved: 1 incidents
```

## Future Enhancements
- Send notifications through alert channels when incidents are created/resolved
- Add incident acknowledgment support
- Implement incident escalation policies
- Add incident notes/comments
- Support for complex alert conditions (AND/OR logic)
