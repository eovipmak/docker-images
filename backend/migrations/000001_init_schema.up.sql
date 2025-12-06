-- Enable UUID extension if not already enabled
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Create users table
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    role VARCHAR(50) NOT NULL DEFAULT 'user',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create index on email for faster lookups
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);

-- Create function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create triggers to automatically update updated_at
CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Create monitors table
CREATE TABLE IF NOT EXISTS monitors (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id INTEGER NOT NULL,
    name VARCHAR(255) NOT NULL,
    url VARCHAR(512) NOT NULL,
    type VARCHAR(10) DEFAULT 'http',
    keyword TEXT NOT NULL DEFAULT '',
    check_interval INTEGER NOT NULL DEFAULT 300,  -- seconds
    timeout INTEGER NOT NULL DEFAULT 30,  -- seconds
    enabled BOOLEAN NOT NULL DEFAULT true,
    check_ssl BOOLEAN NOT NULL DEFAULT true,
    ssl_alert_days INTEGER NOT NULL DEFAULT 30,
    tags TEXT[] DEFAULT '{}',
    expected_status_codes INTEGER[] DEFAULT '{200}',
    last_checked_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_monitors_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Create index on user_id for faster user-scoped queries
CREATE INDEX IF NOT EXISTS idx_monitors_user_id ON monitors(user_id);

-- Create index on enabled for filtering active monitors
CREATE INDEX IF NOT EXISTS idx_monitors_enabled ON monitors(enabled);

-- Create index on check_ssl for efficient filtering
CREATE INDEX IF NOT EXISTS idx_monitors_check_ssl ON monitors(check_ssl);

-- Create index for efficient queries on monitors that need checking
CREATE INDEX IF NOT EXISTS idx_monitors_enabled_last_checked ON monitors(enabled, last_checked_at);

-- Create index for tags (GIN index for array searching)
CREATE INDEX IF NOT EXISTS idx_monitors_tags ON monitors USING GIN (tags);

-- Create trigger to automatically update updated_at on monitors
CREATE TRIGGER update_monitors_updated_at BEFORE UPDATE ON monitors
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

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

-- Create alert_rules table
CREATE TABLE IF NOT EXISTS alert_rules (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id INTEGER NOT NULL,
    monitor_id UUID,
    name VARCHAR(255) NOT NULL,
    trigger_type VARCHAR(50) NOT NULL CHECK (trigger_type IN ('down', 'ssl_expiry', 'slow_response')),
    threshold_value INTEGER NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_alert_rules_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT fk_alert_rules_monitor FOREIGN KEY (monitor_id) REFERENCES monitors(id) ON DELETE CASCADE
);

-- Create index on user_id for faster user-scoped queries
CREATE INDEX IF NOT EXISTS idx_alert_rules_user_id ON alert_rules(user_id);

-- Create index on monitor_id for faster monitor-specific queries
CREATE INDEX IF NOT EXISTS idx_alert_rules_monitor_id ON alert_rules(monitor_id);

-- Create index on enabled for filtering active rules
CREATE INDEX IF NOT EXISTS idx_alert_rules_enabled ON alert_rules(enabled);

-- Create trigger to automatically update updated_at on alert_rules
CREATE TRIGGER update_alert_rules_updated_at BEFORE UPDATE ON alert_rules
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Create alert_channels table
CREATE TABLE IF NOT EXISTS alert_channels (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id INTEGER NOT NULL,
    type VARCHAR(50) NOT NULL CHECK (type IN ('webhook', 'discord', 'email')),
    name VARCHAR(255) NOT NULL,
    config JSONB NOT NULL DEFAULT '{}'::jsonb,
    enabled BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_alert_channels_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Create index on user_id for faster user-scoped queries
CREATE INDEX IF NOT EXISTS idx_alert_channels_user_id ON alert_channels(user_id);

-- Create index on enabled for filtering active channels
CREATE INDEX IF NOT EXISTS idx_alert_channels_enabled ON alert_channels(enabled);

-- Create trigger to automatically update updated_at on alert_channels
CREATE TRIGGER update_alert_channels_updated_at BEFORE UPDATE ON alert_channels
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

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
    user_id INTEGER NOT NULL,
    started_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    resolved_at TIMESTAMP,
    status VARCHAR(50) NOT NULL CHECK (status IN ('open', 'resolved')) DEFAULT 'open',
    trigger_value TEXT,
    notified_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_incidents_monitor FOREIGN KEY (monitor_id) REFERENCES monitors(id) ON DELETE CASCADE,
    CONSTRAINT fk_incidents_alert_rule FOREIGN KEY (alert_rule_id) REFERENCES alert_rules(id) ON DELETE CASCADE,
    CONSTRAINT fk_incidents_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Create index on monitor_id for faster monitor-specific queries
CREATE INDEX IF NOT EXISTS idx_incidents_monitor_id ON incidents(monitor_id);

-- Create index on alert_rule_id for faster rule-specific queries
CREATE INDEX IF NOT EXISTS idx_incidents_alert_rule_id ON incidents(alert_rule_id);

-- Create index on status for filtering open/resolved incidents
CREATE INDEX IF NOT EXISTS idx_incidents_status ON incidents(status);

-- Create index on notified_at for efficient querying of unnotified incidents
CREATE INDEX IF NOT EXISTS idx_incidents_notified_at ON incidents(notified_at);

-- Create composite index for efficient recent incidents queries
CREATE INDEX IF NOT EXISTS idx_incidents_monitor_started ON incidents(monitor_id, started_at DESC);

-- Create status_pages table
CREATE TABLE IF NOT EXISTS status_pages (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id INTEGER NOT NULL,
    slug VARCHAR(255) NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL,
    public_enabled BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_status_pages_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Create index on user_id for faster user-scoped queries
CREATE INDEX IF NOT EXISTS idx_status_pages_user_id ON status_pages(user_id);

-- Create index on slug for fast lookups
CREATE INDEX IF NOT EXISTS idx_status_pages_slug ON status_pages(slug);

-- Create trigger to automatically update updated_at on status_pages
CREATE TRIGGER update_status_pages_updated_at BEFORE UPDATE ON status_pages
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Create status_page_monitors linking table
CREATE TABLE IF NOT EXISTS status_page_monitors (
    status_page_id UUID NOT NULL,
    monitor_id UUID NOT NULL,
    PRIMARY KEY (status_page_id, monitor_id),
    CONSTRAINT fk_status_page_monitors_status_page FOREIGN KEY (status_page_id) REFERENCES status_pages(id) ON DELETE CASCADE,
    CONSTRAINT fk_status_page_monitors_monitor FOREIGN KEY (monitor_id) REFERENCES monitors(id) ON DELETE CASCADE
);

-- Create index on status_page_id for faster queries
CREATE INDEX IF NOT EXISTS idx_status_page_monitors_status_page_id ON status_page_monitors(status_page_id);

-- Create index on monitor_id for faster queries
CREATE INDEX IF NOT EXISTS idx_status_page_monitors_monitor_id ON status_page_monitors(monitor_id);

-- Create maintenance_windows table
CREATE TABLE IF NOT EXISTS maintenance_windows (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    start_time TIMESTAMP WITH TIME ZONE NOT NULL,
    end_time TIMESTAMP WITH TIME ZONE NOT NULL,
    repeat_interval INTEGER DEFAULT 0, -- in seconds, 0 = one-time
    monitor_ids UUID[] DEFAULT '{}', -- which monitors does this apply to (empty = all)
    tags TEXT[] DEFAULT '{}', -- filter by tags
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create index for user_id lookups
CREATE INDEX idx_maintenance_windows_user_id ON maintenance_windows(user_id);

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

-- Seed demo user (optional, can be removed if not needed)
INSERT INTO users (email, password_hash, role)
VALUES ('test@gmail.com', '$2a$10$3XjX/./././././././././', 'admin')
ON CONFLICT (email) DO NOTHING;
