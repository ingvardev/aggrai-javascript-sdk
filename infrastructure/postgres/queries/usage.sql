-- name: GetUsage :one
SELECT * FROM usage WHERE id = $1 LIMIT 1;

-- name: GetUsageByJobID :one
SELECT * FROM usage WHERE job_id = $1 LIMIT 1;

-- name: ListUsageByTenant :many
SELECT * FROM usage
WHERE tenant_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: CreateUsage :one
INSERT INTO usage (tenant_id, job_id, provider, model, tokens_in, tokens_out, cost)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: GetUsageSummaryByTenant :many
SELECT
    tenant_id,
    provider,
    SUM(tokens_in)::INTEGER as total_tokens_in,
    SUM(tokens_out)::INTEGER as total_tokens_out,
    SUM(cost)::DECIMAL as total_cost,
    COUNT(*)::INTEGER as job_count
FROM usage
WHERE tenant_id = $1
GROUP BY tenant_id, provider;

-- name: GetUsageSummaryByProvider :many
SELECT
    provider,
    SUM(tokens_in)::INTEGER as total_tokens_in,
    SUM(tokens_out)::INTEGER as total_tokens_out,
    SUM(cost)::DECIMAL as total_cost,
    COUNT(*)::INTEGER as job_count
FROM usage
GROUP BY provider;
