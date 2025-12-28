// Package adapters contains repository implementations.
package adapters

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/ingvar/aiaggregator/packages/domain"
	"github.com/ingvar/aiaggregator/packages/usecases"
)

// InMemoryUsageRepository is an in-memory implementation of UsageRepository.
type InMemoryUsageRepository struct {
	mu     sync.RWMutex
	usages map[uuid.UUID]*domain.Usage
}

// Ensure InMemoryUsageRepository implements UsageRepository
var _ usecases.UsageRepository = (*InMemoryUsageRepository)(nil)

// NewInMemoryUsageRepository creates a new in-memory usage repository.
func NewInMemoryUsageRepository() *InMemoryUsageRepository {
	return &InMemoryUsageRepository{
		usages: make(map[uuid.UUID]*domain.Usage),
	}
}

// Create saves a new usage record.
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

// GetByID retrieves a usage record by ID.
func (r *InMemoryUsageRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Usage, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	usage, ok := r.usages[id]
	if !ok {
		return nil, domain.ErrJobNotFound // Using job not found as generic not found
	}
	return usage, nil
}

// GetByTenantID retrieves usage records for a tenant.
func (r *InMemoryUsageRepository) GetByTenantID(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*domain.Usage, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*domain.Usage
	for _, usage := range r.usages {
		if usage.TenantID == tenantID {
			result = append(result, usage)
		}
	}

	// Apply pagination
	if offset >= len(result) {
		return []*domain.Usage{}, nil
	}

	end := offset + limit
	if end > len(result) {
		end = len(result)
	}

	return result[offset:end], nil
}

// GetByJobID retrieves usage records for a job.
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

// GetSummaryByTenant returns aggregated usage statistics for a tenant.
func (r *InMemoryUsageRepository) GetSummary(ctx context.Context, tenantID uuid.UUID) ([]*domain.UsageSummary, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Group by provider
	summaryMap := make(map[string]*domain.UsageSummary)

	for _, usage := range r.usages {
		if usage.TenantID != tenantID {
			continue
		}

		summary, ok := summaryMap[usage.Provider]
		if !ok {
			summary = &domain.UsageSummary{
				Provider: usage.Provider,
			}
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
