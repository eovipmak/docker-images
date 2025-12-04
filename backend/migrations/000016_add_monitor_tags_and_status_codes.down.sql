-- Drop index
DROP INDEX IF EXISTS idx_monitors_tags;

-- Remove columns
ALTER TABLE monitors DROP COLUMN IF EXISTS expected_status_codes;
ALTER TABLE monitors DROP COLUMN IF EXISTS tags;
