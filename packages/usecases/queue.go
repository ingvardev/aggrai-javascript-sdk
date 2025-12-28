// Package usecases contains application business logic and use case implementations.
package usecases

import (
	"context"

	"github.com/google/uuid"
)

// JobQueuePayload represents the payload for job queue messages.
type JobQueuePayload struct {
	JobID    uuid.UUID `json:"job_id"`
	TenantID uuid.UUID `json:"tenant_id"`
	Type     string    `json:"type"`
}

// JobQueue defines the interface for job queue operations.
type JobQueue interface {
	// Enqueue adds a job to the processing queue.
	Enqueue(ctx context.Context, jobID uuid.UUID) error
	// Close closes the queue connection.
	Close() error
}
