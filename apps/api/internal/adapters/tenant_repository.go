// Package adapters contains infrastructure adapter implementations.
package adapters

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/ingvar/aiaggregator/packages/domain"
	"github.com/ingvar/aiaggregator/packages/usecases"
)

// InMemoryTenantRepository is an in-memory implementation of TenantRepository.
type InMemoryTenantRepository struct {
	mu      sync.RWMutex
	tenants map[uuid.UUID]*domain.Tenant
	byKey   map[string]*domain.Tenant
}

// Ensure InMemoryTenantRepository implements TenantRepository
var _ usecases.TenantRepository = (*InMemoryTenantRepository)(nil)

// NewInMemoryTenantRepository creates a new in-memory tenant repository.
func NewInMemoryTenantRepository() *InMemoryTenantRepository {
	return &InMemoryTenantRepository{
		tenants: make(map[uuid.UUID]*domain.Tenant),
		byKey:   make(map[string]*domain.Tenant),
	}
}

// Create saves a new tenant.
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

// GetByID retrieves a tenant by ID.
func (r *InMemoryTenantRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Tenant, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tenant, ok := r.tenants[id]
	if !ok {
		return nil, domain.ErrTenantNotFound
	}
	return tenant, nil
}

// GetByAPIKey retrieves a tenant by API key.
func (r *InMemoryTenantRepository) GetByAPIKey(ctx context.Context, apiKey string) (*domain.Tenant, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tenant, ok := r.byKey[apiKey]
	if !ok {
		return nil, domain.ErrTenantNotFound
	}
	return tenant, nil
}

// Update updates a tenant.
func (r *InMemoryTenantRepository) Update(ctx context.Context, tenant *domain.Tenant) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	old, ok := r.tenants[tenant.ID]
	if !ok {
		return domain.ErrTenantNotFound
	}

	// Remove old API key mapping
	delete(r.byKey, old.APIKey)

	tenant.UpdatedAt = time.Now().UTC()
	r.tenants[tenant.ID] = tenant
	r.byKey[tenant.APIKey] = tenant
	return nil
}

// Delete removes a tenant.
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

// List returns all tenants with pagination.
func (r *InMemoryTenantRepository) List(ctx context.Context, limit, offset int) ([]*domain.Tenant, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]*domain.Tenant, 0, len(r.tenants))
	for _, t := range r.tenants {
		result = append(result, t)
	}

	// Apply pagination
	if offset >= len(result) {
		return []*domain.Tenant{}, nil
	}

	end := offset + limit
	if end > len(result) {
		end = len(result)
	}

	return result[offset:end], nil
}

// SeedTestTenant creates a test tenant for development.
func (r *InMemoryTenantRepository) SeedTestTenant() *domain.Tenant {
	tenant := domain.NewTenant("Test Tenant", "test-api-key-12345")
	_ = r.Create(context.Background(), tenant)
	return tenant
}
