-- Provider pricing table for storing token costs
CREATE TABLE IF NOT EXISTS provider_pricing (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    provider VARCHAR(50) NOT NULL,
    model VARCHAR(100) NOT NULL,
    input_price_per_million DECIMAL(10, 6) NOT NULL DEFAULT 0,
    output_price_per_million DECIMAL(10, 6) NOT NULL DEFAULT 0,
    image_price DECIMAL(10, 4) DEFAULT NULL,
    is_default BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(provider, model)
);

-- Create index for faster lookups
CREATE INDEX idx_provider_pricing_provider ON provider_pricing(provider);
CREATE INDEX idx_provider_pricing_model ON provider_pricing(provider, model);

-- Insert default pricing based on current hardcoded values
INSERT INTO provider_pricing (provider, model, input_price_per_million, output_price_per_million, image_price, is_default) VALUES
    -- OpenAI models
    ('openai', 'gpt-4o-mini', 0.15, 0.60, NULL, true),
    ('openai', 'gpt-4o', 2.50, 10.00, NULL, false),
    ('openai', 'gpt-4-turbo', 10.00, 30.00, NULL, false),
    ('openai', 'gpt-3.5-turbo', 0.50, 1.50, NULL, false),
    ('openai', 'dall-e-3', 0, 0, 0.04, true),
    ('openai', 'dall-e-2', 0, 0, 0.02, false),
    -- Claude models
    ('claude', 'claude-3-haiku-20240307', 0.25, 1.25, NULL, true),
    ('claude', 'claude-3-5-sonnet-20241022', 3.00, 15.00, NULL, false),
    ('claude', 'claude-3-opus-20240229', 15.00, 75.00, NULL, false),
    -- Ollama (free, local)
    ('ollama', 'llama2', 0, 0, NULL, true),
    ('ollama', 'mistral', 0, 0, NULL, false),
    ('ollama', 'codellama', 0, 0, NULL, false),
    -- Stub provider (free, for testing)
    ('stub', 'stub-model', 0, 0, NULL, true);

-- Trigger to update updated_at
CREATE OR REPLACE FUNCTION update_provider_pricing_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_provider_pricing_updated_at
    BEFORE UPDATE ON provider_pricing
    FOR EACH ROW
    EXECUTE FUNCTION update_provider_pricing_updated_at();
