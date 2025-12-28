-- +migrate Up
-- SQL for migration up

-- Usage table
CREATE TABLE IF NOT EXISTS usage (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    job_id UUID NOT NULL REFERENCES jobs(id) ON DELETE CASCADE,
    provider VARCHAR(100) NOT NULL,
    model VARCHAR(255),
    tokens_in INTEGER DEFAULT 0,
    tokens_out INTEGER DEFAULT 0,
    cost DECIMAL(10, 6) DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_usage_tenant_id ON usage(tenant_id);
CREATE INDEX idx_usage_job_id ON usage(job_id);
CREATE INDEX idx_usage_provider ON usage(provider);
CREATE INDEX idx_usage_created_at ON usage(created_at DESC);
CREATE INDEX idx_usage_tenant_provider ON usage(tenant_id, provider);

-- +migrate Down
DROP TABLE IF EXISTS usage;
