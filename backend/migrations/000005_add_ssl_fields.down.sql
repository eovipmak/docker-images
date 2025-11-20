-- Remove SSL checking fields from monitors table
DROP INDEX IF EXISTS idx_monitors_check_ssl;

ALTER TABLE monitors
DROP COLUMN IF EXISTS check_ssl,
DROP COLUMN IF EXISTS ssl_alert_days;
