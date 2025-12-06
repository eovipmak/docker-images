-- Revert demo user role to owner
UPDATE tenant_users
SET role = 'owner'
WHERE user_id = (SELECT id FROM users WHERE email = 'test@gmail.com')
  AND tenant_id = (SELECT id FROM tenants WHERE slug = 'demo-tenant');