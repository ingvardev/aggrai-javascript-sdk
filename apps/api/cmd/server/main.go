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
	"github.com/joho/godotenv"

	"github.com/ingvar/aiaggregator/apps/api/internal/graph"
	"github.com/ingvar/aiaggregator/apps/api/internal/handlers"
	appMiddleware "github.com/ingvar/aiaggregator/apps/api/internal/middleware"
	"github.com/ingvar/aiaggregator/packages/adapters"
	"github.com/ingvar/aiaggregator/packages/providers"
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

	// Initialize in-memory repositories (for development)
	// TODO: Switch to Postgres adapters in production
	jobRepo := adapters.NewInMemoryJobRepository()
	tenantRepo := adapters.NewInMemoryTenantRepository()
	usageRepo := adapters.NewInMemoryUsageRepository()

	// Seed test tenant for development
	testTenant := tenantRepo.SeedTestTenant()
	log.Info().
		Str("tenant_id", testTenant.ID.String()).
		Str("api_key", testTenant.APIKey).
		Msg("Test tenant created")

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

	// Initialize provider registry
	providerRegistry := providers.NewProviderRegistry()
	providerRegistry.Register(providers.NewStubProvider("stub-provider"))
	log.Info().Msg("Stub provider registered")

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
	graphResolver := graph.NewResolver(jobService, authService, providerRegistry)
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
		Str("test_api_key", testTenant.APIKey).
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
