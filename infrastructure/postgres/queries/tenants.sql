-- name: GetTenant :one
SELECT * FROM tenants WHERE id = $1 LIMIT 1;

-- name: GetTenantByAPIKey :one
SELECT * FROM tenants WHERE api_key = $1 AND active = true LIMIT 1;

-- name: ListTenants :many
SELECT * FROM tenants ORDER BY created_at DESC LIMIT $1 OFFSET $2;

-- name: CreateTenant :one
INSERT INTO tenants (name, api_key, active)
VALUES ($1, $2, $3)
RETURNING *;

-- name: UpdateTenant :one
UPDATE tenants
SET name = $2, active = $3, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteTenant :exec
DELETE FROM tenants WHERE id = $1;
