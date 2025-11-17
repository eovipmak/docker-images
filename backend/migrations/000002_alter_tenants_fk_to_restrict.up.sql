-- Drop the existing CASCADE constraint
ALTER TABLE tenants DROP CONSTRAINT IF EXISTS fk_tenants_owner;

-- Add new RESTRICT constraint to prevent deletion of users with dependent tenants
ALTER TABLE tenants
    ADD CONSTRAINT fk_tenants_owner 
    FOREIGN KEY (owner_id) 
    REFERENCES users(id) 
    ON DELETE RESTRICT;
