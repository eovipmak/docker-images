-- Drop metrics optimization indexes

DROP INDEX IF EXISTS idx_monitor_checks_metrics_uptime;
DROP INDEX IF EXISTS idx_monitor_checks_metrics_response_time;
DROP INDEX IF EXISTS idx_monitor_checks_status_code;
DROP INDEX IF EXISTS idx_monitor_checks_tenant_metrics;
