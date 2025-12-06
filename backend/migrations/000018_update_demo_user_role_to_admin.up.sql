-- Update demo user role to admin
UPDATE users
SET role = 'admin'
WHERE email = 'test@gmail.com';