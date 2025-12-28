// Package usecases contains application business logic and use case implementations.
package usecases

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/ingvar/aiaggregator/packages/domain"
)

// ProcessJobService handles job processing logic.
type ProcessJobService struct {
	jobRepo    JobRepository
	usageRepo  UsageRepository
	providerFn func(name string) (AIProvider, bool)
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
		return s.jobRepo.Update(ctx, job)
	}

	// Check provider availability
	if !provider.IsAvailable(ctx) {
		job.MarkFailed("provider not available: " + providerName)
		return s.jobRepo.Update(ctx, job)
	}

	// Mark job as processing
	job.MarkProcessing(providerName)
	if err := s.jobRepo.Update(ctx, job); err != nil {
		return err
	}

	// Execute the job
	result, err := provider.Execute(ctx, job)
	if err != nil {
		job.MarkFailed(err.Error())
		return s.jobRepo.Update(ctx, job)
	}

	// Mark job as completed
	job.MarkCompleted(result.Result, result.TokensIn, result.TokensOut, result.Cost)
	if err := s.jobRepo.Update(ctx, job); err != nil {
		return err
	}

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
