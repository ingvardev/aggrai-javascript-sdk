// Package usecases contains application business logic and use case implementations.
package usecases

import (
	"context"

	"github.com/google/uuid"
	"github.com/ingvar/aiaggregator/packages/domain"
)

// AIProvider defines the interface for AI provider drivers.
// This is the simplified interface used for job processing.
type AIProvider interface {
	// Name returns the provider name.
	Name() string
	// Type returns the provider type.
	Type() string
	// Execute processes a job and returns the result.
	Execute(ctx context.Context, job *domain.Job) (*ProviderResult, error)
	// IsAvailable checks if the provider is currently available.
	IsAvailable(ctx context.Context) bool
}

// ProviderResult represents the result from an AI provider.
type ProviderResult struct {
	Result    string
	TokensIn  int
	TokensOut int
	Cost      float64
	Model     string
}

// CompletionRequest represents a text completion request.
type CompletionRequest struct {
	JobID     uuid.UUID
	Prompt    string
	Model     string
	MaxTokens int
	Options   map[string]interface{}
}

// CompletionResponse represents a text completion response.
type CompletionResponse struct {
	Content   string
	Model     string
	TokensIn  int
	TokensOut int
	Cost      float64
}

// ImageRequest represents an image generation request.
type ImageRequest struct {
	JobID   uuid.UUID
	Prompt  string
	Model   string
	Size    string
	Options map[string]interface{}
}

// ImageResponse represents an image generation response.
type ImageResponse struct {
	URL   string
	Model string
	Cost  float64
}

// ProviderSelector selects the best available provider for a request.
type ProviderSelector interface {
	SelectProvider(ctx context.Context, jobType domain.JobType) (AIProvider, error)
}
