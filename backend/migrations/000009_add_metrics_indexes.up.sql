-- Add composite indexes for metrics queries optimization

-- Composite index for uptime and status code distribution queries
-- Optimizes queries that filter by monitor_id, checked_at, and success
CREATE INDEX IF NOT EXISTS idx_monitor_checks_metrics_uptime 
ON monitor_checks(monitor_id, checked_at DESC, success);

-- Composite index for response time queries
-- Optimizes queries that filter by monitor_id, checked_at, and need response_time_ms
CREATE INDEX IF NOT EXISTS idx_monitor_checks_metrics_response_time 
ON monitor_checks(monitor_id, checked_at DESC, response_time_ms) 
WHERE response_time_ms IS NOT NULL;

-- Index for status code distribution queries
CREATE INDEX IF NOT EXISTS idx_monitor_checks_status_code 
ON monitor_checks(monitor_id, checked_at DESC, status_code) 
WHERE status_code IS NOT NULL;

-- Composite index for global tenant-wide metrics
-- Optimizes dashboard queries that join monitors with checks
CREATE INDEX IF NOT EXISTS idx_monitor_checks_tenant_metrics 
ON monitor_checks(monitor_id, checked_at DESC) 
INCLUDE (response_time_ms, success);
