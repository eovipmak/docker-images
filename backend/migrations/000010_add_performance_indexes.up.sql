-- Add performance optimization indexes

-- Composite index for efficient open incident lookups
-- Optimizes GetOpenIncident query which filters by monitor_id, alert_rule_id, and status
CREATE INDEX IF NOT EXISTS idx_incidents_open_lookup
ON incidents(monitor_id, alert_rule_id, status)
WHERE status = 'open';

-- Composite index for monitors needing check
-- Optimizes GetMonitorsNeedingCheck query which filters by enabled and orders by last_checked_at
CREATE INDEX IF NOT EXISTS idx_monitors_check_schedule
ON monitors(enabled, last_checked_at ASC NULLS FIRST)
WHERE enabled = true;

-- Index for tenant-scoped incident queries via monitor join
-- Optimizes List query which joins incidents with monitors and filters by tenant_id
CREATE INDEX IF NOT EXISTS idx_monitors_tenant_id_id
ON monitors(tenant_id, id);
