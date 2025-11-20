-- Drop triggers
DROP TRIGGER IF EXISTS update_monitors_updated_at ON monitors;

-- Drop indexes
DROP INDEX IF EXISTS idx_monitor_checks_monitor_checked;
DROP INDEX IF EXISTS idx_monitor_checks_checked_at;
DROP INDEX IF EXISTS idx_monitor_checks_monitor_id;
DROP INDEX IF EXISTS idx_monitors_enabled;
DROP INDEX IF EXISTS idx_monitors_tenant_id;

-- Drop tables (cascade will remove foreign key constraints)
DROP TABLE IF EXISTS monitor_checks CASCADE;
DROP TABLE IF EXISTS monitors CASCADE;
