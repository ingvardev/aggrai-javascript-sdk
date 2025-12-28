// Package usecases contains application business logic and use case implementations.
package usecases

import (
	"context"

	"github.com/google/uuid"
	"github.com/ingvar/aiaggregator/packages/domain"
)

// CreateJobInput represents input for creating a job.
type CreateJobInput struct {
	TenantID uuid.UUID
	Type     domain.JobType
	Input    string
}

// JobService handles job business logic.
type JobService struct {
	jobRepo  JobRepository
	jobQueue JobQueue
}

// NewJobService creates a new job service.
func NewJobService(jobRepo JobRepository, jobQueue JobQueue) *JobService {
	return &JobService{
		jobRepo:  jobRepo,
		jobQueue: jobQueue,
	}
}

// CreateJob creates a new job and enqueues it for processing.
func (s *JobService) CreateJob(ctx context.Context, input *CreateJobInput) (*domain.Job, error) {
	if input.Input == "" {
		return nil, domain.ErrInvalidInput
	}

	job := domain.NewJob(input.TenantID, input.Type, input.Input)

	if err := s.jobRepo.Create(ctx, job); err != nil {
		return nil, err
	}

	// Enqueue the job for async processing
	if s.jobQueue != nil {
		if err := s.jobQueue.Enqueue(ctx, job.ID); err != nil {
			// Log error but don't fail - job is already created
			// TODO: implement retry mechanism
		}
	}

	return job, nil
}

// GetJob retrieves a job by ID.
func (s *JobService) GetJob(ctx context.Context, id uuid.UUID) (*domain.Job, error) {
	return s.jobRepo.GetByID(ctx, id)
}

// ListJobs retrieves jobs for a tenant with pagination.
func (s *JobService) ListJobs(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*domain.Job, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	return s.jobRepo.GetByTenantID(ctx, tenantID, limit, offset)
}

// CountJobs returns the total count of jobs for a tenant.
func (s *JobService) CountJobs(ctx context.Context, tenantID uuid.UUID) (int, error) {
	return s.jobRepo.Count(ctx, tenantID)
}

// UpdateJob updates a job.
func (s *JobService) UpdateJob(ctx context.Context, job *domain.Job) error {
	return s.jobRepo.Update(ctx, job)
}

// CancelJob attempts to cancel a pending job.
func (s *JobService) CancelJob(ctx context.Context, id uuid.UUID) (*domain.Job, error) {
	job, err := s.jobRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if job.Status != domain.JobStatusPending {
		return nil, domain.ErrInvalidInput
	}

	job.MarkFailed("cancelled by user")
	if err := s.jobRepo.Update(ctx, job); err != nil {
		return nil, err
	}

	return job, nil
}
