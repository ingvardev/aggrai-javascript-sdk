-- +migrate Up
-- SQL for migration up

-- Providers table
CREATE TABLE IF NOT EXISTS providers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) UNIQUE NOT NULL,
    type VARCHAR(50) NOT NULL,
    endpoint VARCHAR(500),
    api_key VARCHAR(500),
    model VARCHAR(255),
    enabled BOOLEAN DEFAULT true,
    priority INTEGER DEFAULT 0,
    config JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_providers_type ON providers(type);
CREATE INDEX idx_providers_enabled ON providers(enabled);
CREATE INDEX idx_providers_priority ON providers(priority);

-- +migrate Down
DROP TABLE IF EXISTS providers;
