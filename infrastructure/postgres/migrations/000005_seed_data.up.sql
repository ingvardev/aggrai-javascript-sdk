-- Seed data for development

-- Insert default tenant
INSERT INTO tenants (id, name, api_key, active) VALUES
    ('00000000-0000-0000-0000-000000000001', 'Default Tenant', 'dev-api-key-12345', true)
ON CONFLICT (id) DO NOTHING;

-- Insert default stub provider
INSERT INTO providers (id, name, type, endpoint, enabled, priority) VALUES
    ('00000000-0000-0000-0000-000000000001', 'Stub Provider', 'local', 'http://localhost', true, 100)
ON CONFLICT (id) DO NOTHING;
