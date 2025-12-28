package adapters

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/ingvar/aiaggregator/packages/domain"
	"github.com/ingvar/aiaggregator/packages/usecases"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Ensure PostgresAPIUserRepository implements the interface.
var _ usecases.APIUserRepository = (*PostgresAPIUserRepository)(nil)

// PostgresAPIUserRepository implements APIUserRepository using PostgreSQL.
type PostgresAPIUserRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresAPIUserRepository creates a new PostgreSQL API user repository.
func NewPostgresAPIUserRepository(pool *pgxpool.Pool) *PostgresAPIUserRepository {
	return &PostgresAPIUserRepository{pool: pool}
}

// Create inserts a new API user.
func (r *PostgresAPIUserRepository) Create(ctx context.Context, user *domain.APIUser) error {
	query := `
		INSERT INTO api_users (id, tenant_id, name, description, active, metadata, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	_, err := r.pool.Exec(ctx, query,
		user.ID, user.TenantID, user.Name, user.Description,
		user.Active, user.Metadata, user.CreatedAt, user.UpdatedAt)
	return err
}

// GetByID retrieves an API user by ID.
func (r *PostgresAPIUserRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.APIUser, error) {
	query := `
		SELECT id, tenant_id, name, description, active, metadata, created_at, updated_at
		FROM api_users WHERE id = $1`

	var user domain.APIUser
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&user.ID, &user.TenantID, &user.Name, &user.Description,
		&user.Active, &user.Metadata, &user.CreatedAt, &user.UpdatedAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrAPIUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetByTenantID retrieves all API users for a tenant.
func (r *PostgresAPIUserRepository) GetByTenantID(ctx context.Context, tenantID uuid.UUID) ([]*domain.APIUser, error) {
	query := `
		SELECT id, tenant_id, name, description, active, metadata, created_at, updated_at
		FROM api_users
		WHERE tenant_id = $1
		ORDER BY created_at DESC`

	rows, err := r.pool.Query(ctx, query, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*domain.APIUser
	for rows.Next() {
		var user domain.APIUser
		if err := rows.Scan(
			&user.ID, &user.TenantID, &user.Name, &user.Description,
			&user.Active, &user.Metadata, &user.CreatedAt, &user.UpdatedAt); err != nil {
			return nil, err
		}
		users = append(users, &user)
	}
	return users, rows.Err()
}

// Update updates an existing API user.
func (r *PostgresAPIUserRepository) Update(ctx context.Context, user *domain.APIUser) error {
	query := `
		UPDATE api_users
		SET name = $2, description = $3, active = $4, metadata = $5, updated_at = NOW()
		WHERE id = $1`

	result, err := r.pool.Exec(ctx, query,
		user.ID, user.Name, user.Description, user.Active, user.Metadata)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return domain.ErrAPIUserNotFound
	}
	return nil
}

// Delete removes an API user and cascades to their keys.
func (r *PostgresAPIUserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result, err := r.pool.Exec(ctx, "DELETE FROM api_users WHERE id = $1", id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return domain.ErrAPIUserNotFound
	}
	return nil
}
