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
