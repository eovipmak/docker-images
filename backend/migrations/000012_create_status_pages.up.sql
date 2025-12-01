-- Create status_pages table
CREATE TABLE IF NOT EXISTS status_pages (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id INTEGER NOT NULL,
    slug VARCHAR(255) NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL,
    public_enabled BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_status_pages_tenant FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE
);

-- Create index on tenant_id for faster tenant-scoped queries
CREATE INDEX IF NOT EXISTS idx_status_pages_tenant_id ON status_pages(tenant_id);

-- Create index on slug for fast lookups
CREATE INDEX IF NOT EXISTS idx_status_pages_slug ON status_pages(slug);

-- Create trigger to automatically update updated_at on status_pages
CREATE TRIGGER update_status_pages_updated_at BEFORE UPDATE ON status_pages
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();