package adapters

import (
	"context"
	"errors"
	"net"

	"github.com/google/uuid"
	"github.com/ingvar/aiaggregator/packages/domain"
	"github.com/ingvar/aiaggregator/packages/usecases"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Ensure PostgresAPIKeyRepository implements the interface.
var _ usecases.APIKeyRepository = (*PostgresAPIKeyRepository)(nil)

// PostgresAPIKeyRepository implements APIKeyRepository using PostgreSQL.
type PostgresAPIKeyRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresAPIKeyRepository creates a new PostgreSQL API key repository.
func NewPostgresAPIKeyRepository(pool *pgxpool.Pool) *PostgresAPIKeyRepository {
	return &PostgresAPIKeyRepository{pool: pool}
}

// Create inserts a new API key.
func (r *PostgresAPIKeyRepository) Create(ctx context.Context, key *domain.APIKey) error {
	query := `
		INSERT INTO api_keys (id, api_user_id, key_hash, key_prefix, name, scopes, active, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	_, err := r.pool.Exec(ctx, query,
		key.ID, key.APIUserID, key.KeyHash, key.KeyPrefix,
		key.Name, key.Scopes, key.Active, key.ExpiresAt, key.CreatedAt)
	return err
}

// GetByHash retrieves an API key by its hash (for authentication).
func (r *PostgresAPIKeyRepository) GetByHash(ctx context.Context, keyHash string) (*domain.APIKey, error) {
	query := `
		SELECT id, api_user_id, key_hash, key_prefix, name, scopes, active,
		       expires_at, last_used_at, last_used_ip, usage_count, created_at, revoked_at, revoked_by
		FROM api_keys
		WHERE key_hash = $1`

	var key domain.APIKey
	var lastUsedIP *string
	err := r.pool.QueryRow(ctx, query, keyHash).Scan(
		&key.ID, &key.APIUserID, &key.KeyHash, &key.KeyPrefix,
		&key.Name, &key.Scopes, &key.Active, &key.ExpiresAt,
		&key.LastUsedAt, &lastUsedIP, &key.UsageCount, &key.CreatedAt,
		&key.RevokedAt, &key.RevokedBy)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrAPIKeyNotFound
	}
	if err != nil {
		return nil, err
	}

	if lastUsedIP != nil {
		key.LastUsedIP = net.ParseIP(*lastUsedIP)
	}

	return &key, nil
}

// GetByID retrieves an API key by its ID.
func (r *PostgresAPIKeyRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.APIKey, error) {
	query := `
		SELECT id, api_user_id, key_hash, key_prefix, name, scopes, active,
		       expires_at, last_used_at, last_used_ip, usage_count, created_at, revoked_at, revoked_by
		FROM api_keys
		WHERE id = $1`

	var key domain.APIKey
	var lastUsedIP *string
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&key.ID, &key.APIUserID, &key.KeyHash, &key.KeyPrefix,
		&key.Name, &key.Scopes, &key.Active, &key.ExpiresAt,
		&key.LastUsedAt, &lastUsedIP, &key.UsageCount, &key.CreatedAt,
		&key.RevokedAt, &key.RevokedBy)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrAPIKeyNotFound
	}
	if err != nil {
		return nil, err
	}

	if lastUsedIP != nil {
		key.LastUsedIP = net.ParseIP(*lastUsedIP)
	}

	return &key, nil
}

// GetByUserID retrieves all API keys for a user.
func (r *PostgresAPIKeyRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.APIKey, error) {
	query := `
		SELECT id, api_user_id, key_hash, key_prefix, name, scopes, active,
		       expires_at, last_used_at, last_used_ip, usage_count, created_at, revoked_at, revoked_by
		FROM api_keys
		WHERE api_user_id = $1
		ORDER BY created_at DESC`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var keys []*domain.APIKey
	for rows.Next() {
		var key domain.APIKey
		var lastUsedIP *string
		if err := rows.Scan(
			&key.ID, &key.APIUserID, &key.KeyHash, &key.KeyPrefix,
			&key.Name, &key.Scopes, &key.Active, &key.ExpiresAt,
			&key.LastUsedAt, &lastUsedIP, &key.UsageCount, &key.CreatedAt,
			&key.RevokedAt, &key.RevokedBy); err != nil {
			return nil, err
		}
		if lastUsedIP != nil {
			key.LastUsedIP = net.ParseIP(*lastUsedIP)
		}
		keys = append(keys, &key)
	}
	return keys, rows.Err()
}

// UpdateLastUsed updates usage tracking for a key (called asynchronously).
func (r *PostgresAPIKeyRepository) UpdateLastUsed(ctx context.Context, id uuid.UUID, clientIP string) error {
	// Use the database function for atomic update
	_, err := r.pool.Exec(ctx, "SELECT increment_api_key_usage($1, $2::inet)", id, clientIP)
	return err
}

// Revoke marks a key as revoked (soft delete).
func (r *PostgresAPIKeyRepository) Revoke(ctx context.Context, id uuid.UUID, revokedBy uuid.UUID) error {
	query := `
		UPDATE api_keys
		SET active = false, revoked_at = NOW(), revoked_by = $2
		WHERE id = $1 AND active = true`

	result, err := r.pool.Exec(ctx, query, id, revokedBy)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return domain.ErrAPIKeyNotFound
	}
	return nil
}

// RevokeWithTenantCheck revokes a key only if it belongs to the specified tenant.
// This prevents cross-tenant key revocation.
func (r *PostgresAPIKeyRepository) RevokeWithTenantCheck(ctx context.Context, keyID, tenantID uuid.UUID, revokedBy uuid.UUID) error {
	query := `
		UPDATE api_keys
		SET active = false, revoked_at = NOW(), revoked_by = $3
		WHERE id = $1
		  AND active = true
		  AND api_user_id IN (SELECT id FROM api_users WHERE tenant_id = $2)`

	result, err := r.pool.Exec(ctx, query, keyID, tenantID, revokedBy)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return domain.ErrAPIKeyNotFound
	}
	return nil
}

// Delete permanently removes an API key.
func (r *PostgresAPIKeyRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result, err := r.pool.Exec(ctx, "DELETE FROM api_keys WHERE id = $1", id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return domain.ErrAPIKeyNotFound
	}
	return nil
}
