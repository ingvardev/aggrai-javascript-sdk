-- Seed data for development

-- Insert default tenant
INSERT INTO tenants (id, name, api_key, active) VALUES
    ('00000000-0000-0000-0000-000000000001', 'Default Tenant', 'dev-api-key-12345', true)
ON CONFLICT (id) DO NOTHING;

-- Insert default stub provider
INSERT INTO providers (id, name, type, endpoint, enabled, priority) VALUES
    ('00000000-0000-0000-0000-000000000001', 'Stub Provider', 'local', 'http://localhost', true, 100)
ON CONFLICT (id) DO NOTHING;

-- Insert default admin user for dashboard
-- Password: admin123 (bcrypt hash)
INSERT INTO tenant_owners (id, tenant_id, email, password_hash, name, role, active, email_verified) VALUES
    ('00000000-0000-0000-0000-000000000002', '00000000-0000-0000-0000-000000000001',
     'admin@localhost', '$2a$10$aDedOTW9djhhJ5ZFmysIN.lqCb6EfJGfb/J6p8nE.fb/HOEkXeGDa',
     'Admin', 'owner', true, true)
ON CONFLICT (id) DO NOTHING;
