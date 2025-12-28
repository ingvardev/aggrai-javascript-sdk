package handlers

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/ingvar/aiaggregator/packages/domain"
	"github.com/ingvar/aiaggregator/packages/usecases"
)

// Mock implementations for testing

type mockJobRepo struct {
	jobs map[uuid.UUID]*domain.Job
}

func newMockJobRepo() *mockJobRepo {
	return &mockJobRepo{jobs: make(map[uuid.UUID]*domain.Job)}
}

func (r *mockJobRepo) Create(ctx context.Context, job *domain.Job) error {
	r.jobs[job.ID] = job
	return nil
}

func (r *mockJobRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Job, error) {
	job, ok := r.jobs[id]
	if !ok {
		return nil, domain.ErrJobNotFound
	}
	return job, nil
}

func (r *mockJobRepo) GetByTenantID(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*domain.Job, error) {
	return nil, nil
}

func (r *mockJobRepo) Update(ctx context.Context, job *domain.Job) error {
	r.jobs[job.ID] = job
	return nil
}

func (r *mockJobRepo) Delete(ctx context.Context, id uuid.UUID) error {
	delete(r.jobs, id)
	return nil
}

func (r *mockJobRepo) Count(ctx context.Context, tenantID uuid.UUID) (int, error) {
	return len(r.jobs), nil
}

type mockUsageRepo struct {
	usages []*domain.Usage
}

func (r *mockUsageRepo) Create(ctx context.Context, usage *domain.Usage) error {
	r.usages = append(r.usages, usage)
	return nil
}

func (r *mockUsageRepo) GetByJobID(ctx context.Context, jobID uuid.UUID) (*domain.Usage, error) {
	return nil, nil
}

func (r *mockUsageRepo) GetByTenantID(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*domain.Usage, error) {
	return r.usages, nil
}

func (r *mockUsageRepo) GetSummary(ctx context.Context, tenantID uuid.UUID) ([]*domain.UsageSummary, error) {
	return nil, nil
}

type mockAIProvider struct {
	name     string
	result   *usecases.ProviderResult
	err      error
	available bool
}

func (p *mockAIProvider) Name() string {
	return p.name
}

func (p *mockAIProvider) Type() string {
	return "mock"
}

func (p *mockAIProvider) Execute(ctx context.Context, job *domain.Job) (*usecases.ProviderResult, error) {
	if p.err != nil {
		return nil, p.err
	}
	return p.result, nil
}

func (p *mockAIProvider) IsAvailable(ctx context.Context) bool {
	return p.available
}

func TestJobHandler_HandleProcessJob(t *testing.T) {
	ctx := context.Background()

	t.Run("successful job processing", func(t *testing.T) {
		jobRepo := newMockJobRepo()
		usageRepo := &mockUsageRepo{}
		provider := &mockAIProvider{
			name: "test-provider",
			result: &usecases.ProviderResult{
				Result:    "Test response",
				Model:     "test-model",
				TokensIn:  10,
				TokensOut: 20,
				Cost:      0.001,
			},
			available: true,
		}

		// Create a test job
		tenantID := uuid.New()
		job := domain.NewJob(tenantID, domain.JobTypeText, "Hello AI")
		_ = jobRepo.Create(ctx, job)

		providerFn := func(name string) (usecases.AIProvider, bool) {
			if name == "test-provider" {
				return provider, true
			}
			return nil, false
		}

		processService := usecases.NewProcessJobService(jobRepo, usageRepo, providerFn)
		handler := NewJobHandler(processService, "test-provider")

		// Create task payload
		payload := usecases.JobQueuePayload{
			JobID:    job.ID,
			TenantID: job.TenantID,
			Type:     string(job.Type),
		}
		payloadBytes, _ := json.Marshal(payload)

		task := asynq.NewTask("ai:process", payloadBytes)

		err := handler.HandleProcessJob(ctx, task)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Verify job was updated
		updatedJob, _ := jobRepo.GetByID(ctx, job.ID)
		if updatedJob.Status != domain.JobStatusCompleted {
			t.Errorf("expected status Completed, got %v", updatedJob.Status)
		}

		if updatedJob.Result == nil || *updatedJob.Result != "Test response" {
			t.Errorf("expected result %q, got %v", "Test response", updatedJob.Result)
		}
	})

	t.Run("job not found", func(t *testing.T) {
		jobRepo := newMockJobRepo()
		usageRepo := &mockUsageRepo{}

		providerFn := func(name string) (usecases.AIProvider, bool) {
			return nil, false
		}

		processService := usecases.NewProcessJobService(jobRepo, usageRepo, providerFn)
		handler := NewJobHandler(processService, "test-provider")

		payload := usecases.JobQueuePayload{
			JobID:    uuid.New(), // Non-existent job
			TenantID: uuid.New(),
			Type:     "text",
		}
		payloadBytes, _ := json.Marshal(payload)

		task := asynq.NewTask("ai:process", payloadBytes)

		err := handler.HandleProcessJob(ctx, task)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("invalid payload", func(t *testing.T) {
		jobRepo := newMockJobRepo()
		usageRepo := &mockUsageRepo{}

		providerFn := func(name string) (usecases.AIProvider, bool) {
			return nil, false
		}

		processService := usecases.NewProcessJobService(jobRepo, usageRepo, providerFn)
		handler := NewJobHandler(processService, "test-provider")

		// Invalid JSON
		task := asynq.NewTask("ai:process", []byte("invalid json"))

		err := handler.HandleProcessJob(ctx, task)
		if err == nil {
			t.Fatal("expected error for invalid payload")
		}
	})
}
