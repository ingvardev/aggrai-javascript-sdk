// Package providers contains AI provider implementations.
package providers

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/ingvar/aiaggregator/packages/domain"
	"github.com/ingvar/aiaggregator/packages/usecases"
)

// StubProvider is a test provider that returns mock responses.
type StubProvider struct {
	name      string
	available bool
	delay     time.Duration
}

// Ensure StubProvider implements AIProvider
var _ usecases.AIProvider = (*StubProvider)(nil)

// NewStubProvider creates a new stub provider.
func NewStubProvider(name string) *StubProvider {
	return &StubProvider{
		name:      name,
		available: true,
		delay:     100 * time.Millisecond,
	}
}

// Name returns the provider name.
func (p *StubProvider) Name() string {
	return p.name
}

// Type returns the provider type.
func (p *StubProvider) Type() string {
	return "LOCAL"
}

// IsAvailable checks if the provider is available.
func (p *StubProvider) IsAvailable(ctx context.Context) bool {
	return p.available
}

// SetAvailable sets the provider availability (for testing).
func (p *StubProvider) SetAvailable(available bool) {
	p.available = available
}

// SetDelay sets the processing delay (for testing).
func (p *StubProvider) SetDelay(delay time.Duration) {
	p.delay = delay
}

// Execute processes a job and returns a mock result.
func (p *StubProvider) Execute(ctx context.Context, job *domain.Job) (*usecases.ProviderResult, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(p.delay):
		// Continue processing
	}

	// Generate mock response based on job type
	var response string
	switch job.Type {
	case domain.JobTypeText:
		response = fmt.Sprintf("Stub response for input: %s", truncateInput(job.Input, 50))
	case domain.JobTypeImage:
		response = "https://example.com/stub-image.png"
	default:
		response = "Unknown job type"
	}

	// Generate mock token counts
	tokensIn := len(job.Input) / 4 // Rough approximation
	tokensOut := len(response) / 4

	// Random cost between 0.001 and 0.01
	cost := 0.001 + rand.Float64()*0.009

	return &usecases.ProviderResult{
		Result:    response,
		TokensIn:  tokensIn,
		TokensOut: tokensOut,
		Cost:      cost,
		Model:     "stub-model-v1",
	}, nil
}

// truncateInput truncates input to a maximum length.
func truncateInput(input string, maxLen int) string {
	if len(input) <= maxLen {
		return input
	}
	return input[:maxLen] + "..."
}
