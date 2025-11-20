-- Enable UUID extension if not already enabled
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Create monitors table
CREATE TABLE IF NOT EXISTS monitors (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id INTEGER NOT NULL,
    name VARCHAR(255) NOT NULL,
    url VARCHAR(512) NOT NULL,
    check_interval INTEGER NOT NULL DEFAULT 300,  -- seconds
    timeout INTEGER NOT NULL DEFAULT 30,  -- seconds
    enabled BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_monitors_tenant FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE
);

-- Create index on tenant_id for faster tenant-scoped queries
CREATE INDEX IF NOT EXISTS idx_monitors_tenant_id ON monitors(tenant_id);

-- Create index on enabled for filtering active monitors
CREATE INDEX IF NOT EXISTS idx_monitors_enabled ON monitors(enabled);

-- Create monitor_checks table
CREATE TABLE IF NOT EXISTS monitor_checks (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    monitor_id UUID NOT NULL,
    checked_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    status_code INTEGER,
    response_time_ms INTEGER,
    ssl_valid BOOLEAN,
    ssl_expires_at TIMESTAMP,
    error_message TEXT,
    success BOOLEAN NOT NULL DEFAULT false,
    CONSTRAINT fk_monitor_checks_monitor FOREIGN KEY (monitor_id) REFERENCES monitors(id) ON DELETE CASCADE
);

-- Create index on monitor_id for faster monitor-specific queries
CREATE INDEX IF NOT EXISTS idx_monitor_checks_monitor_id ON monitor_checks(monitor_id);

-- Create index on checked_at for time-based queries
CREATE INDEX IF NOT EXISTS idx_monitor_checks_checked_at ON monitor_checks(checked_at);

-- Create composite index for efficient monitor history queries
CREATE INDEX IF NOT EXISTS idx_monitor_checks_monitor_checked ON monitor_checks(monitor_id, checked_at DESC);

-- Create trigger to automatically update updated_at on monitors
CREATE TRIGGER update_monitors_updated_at BEFORE UPDATE ON monitors
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
