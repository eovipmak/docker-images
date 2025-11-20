-- Add SSL checking fields to monitors table
ALTER TABLE monitors
ADD COLUMN check_ssl BOOLEAN NOT NULL DEFAULT true,
ADD COLUMN ssl_alert_days INTEGER NOT NULL DEFAULT 30;

-- Create index on check_ssl for efficient filtering
CREATE INDEX IF NOT EXISTS idx_monitors_check_ssl ON monitors(check_ssl);
