package usecases

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/ingvar/aiaggregator/packages/domain"
)

func TestJobService_CreateJob(t *testing.T) {
	tests := []struct {
		name      string
		input     *CreateJobInput
		setupMock func(*MockJobRepository, *MockJobQueue)
		wantErr   bool
	}{
		{
			name: "successful job creation",
			input: &CreateJobInput{
				TenantID: uuid.New(),
				Type:     domain.JobTypeText,
				Input:    "Hello AI",
			},
			setupMock: func(r *MockJobRepository, q *MockJobQueue) {},
			wantErr:   false,
		},
		{
			name: "empty input",
			input: &CreateJobInput{
				TenantID: uuid.New(),
				Type:     domain.JobTypeText,
				Input:    "",
			},
			setupMock: func(r *MockJobRepository, q *MockJobQueue) {},
			wantErr:   true,
		},
		{
			name: "repository error",
			input: &CreateJobInput{
				TenantID: uuid.New(),
				Type:     domain.JobTypeText,
				Input:    "Hello AI",
			},
			setupMock: func(r *MockJobRepository, q *MockJobQueue) {
				r.CreateErr = errors.New("db error")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewMockJobRepository()
			queue := NewMockJobQueue()
			tt.setupMock(repo, queue)

			svc := NewJobService(repo, queue)
			job, err := svc.CreateJob(context.Background(), tt.input)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if job == nil {
				t.Fatal("expected job, got nil")
			}

			if job.Status != domain.JobStatusPending {
				t.Errorf("expected status Pending, got %v", job.Status)
			}

			if queue.EnqueuedCount() != 1 {
				t.Errorf("expected 1 enqueued job, got %d", queue.EnqueuedCount())
			}
		})
	}
}

func TestJobService_GetJob(t *testing.T) {
	ctx := context.Background()

	t.Run("existing job", func(t *testing.T) {
		repo := NewMockJobRepository()
		queue := NewMockJobQueue()

		tenantID := uuid.New()
		job := domain.NewJob(tenantID, domain.JobTypeText, "Test input")
		_ = repo.Create(ctx, job)

		svc := NewJobService(repo, queue)
		found, err := svc.GetJob(ctx, job.ID)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if found.ID != job.ID {
			t.Errorf("expected job ID %v, got %v", job.ID, found.ID)
		}
	})

	t.Run("non-existing job", func(t *testing.T) {
		repo := NewMockJobRepository()
		queue := NewMockJobQueue()

		svc := NewJobService(repo, queue)
		_, err := svc.GetJob(ctx, uuid.New())

		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if !errors.Is(err, domain.ErrJobNotFound) {
			t.Errorf("expected ErrJobNotFound, got %v", err)
		}
	})
}

func TestJobService_ListJobs(t *testing.T) {
	ctx := context.Background()
	repo := NewMockJobRepository()
	queue := NewMockJobQueue()

	tenantID := uuid.New()
	otherTenantID := uuid.New()

	// Create jobs for different tenants
	for i := 0; i < 5; i++ {
		job := domain.NewJob(tenantID, domain.JobTypeText, "Input")
		_ = repo.Create(ctx, job)
	}
	for i := 0; i < 3; i++ {
		job := domain.NewJob(otherTenantID, domain.JobTypeText, "Input")
		_ = repo.Create(ctx, job)
	}

	svc := NewJobService(repo, queue)
	jobs, err := svc.ListJobs(ctx, tenantID, 10, 0)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(jobs) != 5 {
		t.Errorf("expected 5 jobs, got %d", len(jobs))
	}

	for _, job := range jobs {
		if job.TenantID != tenantID {
			t.Errorf("expected tenant ID %v, got %v", tenantID, job.TenantID)
		}
	}
}

func TestJobService_CountJobs(t *testing.T) {
	ctx := context.Background()
	repo := NewMockJobRepository()
	queue := NewMockJobQueue()

	tenantID := uuid.New()

	for i := 0; i < 7; i++ {
		job := domain.NewJob(tenantID, domain.JobTypeText, "Input")
		_ = repo.Create(ctx, job)
	}

	svc := NewJobService(repo, queue)
	count, err := svc.CountJobs(ctx, tenantID)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if count != 7 {
		t.Errorf("expected count 7, got %d", count)
	}
}

func TestJobService_CancelJob(t *testing.T) {
	ctx := context.Background()

	t.Run("cancel pending job", func(t *testing.T) {
		repo := NewMockJobRepository()
		queue := NewMockJobQueue()

		tenantID := uuid.New()
		job := domain.NewJob(tenantID, domain.JobTypeText, "Test")
		_ = repo.Create(ctx, job)

		svc := NewJobService(repo, queue)
		cancelled, err := svc.CancelJob(ctx, job.ID)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if cancelled.Status != domain.JobStatusFailed {
			t.Errorf("expected status Failed, got %v", cancelled.Status)
		}
	})

	t.Run("cannot cancel processing job", func(t *testing.T) {
		repo := NewMockJobRepository()
		queue := NewMockJobQueue()

		tenantID := uuid.New()
		job := domain.NewJob(tenantID, domain.JobTypeText, "Test")
		job.MarkProcessing("openai")
		_ = repo.Create(ctx, job)

		svc := NewJobService(repo, queue)
		_, err := svc.CancelJob(ctx, job.ID)

		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}
