package adapters

import (
	"context"
	"errors"
	"net"
	"net/netip"

	"github.com/google/uuid"
	"github.com/ingvar/aiaggregator/packages/domain"
	"github.com/ingvar/aiaggregator/packages/usecases"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Ensure PostgresOwnerSessionRepository implements the interface.
var _ usecases.OwnerSessionRepository = (*PostgresOwnerSessionRepository)(nil)

// PostgresOwnerSessionRepository implements OwnerSessionRepository using PostgreSQL.
type PostgresOwnerSessionRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresOwnerSessionRepository creates a new PostgreSQL owner session repository.
func NewPostgresOwnerSessionRepository(pool *pgxpool.Pool) *PostgresOwnerSessionRepository {
	return &PostgresOwnerSessionRepository{pool: pool}
}

// Create inserts a new session.
func (r *PostgresOwnerSessionRepository) Create(ctx context.Context, session *domain.OwnerSession) error {
	query := `
		INSERT INTO owner_sessions (id, owner_id, token_hash, user_agent, ip_address, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5::inet, $6, $7)`

	var ipStr *string
	if session.IPAddress != nil {
		s := session.IPAddress.String()
		ipStr = &s
	}

	_, err := r.pool.Exec(ctx, query,
		session.ID, session.OwnerID, session.TokenHash,
		session.UserAgent, ipStr, session.ExpiresAt, session.CreatedAt)
	return err
}

// GetByTokenHash retrieves a session by its token hash.
func (r *PostgresOwnerSessionRepository) GetByTokenHash(ctx context.Context, tokenHash string) (*domain.OwnerSession, error) {
	query := `
		SELECT id, owner_id, token_hash, user_agent, ip_address, expires_at, created_at
		FROM owner_sessions
		WHERE token_hash = $1`

	var session domain.OwnerSession
	var ipAddr *netip.Addr
	err := r.pool.QueryRow(ctx, query, tokenHash).Scan(
		&session.ID, &session.OwnerID, &session.TokenHash,
		&session.UserAgent, &ipAddr, &session.ExpiresAt, &session.CreatedAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrSessionNotFound
	}
	if err != nil {
		return nil, err
	}

	if ipAddr != nil {
		session.IPAddress = net.IP(ipAddr.AsSlice())
	}

	return &session, nil
}

// GetByOwnerID retrieves all sessions for an owner.
func (r *PostgresOwnerSessionRepository) GetByOwnerID(ctx context.Context, ownerID uuid.UUID) ([]*domain.OwnerSession, error) {
	query := `
		SELECT id, owner_id, token_hash, user_agent, ip_address, expires_at, created_at
		FROM owner_sessions
		WHERE owner_id = $1
		ORDER BY created_at DESC`

	rows, err := r.pool.Query(ctx, query, ownerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []*domain.OwnerSession
	for rows.Next() {
		var session domain.OwnerSession
		var ipAddr *netip.Addr
		if err := rows.Scan(
			&session.ID, &session.OwnerID, &session.TokenHash,
			&session.UserAgent, &ipAddr, &session.ExpiresAt, &session.CreatedAt); err != nil {
			return nil, err
		}
		if ipAddr != nil {
			session.IPAddress = net.IP(ipAddr.AsSlice())
		}
		sessions = append(sessions, &session)
	}
	return sessions, rows.Err()
}

// Delete removes a session.
func (r *PostgresOwnerSessionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM owner_sessions WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id)
	return err
}

// DeleteByOwnerID removes all sessions for an owner.
func (r *PostgresOwnerSessionRepository) DeleteByOwnerID(ctx context.Context, ownerID uuid.UUID) error {
	query := `DELETE FROM owner_sessions WHERE owner_id = $1`
	_, err := r.pool.Exec(ctx, query, ownerID)
	return err
}

// DeleteExpired removes all expired sessions.
func (r *PostgresOwnerSessionRepository) DeleteExpired(ctx context.Context) (int64, error) {
	query := `DELETE FROM owner_sessions WHERE expires_at < NOW()`
	result, err := r.pool.Exec(ctx, query)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected(), nil
}
