-- +migrate Up
-- Seed data for development

-- Insert default tenant
INSERT INTO tenants (id, name, api_key, active) VALUES
    ('00000000-0000-0000-0000-000000000001', 'Default Tenant', 'dev-api-key-12345', true);

-- Insert default stub provider
INSERT INTO providers (id, name, type, endpoint, enabled, priority) VALUES
    ('00000000-0000-0000-0000-000000000001', 'Stub Provider', 'local', 'http://localhost', true, 100);

-- +migrate Down
DELETE FROM providers WHERE id = '00000000-0000-0000-0000-000000000001';
DELETE FROM tenants WHERE id = '00000000-0000-0000-0000-000000000001';
