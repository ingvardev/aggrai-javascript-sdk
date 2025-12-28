// Package providers contains AI provider implementations.
package providers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/ingvar/aiaggregator/packages/domain"
	"github.com/ingvar/aiaggregator/packages/usecases"
)

// ClaudeProvider implements AI provider using Anthropic Claude API.
type ClaudeProvider struct {
	apiKey   string
	model    string
	endpoint string
	client   *http.Client
}

// ClaudeConfig holds configuration for Claude provider.
type ClaudeConfig struct {
	APIKey   string
	Model    string
	Endpoint string
}

// NewClaudeProvider creates a new Claude provider.
func NewClaudeProvider(cfg ClaudeConfig) *ClaudeProvider {
	if cfg.Endpoint == "" {
		cfg.Endpoint = "https://api.anthropic.com/v1"
	}
	if cfg.Model == "" {
		cfg.Model = "claude-3-haiku-20240307"
	}

	return &ClaudeProvider{
		apiKey:   cfg.APIKey,
		model:    cfg.Model,
		endpoint: cfg.Endpoint,
		client:   &http.Client{},
	}
}

// Name returns the provider name.
func (p *ClaudeProvider) Name() string {
	return "claude"
}

// Type returns the provider type.
func (p *ClaudeProvider) Type() domain.ProviderType {
	return domain.ProviderTypeClaude
}

// claudeRequest represents Claude API request.
type claudeRequest struct {
	Model     string          `json:"model"`
	MaxTokens int             `json:"max_tokens"`
	Messages  []claudeMessage `json:"messages"`
}

type claudeMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type claudeResponse struct {
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	Usage struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
}

// Complete performs a text completion request.
func (p *ClaudeProvider) Complete(ctx context.Context, request *usecases.CompletionRequest) (*usecases.CompletionResponse, error) {
	maxTokens := request.MaxTokens
	if maxTokens == 0 {
		maxTokens = 2048
	}

	reqBody := claudeRequest{
		Model:     p.model,
		MaxTokens: maxTokens,
		Messages: []claudeMessage{
			{Role: "user", Content: request.Prompt},
		},
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", p.endpoint+"/messages", bytes.NewReader(jsonBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", p.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Claude API error: %s", string(body))
	}

	var claudeResp claudeResponse
	if err := json.NewDecoder(resp.Body).Decode(&claudeResp); err != nil {
		return nil, err
	}

	if len(claudeResp.Content) == 0 {
		return nil, fmt.Errorf("no response from Claude")
	}

	// Calculate cost (Haiku pricing)
	inputCost := float64(claudeResp.Usage.InputTokens) * 0.00000025
	outputCost := float64(claudeResp.Usage.OutputTokens) * 0.00000125
	totalCost := inputCost + outputCost

	return &usecases.CompletionResponse{
		Content:   claudeResp.Content[0].Text,
		Model:     p.model,
		TokensIn:  claudeResp.Usage.InputTokens,
		TokensOut: claudeResp.Usage.OutputTokens,
		Cost:      totalCost,
	}, nil
}

// GenerateImage is not supported by Claude.
func (p *ClaudeProvider) GenerateImage(ctx context.Context, request *usecases.ImageRequest) (*usecases.ImageResponse, error) {
	return nil, fmt.Errorf("Claude does not support image generation")
}

// IsAvailable checks if the provider is available.
func (p *ClaudeProvider) IsAvailable(ctx context.Context) bool {
	return p.apiKey != ""
}
