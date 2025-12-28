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
