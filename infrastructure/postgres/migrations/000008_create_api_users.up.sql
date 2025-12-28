-- API Users table - represents API users within a tenant
CREATE TABLE IF NOT EXISTS api_users (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name        VARCHAR(255) NOT NULL,
    description TEXT,
    active      BOOLEAN NOT NULL DEFAULT true,
    metadata    JSONB DEFAULT '{}',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_api_users_tenant_id ON api_users(tenant_id);
CREATE INDEX idx_api_users_tenant_active ON api_users(tenant_id, active);

-- API Keys table - stores hashed API keys
-- Uses HMAC-SHA256 for secure hashing
CREATE TABLE IF NOT EXISTS api_keys (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    api_user_id  UUID NOT NULL REFERENCES api_users(id) ON DELETE CASCADE,
    key_hash     VARCHAR(64) NOT NULL,      -- HMAC-SHA256 = 64 hex chars
    key_prefix   VARCHAR(12) NOT NULL,      -- For display: "sk-abc123..."
    name         VARCHAR(255) NOT NULL DEFAULT 'Default',
    -- Scopes/permissions for the key
    scopes       TEXT[] NOT NULL DEFAULT ARRAY['read', 'write'],
    active       BOOLEAN NOT NULL DEFAULT true,
    expires_at   TIMESTAMPTZ,
    last_used_at TIMESTAMPTZ,
    last_used_ip INET,                      -- IP tracking
    usage_count  BIGINT NOT NULL DEFAULT 0, -- Request counter
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    revoked_at   TIMESTAMPTZ,               -- When key was revoked
    revoked_by   UUID                        -- Who revoked (api_user_id)
);

CREATE UNIQUE INDEX idx_api_keys_hash ON api_keys(key_hash);
CREATE INDEX idx_api_keys_user_id ON api_keys(api_user_id);
CREATE INDEX idx_api_keys_prefix ON api_keys(key_prefix);
CREATE INDEX idx_api_keys_active ON api_keys(active) WHERE active = true;

-- Audit log for security events
CREATE TABLE IF NOT EXISTS api_audit_log (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    api_user_id UUID REFERENCES api_users(id) ON DELETE SET NULL,
    api_key_id  UUID REFERENCES api_keys(id) ON DELETE SET NULL,
    action      VARCHAR(50) NOT NULL,       -- 'key_created', 'key_revoked', 'user_created', 'auth_failed'
    details     JSONB DEFAULT '{}',
    ip_address  INET,
    user_agent  TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_api_audit_log_tenant ON api_audit_log(tenant_id);
CREATE INDEX idx_api_audit_log_user ON api_audit_log(api_user_id);
CREATE INDEX idx_api_audit_log_action ON api_audit_log(action);
CREATE INDEX idx_api_audit_log_created ON api_audit_log(created_at DESC);

-- Add api_user_id to jobs (nullable for backward compatibility)
ALTER TABLE jobs ADD COLUMN IF NOT EXISTS api_user_id UUID REFERENCES api_users(id) ON DELETE SET NULL;
CREATE INDEX IF NOT EXISTS idx_jobs_api_user_id ON jobs(api_user_id);

-- Add api_user_id to usage (nullable for backward compatibility)
ALTER TABLE usage ADD COLUMN IF NOT EXISTS api_user_id UUID REFERENCES api_users(id) ON DELETE SET NULL;
CREATE INDEX IF NOT EXISTS idx_usage_api_user_id ON usage(api_user_id);

-- Trigger for api_users.updated_at
CREATE OR REPLACE FUNCTION update_api_users_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_api_users_updated_at
    BEFORE UPDATE ON api_users
    FOR EACH ROW
    EXECUTE FUNCTION update_api_users_updated_at();

-- Function to increment api_key usage_count (for atomic updates)
CREATE OR REPLACE FUNCTION increment_api_key_usage(key_id UUID, client_ip INET)
RETURNS VOID AS $$
BEGIN
    UPDATE api_keys
    SET usage_count = usage_count + 1,
        last_used_at = NOW(),
        last_used_ip = client_ip
    WHERE id = key_id;
END;
$$ LANGUAGE plpgsql;
