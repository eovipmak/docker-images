-- Drop index on notified_at
DROP INDEX IF EXISTS idx_incidents_notified_at;

-- Remove notified_at column from incidents table
ALTER TABLE incidents
DROP COLUMN notified_at;
