-- Drop trigger and function
DROP TRIGGER IF EXISTS trigger_provider_pricing_updated_at ON provider_pricing;
DROP FUNCTION IF EXISTS update_provider_pricing_updated_at();

-- Drop indexes
DROP INDEX IF EXISTS idx_provider_pricing_model;
DROP INDEX IF EXISTS idx_provider_pricing_provider;

-- Drop table
DROP TABLE IF EXISTS provider_pricing;
