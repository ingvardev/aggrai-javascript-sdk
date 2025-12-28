-- +migrate Up
-- SQL for migration up

-- Tenants table
CREATE TABLE IF NOT EXISTS tenants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    api_key VARCHAR(255) UNIQUE NOT NULL,
    active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_tenants_api_key ON tenants(api_key);
CREATE INDEX idx_tenants_active ON tenants(active);

-- +migrate Down
DROP TABLE IF EXISTS tenants;
