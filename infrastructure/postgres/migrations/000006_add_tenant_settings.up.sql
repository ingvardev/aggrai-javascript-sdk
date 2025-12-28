-- Add settings and default_provider to tenants
ALTER TABLE tenants ADD COLUMN IF NOT EXISTS default_provider VARCHAR(255);
ALTER TABLE tenants ADD COLUMN IF NOT EXISTS settings JSONB DEFAULT '{"darkMode": true, "notifications": {"jobCompleted": true, "jobFailed": true, "providerOffline": true, "usageThreshold": false, "weeklySummary": false, "marketingEmails": false}}'::jsonb;
