-- Revert back to CASCADE constraint
ALTER TABLE tenants DROP CONSTRAINT IF EXISTS fk_tenants_owner;

ALTER TABLE tenants
    ADD CONSTRAINT fk_tenants_owner 
    FOREIGN KEY (owner_id) 
    REFERENCES users(id) 
    ON DELETE CASCADE;
