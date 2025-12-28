-- Tenant Owners - human users who log in to dashboard
CREATE TABLE IF NOT EXISTS tenant_owners (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id      UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    email          VARCHAR(255) NOT NULL,
    password_hash  VARCHAR(255) NOT NULL,
    name           VARCHAR(255) NOT NULL,
    role           VARCHAR(50) NOT NULL DEFAULT 'owner',  -- owner, admin, member
    active         BOOLEAN NOT NULL DEFAULT true,
    email_verified BOOLEAN NOT NULL DEFAULT false,
    last_login_at  TIMESTAMPTZ,
    last_login_ip  INET,
    failed_attempts INT NOT NULL DEFAULT 0,
    locked_until   TIMESTAMPTZ,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Unique email globally (one account per email)
CREATE UNIQUE INDEX idx_tenant_owners_email ON tenant_owners(email);
CREATE INDEX idx_tenant_owners_tenant_id ON tenant_owners(tenant_id);
CREATE INDEX idx_tenant_owners_active ON tenant_owners(active) WHERE active = true;

-- Sessions for web authentication
CREATE TABLE IF NOT EXISTS owner_sessions (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id    UUID NOT NULL REFERENCES tenant_owners(id) ON DELETE CASCADE,
    token_hash  VARCHAR(64) NOT NULL,
    user_agent  TEXT,
    ip_address  INET,
    expires_at  TIMESTAMPTZ NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_owner_sessions_token ON owner_sessions(token_hash);
CREATE INDEX idx_owner_sessions_owner ON owner_sessions(owner_id);
CREATE INDEX idx_owner_sessions_expires ON owner_sessions(expires_at);

-- Password reset tokens
CREATE TABLE IF NOT EXISTS password_reset_tokens (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id   UUID NOT NULL REFERENCES tenant_owners(id) ON DELETE CASCADE,
    token_hash VARCHAR(64) NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    used_at    TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_password_reset_token_hash ON password_reset_tokens(token_hash);
CREATE INDEX idx_password_reset_owner ON password_reset_tokens(owner_id);

-- Trigger for updated_at (reuse existing function)
CREATE TRIGGER trigger_tenant_owners_updated_at
    BEFORE UPDATE ON tenant_owners
    FOR EACH ROW
    EXECUTE FUNCTION update_api_users_updated_at();
