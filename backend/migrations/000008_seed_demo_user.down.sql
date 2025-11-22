-- Remove demo user and associated tenant
DELETE FROM users WHERE email = 'test@gmail.com';
-- Cascade will automatically remove tenant and tenant_users entries
