-- Revert changes is complex because we destroyed data (tenants table).
-- We will try to restore structure but data loss is expected for tenant specific info if not backed up.
-- For this exercise, we will just recreate the tables and columns.

-- Create tenants table
CREATE TABLE IF NOT EXISTS tenants (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(255) NOT NULL UNIQUE,
    owner_id INT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create tenant_users table
CREATE TABLE IF NOT EXISTS tenant_users (
    tenant_id INT NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role VARCHAR(50) NOT NULL DEFAULT 'member',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    PRIMARY KEY (tenant_id, user_id)
);

-- Add tenant_id back to tables
ALTER TABLE monitors ADD COLUMN IF NOT EXISTS tenant_id INT REFERENCES tenants(id) ON DELETE CASCADE;
ALTER TABLE alert_rules ADD COLUMN IF NOT EXISTS tenant_id INT REFERENCES tenants(id) ON DELETE CASCADE;
ALTER TABLE alert_channels ADD COLUMN IF NOT EXISTS tenant_id INT REFERENCES tenants(id) ON DELETE CASCADE;
ALTER TABLE status_pages ADD COLUMN IF NOT EXISTS tenant_id INT REFERENCES tenants(id) ON DELETE CASCADE;
ALTER TABLE maintenance_windows ADD COLUMN IF NOT EXISTS tenant_id INT REFERENCES tenants(id) ON DELETE CASCADE;

-- Drop user_id columns
ALTER TABLE monitors DROP COLUMN IF EXISTS user_id;
ALTER TABLE alert_rules DROP COLUMN IF EXISTS user_id;
ALTER TABLE alert_channels DROP COLUMN IF EXISTS user_id;
ALTER TABLE incidents DROP COLUMN IF EXISTS user_id;
ALTER TABLE status_pages DROP COLUMN IF EXISTS user_id;
ALTER TABLE maintenance_windows DROP COLUMN IF EXISTS user_id;

-- Drop role from users
ALTER TABLE users DROP COLUMN IF EXISTS role;
