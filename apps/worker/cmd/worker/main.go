// Package main is the entry point for the asynq worker.
package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/hibiken/asynq"
	"github.com/joho/godotenv"

	"github.com/ingvar/aiaggregator/apps/worker/internal/handlers"
	"github.com/ingvar/aiaggregator/packages/providers"
	"github.com/ingvar/aiaggregator/packages/shared"
)

const (
	// TaskTypeProcessJob is the task type for processing jobs.
	TaskTypeProcessJob = "job:process"
)

func main() {
	// Load .env file if exists
	_ = godotenv.Load()

	log := shared.NewLogger("worker")
	cfg := shared.LoadConfig()

	log.Info().
		Str("redis", cfg.RedisURL).
		Msg("Starting AI Aggregator Worker")

	// Parse Redis URL
	redisOpt, err := asynq.ParseRedisURI(cfg.RedisURL)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to parse Redis URL")
	}

	// Initialize provider registry
	registry := providers.NewProviderRegistry()
	registry.Register(providers.NewStubProvider("stub-provider"))

	// Create job handler
	jobHandler := handlers.NewJobHandler(registry)

	// Create asynq server
	srv := asynq.NewServer(
		redisOpt,
		asynq.Config{
			Concurrency: 10,
			Queues: map[string]int{
				"critical": 6,
				"default":  3,
				"low":      1,
			},
			ErrorHandler: asynq.ErrorHandlerFunc(func(ctx context.Context, task *asynq.Task, err error) {
				log.Error().
					Err(err).
					Str("type", task.Type()).
					Msg("Task processing failed")
			}),
		},
	)

	// Register handlers
	mux := asynq.NewServeMux()
	mux.HandleFunc(TaskTypeProcessJob, jobHandler.HandleProcessJob)

	// Graceful shutdown
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := srv.Run(mux); err != nil {
			log.Fatal().Err(err).Msg("Failed to start worker")
		}
	}()

	log.Info().Msg("Worker is running")

	<-done
	log.Info().Msg("Shutting down worker...")

	srv.Shutdown()

	log.Info().Msg("Worker exited properly")
}
