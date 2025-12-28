package adapters

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/ingvar/aiaggregator/packages/domain"
	"github.com/ingvar/aiaggregator/packages/usecases"
)

type InMemoryJobRepository struct {
	mu   sync.RWMutex
	jobs map[uuid.UUID]*domain.Job
}

var _ usecases.JobRepository = (*InMemoryJobRepository)(nil)

func NewInMemoryJobRepository() *InMemoryJobRepository {
	return &InMemoryJobRepository{jobs: make(map[uuid.UUID]*domain.Job)}
}

func (r *InMemoryJobRepository) Create(ctx context.Context, job *domain.Job) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if job.ID == uuid.Nil {
		job.ID = uuid.New()
	}
	now := time.Now().UTC()
	if job.CreatedAt.IsZero() {
		job.CreatedAt = now
	}
	job.UpdatedAt = now
	r.jobs[job.ID] = job
	return nil
}

func (r *InMemoryJobRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Job, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	job, ok := r.jobs[id]
	if !ok {
		return nil, domain.ErrJobNotFound
	}
	return job, nil
}

func (r *InMemoryJobRepository) GetByTenantID(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*domain.Job, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []*domain.Job
	for _, job := range r.jobs {
		if job.TenantID == tenantID {
			result = append(result, job)
		}
	}
	if offset >= len(result) {
		return []*domain.Job{}, nil
	}
	end := offset + limit
	if end > len(result) {
		end = len(result)
	}
	return result[offset:end], nil
}

func (r *InMemoryJobRepository) Update(ctx context.Context, job *domain.Job) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.jobs[job.ID]; !ok {
		return domain.ErrJobNotFound
	}
	job.UpdatedAt = time.Now().UTC()
	r.jobs[job.ID] = job
	return nil
}

func (r *InMemoryJobRepository) Delete(ctx context.Context, id uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.jobs[id]; !ok {
		return domain.ErrJobNotFound
	}
	delete(r.jobs, id)
	return nil
}

func (r *InMemoryJobRepository) Count(ctx context.Context, tenantID uuid.UUID) (int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	count := 0
	for _, job := range r.jobs {
		if job.TenantID == tenantID {
			count++
		}
	}
	return count, nil
}

type InMemoryTenantRepository struct {
	mu      sync.RWMutex
	tenants map[uuid.UUID]*domain.Tenant
	byKey   map[string]*domain.Tenant
}

var _ usecases.TenantRepository = (*InMemoryTenantRepository)(nil)

func NewInMemoryTenantRepository() *InMemoryTenantRepository {
	return &InMemoryTenantRepository{
		tenants: make(map[uuid.UUID]*domain.Tenant),
		byKey:   make(map[string]*domain.Tenant),
	}
}

func (r *InMemoryTenantRepository) Create(ctx context.Context, tenant *domain.Tenant) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if tenant.ID == uuid.Nil {
		tenant.ID = uuid.New()
	}
	now := time.Now().UTC()
	if tenant.CreatedAt.IsZero() {
		tenant.CreatedAt = now
	}
	tenant.UpdatedAt = now
	r.tenants[tenant.ID] = tenant
	r.byKey[tenant.APIKey] = tenant
	return nil
}

func (r *InMemoryTenantRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Tenant, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	tenant, ok := r.tenants[id]
	if !ok {
		return nil, domain.ErrTenantNotFound
	}
	return tenant, nil
}

func (r *InMemoryTenantRepository) GetByAPIKey(ctx context.Context, apiKey string) (*domain.Tenant, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	tenant, ok := r.byKey[apiKey]
	if !ok {
		return nil, domain.ErrTenantNotFound
	}
	return tenant, nil
}

func (r *InMemoryTenantRepository) Update(ctx context.Context, tenant *domain.Tenant) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	old, ok := r.tenants[tenant.ID]
	if !ok {
		return domain.ErrTenantNotFound
	}
	delete(r.byKey, old.APIKey)
	tenant.UpdatedAt = time.Now().UTC()
	r.tenants[tenant.ID] = tenant
	r.byKey[tenant.APIKey] = tenant
	return nil
}

func (r *InMemoryTenantRepository) Delete(ctx context.Context, id uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	tenant, ok := r.tenants[id]
	if !ok {
		return domain.ErrTenantNotFound
	}
	delete(r.byKey, tenant.APIKey)
	delete(r.tenants, id)
	return nil
}

func (r *InMemoryTenantRepository) List(ctx context.Context, limit, offset int) ([]*domain.Tenant, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]*domain.Tenant, 0, len(r.tenants))
	for _, t := range r.tenants {
		result = append(result, t)
	}
	if offset >= len(result) {
		return []*domain.Tenant{}, nil
	}
	end := offset + limit
	if end > len(result) {
		end = len(result)
	}
	return result[offset:end], nil
}

func (r *InMemoryTenantRepository) SeedTestTenant() *domain.Tenant {
	tenant := domain.NewTenant("Test Tenant", "test-api-key-12345")
	r.Create(context.Background(), tenant)
	return tenant
}

type InMemoryUsageRepository struct {
	mu     sync.RWMutex
	usages map[uuid.UUID]*domain.Usage
}

var _ usecases.UsageRepository = (*InMemoryUsageRepository)(nil)

func NewInMemoryUsageRepository() *InMemoryUsageRepository {
	return &InMemoryUsageRepository{usages: make(map[uuid.UUID]*domain.Usage)}
}

func (r *InMemoryUsageRepository) Create(ctx context.Context, usage *domain.Usage) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if usage.ID == uuid.Nil {
		usage.ID = uuid.New()
	}
	if usage.CreatedAt.IsZero() {
		usage.CreatedAt = time.Now().UTC()
	}
	r.usages[usage.ID] = usage
	return nil
}

func (r *InMemoryUsageRepository) GetByJobID(ctx context.Context, jobID uuid.UUID) (*domain.Usage, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, usage := range r.usages {
		if usage.JobID == jobID {
			return usage, nil
		}
	}
	return nil, domain.ErrJobNotFound
}

func (r *InMemoryUsageRepository) GetByTenantID(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*domain.Usage, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []*domain.Usage
	for _, usage := range r.usages {
		if usage.TenantID == tenantID {
			result = append(result, usage)
		}
	}
	if offset >= len(result) {
		return []*domain.Usage{}, nil
	}
	end := offset + limit
	if end > len(result) {
		end = len(result)
	}
	return result[offset:end], nil
}

func (r *InMemoryUsageRepository) GetSummary(ctx context.Context, tenantID uuid.UUID) ([]*domain.UsageSummary, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	summaryMap := make(map[string]*domain.UsageSummary)
	for _, usage := range r.usages {
		if usage.TenantID != tenantID {
			continue
		}
		summary, ok := summaryMap[usage.Provider]
		if !ok {
			summary = &domain.UsageSummary{Provider: usage.Provider}
			summaryMap[usage.Provider] = summary
		}
		summary.TotalTokensIn += usage.TokensIn
		summary.TotalTokensOut += usage.TokensOut
		summary.TotalCost += usage.Cost
		summary.JobCount++
	}
	result := make([]*domain.UsageSummary, 0, len(summaryMap))
	for _, s := range summaryMap {
		result = append(result, s)
	}
	return result, nil
}
