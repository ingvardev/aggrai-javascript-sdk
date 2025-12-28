package usecases

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/ingvar/aiaggregator/packages/domain"
)

func TestProcessJobService_ProcessJob(t *testing.T) {
	ctx := context.Background()

	t.Run("successful job processing", func(t *testing.T) {
		jobRepo := NewMockJobRepository()
		usageRepo := NewMockUsageRepository()
		provider := NewMockAIProvider("test-provider")

		tenantID := uuid.New()
		job := domain.NewJob(tenantID, domain.JobTypeText, "Hello AI")
		_ = jobRepo.Create(ctx, job)

		providerFn := func(name string) (AIProvider, bool) {
			if name == "test-provider" {
				return provider, true
			}
			return nil, false
		}

		svc := NewProcessJobService(jobRepo, usageRepo, providerFn)

		err := svc.ProcessJob(ctx, job.ID, "test-provider")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Verify job was updated
		updatedJob, _ := jobRepo.GetByID(ctx, job.ID)
		if updatedJob.Status != domain.JobStatusCompleted {
			t.Errorf("expected status Completed, got %v", updatedJob.Status)
		}

		if updatedJob.Result == nil {
			t.Error("expected result to be set")
		}

		// Verify usage was recorded
		if usageRepo.CreatedCount() != 1 {
			t.Errorf("expected 1 usage record, got %d", usageRepo.CreatedCount())
		}
	})

	t.Run("job not found", func(t *testing.T) {
		jobRepo := NewMockJobRepository()
		usageRepo := NewMockUsageRepository()

		providerFn := func(name string) (AIProvider, bool) {
			return nil, false
		}

		svc := NewProcessJobService(jobRepo, usageRepo, providerFn)

		err := svc.ProcessJob(ctx, uuid.New(), "test-provider")
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if !errors.Is(err, domain.ErrJobNotFound) {
			t.Errorf("expected ErrJobNotFound, got %v", err)
		}
	})

	t.Run("provider not found", func(t *testing.T) {
		jobRepo := NewMockJobRepository()
		usageRepo := NewMockUsageRepository()

		tenantID := uuid.New()
		job := domain.NewJob(tenantID, domain.JobTypeText, "Hello AI")
		_ = jobRepo.Create(ctx, job)

		providerFn := func(name string) (AIProvider, bool) {
			return nil, false
		}

		svc := NewProcessJobService(jobRepo, usageRepo, providerFn)

		err := svc.ProcessJob(ctx, job.ID, "unknown-provider")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Verify job was marked as failed
		updatedJob, _ := jobRepo.GetByID(ctx, job.ID)
		if updatedJob.Status != domain.JobStatusFailed {
			t.Errorf("expected status Failed, got %v", updatedJob.Status)
		}
	})

	t.Run("provider not available", func(t *testing.T) {
		jobRepo := NewMockJobRepository()
		usageRepo := NewMockUsageRepository()
		provider := NewMockAIProvider("test-provider")
		provider.available = false

		tenantID := uuid.New()
		job := domain.NewJob(tenantID, domain.JobTypeText, "Hello AI")
		_ = jobRepo.Create(ctx, job)

		providerFn := func(name string) (AIProvider, bool) {
			if name == "test-provider" {
				return provider, true
			}
			return nil, false
		}

		svc := NewProcessJobService(jobRepo, usageRepo, providerFn)

		err := svc.ProcessJob(ctx, job.ID, "test-provider")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Verify job was marked as failed
		updatedJob, _ := jobRepo.GetByID(ctx, job.ID)
		if updatedJob.Status != domain.JobStatusFailed {
			t.Errorf("expected status Failed, got %v", updatedJob.Status)
		}
	})

	t.Run("provider execution error", func(t *testing.T) {
		jobRepo := NewMockJobRepository()
		usageRepo := NewMockUsageRepository()
		provider := NewMockAIProvider("test-provider")
		provider.ExecuteErr = errors.New("provider error")

		tenantID := uuid.New()
		job := domain.NewJob(tenantID, domain.JobTypeText, "Hello AI")
		_ = jobRepo.Create(ctx, job)

		providerFn := func(name string) (AIProvider, bool) {
			if name == "test-provider" {
				return provider, true
			}
			return nil, false
		}

		svc := NewProcessJobService(jobRepo, usageRepo, providerFn)

		err := svc.ProcessJob(ctx, job.ID, "test-provider")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Verify job was marked as failed
		updatedJob, _ := jobRepo.GetByID(ctx, job.ID)
		if updatedJob.Status != domain.JobStatusFailed {
			t.Errorf("expected status Failed, got %v", updatedJob.Status)
		}

		if updatedJob.Error == nil || *updatedJob.Error != "provider error" {
			t.Errorf("expected error message 'provider error', got %v", updatedJob.Error)
		}
	})
}
