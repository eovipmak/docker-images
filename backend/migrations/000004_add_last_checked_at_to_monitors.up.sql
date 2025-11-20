-- Add last_checked_at column to monitors table
ALTER TABLE monitors ADD COLUMN last_checked_at TIMESTAMP;

-- Create index for efficient queries on monitors that need checking
CREATE INDEX IF NOT EXISTS idx_monitors_enabled_last_checked ON monitors(enabled, last_checked_at);
