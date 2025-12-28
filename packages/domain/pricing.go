package domain

import (
	"time"

	"github.com/google/uuid"
)

// ProviderPricing represents pricing configuration for a provider model.
type ProviderPricing struct {
	ID                    uuid.UUID
	Provider              string
	Model                 string
	InputPricePerMillion  float64 // Price per 1 million input tokens
	OutputPricePerMillion float64 // Price per 1 million output tokens
	ImagePrice            *float64 // Price per image (for image models)
	IsDefault             bool
	CreatedAt             time.Time
	UpdatedAt             time.Time
}

// CalculateCost calculates the cost based on token usage.
func (p *ProviderPricing) CalculateCost(tokensIn, tokensOut int) float64 {
	inputCost := float64(tokensIn) * (p.InputPricePerMillion / 1_000_000)
	outputCost := float64(tokensOut) * (p.OutputPricePerMillion / 1_000_000)
	return inputCost + outputCost
}

// CalculateImageCost returns the image cost or 0 if not an image model.
func (p *ProviderPricing) CalculateImageCost() float64 {
	if p.ImagePrice != nil {
		return *p.ImagePrice
	}
	return 0
}
