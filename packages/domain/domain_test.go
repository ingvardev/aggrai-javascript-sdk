package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestNewJob(t *testing.T) {
	tenantID := uuid.New()
	job := NewJob(tenantID, JobTypeText, "Test input")

	if job.ID == uuid.Nil {
		t.Error("expected job ID to be set")
	}

	if job.TenantID != tenantID {
		t.Errorf("expected tenant ID %v, got %v", tenantID, job.TenantID)
	}

	if job.Type != JobTypeText {
		t.Errorf("expected type %v, got %v", JobTypeText, job.Type)
	}

	if job.Input != "Test input" {
		t.Errorf("expected input %q, got %q", "Test input", job.Input)
	}

	if job.Status != JobStatusPending {
		t.Errorf("expected status %v, got %v", JobStatusPending, job.Status)
	}

	if job.CreatedAt.IsZero() {
		t.Error("expected CreatedAt to be set")
	}
}

func TestJob_MarkProcessing(t *testing.T) {
	job := NewJob(uuid.New(), JobTypeText, "Test")

	job.MarkProcessing("openai")

	if job.Status != JobStatusProcessing {
		t.Errorf("expected status %v, got %v", JobStatusProcessing, job.Status)
	}

	if job.Provider == nil || *job.Provider != "openai" {
		t.Errorf("expected provider %q, got %v", "openai", job.Provider)
	}

	if job.StartedAt == nil {
		t.Error("expected StartedAt to be set")
	}
}

func TestJob_MarkCompleted(t *testing.T) {
	job := NewJob(uuid.New(), JobTypeText, "Test")
	job.MarkProcessing("openai")

	job.MarkCompleted("AI response", 100, 200, 0.5)

	if job.Status != JobStatusCompleted {
		t.Errorf("expected status %v, got %v", JobStatusCompleted, job.Status)
	}

	if job.Result == nil || *job.Result != "AI response" {
		t.Errorf("expected result %q, got %v", "AI response", job.Result)
	}

	if job.TokensIn != 100 {
		t.Errorf("expected TokensIn %d, got %d", 100, job.TokensIn)
	}

	if job.TokensOut != 200 {
		t.Errorf("expected TokensOut %d, got %d", 200, job.TokensOut)
	}

	if job.Cost != 0.5 {
		t.Errorf("expected Cost %f, got %f", 0.5, job.Cost)
	}

	if job.FinishedAt == nil {
		t.Error("expected FinishedAt to be set")
	}
}

func TestJob_MarkFailed(t *testing.T) {
	job := NewJob(uuid.New(), JobTypeText, "Test")
	job.MarkProcessing("openai")

	job.MarkFailed("something went wrong")

	if job.Status != JobStatusFailed {
		t.Errorf("expected status %v, got %v", JobStatusFailed, job.Status)
	}

	if job.Error == nil || *job.Error != "something went wrong" {
		t.Errorf("expected error %q, got %v", "something went wrong", job.Error)
	}

	if job.FinishedAt == nil {
		t.Error("expected FinishedAt to be set")
	}
}

func TestJob_IsTerminal(t *testing.T) {
	tests := []struct {
		name     string
		status   JobStatus
		expected bool
	}{
		{"pending", JobStatusPending, false},
		{"processing", JobStatusProcessing, false},
		{"completed", JobStatusCompleted, true},
		{"failed", JobStatusFailed, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			job := NewJob(uuid.New(), JobTypeText, "Test")
			job.Status = tt.status

			if job.IsTerminal() != tt.expected {
				t.Errorf("expected IsTerminal() = %v, got %v", tt.expected, job.IsTerminal())
			}
		})
	}
}

func TestNewTenant(t *testing.T) {
	tenant := NewTenant("Test Corp", "api-key-123")

	if tenant.ID == uuid.Nil {
		t.Error("expected tenant ID to be set")
	}

	if tenant.Name != "Test Corp" {
		t.Errorf("expected name %q, got %q", "Test Corp", tenant.Name)
	}

	if tenant.APIKey != "api-key-123" {
		t.Errorf("expected API key %q, got %q", "api-key-123", tenant.APIKey)
	}

	if !tenant.Active {
		t.Error("expected tenant to be active")
	}

	if tenant.CreatedAt.IsZero() {
		t.Error("expected CreatedAt to be set")
	}
}

func TestTenant_Deactivate(t *testing.T) {
	tenant := NewTenant("Test Corp", "api-key")

	if !tenant.Active {
		t.Error("expected tenant to be active initially")
	}

	oldUpdatedAt := tenant.UpdatedAt
	time.Sleep(1 * time.Millisecond)
	tenant.Deactivate()

	if tenant.Active {
		t.Error("expected tenant to be inactive")
	}

	if !tenant.UpdatedAt.After(oldUpdatedAt) {
		t.Error("expected UpdatedAt to be updated")
	}
}

func TestNewUsage(t *testing.T) {
	tenantID := uuid.New()
	jobID := uuid.New()

	usage := NewUsage(tenantID, jobID, "openai", "gpt-4", 100, 200, 0.5)

	if usage.ID == uuid.Nil {
		t.Error("expected usage ID to be set")
	}

	if usage.TenantID != tenantID {
		t.Errorf("expected tenant ID %v, got %v", tenantID, usage.TenantID)
	}

	if usage.JobID != jobID {
		t.Errorf("expected job ID %v, got %v", jobID, usage.JobID)
	}

	if usage.Provider != "openai" {
		t.Errorf("expected provider %q, got %q", "openai", usage.Provider)
	}

	if usage.Model != "gpt-4" {
		t.Errorf("expected model %q, got %q", "gpt-4", usage.Model)
	}

	if usage.TokensIn != 100 {
		t.Errorf("expected input tokens %d, got %d", 100, usage.TokensIn)
	}

	if usage.TokensOut != 200 {
		t.Errorf("expected output tokens %d, got %d", 200, usage.TokensOut)
	}

	if usage.TotalTokens() != 300 {
		t.Errorf("expected total tokens %d, got %d", 300, usage.TotalTokens())
	}
}
