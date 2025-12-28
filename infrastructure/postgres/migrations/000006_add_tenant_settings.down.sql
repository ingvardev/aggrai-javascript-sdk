-- Remove settings and default_provider from tenants
ALTER TABLE tenants DROP COLUMN IF EXISTS settings;
ALTER TABLE tenants DROP COLUMN IF EXISTS default_provider;
