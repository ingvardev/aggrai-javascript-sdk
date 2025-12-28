-- name: GetJob :one
SELECT * FROM jobs WHERE id = $1 LIMIT 1;

-- name: ListJobsByTenant :many
SELECT * FROM jobs
WHERE tenant_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: CountJobsByTenant :one
SELECT COUNT(*) FROM jobs WHERE tenant_id = $1;

-- name: CreateJob :one
INSERT INTO jobs (tenant_id, type, input, status)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: UpdateJob :one
UPDATE jobs
SET status = $2, result = $3, error = $4, provider = $5,
    tokens_in = $6, tokens_out = $7, cost = $8,
    started_at = $9, finished_at = $10, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteJob :exec
DELETE FROM jobs WHERE id = $1;

-- name: ListJobsByStatus :many
SELECT * FROM jobs
WHERE status = $1
ORDER BY created_at ASC
LIMIT $2;
