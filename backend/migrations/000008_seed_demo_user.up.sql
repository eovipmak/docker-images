-- Seed demo user for automated testing
-- Email: test@gmail.com
-- Password: Password!
-- Pre-computed bcrypt hash (cost=10): $2a$10$YourHashHere

DO $$
DECLARE
    demo_user_id INTEGER;
    demo_tenant_id INTEGER;
BEGIN
    -- Check if demo user already exists
    IF NOT EXISTS (SELECT 1 FROM users WHERE email = 'test@gmail.com') THEN
        -- Insert demo user
        -- Password hash for "Password!" using bcrypt cost 10
        -- Generated using: bcrypt.GenerateFromPassword([]byte("Password!"), bcrypt.DefaultCost)
        INSERT INTO users (email, password_hash, created_at, updated_at)
        VALUES (
            'test@gmail.com',
            '$2a$10$N9qo8uLOickgx2ZMRZoMye7.9VQ/T7dcCXL0V8LlQIl2iu0fvO1d6',
            CURRENT_TIMESTAMP,
            CURRENT_TIMESTAMP
        )
        RETURNING id INTO demo_user_id;

        -- Create demo tenant
        INSERT INTO tenants (name, slug, owner_id, created_at, updated_at)
        VALUES (
            'Demo Tenant',
            'demo-tenant',
            demo_user_id,
            CURRENT_TIMESTAMP,
            CURRENT_TIMESTAMP
        )
        RETURNING id INTO demo_tenant_id;

        -- Add user to tenant_users
        INSERT INTO tenant_users (tenant_id, user_id, role, created_at)
        VALUES (
            demo_tenant_id,
            demo_user_id,
            'owner',
            CURRENT_TIMESTAMP
        );

        RAISE NOTICE 'Demo user created: test@gmail.com / Password!';
    ELSE
        RAISE NOTICE 'Demo user already exists, skipping...';
    END IF;
END $$;
