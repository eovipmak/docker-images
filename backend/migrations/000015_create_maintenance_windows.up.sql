-- Create maintenance_windows table
CREATE TABLE IF NOT EXISTS maintenance_windows (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id INTEGER NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    start_time TIMESTAMP WITH TIME ZONE NOT NULL,
    end_time TIMESTAMP WITH TIME ZONE NOT NULL,
    repeat_interval INTEGER DEFAULT 0, -- in seconds, 0 = one-time
    monitor_ids UUID[] DEFAULT '{}', -- which monitors does this apply to (empty = all)
    tags TEXT[] DEFAULT '{}', -- filter by tags
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create index for tenant_id lookups
CREATE INDEX idx_maintenance_windows_tenant_id ON maintenance_windows(tenant_id);

-- Create index for active window queries (start_time, end_time)
CREATE INDEX idx_maintenance_windows_time_range ON maintenance_windows(start_time, end_time);

-- Create trigger to auto-update updated_at
CREATE OR REPLACE FUNCTION update_maintenance_windows_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_maintenance_windows_updated_at
    BEFORE UPDATE ON maintenance_windows
    FOR EACH ROW
    EXECUTE FUNCTION update_maintenance_windows_updated_at();
