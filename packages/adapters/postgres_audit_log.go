package adapters

import (
	"context"
	"net"

	"github.com/google/uuid"
	"github.com/ingvar/aiaggregator/packages/domain"
	"github.com/ingvar/aiaggregator/packages/usecases"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Ensure PostgresAuditLogRepository implements the interface.
var _ usecases.AuditLogRepository = (*PostgresAuditLogRepository)(nil)

// PostgresAuditLogRepository implements AuditLogRepository using PostgreSQL.
type PostgresAuditLogRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresAuditLogRepository creates a new PostgreSQL audit log repository.
func NewPostgresAuditLogRepository(pool *pgxpool.Pool) *PostgresAuditLogRepository {
	return &PostgresAuditLogRepository{pool: pool}
}

// Create inserts a new audit log entry.
func (r *PostgresAuditLogRepository) Create(ctx context.Context, entry *domain.AuditLogEntry) error {
	query := `
		INSERT INTO api_audit_log (id, tenant_id, api_user_id, api_key_id, action, details, ip_address, user_agent, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	var ipStr *string
	if entry.IPAddress != nil {
		s := entry.IPAddress.String()
		ipStr = &s
	}

	_, err := r.pool.Exec(ctx, query,
		entry.ID, entry.TenantID, entry.APIUserID, entry.APIKeyID,
		string(entry.Action), entry.Details, ipStr, entry.UserAgent, entry.CreatedAt)
	return err
}

// GetByTenantID retrieves audit log entries for a tenant.
func (r *PostgresAuditLogRepository) GetByTenantID(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*domain.AuditLogEntry, error) {
	query := `
		SELECT id, tenant_id, api_user_id, api_key_id, action, details, ip_address, user_agent, created_at
		FROM api_audit_log
		WHERE tenant_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.pool.Query(ctx, query, tenantID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []*domain.AuditLogEntry
	for rows.Next() {
		var entry domain.AuditLogEntry
		var actionStr string
		var ipStr *string
		if err := rows.Scan(
			&entry.ID, &entry.TenantID, &entry.APIUserID, &entry.APIKeyID,
			&actionStr, &entry.Details, &ipStr, &entry.UserAgent, &entry.CreatedAt); err != nil {
			return nil, err
		}
		entry.Action = domain.AuditAction(actionStr)
		if ipStr != nil {
			entry.IPAddress = net.ParseIP(*ipStr)
		}
		entries = append(entries, &entry)
	}
	return entries, rows.Err()
}

// GetByAPIUserID retrieves audit log entries for a specific API user.
func (r *PostgresAuditLogRepository) GetByAPIUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*domain.AuditLogEntry, error) {
	query := `
		SELECT id, tenant_id, api_user_id, api_key_id, action, details, ip_address, user_agent, created_at
		FROM api_audit_log
		WHERE api_user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.pool.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []*domain.AuditLogEntry
	for rows.Next() {
		var entry domain.AuditLogEntry
		var actionStr string
		var ipStr *string
		if err := rows.Scan(
			&entry.ID, &entry.TenantID, &entry.APIUserID, &entry.APIKeyID,
			&actionStr, &entry.Details, &ipStr, &entry.UserAgent, &entry.CreatedAt); err != nil {
			return nil, err
		}
		entry.Action = domain.AuditAction(actionStr)
		if ipStr != nil {
			entry.IPAddress = net.ParseIP(*ipStr)
		}
		entries = append(entries, &entry)
	}
	return entries, rows.Err()
}
