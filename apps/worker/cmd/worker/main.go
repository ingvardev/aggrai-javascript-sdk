// Package main is the entry point for the asynq worker.
package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"

	"github.com/ingvar/aiaggregator/apps/worker/internal/handlers"
	"github.com/ingvar/aiaggregator/packages/adapters"
	"github.com/ingvar/aiaggregator/packages/providers"
	"github.com/ingvar/aiaggregator/packages/pubsub"
	"github.com/ingvar/aiaggregator/packages/shared"
	"github.com/ingvar/aiaggregator/packages/usecases"
)

const (
	// TaskTypeProcessJob is the task type for processing jobs.
	TaskTypeProcessJob = "job:process"
)

func main() {
	// Load .env file if exists (try multiple locations)
	_ = godotenv.Load()           // Current directory
	_ = godotenv.Load("../../.env") // From apps/worker to root
	_ = godotenv.Load("../../../.env") // From apps/worker/cmd/worker to root

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

	// Initialize repositories based on configuration
	var jobRepo usecases.JobRepository
	var usageRepo usecases.UsageRepository

	// Try to connect to PostgreSQL
	pool, err := pgxpool.New(context.Background(), cfg.DatabaseURL)
	if err != nil {
		log.Warn().Err(err).Msg("PostgreSQL not available, using in-memory repositories")
		// Fallback to in-memory (warning: data won't be shared with API!)
		jobRepo = adapters.NewInMemoryJobRepository()
		usageRepo = adapters.NewInMemoryUsageRepository()
	} else {
		defer pool.Close()
		log.Info().Msg("Connected to PostgreSQL")
		// Use PostgreSQL repositories (shared with API)
		jobRepo = adapters.NewPostgresJobRepository(pool)
		usageRepo = adapters.NewPostgresUsageRepository(pool)
	}

	// Initialize pricing service
	var pricingService *usecases.PricingService
	if pool != nil {
		pricingRepo := adapters.NewPostgresPricingRepository(pool)
		pricingService = usecases.NewPricingService(pricingRepo)
		log.Info().Msg("Pricing service initialized")
	}

	// Initialize provider registry with available providers
	registry := providers.NewProviderRegistry()

	// Register stub provider only in dev mode
	if cfg.EnableStubProvider {
		registry.Register(providers.NewStubProvider("stub-provider"))
		log.Info().Msg("Stub provider registered (dev mode)")
	}

	// Register OpenAI if configured
	if cfg.OpenAIAPIKey != "" {
		openai := providers.NewOpenAIProvider(providers.OpenAIConfig{
			APIKey:         cfg.OpenAIAPIKey,
			PricingService: pricingService,
		})
		registry.Register(openai)
		log.Info().Msg("OpenAI provider registered")
	}

	// Register Claude if configured
	if cfg.AnthropicAPIKey != "" {
		claude := providers.NewClaudeProvider(providers.ClaudeConfig{
			APIKey:         cfg.AnthropicAPIKey,
			PricingService: pricingService,
		})
		registry.Register(claude)
		log.Info().Msg("Claude provider registered")
	}

	// Register Ollama if available
	ollamaProvider := providers.NewOllamaProvider(providers.OllamaConfig{
		Endpoint: cfg.OllamaURL,
	})
	if ollamaProvider.IsAvailable(context.Background()) {
		registry.Register(ollamaProvider)
		log.Info().Str("endpoint", cfg.OllamaURL).Msg("Ollama provider registered")
	}

	// Determine default provider
	defaultProvider := "stub-provider"
	if cfg.OpenAIAPIKey != "" {
		defaultProvider = "openai"
	} else if cfg.AnthropicAPIKey != "" {
		defaultProvider = "claude"
	} else if ollamaProvider.IsAvailable(context.Background()) {
		defaultProvider = "ollama"
	}
	log.Info().Str("default_provider", defaultProvider).Msg("Default provider selected")

	// Initialize Redis publisher for job updates
	publisher, err := pubsub.NewPublisher(cfg.RedisURL)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to create Redis publisher, subscriptions will not work")
	} else {
		defer publisher.Close()
		log.Info().Msg("Redis publisher initialized for job updates")
	}

	// Initialize process job service
	processService := usecases.NewProcessJobService(jobRepo, usageRepo, registry.Get)
	if publisher != nil {
		processService.SetPublisher(publisher)
	}

	// Create job handler
	jobHandler := handlers.NewJobHandler(processService, defaultProvider)

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
