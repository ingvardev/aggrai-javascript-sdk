// Package domain contains core business entities and value objects.
// These types have no external dependencies and represent the heart of the application.
package domain

import (
	"time"

	"github.com/google/uuid"
)

// JobStatus represents the current state of a job.
type JobStatus string

const (
	JobStatusPending    JobStatus = "pending"
	JobStatusProcessing JobStatus = "processing"
	JobStatusCompleted  JobStatus = "completed"
	JobStatusFailed     JobStatus = "failed"
)

// JobType represents the type of AI request.
type JobType string

const (
	JobTypeText  JobType = "text"
	JobTypeImage JobType = "image"
)

// Job represents an AI processing request.
type Job struct {
	ID         uuid.UUID
	TenantID   uuid.UUID
	Type       JobType
	Input      string
	Status     JobStatus
	Result     *string
	Error      *string
	Provider   *string
	TokensIn   int
	TokensOut  int
	Cost       float64
	CreatedAt  time.Time
	UpdatedAt  time.Time
	StartedAt  *time.Time
	FinishedAt *time.Time
}

// NewJob creates a new job with pending status.
func NewJob(tenantID uuid.UUID, jobType JobType, input string) *Job {
	now := time.Now()
	return &Job{
		ID:        uuid.New(),
		TenantID:  tenantID,
		Type:      jobType,
		Input:     input,
		Status:    JobStatusPending,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// MarkProcessing marks the job as being processed.
func (j *Job) MarkProcessing(provider string) {
	j.Status = JobStatusProcessing
	j.Provider = &provider
	now := time.Now()
	j.StartedAt = &now
	j.UpdatedAt = now
}

// MarkCompleted marks the job as successfully completed.
func (j *Job) MarkCompleted(result string, tokensIn, tokensOut int, cost float64) {
	j.Status = JobStatusCompleted
	j.Result = &result
	j.TokensIn = tokensIn
	j.TokensOut = tokensOut
	j.Cost = cost
	now := time.Now()
	j.FinishedAt = &now
	j.UpdatedAt = now
}

// MarkFailed marks the job as failed.
func (j *Job) MarkFailed(errMsg string) {
	j.Status = JobStatusFailed
	j.Error = &errMsg
	now := time.Now()
	j.FinishedAt = &now
	j.UpdatedAt = now
}

// IsTerminal returns true if the job is in a terminal state.
func (j *Job) IsTerminal() bool {
	return j.Status == JobStatusCompleted || j.Status == JobStatusFailed
}
