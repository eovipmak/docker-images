-- Drop index
DROP INDEX IF EXISTS idx_monitors_enabled_last_checked;

-- Remove last_checked_at column from monitors table
ALTER TABLE monitors DROP COLUMN IF EXISTS last_checked_at;
