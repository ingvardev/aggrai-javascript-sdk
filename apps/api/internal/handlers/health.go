// Package handlers contains HTTP handlers for the API.
package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

// HealthResponse represents the health check response.
type HealthResponse struct {
	Status    string                   `json:"status"`
	Service   string                   `json:"service"`
	Version   string                   `json:"version"`
	Timestamp string                   `json:"timestamp"`
	Checks    map[string]*HealthCheck  `json:"checks,omitempty"`
}

// HealthCheck represents an individual health check result.
type HealthCheck struct {
	Status  string `json:"status"`
	Latency string `json:"latency,omitempty"`
	Error   string `json:"error,omitempty"`
}

// HealthChecker provides health check endpoints with dependency verification.
type HealthChecker struct {
	pool       *pgxpool.Pool
	redisURL   string
	redisOnce  sync.Once
	redisClient *redis.Client
}

// NewHealthChecker creates a new health checker with optional dependencies.
func NewHealthChecker(pool *pgxpool.Pool, redisURL string) *HealthChecker {
	return &HealthChecker{
		pool:     pool,
		redisURL: redisURL,
	}
}

// getRedisClient lazily initializes Redis client.
func (h *HealthChecker) getRedisClient() *redis.Client {
	h.redisOnce.Do(func() {
		if h.redisURL != "" {
			opts, err := redis.ParseURL(h.redisURL)
			if err == nil {
				h.redisClient = redis.NewClient(opts)
			}
		}
	})
	return h.redisClient
}

// LiveHandler returns basic liveness status (for Kubernetes liveness probe).
// This should always return 200 if the process is running.
func (h *HealthChecker) LiveHandler(w http.ResponseWriter, r *http.Request) {
	response := HealthResponse{
		Status:    "alive",
		Service:   "ai-aggregator-api",
		Version:   "0.1.0",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// ReadyHandler checks if the service is ready to accept traffic (for Kubernetes readiness probe).
// This verifies all critical dependencies are available.
func (h *HealthChecker) ReadyHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	checks := make(map[string]*HealthCheck)
	allHealthy := true

	// Check PostgreSQL
	if h.pool != nil {
		checks["postgres"] = h.checkPostgres(ctx)
		if checks["postgres"].Status != "healthy" {
			allHealthy = false
		}
	}

	// Check Redis
	if h.redisURL != "" {
		checks["redis"] = h.checkRedis(ctx)
		if checks["redis"].Status != "healthy" {
			allHealthy = false
		}
	}

	status := "ready"
	httpStatus := http.StatusOK
	if !allHealthy {
		status = "not_ready"
		httpStatus = http.StatusServiceUnavailable
	}

	response := HealthResponse{
		Status:    status,
		Service:   "ai-aggregator-api",
		Version:   "0.1.0",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Checks:    checks,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatus)
	json.NewEncoder(w).Encode(response)
}

// HealthHandler returns comprehensive health status with all dependency checks.
func (h *HealthChecker) HealthHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	checks := make(map[string]*HealthCheck)
	allHealthy := true

	// Check PostgreSQL
	if h.pool != nil {
		checks["postgres"] = h.checkPostgres(ctx)
		if checks["postgres"].Status != "healthy" {
			allHealthy = false
		}
	} else {
		checks["postgres"] = &HealthCheck{Status: "disabled"}
	}

	// Check Redis
	if h.redisURL != "" {
		checks["redis"] = h.checkRedis(ctx)
		if checks["redis"].Status != "healthy" {
			allHealthy = false
		}
	} else {
		checks["redis"] = &HealthCheck{Status: "disabled"}
	}

	status := "healthy"
	httpStatus := http.StatusOK
	if !allHealthy {
		status = "unhealthy"
		httpStatus = http.StatusServiceUnavailable
	}

	response := HealthResponse{
		Status:    status,
		Service:   "ai-aggregator-api",
		Version:   "0.1.0",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Checks:    checks,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatus)
	json.NewEncoder(w).Encode(response)
}

func (h *HealthChecker) checkPostgres(ctx context.Context) *HealthCheck {
	start := time.Now()
	err := h.pool.Ping(ctx)
	latency := time.Since(start)

	if err != nil {
		return &HealthCheck{
			Status:  "unhealthy",
			Latency: latency.String(),
			Error:   err.Error(),
		}
	}

	return &HealthCheck{
		Status:  "healthy",
		Latency: latency.String(),
	}
}

func (h *HealthChecker) checkRedis(ctx context.Context) *HealthCheck {
	client := h.getRedisClient()
	if client == nil {
		return &HealthCheck{
			Status: "unhealthy",
			Error:  "failed to parse redis URL",
		}
	}

	start := time.Now()
	_, err := client.Ping(ctx).Result()
	latency := time.Since(start)

	if err != nil {
		return &HealthCheck{
			Status:  "unhealthy",
			Latency: latency.String(),
			Error:   err.Error(),
		}
	}

	return &HealthCheck{
		Status:  "healthy",
		Latency: latency.String(),
	}
}

// HealthHandler returns the health status of the API (legacy simple handler).
func HealthHandler(w http.ResponseWriter, r *http.Request) {
	response := HealthResponse{
		Status:    "healthy",
		Service:   "ai-aggregator-api",
		Version:   "0.1.0",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

