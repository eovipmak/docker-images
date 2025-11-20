-- Enable UUID extension if not already enabled
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Create alert_rules table
CREATE TABLE IF NOT EXISTS alert_rules (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id INTEGER NOT NULL,
    monitor_id UUID,
    name VARCHAR(255) NOT NULL,
    trigger_type VARCHAR(50) NOT NULL CHECK (trigger_type IN ('down', 'ssl_expiry', 'slow_response')),
    threshold_value INTEGER NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_alert_rules_tenant FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE,
    CONSTRAINT fk_alert_rules_monitor FOREIGN KEY (monitor_id) REFERENCES monitors(id) ON DELETE CASCADE
);

-- Create index on tenant_id for faster tenant-scoped queries
CREATE INDEX IF NOT EXISTS idx_alert_rules_tenant_id ON alert_rules(tenant_id);

-- Create index on monitor_id for faster monitor-specific queries
CREATE INDEX IF NOT EXISTS idx_alert_rules_monitor_id ON alert_rules(monitor_id);

-- Create index on enabled for filtering active rules
CREATE INDEX IF NOT EXISTS idx_alert_rules_enabled ON alert_rules(enabled);

-- Create alert_channels table
CREATE TABLE IF NOT EXISTS alert_channels (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id INTEGER NOT NULL,
    type VARCHAR(50) NOT NULL CHECK (type IN ('webhook', 'discord', 'email')),
    name VARCHAR(255) NOT NULL,
    config JSONB NOT NULL DEFAULT '{}'::jsonb,
    enabled BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_alert_channels_tenant FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE
);

-- Create index on tenant_id for faster tenant-scoped queries
CREATE INDEX IF NOT EXISTS idx_alert_channels_tenant_id ON alert_channels(tenant_id);

-- Create index on enabled for filtering active channels
CREATE INDEX IF NOT EXISTS idx_alert_channels_enabled ON alert_channels(enabled);

-- Create alert_rule_channels junction table (many-to-many)
CREATE TABLE IF NOT EXISTS alert_rule_channels (
    alert_rule_id UUID NOT NULL,
    alert_channel_id UUID NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (alert_rule_id, alert_channel_id),
    CONSTRAINT fk_arc_alert_rule FOREIGN KEY (alert_rule_id) REFERENCES alert_rules(id) ON DELETE CASCADE,
    CONSTRAINT fk_arc_alert_channel FOREIGN KEY (alert_channel_id) REFERENCES alert_channels(id) ON DELETE CASCADE
);

-- Create index on alert_channel_id for reverse lookups
CREATE INDEX IF NOT EXISTS idx_alert_rule_channels_channel_id ON alert_rule_channels(alert_channel_id);

-- Create incidents table
CREATE TABLE IF NOT EXISTS incidents (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    monitor_id UUID NOT NULL,
    alert_rule_id UUID NOT NULL,
    started_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    resolved_at TIMESTAMP,
    status VARCHAR(50) NOT NULL CHECK (status IN ('open', 'resolved')) DEFAULT 'open',
    trigger_value TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_incidents_monitor FOREIGN KEY (monitor_id) REFERENCES monitors(id) ON DELETE CASCADE,
    CONSTRAINT fk_incidents_alert_rule FOREIGN KEY (alert_rule_id) REFERENCES alert_rules(id) ON DELETE CASCADE
);

-- Create index on monitor_id for faster monitor-specific queries
CREATE INDEX IF NOT EXISTS idx_incidents_monitor_id ON incidents(monitor_id);

-- Create index on alert_rule_id for faster rule-specific queries
CREATE INDEX IF NOT EXISTS idx_incidents_alert_rule_id ON incidents(alert_rule_id);

-- Create index on status for filtering open/resolved incidents
CREATE INDEX IF NOT EXISTS idx_incidents_status ON incidents(status);

-- Create composite index for efficient recent incidents queries
CREATE INDEX IF NOT EXISTS idx_incidents_monitor_started ON incidents(monitor_id, started_at DESC);

-- Create trigger to automatically update updated_at on alert_rules
CREATE TRIGGER update_alert_rules_updated_at BEFORE UPDATE ON alert_rules
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Create trigger to automatically update updated_at on alert_channels
CREATE TRIGGER update_alert_channels_updated_at BEFORE UPDATE ON alert_channels
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
