-- Drop trigger first
DROP TRIGGER IF EXISTS trigger_maintenance_windows_updated_at ON maintenance_windows;

-- Drop function
DROP FUNCTION IF EXISTS update_maintenance_windows_updated_at();

-- Drop indexes
DROP INDEX IF EXISTS idx_maintenance_windows_time_range;
DROP INDEX IF EXISTS idx_maintenance_windows_tenant_id;

-- Drop table
DROP TABLE IF EXISTS maintenance_windows;
