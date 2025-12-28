// Package domain contains core business entities and value objects.
package domain

import (
	"time"

	"github.com/google/uuid"
)

// Usage represents resource consumption for a job.
type Usage struct {
	ID        uuid.UUID
	TenantID  uuid.UUID
	JobID     uuid.UUID
	Provider  string
	Model     string
	TokensIn  int
	TokensOut int
	Cost      float64
	CreatedAt time.Time
}

// NewUsage creates a new usage record.
func NewUsage(tenantID, jobID uuid.UUID, provider, model string, tokensIn, tokensOut int, cost float64) *Usage {
	return &Usage{
		ID:        uuid.New(),
		TenantID:  tenantID,
		JobID:     jobID,
		Provider:  provider,
		Model:     model,
		TokensIn:  tokensIn,
		TokensOut: tokensOut,
		Cost:      cost,
		CreatedAt: time.Now(),
	}
}

// TotalTokens returns the sum of input and output tokens.
func (u *Usage) TotalTokens() int {
	return u.TokensIn + u.TokensOut
}

// UsageSummary represents aggregated usage statistics.
type UsageSummary struct {
	TenantID       uuid.UUID
	Provider       string
	TotalTokensIn  int
	TotalTokensOut int
	TotalCost      float64
	JobCount       int
	Period         time.Time
}
