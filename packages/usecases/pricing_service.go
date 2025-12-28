package usecases

import (
	"context"
	"sync"

	"github.com/google/uuid"
	"github.com/ingvar/aiaggregator/packages/domain"
)

// PricingService provides pricing information for providers.
type PricingService struct {
	repo  PricingRepository
	cache map[string]*domain.ProviderPricing // provider:model -> pricing
	mu    sync.RWMutex
}

// NewPricingService creates a new PricingService.
func NewPricingService(repo PricingRepository) *PricingService {
	return &PricingService{
		repo:  repo,
		cache: make(map[string]*domain.ProviderPricing),
	}
}

// GetPricing returns pricing for a specific provider and model.
// Falls back to default pricing for provider if model not found.
func (s *PricingService) GetPricing(ctx context.Context, provider, model string) (*domain.ProviderPricing, error) {
	cacheKey := provider + ":" + model

	// Check cache first
	s.mu.RLock()
	if pricing, ok := s.cache[cacheKey]; ok {
		s.mu.RUnlock()
		return pricing, nil
	}
	s.mu.RUnlock()

	// Try to get exact match
	pricing, err := s.repo.GetByProviderModel(ctx, provider, model)
	if err != nil {
		return nil, err
	}

	// If not found, try default for provider
	if pricing == nil {
		pricing, err = s.repo.GetDefaultByProvider(ctx, provider)
		if err != nil {
			return nil, err
		}
	}

	// If still not found, return zero pricing
	if pricing == nil {
		pricing = &domain.ProviderPricing{
			Provider:              provider,
			Model:                 model,
			InputPricePerMillion:  0,
			OutputPricePerMillion: 0,
		}
	}

	// Cache the result
	s.mu.Lock()
	s.cache[cacheKey] = pricing
	s.mu.Unlock()

	return pricing, nil
}

// CalculateCost calculates cost for given provider, model and token counts.
func (s *PricingService) CalculateCost(ctx context.Context, provider, model string, tokensIn, tokensOut int) (float64, error) {
	pricing, err := s.GetPricing(ctx, provider, model)
	if err != nil {
		return 0, err
	}
	return pricing.CalculateCost(tokensIn, tokensOut), nil
}

// CalculateImageCost returns the image generation cost for a provider/model.
func (s *PricingService) CalculateImageCost(ctx context.Context, provider, model string) (float64, error) {
	pricing, err := s.GetPricing(ctx, provider, model)
	if err != nil {
		return 0, err
	}
	return pricing.CalculateImageCost(), nil
}

// InvalidateCache clears the pricing cache.
func (s *PricingService) InvalidateCache() {
	s.mu.Lock()
	s.cache = make(map[string]*domain.ProviderPricing)
	s.mu.Unlock()
}

// List returns all pricing configurations.
func (s *PricingService) List(ctx context.Context) ([]*domain.ProviderPricing, error) {
	return s.repo.List(ctx)
}

// ListByProvider returns pricing configurations for a specific provider.
func (s *PricingService) ListByProvider(ctx context.Context, provider string) ([]*domain.ProviderPricing, error) {
	return s.repo.ListByProvider(ctx, provider)
}

// Create creates a new pricing configuration.
func (s *PricingService) Create(ctx context.Context, pricing *domain.ProviderPricing) error {
	err := s.repo.Create(ctx, pricing)
	if err != nil {
		return err
	}
	s.InvalidateCache()
	return nil
}

// Update updates a pricing configuration.
func (s *PricingService) Update(ctx context.Context, pricing *domain.ProviderPricing) error {
	err := s.repo.Update(ctx, pricing)
	if err != nil {
		return err
	}
	s.InvalidateCache()
	return nil
}

// Delete deletes a pricing configuration.
func (s *PricingService) Delete(ctx context.Context, id uuid.UUID) error {
	err := s.repo.Delete(ctx, id)
	if err != nil {
		return err
	}
	s.InvalidateCache()
	return nil
}
