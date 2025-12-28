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

// InMemoryJobRepository is an in-memory implementation of JobRepository.
type InMemoryJobRepository struct {
	mu   sync.RWMutex
	jobs map[uuid.UUID]*domain.Job
}

// Ensure InMemoryJobRepository implements JobRepository
var _ usecases.JobRepository = (*InMemoryJobRepository)(nil)

// NewInMemoryJobRepository creates a new in-memory job repository.
func NewInMemoryJobRepository() *InMemoryJobRepository {
	return &InMemoryJobRepository{
		jobs: make(map[uuid.UUID]*domain.Job),
	}
}

// Create saves a new job.
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

// GetByID retrieves a job by ID.
func (r *InMemoryJobRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Job, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	job, ok := r.jobs[id]
	if !ok {
		return nil, domain.ErrJobNotFound
	}
	return job, nil
}

// GetByTenantID retrieves jobs for a tenant with pagination.
func (r *InMemoryJobRepository) GetByTenantID(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*domain.Job, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*domain.Job
	for _, job := range r.jobs {
		if job.TenantID == tenantID {
			result = append(result, job)
		}
	}

	// Sort by CreatedAt descending (newest first) - simplified
	// In production, use proper sorting

	// Apply pagination
	if offset >= len(result) {
		return []*domain.Job{}, nil
	}

	end := offset + limit
	if end > len(result) {
		end = len(result)
	}

	return result[offset:end], nil
}

// Update updates a job.
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

// Delete removes a job.
func (r *InMemoryJobRepository) Delete(ctx context.Context, id uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.jobs[id]; !ok {
		return domain.ErrJobNotFound
	}

	delete(r.jobs, id)
	return nil
}

// Count returns the total number of jobs for a tenant.
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
