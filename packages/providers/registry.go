// Package providers contains AI provider implementations.
package providers

import (
	"context"
	"sync"

	"github.com/ingvar/aiaggregator/packages/domain"
	"github.com/ingvar/aiaggregator/packages/usecases"
)

// ProviderRegistry manages available AI providers.
type ProviderRegistry struct {
	mu        sync.RWMutex
	providers map[string]usecases.AIProvider
}

// NewProviderRegistry creates a new provider registry.
func NewProviderRegistry() *ProviderRegistry {
	return &ProviderRegistry{
		providers: make(map[string]usecases.AIProvider),
	}
}

// Register adds a provider to the registry.
func (r *ProviderRegistry) Register(provider usecases.AIProvider) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.providers[provider.Name()] = provider
}

// Get retrieves a provider by name.
func (r *ProviderRegistry) Get(name string) (usecases.AIProvider, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	provider, ok := r.providers[name]
	return provider, ok
}

// List returns all registered providers.
func (r *ProviderRegistry) List() []usecases.AIProvider {
	r.mu.RLock()
	defer r.mu.RUnlock()

	providers := make([]usecases.AIProvider, 0, len(r.providers))
	for _, p := range r.providers {
		providers = append(providers, p)
	}
	return providers
}

// SelectProvider selects the best available provider for the job type.
func (r *ProviderRegistry) SelectProvider(ctx context.Context, jobType domain.JobType) (usecases.AIProvider, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, provider := range r.providers {
		if provider.IsAvailable(ctx) {
			return provider, nil
		}
	}

	return nil, domain.ErrProviderUnavailable
}
