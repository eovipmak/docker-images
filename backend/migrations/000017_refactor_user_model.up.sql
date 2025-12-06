-- Add role to users table
ALTER TABLE users ADD COLUMN IF NOT EXISTS role VARCHAR(50) NOT NULL DEFAULT 'user';

-- Add user_id to monitors
ALTER TABLE monitors ADD COLUMN IF NOT EXISTS user_id INT;
ALTER TABLE monitors ADD CONSTRAINT fk_monitors_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

-- Add user_id to alert_rules
ALTER TABLE alert_rules ADD COLUMN IF NOT EXISTS user_id INT;
ALTER TABLE alert_rules ADD CONSTRAINT fk_alert_rules_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

-- Add user_id to alert_channels
ALTER TABLE alert_channels ADD COLUMN IF NOT EXISTS user_id INT;
ALTER TABLE alert_channels ADD CONSTRAINT fk_alert_channels_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

-- Add user_id to incidents
ALTER TABLE incidents ADD COLUMN IF NOT EXISTS user_id INT;
ALTER TABLE incidents ADD CONSTRAINT fk_incidents_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

-- Add user_id to status_pages
ALTER TABLE status_pages ADD COLUMN IF NOT EXISTS user_id INT;
ALTER TABLE status_pages ADD CONSTRAINT fk_status_pages_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

-- Add user_id to maintenance_windows
ALTER TABLE maintenance_windows ADD COLUMN IF NOT EXISTS user_id INT;
ALTER TABLE maintenance_windows ADD CONSTRAINT fk_maintenance_windows_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

-- Data Migration: Migrate data from tenants to users (using owner_id from tenants)

-- Update monitors
UPDATE monitors m
SET user_id = t.owner_id
FROM tenants t
WHERE m.tenant_id = t.id;

-- Update alert_rules
UPDATE alert_rules ar
SET user_id = t.owner_id
FROM tenants t
WHERE ar.tenant_id = t.id;

-- Update alert_channels
UPDATE alert_channels ac
SET user_id = t.owner_id
FROM tenants t
WHERE ac.tenant_id = t.id;

-- Update incidents (JOIN via monitors because incidents don't have tenant_id)
UPDATE incidents i
SET user_id = t.owner_id
FROM monitors m
JOIN tenants t ON m.tenant_id = t.id
WHERE i.monitor_id = m.id;

-- Update status_pages
UPDATE status_pages sp
SET user_id = t.owner_id
FROM tenants t
WHERE sp.tenant_id = t.id;

-- Update maintenance_windows
UPDATE maintenance_windows mw
SET user_id = t.owner_id
FROM tenants t
WHERE mw.tenant_id = t.id;

-- Drop tenant_id columns and tenant-related tables

ALTER TABLE monitors DROP COLUMN tenant_id;
ALTER TABLE alert_rules DROP COLUMN tenant_id;
ALTER TABLE alert_channels DROP COLUMN tenant_id;
-- Incidents does not have tenant_id
ALTER TABLE status_pages DROP COLUMN tenant_id;
ALTER TABLE maintenance_windows DROP COLUMN tenant_id;

DROP TABLE IF EXISTS tenant_users;
DROP TABLE IF EXISTS tenants;
