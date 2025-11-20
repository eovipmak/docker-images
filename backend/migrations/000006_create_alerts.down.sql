-- Drop triggers
DROP TRIGGER IF EXISTS update_alert_channels_updated_at ON alert_channels;
DROP TRIGGER IF EXISTS update_alert_rules_updated_at ON alert_rules;

-- Drop indexes
DROP INDEX IF EXISTS idx_incidents_monitor_started;
DROP INDEX IF EXISTS idx_incidents_status;
DROP INDEX IF EXISTS idx_incidents_alert_rule_id;
DROP INDEX IF EXISTS idx_incidents_monitor_id;
DROP INDEX IF EXISTS idx_alert_rule_channels_channel_id;
DROP INDEX IF EXISTS idx_alert_channels_enabled;
DROP INDEX IF EXISTS idx_alert_channels_tenant_id;
DROP INDEX IF EXISTS idx_alert_rules_enabled;
DROP INDEX IF EXISTS idx_alert_rules_monitor_id;
DROP INDEX IF EXISTS idx_alert_rules_tenant_id;

-- Drop tables (cascade will remove foreign key constraints)
DROP TABLE IF EXISTS incidents CASCADE;
DROP TABLE IF EXISTS alert_rule_channels CASCADE;
DROP TABLE IF EXISTS alert_channels CASCADE;
DROP TABLE IF EXISTS alert_rules CASCADE;
