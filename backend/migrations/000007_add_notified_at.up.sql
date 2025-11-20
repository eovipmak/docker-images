-- Add notified_at column to incidents table to track when notifications were sent
ALTER TABLE incidents
ADD COLUMN notified_at TIMESTAMP;

-- Create index on notified_at for efficient querying of unnotified incidents
CREATE INDEX IF NOT EXISTS idx_incidents_notified_at ON incidents(notified_at);
