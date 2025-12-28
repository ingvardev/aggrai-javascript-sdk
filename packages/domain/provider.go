// Package domain contains core business entities and value objects.
package domain

import (
	"time"

	"github.com/google/uuid"
)

// ProviderType represents the type of AI provider.
type ProviderType string

const (
	ProviderTypeOpenAI ProviderType = "openai"
	ProviderTypeClaude ProviderType = "claude"
	ProviderTypeLocal  ProviderType = "local"
	ProviderTypeOllama ProviderType = "ollama"
)

// Provider represents an AI provider configuration.
type Provider struct {
	ID        uuid.UUID
	Name      string
	Type      ProviderType
	Endpoint  string
	APIKey    string
	Model     string
	Enabled   bool
	Priority  int
	Config    map[string]interface{}
	CreatedAt time.Time
	UpdatedAt time.Time
}

// NewProvider creates a new provider.
func NewProvider(name string, providerType ProviderType, endpoint, apiKey, model string) *Provider {
	now := time.Now()
	return &Provider{
		ID:        uuid.New(),
		Name:      name,
		Type:      providerType,
		Endpoint:  endpoint,
		APIKey:    apiKey,
		Model:     model,
		Enabled:   true,
		Priority:  0,
		Config:    make(map[string]interface{}),
		CreatedAt: now,
		UpdatedAt: now,
	}
}
