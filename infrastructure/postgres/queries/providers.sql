-- name: GetProvider :one
SELECT * FROM providers WHERE id = $1 LIMIT 1;

-- name: GetProviderByName :one
SELECT * FROM providers WHERE name = $1 LIMIT 1;

-- name: ListProviders :many
SELECT * FROM providers ORDER BY priority DESC, name ASC;

-- name: ListEnabledProviders :many
SELECT * FROM providers WHERE enabled = true ORDER BY priority DESC, name ASC;

-- name: CreateProvider :one
INSERT INTO providers (name, type, endpoint, api_key, model, enabled, priority, config)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: UpdateProvider :one
UPDATE providers
SET name = $2, type = $3, endpoint = $4, api_key = $5, model = $6,
    enabled = $7, priority = $8, config = $9, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteProvider :exec
DELETE FROM providers WHERE id = $1;
