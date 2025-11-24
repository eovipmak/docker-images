-- Drop performance optimization indexes

DROP INDEX IF EXISTS idx_incidents_open_lookup;
DROP INDEX IF EXISTS idx_monitors_check_schedule;
DROP INDEX IF EXISTS idx_monitors_tenant_id_id;
