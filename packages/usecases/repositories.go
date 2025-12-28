// Package usecases contains application business logic and use case implementations.
package usecases

import (
	"context"

	"github.com/google/uuid"
	"github.com/ingvar/aiaggregator/packages/domain"
)

// JobRepository defines the interface for job persistence.
type JobRepository interface {
	Create(ctx context.Context, job *domain.Job) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Job, error)
	GetByTenantID(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*domain.Job, error)
	Update(ctx context.Context, job *domain.Job) error
	Delete(ctx context.Context, id uuid.UUID) error
	Count(ctx context.Context, tenantID uuid.UUID) (int, error)
}

// TenantRepository defines the interface for tenant persistence.
type TenantRepository interface {
	Create(ctx context.Context, tenant *domain.Tenant) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Tenant, error)
	GetByAPIKey(ctx context.Context, apiKey string) (*domain.Tenant, error)
	Update(ctx context.Context, tenant *domain.Tenant) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*domain.Tenant, error)
}

// ProviderRepository defines the interface for provider persistence.
type ProviderRepository interface {
	Create(ctx context.Context, provider *domain.Provider) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Provider, error)
	GetEnabled(ctx context.Context) ([]*domain.Provider, error)
	Update(ctx context.Context, provider *domain.Provider) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context) ([]*domain.Provider, error)
}

// UsageRepository defines the interface for usage persistence.
type UsageRepository interface {
	Create(ctx context.Context, usage *domain.Usage) error
	GetByJobID(ctx context.Context, jobID uuid.UUID) (*domain.Usage, error)
	GetByTenantID(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*domain.Usage, error)
	GetSummary(ctx context.Context, tenantID uuid.UUID) ([]*domain.UsageSummary, error)
}

// PricingRepository defines the interface for provider pricing persistence.
type PricingRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*domain.ProviderPricing, error)
	GetByProviderModel(ctx context.Context, provider, model string) (*domain.ProviderPricing, error)
	GetDefaultByProvider(ctx context.Context, provider string) (*domain.ProviderPricing, error)
	List(ctx context.Context) ([]*domain.ProviderPricing, error)
	ListByProvider(ctx context.Context, provider string) ([]*domain.ProviderPricing, error)
	Create(ctx context.Context, pricing *domain.ProviderPricing) error
	Update(ctx context.Context, pricing *domain.ProviderPricing) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// APIUserRepository defines the interface for API user persistence.
type APIUserRepository interface {
	Create(ctx context.Context, user *domain.APIUser) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.APIUser, error)
	GetByTenantID(ctx context.Context, tenantID uuid.UUID) ([]*domain.APIUser, error)
	Update(ctx context.Context, user *domain.APIUser) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// APIKeyRepository defines the interface for API key persistence.
type APIKeyRepository interface {
	Create(ctx context.Context, key *domain.APIKey) error
	GetByHash(ctx context.Context, keyHash string) (*domain.APIKey, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.APIKey, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.APIKey, error)
	// UpdateLastUsed updates usage tracking (called async)
	UpdateLastUsed(ctx context.Context, id uuid.UUID, clientIP string) error
	// Revoke marks a key as revoked (soft delete)
	Revoke(ctx context.Context, id uuid.UUID, revokedBy uuid.UUID) error
	// RevokeWithTenantCheck revokes a key only if it belongs to the tenant
	RevokeWithTenantCheck(ctx context.Context, keyID, tenantID uuid.UUID, revokedBy uuid.UUID) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// AuditLogRepository defines the interface for audit log persistence.
type AuditLogRepository interface {
	Create(ctx context.Context, entry *domain.AuditLogEntry) error
	GetByTenantID(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*domain.AuditLogEntry, error)
	GetByAPIUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*domain.AuditLogEntry, error)
}
