-- name: GetPricing :one
SELECT * FROM provider_pricing WHERE id = $1 LIMIT 1;

-- name: GetPricingByProviderModel :one
SELECT * FROM provider_pricing
WHERE provider = $1 AND model = $2
LIMIT 1;

-- name: GetDefaultPricingByProvider :one
SELECT * FROM provider_pricing
WHERE provider = $1 AND is_default = true
LIMIT 1;

-- name: ListPricing :many
SELECT * FROM provider_pricing
ORDER BY provider, model;

-- name: ListPricingByProvider :many
SELECT * FROM provider_pricing
WHERE provider = $1
ORDER BY is_default DESC, model;

-- name: CreatePricing :one
INSERT INTO provider_pricing (provider, model, input_price_per_million, output_price_per_million, image_price, is_default)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: UpdatePricing :one
UPDATE provider_pricing
SET
    input_price_per_million = $2,
    output_price_per_million = $3,
    image_price = $4,
    is_default = $5,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeletePricing :exec
DELETE FROM provider_pricing WHERE id = $1;

-- name: SetDefaultPricing :exec
UPDATE provider_pricing
SET is_default = (id = $2)
WHERE provider = $1;
