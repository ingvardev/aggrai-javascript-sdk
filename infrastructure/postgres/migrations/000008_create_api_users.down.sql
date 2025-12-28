-- Remove triggers and functions
DROP TRIGGER IF EXISTS trigger_api_users_updated_at ON api_users;
DROP FUNCTION IF EXISTS update_api_users_updated_at();
DROP FUNCTION IF EXISTS increment_api_key_usage(UUID, INET);

-- Remove columns from existing tables
ALTER TABLE usage DROP COLUMN IF EXISTS api_user_id;
ALTER TABLE jobs DROP COLUMN IF EXISTS api_user_id;

-- Drop new tables (order matters due to FK constraints)
DROP TABLE IF EXISTS api_audit_log;
DROP TABLE IF EXISTS api_keys;
DROP TABLE IF EXISTS api_users;
