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

// OpenAIProvider implements AI provider using OpenAI API.
type OpenAIProvider struct {
	apiKey   string
	model    string
	endpoint string
	client   *http.Client
}

// OpenAIConfig holds configuration for OpenAI provider.
type OpenAIConfig struct {
	APIKey   string
	Model    string
	Endpoint string
}

// NewOpenAIProvider creates a new OpenAI provider.
func NewOpenAIProvider(cfg OpenAIConfig) *OpenAIProvider {
	if cfg.Endpoint == "" {
		cfg.Endpoint = "https://api.openai.com/v1"
	}
	if cfg.Model == "" {
		cfg.Model = "gpt-4o-mini"
	}

	return &OpenAIProvider{
		apiKey:   cfg.APIKey,
		model:    cfg.Model,
		endpoint: cfg.Endpoint,
		client:   &http.Client{},
	}
}

// Name returns the provider name.
func (p *OpenAIProvider) Name() string {
	return "openai"
}

// Type returns the provider type.
func (p *OpenAIProvider) Type() domain.ProviderType {
	return domain.ProviderTypeOpenAI
}

// openAIChatRequest represents OpenAI chat completion request.
type openAIChatRequest struct {
	Model     string              `json:"model"`
	Messages  []openAIChatMessage `json:"messages"`
	MaxTokens int                 `json:"max_tokens,omitempty"`
}

type openAIChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openAIChatResponse struct {
	ID      string `json:"id"`
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
	} `json:"usage"`
}

// Complete performs a text completion request.
func (p *OpenAIProvider) Complete(ctx context.Context, request *usecases.CompletionRequest) (*usecases.CompletionResponse, error) {
	reqBody := openAIChatRequest{
		Model: p.model,
		Messages: []openAIChatMessage{
			{Role: "user", Content: request.Prompt},
		},
		MaxTokens: request.MaxTokens,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", p.endpoint+"/chat/completions", bytes.NewReader(jsonBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("OpenAI API error: %s", string(body))
	}

	var openAIResp openAIChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&openAIResp); err != nil {
		return nil, err
	}

	if len(openAIResp.Choices) == 0 {
		return nil, fmt.Errorf("no response from OpenAI")
	}

	// Calculate cost (approximate pricing)
	inputCost := float64(openAIResp.Usage.PromptTokens) * 0.00000015
	outputCost := float64(openAIResp.Usage.CompletionTokens) * 0.0000006
	totalCost := inputCost + outputCost

	return &usecases.CompletionResponse{
		Content:   openAIResp.Choices[0].Message.Content,
		Model:     p.model,
		TokensIn:  openAIResp.Usage.PromptTokens,
		TokensOut: openAIResp.Usage.CompletionTokens,
		Cost:      totalCost,
	}, nil
}

// GenerateImage generates an image from a prompt.
func (p *OpenAIProvider) GenerateImage(ctx context.Context, request *usecases.ImageRequest) (*usecases.ImageResponse, error) {
	reqBody := map[string]interface{}{
		"model":  "dall-e-3",
		"prompt": request.Prompt,
		"n":      1,
		"size":   request.Size,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", p.endpoint+"/images/generations", bytes.NewReader(jsonBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("OpenAI API error: %s", string(body))
	}

	var result struct {
		Data []struct {
			URL string `json:"url"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if len(result.Data) == 0 {
		return nil, fmt.Errorf("no image generated")
	}

	return &usecases.ImageResponse{
		URL:   result.Data[0].URL,
		Model: "dall-e-3",
		Cost:  0.04, // DALL-E 3 standard pricing
	}, nil
}

// IsAvailable checks if the provider is available.
func (p *OpenAIProvider) IsAvailable(ctx context.Context) bool {
	return p.apiKey != ""
}
