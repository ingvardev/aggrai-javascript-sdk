// Package usecases contains application business logic and use case implementations.
package usecases

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/ingvar/aiaggregator/packages/domain"
	"github.com/ingvar/aiaggregator/packages/pubsub"
)

// ProcessJobService handles job processing logic.
type ProcessJobService struct {
	jobRepo    JobRepository
	usageRepo  UsageRepository
	providerFn func(name string) (AIProvider, bool)
	publisher  *pubsub.Publisher
}

// NewProcessJobService creates a new process job service.
func NewProcessJobService(
	jobRepo JobRepository,
	usageRepo UsageRepository,
	providerFn func(name string) (AIProvider, bool),
) *ProcessJobService {
	return &ProcessJobService{
		jobRepo:    jobRepo,
		usageRepo:  usageRepo,
		providerFn: providerFn,
	}
}

// SetPublisher sets the Redis publisher for job updates.
func (s *ProcessJobService) SetPublisher(publisher *pubsub.Publisher) {
	s.publisher = publisher
}

// publishUpdate publishes a job update event via Redis.
func (s *ProcessJobService) publishUpdate(ctx context.Context, job *domain.Job) {
	if s.publisher == nil {
		return
	}

	update := &pubsub.JobUpdate{
		JobID:      job.ID.String(),
		TenantID:   job.TenantID.String(),
		Type:       string(job.Type),
		Input:      job.Input,
		Status:     string(job.Status),
		Result:     job.Result,
		Error:      job.Error,
		Provider:   job.Provider,
		TokensIn:   job.TokensIn,
		TokensOut:  job.TokensOut,
		Cost:       job.Cost,
		CreatedAt:  job.CreatedAt,
		UpdatedAt:  job.UpdatedAt,
		StartedAt:  job.StartedAt,
		FinishedAt: job.FinishedAt,
	}

	// Fire and forget - don't block on publish errors
	go func() {
		_ = s.publisher.Publish(context.Background(), update)
	}()
}

// ProcessJob processes a job with the specified provider.
func (s *ProcessJobService) ProcessJob(ctx context.Context, jobID uuid.UUID, providerName string) error {
	job, err := s.jobRepo.GetByID(ctx, jobID)
	if err != nil {
		return err
	}

	// Get the provider
	provider, ok := s.providerFn(providerName)
	if !ok {
		job.MarkFailed("provider not found: " + providerName)
		_ = s.jobRepo.Update(ctx, job)
		s.publishUpdate(ctx, job)
		return nil
	}

	// Check provider availability
	if !provider.IsAvailable(ctx) {
		job.MarkFailed("provider not available: " + providerName)
		_ = s.jobRepo.Update(ctx, job)
		s.publishUpdate(ctx, job)
		return nil
	}

	// Mark job as processing
	job.MarkProcessing(providerName)
	if err := s.jobRepo.Update(ctx, job); err != nil {
		return err
	}
	s.publishUpdate(ctx, job) // Notify: PROCESSING

	// Execute the job
	result, err := provider.Execute(ctx, job)
	if err != nil {
		job.MarkFailed(err.Error())
		_ = s.jobRepo.Update(ctx, job)
		s.publishUpdate(ctx, job) // Notify: FAILED
		return nil
	}

	// Mark job as completed
	job.MarkCompleted(result.Result, result.TokensIn, result.TokensOut, result.Cost)
	if err := s.jobRepo.Update(ctx, job); err != nil {
		return err
	}
	s.publishUpdate(ctx, job) // Notify: COMPLETED

	// Record usage
	usage := &domain.Usage{
		ID:        uuid.New(),
		TenantID:  job.TenantID,
		JobID:     job.ID,
		Provider:  providerName,
		Model:     result.Model,
		TokensIn:  result.TokensIn,
		TokensOut: result.TokensOut,
		Cost:      result.Cost,
		CreatedAt: time.Now().UTC(),
	}

	if s.usageRepo != nil {
		if err := s.usageRepo.Create(ctx, usage); err != nil {
			// Log error but don't fail the job
			// TODO: add logging
		}
	}

	return nil
}

// ProcessJobWithAutoProvider processes a job, automatically selecting the best available provider.
func (s *ProcessJobService) ProcessJobWithAutoProvider(ctx context.Context, jobID uuid.UUID, providers []string) error {
	// Try providers in order until one succeeds
	for _, providerName := range providers {
		provider, ok := s.providerFn(providerName)
		if !ok {
			continue
		}

		if provider.IsAvailable(ctx) {
			return s.ProcessJob(ctx, jobID, providerName)
		}
	}

	// No provider available
	job, err := s.jobRepo.GetByID(ctx, jobID)
	if err != nil {
		return err
	}

	job.MarkFailed("no available providers")
	return s.jobRepo.Update(ctx, job)
}
