-- Add role column to users table
ALTER TABLE users ADD COLUMN IF NOT EXISTS role VARCHAR(50) NOT NULL DEFAULT 'user';

-- Update demo user to admin
UPDATE users SET role = 'admin' WHERE email = 'test@gmail.com';