-- 006_seed_roles.sql
-- Built-in roles (ARCHITECTURE.md §9, PRD §5). `name` is the user_role_type enum
-- and is unique, so this is safe to run repeatedly.

INSERT INTO roles (name, description) VALUES
    ('user',   'Registered customer'),
    ('vendor', 'Marketplace vendor'),
    ('admin',  'Platform administrator')
ON CONFLICT (name) DO NOTHING;
