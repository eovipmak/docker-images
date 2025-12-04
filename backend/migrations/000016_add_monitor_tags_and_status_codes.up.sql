-- Add tags and expected_status_codes columns to monitors table
ALTER TABLE monitors ADD COLUMN IF NOT EXISTS tags TEXT[] DEFAULT '{}';
ALTER TABLE monitors ADD COLUMN IF NOT EXISTS expected_status_codes INTEGER[] DEFAULT '{200}';

-- Create index for tags (GIN index for array searching)
CREATE INDEX IF NOT EXISTS idx_monitors_tags ON monitors USING GIN (tags);
