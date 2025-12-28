// Package handlers contains task handlers for the worker.
package handlers

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"

	"github.com/ingvar/aiaggregator/packages/shared"
	"github.com/ingvar/aiaggregator/packages/usecases"
)

var log = shared.NewLogger("job-handler")

// JobPayload represents the job processing payload.
type JobPayload struct {
	JobID    uuid.UUID `json:"job_id"`
	TenantID uuid.UUID `json:"tenant_id"`
	Type     string    `json:"type"`
	Input    string    `json:"input"`
}

// JobHandler handles job processing tasks.
type JobHandler struct {
	processService  *usecases.ProcessJobService
	defaultProvider string
}

// NewJobHandler creates a new job handler.
func NewJobHandler(processService *usecases.ProcessJobService, defaultProvider string) *JobHandler {
	return &JobHandler{
		processService:  processService,
		defaultProvider: defaultProvider,
	}
}

// HandleProcessJob processes a job task.
func (h *JobHandler) HandleProcessJob(ctx context.Context, task *asynq.Task) error {
	var payload usecases.JobQueuePayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		log.Error().Err(err).Msg("Failed to unmarshal task payload")
		return err
	}

	log.Info().
		Str("job_id", payload.JobID.String()).
		Msg("Processing job")

	// Process the job using the default provider
	err := h.processService.ProcessJob(ctx, payload.JobID, h.defaultProvider)
	if err != nil {
		log.Error().Err(err).
			Str("job_id", payload.JobID.String()).
			Msg("Job processing failed")
		return err
	}

	log.Info().
		Str("job_id", payload.JobID.String()).
		Msg("Job completed successfully")

	return nil
}
