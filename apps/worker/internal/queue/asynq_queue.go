// Package queue provides the job queue implementation using asynq.
package queue

import (
	"context"
	"encoding/json"

	"github.com/hibiken/asynq"

	"github.com/ingvar/aiaggregator/packages/usecases"
)

const (
	// TaskTypeProcessJob is the task type for processing jobs.
	TaskTypeProcessJob = "job:process"
)

// AsynqQueue implements the JobQueue interface using asynq.
type AsynqQueue struct {
	client *asynq.Client
}

// NewAsynqQueue creates a new asynq-based job queue.
func NewAsynqQueue(redisURL string) (*AsynqQueue, error) {
	opt, err := asynq.ParseRedisURI(redisURL)
	if err != nil {
		return nil, err
	}

	client := asynq.NewClient(opt)
	return &AsynqQueue{client: client}, nil
}

// Enqueue adds a job to the processing queue.
func (q *AsynqQueue) Enqueue(ctx context.Context, payload *usecases.JobQueuePayload) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	task := asynq.NewTask(TaskTypeProcessJob, data)
	_, err = q.client.EnqueueContext(ctx, task)
	return err
}

// Close closes the queue connection.
func (q *AsynqQueue) Close() error {
	return q.client.Close()
}
