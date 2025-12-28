package usecases

import (
	"context"
	"sync"

	"github.com/google/uuid"
	"github.com/ingvar/aiaggregator/packages/domain"
)

// MockJobRepository is a mock implementation of JobRepository
type MockJobRepository struct {
	mu   sync.RWMutex
	jobs map[uuid.UUID]*domain.Job

	CreateErr error
	UpdateErr error
	GetErr    error
	DeleteErr error
}

func NewMockJobRepository() *MockJobRepository {
	return &MockJobRepository{
		jobs: make(map[uuid.UUID]*domain.Job),
	}
}

func (r *MockJobRepository) Create(ctx context.Context, job *domain.Job) error {
	if r.CreateErr != nil {
		return r.CreateErr
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.jobs[job.ID] = job
	return nil
}

func (r *MockJobRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Job, error) {
	if r.GetErr != nil {
		return nil, r.GetErr
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	job, ok := r.jobs[id]
	if !ok {
		return nil, domain.ErrJobNotFound
	}
	return job, nil
}

func (r *MockJobRepository) GetByTenantID(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*domain.Job, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []*domain.Job
	for _, job := range r.jobs {
		if job.TenantID == tenantID {
			result = append(result, job)
		}
	}
	return result, nil
}

func (r *MockJobRepository) Update(ctx context.Context, job *domain.Job) error {
	if r.UpdateErr != nil {
		return r.UpdateErr
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.jobs[job.ID] = job
	return nil
}

func (r *MockJobRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if r.DeleteErr != nil {
		return r.DeleteErr
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.jobs, id)
	return nil
}

func (r *MockJobRepository) Count(ctx context.Context, tenantID uuid.UUID) (int, error) {
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

// MockJobQueue is a mock implementation of JobQueue
type MockJobQueue struct {
	mu       sync.Mutex
	enqueued []uuid.UUID

	EnqueueErr error
}

func NewMockJobQueue() *MockJobQueue {
	return &MockJobQueue{}
}

func (q *MockJobQueue) Enqueue(ctx context.Context, jobID uuid.UUID) error {
	if q.EnqueueErr != nil {
		return q.EnqueueErr
	}
	q.mu.Lock()
	defer q.mu.Unlock()
	q.enqueued = append(q.enqueued, jobID)
	return nil
}

func (q *MockJobQueue) Close() error {
	return nil
}

func (q *MockJobQueue) EnqueuedCount() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return len(q.enqueued)
}

// MockUsageRepository is a mock implementation of UsageRepository
type MockUsageRepository struct {
	mu     sync.Mutex
	usages []*domain.Usage

	CreateErr error
}

func NewMockUsageRepository() *MockUsageRepository {
	return &MockUsageRepository{}
}

func (r *MockUsageRepository) Create(ctx context.Context, usage *domain.Usage) error {
	if r.CreateErr != nil {
		return r.CreateErr
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.usages = append(r.usages, usage)
	return nil
}

func (r *MockUsageRepository) GetByJobID(ctx context.Context, jobID uuid.UUID) (*domain.Usage, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, u := range r.usages {
		if u.JobID == jobID {
			return u, nil
		}
	}
	return nil, nil
}

func (r *MockUsageRepository) GetByTenantID(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*domain.Usage, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var result []*domain.Usage
	for _, u := range r.usages {
		if u.TenantID == tenantID {
			result = append(result, u)
		}
	}
	return result, nil
}

func (r *MockUsageRepository) GetSummary(ctx context.Context, tenantID uuid.UUID) ([]*domain.UsageSummary, error) {
	return nil, nil
}

func (r *MockUsageRepository) CreatedCount() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.usages)
}

// MockAIProvider is a mock implementation of AIProvider
type MockAIProvider struct {
	name       string
	provType   string
	Result     *ProviderResult
	ExecuteErr error
	available  bool
}

func NewMockAIProvider(name string) *MockAIProvider {
	return &MockAIProvider{
		name:     name,
		provType: "mock",
		Result: &ProviderResult{
			Result:    "Mock response",
			Model:     "mock-model",
			TokensIn:  10,
			TokensOut: 20,
			Cost:      0.001,
		},
		available: true,
	}
}

func (p *MockAIProvider) Name() string {
	return p.name
}

func (p *MockAIProvider) Type() string {
	return p.provType
}

func (p *MockAIProvider) Execute(ctx context.Context, job *domain.Job) (*ProviderResult, error) {
	if p.ExecuteErr != nil {
		return nil, p.ExecuteErr
	}
	return p.Result, nil
}

func (p *MockAIProvider) IsAvailable(ctx context.Context) bool {
	return p.available
}

// MockTenantRepository is a mock implementation of TenantRepository
type MockTenantRepository struct {
	mu      sync.RWMutex
	tenants map[uuid.UUID]*domain.Tenant
	byKey   map[string]*domain.Tenant

	GetErr      error
	GetByKeyErr error
}

func NewMockTenantRepository() *MockTenantRepository {
	return &MockTenantRepository{
		tenants: make(map[uuid.UUID]*domain.Tenant),
		byKey:   make(map[string]*domain.Tenant),
	}
}

func (r *MockTenantRepository) Create(ctx context.Context, tenant *domain.Tenant) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tenants[tenant.ID] = tenant
	r.byKey[tenant.APIKey] = tenant
	return nil
}

func (r *MockTenantRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Tenant, error) {
	if r.GetErr != nil {
		return nil, r.GetErr
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	tenant, ok := r.tenants[id]
	if !ok {
		return nil, domain.ErrTenantNotFound
	}
	return tenant, nil
}

func (r *MockTenantRepository) GetByAPIKey(ctx context.Context, apiKey string) (*domain.Tenant, error) {
	if r.GetByKeyErr != nil {
		return nil, r.GetByKeyErr
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	tenant, ok := r.byKey[apiKey]
	if !ok {
		return nil, domain.ErrTenantNotFound
	}
	return tenant, nil
}

func (r *MockTenantRepository) Update(ctx context.Context, tenant *domain.Tenant) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tenants[tenant.ID] = tenant
	r.byKey[tenant.APIKey] = tenant
	return nil
}

func (r *MockTenantRepository) Delete(ctx context.Context, id uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if tenant, ok := r.tenants[id]; ok {
		delete(r.byKey, tenant.APIKey)
		delete(r.tenants, id)
	}
	return nil
}

func (r *MockTenantRepository) List(ctx context.Context, limit, offset int) ([]*domain.Tenant, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []*domain.Tenant
	for _, t := range r.tenants {
		result = append(result, t)
	}
	return result, nil
}
