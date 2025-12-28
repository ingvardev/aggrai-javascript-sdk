// Package main is the entry point for the GraphQL API server.
package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"

	"github.com/ingvar/aiaggregator/apps/api/internal/graph"
	"github.com/ingvar/aiaggregator/apps/api/internal/handlers"
	appMiddleware "github.com/ingvar/aiaggregator/apps/api/internal/middleware"
	"github.com/ingvar/aiaggregator/packages/adapters"
	"github.com/ingvar/aiaggregator/packages/providers"
	"github.com/ingvar/aiaggregator/packages/pubsub"
	"github.com/ingvar/aiaggregator/packages/queue"
	"github.com/ingvar/aiaggregator/packages/shared"
	"github.com/ingvar/aiaggregator/packages/usecases"
)

func main() {
	// Load .env file if exists
	_ = godotenv.Load()

	log := shared.NewLogger("api")
	cfg := shared.LoadConfig()

	log.Info().
		Str("port", cfg.APIPort).
		Msg("Starting AI Aggregator API server")

	// Initialize repositories based on configuration
	var jobRepo usecases.JobRepository
	var tenantRepo usecases.TenantRepository
	var usageRepo usecases.UsageRepository
	var testAPIKey string

	// Try to connect to PostgreSQL
	pool, err := pgxpool.New(context.Background(), cfg.DatabaseURL)
	if err != nil {
		log.Warn().Err(err).Msg("PostgreSQL not available, using in-memory repositories")
		// Fallback to in-memory for development
		memJobRepo := adapters.NewInMemoryJobRepository()
		memTenantRepo := adapters.NewInMemoryTenantRepository()
		memUsageRepo := adapters.NewInMemoryUsageRepository()
		jobRepo = memJobRepo
		tenantRepo = memTenantRepo
		usageRepo = memUsageRepo
		// Seed test tenant for development
		testTenant := memTenantRepo.SeedTestTenant()
		testAPIKey = testTenant.APIKey
		log.Info().
			Str("tenant_id", testTenant.ID.String()).
			Str("api_key", testTenant.APIKey).
			Msg("Test tenant created (in-memory)")
	} else {
		defer pool.Close()
		log.Info().Msg("Connected to PostgreSQL")
		// Use PostgreSQL repositories
		jobRepo = adapters.NewPostgresJobRepository(pool)
		tenantRepo = adapters.NewPostgresTenantRepository(pool)
		usageRepo = adapters.NewPostgresUsageRepository(pool)
		testAPIKey = "see database"
		log.Info().Msg("Using PostgreSQL repositories")
	}

	// Initialize pricing repository and service
	var pricingService *usecases.PricingService
	if pool != nil {
		pricingRepo := adapters.NewPostgresPricingRepository(pool)
		pricingService = usecases.NewPricingService(pricingRepo)
		log.Info().Msg("Pricing service initialized")
	}

	// Initialize job queue (optional - gracefully handle Redis unavailability)
	var jobQueue usecases.JobQueue
	q, err := queue.NewAsynqQueue(cfg.RedisURL)
	if err != nil {
		log.Warn().Err(err).Msg("Redis not available, job queue disabled")
	} else {
		jobQueue = q
		defer q.Close()
		log.Info().Msg("Connected to Redis")
	}

	// Initialize Redis subscriber for job updates (for GraphQL subscriptions)
	subscriber, err := pubsub.NewSubscriber(cfg.RedisURL)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to create Redis subscriber, subscriptions will not receive updates")
	} else {
		defer subscriber.Close()

		// Start listening for job updates
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		updates, err := subscriber.Subscribe(ctx)
		if err != nil {
			log.Error().Err(err).Msg("Failed to subscribe to job updates")
		} else {
			log.Info().Msg("Subscribed to Redis for job updates")

			// Create a callback to get usage data for a tenant
			getUsageFunc := func(tenantIDStr string) []*graph.UsageSummary {
				tenantID, err := uuid.Parse(tenantIDStr)
				if err != nil {
					return nil
				}
				usage, err := usageRepo.GetSummary(context.Background(), tenantID)
				if err != nil {
					return nil
				}
				result := make([]*graph.UsageSummary, len(usage))
				for i, u := range usage {
					result[i] = &graph.UsageSummary{
						Provider:       u.Provider,
						TotalTokensIn:  u.TotalTokensIn,
						TotalTokensOut: u.TotalTokensOut,
						TotalCost:      u.TotalCost,
						JobCount:       u.JobCount,
					}
				}
				return result
			}

			// Forward Redis updates to GraphQL subscriptions
			go func() {
				for update := range updates {
					graph.JobPubSub.HandleRedisUpdateWithUsage(update, getUsageFunc)
				}
			}()
		}
	}

	// Initialize provider registry with available providers
	providerRegistry := providers.NewProviderRegistry()

	// Always register stub provider for testing
	providerRegistry.Register(providers.NewStubProvider("stub-provider"))
	log.Info().Msg("Stub provider registered")

	// Register OpenAI if configured
	if cfg.OpenAIAPIKey != "" {
		openai := providers.NewOpenAIProvider(providers.OpenAIConfig{
			APIKey:         cfg.OpenAIAPIKey,
			PricingService: pricingService,
		})
		providerRegistry.Register(openai)
		log.Info().Msg("OpenAI provider registered")
	}

	// Register Claude if configured
	if cfg.AnthropicAPIKey != "" {
		claude := providers.NewClaudeProvider(providers.ClaudeConfig{
			APIKey:         cfg.AnthropicAPIKey,
			PricingService: pricingService,
		})
		providerRegistry.Register(claude)
		log.Info().Msg("Claude provider registered")
	}

	// Register Ollama if available (check at startup)
	ollamaProvider := providers.NewOllamaProvider(providers.OllamaConfig{
		Endpoint: cfg.OllamaURL,
	})
	if ollamaProvider.IsAvailable(context.Background()) {
		providerRegistry.Register(ollamaProvider)
		log.Info().Str("endpoint", cfg.OllamaURL).Msg("Ollama provider registered")
	} else {
		log.Debug().Str("endpoint", cfg.OllamaURL).Msg("Ollama not available, skipping")
	}

	// Initialize services
	authService := usecases.NewAuthService(tenantRepo)
	jobService := usecases.NewJobService(jobRepo, jobQueue)

	// Process job service for background processing
	_ = usecases.NewProcessJobService(jobRepo, usageRepo, providerRegistry.Get)

	// Initialize router
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	// CORS
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000", "http://localhost:*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-API-Key"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Health check (public)
	r.Get("/health", handlers.HealthHandler)

	// GraphQL playground (public in dev)
	if cfg.EnablePlayground {
		r.Get("/", handlers.PlaygroundHandler("/graphql"))
		r.Get("/playground", handlers.PlaygroundHandler("/graphql"))
	}

	// GraphQL endpoint with auth middleware
	graphResolver := graph.NewResolver(jobService, authService, tenantRepo, usageRepo, pricingService, providerRegistry)
	graphServer := graph.NewServer(graphResolver)

	// Apply auth middleware for GraphQL
	r.Group(func(r chi.Router) {
		r.Use(appMiddleware.AuthMiddleware(authService))
		r.Handle("/graphql", graphServer)
	})

	// Create server
	srv := &http.Server{
		Addr:         fmt.Sprintf("%s:%s", cfg.APIHost, cfg.APIPort),
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("Failed to start server")
		}
	}()

	log.Info().
		Str("addr", srv.Addr).
		Str("playground", fmt.Sprintf("http://localhost:%s/playground", cfg.APIPort)).
		Str("test_api_key", testAPIKey).
		Msg("Server is running")

	<-done
	log.Info().Msg("Shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatal().Err(err).Msg("Server forced to shutdown")
	}

	log.Info().Msg("Server exited properly")
}
