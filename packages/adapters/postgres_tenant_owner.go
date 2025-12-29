package adapters

import (
	"context"
	"errors"
	"net"
	"net/netip"
	"time"

	"github.com/google/uuid"
	"github.com/ingvar/aiaggregator/packages/domain"
	"github.com/ingvar/aiaggregator/packages/usecases"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// parseIP safely parses an IP address string
func parseIP(s string) net.IP {
	return net.ParseIP(s)
}

// netipAddrToNetIP converts netip.Addr to net.IP
func netipAddrToNetIP(addr netip.Addr) net.IP {
	if !addr.IsValid() {
		return nil
	}
	return net.IP(addr.AsSlice())
}

// Ensure PostgresTenantOwnerRepository implements the interface.
var _ usecases.TenantOwnerRepository = (*PostgresTenantOwnerRepository)(nil)

// PostgresTenantOwnerRepository implements TenantOwnerRepository using PostgreSQL.
type PostgresTenantOwnerRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresTenantOwnerRepository creates a new PostgreSQL tenant owner repository.
func NewPostgresTenantOwnerRepository(pool *pgxpool.Pool) *PostgresTenantOwnerRepository {
	return &PostgresTenantOwnerRepository{pool: pool}
}

// Create inserts a new tenant owner.
func (r *PostgresTenantOwnerRepository) Create(ctx context.Context, owner *domain.TenantOwner) error {
	query := `
		INSERT INTO tenant_owners (id, tenant_id, email, password_hash, name, role, active, email_verified, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`

	_, err := r.pool.Exec(ctx, query,
		owner.ID, owner.TenantID, owner.Email, owner.PasswordHash,
		owner.Name, owner.Role, owner.Active, owner.EmailVerified,
		owner.CreatedAt, owner.UpdatedAt)
	return err
}

// GetByID retrieves a tenant owner by ID.
func (r *PostgresTenantOwnerRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.TenantOwner, error) {
	query := `
		SELECT id, tenant_id, email, password_hash, name, role, active, email_verified,
		       last_login_at, last_login_ip, failed_attempts, locked_until, created_at, updated_at
		FROM tenant_owners WHERE id = $1`

	var owner domain.TenantOwner
	var lastLoginIP *netip.Addr
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&owner.ID, &owner.TenantID, &owner.Email, &owner.PasswordHash,
		&owner.Name, &owner.Role, &owner.Active, &owner.EmailVerified,
		&owner.LastLoginAt, &lastLoginIP, &owner.FailedAttempts, &owner.LockedUntil,
		&owner.CreatedAt, &owner.UpdatedAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrOwnerNotFound
	}
	if err != nil {
		return nil, err
	}

	if lastLoginIP != nil {
		owner.LastLoginIP = netipAddrToNetIP(*lastLoginIP)
	}

	return &owner, nil
}

// GetByEmail retrieves a tenant owner by email.
func (r *PostgresTenantOwnerRepository) GetByEmail(ctx context.Context, email string) (*domain.TenantOwner, error) {
	query := `
		SELECT id, tenant_id, email, password_hash, name, role, active, email_verified,
		       last_login_at, last_login_ip, failed_attempts, locked_until, created_at, updated_at
		FROM tenant_owners WHERE email = $1`

	var owner domain.TenantOwner
	var lastLoginIP *netip.Addr
	err := r.pool.QueryRow(ctx, query, email).Scan(
		&owner.ID, &owner.TenantID, &owner.Email, &owner.PasswordHash,
		&owner.Name, &owner.Role, &owner.Active, &owner.EmailVerified,
		&owner.LastLoginAt, &lastLoginIP, &owner.FailedAttempts, &owner.LockedUntil,
		&owner.CreatedAt, &owner.UpdatedAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrOwnerNotFound
	}
	if err != nil {
		return nil, err
	}

	if lastLoginIP != nil {
		owner.LastLoginIP = netipAddrToNetIP(*lastLoginIP)
	}

	return &owner, nil
}

// GetByTenantID retrieves all owners for a tenant.
func (r *PostgresTenantOwnerRepository) GetByTenantID(ctx context.Context, tenantID uuid.UUID) ([]*domain.TenantOwner, error) {
	query := `
		SELECT id, tenant_id, email, password_hash, name, role, active, email_verified,
		       last_login_at, last_login_ip, failed_attempts, locked_until, created_at, updated_at
		FROM tenant_owners
		WHERE tenant_id = $1
		ORDER BY created_at ASC`

	rows, err := r.pool.Query(ctx, query, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var owners []*domain.TenantOwner
	for rows.Next() {
		var owner domain.TenantOwner
		var lastLoginIP *netip.Addr
		if err := rows.Scan(
			&owner.ID, &owner.TenantID, &owner.Email, &owner.PasswordHash,
			&owner.Name, &owner.Role, &owner.Active, &owner.EmailVerified,
			&owner.LastLoginAt, &lastLoginIP, &owner.FailedAttempts, &owner.LockedUntil,
			&owner.CreatedAt, &owner.UpdatedAt); err != nil {
			return nil, err
		}
		if lastLoginIP != nil {
			owner.LastLoginIP = netipAddrToNetIP(*lastLoginIP)
		}
		owners = append(owners, &owner)
	}
	return owners, rows.Err()
}

// Update updates an existing tenant owner.
func (r *PostgresTenantOwnerRepository) Update(ctx context.Context, owner *domain.TenantOwner) error {
	query := `
		UPDATE tenant_owners
		SET email = $2, name = $3, role = $4, active = $5, email_verified = $6, updated_at = NOW()
		WHERE id = $1`

	result, err := r.pool.Exec(ctx, query,
		owner.ID, owner.Email, owner.Name, owner.Role, owner.Active, owner.EmailVerified)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return domain.ErrOwnerNotFound
	}
	return nil
}

// Delete removes a tenant owner.
func (r *PostgresTenantOwnerRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM tenant_owners WHERE id = $1`
	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return domain.ErrOwnerNotFound
	}
	return nil
}

// IncrementFailedAttempts increments the failed login attempts counter.
func (r *PostgresTenantOwnerRepository) IncrementFailedAttempts(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE tenant_owners SET failed_attempts = failed_attempts + 1 WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id)
	return err
}

// ResetFailedAttempts resets the failed login attempts counter.
func (r *PostgresTenantOwnerRepository) ResetFailedAttempts(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE tenant_owners SET failed_attempts = 0, locked_until = NULL WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id)
	return err
}

// LockAccount locks the account until the specified time.
func (r *PostgresTenantOwnerRepository) LockAccount(ctx context.Context, id uuid.UUID, until time.Time) error {
	query := `UPDATE tenant_owners SET locked_until = $2 WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id, until)
	return err
}

// UpdateLastLogin updates the last login timestamp and IP.
func (r *PostgresTenantOwnerRepository) UpdateLastLogin(ctx context.Context, id uuid.UUID, ip string) error {
	query := `UPDATE tenant_owners SET last_login_at = NOW(), last_login_ip = $2::inet WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id, ip)
	return err
}
