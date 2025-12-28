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
	apiKey         string
	model          string
	endpoint       string
	client         *http.Client
	pricingService *usecases.PricingService
}

// ClaudeConfig holds configuration for Claude provider.
type ClaudeConfig struct {
	APIKey         string
	Model          string
	Endpoint       string
	PricingService *usecases.PricingService
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
		apiKey:         cfg.APIKey,
		model:          cfg.Model,
		endpoint:       cfg.Endpoint,
		client:         &http.Client{},
		pricingService: cfg.PricingService,
	}
}

// Name returns the provider name.
func (p *ClaudeProvider) Name() string {
	return "claude"
}

// Type returns the provider type.
func (p *ClaudeProvider) Type() string {
	return string(domain.ProviderTypeClaude)
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

	// Calculate cost using pricing service or fallback to defaults
	var totalCost float64
	if p.pricingService != nil {
		cost, err := p.pricingService.CalculateCost(ctx, "claude", p.model, claudeResp.Usage.InputTokens, claudeResp.Usage.OutputTokens)
		if err == nil {
			totalCost = cost
		}
	}
	if totalCost == 0 {
		// Fallback to default Haiku pricing
		inputCost := float64(claudeResp.Usage.InputTokens) * 0.00000025
		outputCost := float64(claudeResp.Usage.OutputTokens) * 0.00000125
		totalCost = inputCost + outputCost
	}

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

// Execute processes a job and returns the result.
// This is the main interface method for AIProvider.
func (p *ClaudeProvider) Execute(ctx context.Context, job *domain.Job) (*usecases.ProviderResult, error) {
	switch job.Type {
	case domain.JobTypeText:
		resp, err := p.Complete(ctx, &usecases.CompletionRequest{
			JobID:     job.ID,
			Prompt:    job.Input,
			MaxTokens: 2048,
		})
		if err != nil {
			return nil, err
		}
		return &usecases.ProviderResult{
			Result:    resp.Content,
			TokensIn:  resp.TokensIn,
			TokensOut: resp.TokensOut,
			Cost:      resp.Cost,
			Model:     resp.Model,
		}, nil

	case domain.JobTypeImage:
		return nil, fmt.Errorf("Claude does not support image generation")

	default:
		return nil, fmt.Errorf("unsupported job type: %s", job.Type)
	}
}

// Ensure ClaudeProvider implements AIProvider
var _ usecases.AIProvider = (*ClaudeProvider)(nil)

// Ensure ClaudeProvider implements StreamingProvider
var _ usecases.StreamingProvider = (*ClaudeProvider)(nil)

// Ensure ClaudeProvider implements ModelListProvider
var _ usecases.ModelListProvider = (*ClaudeProvider)(nil)

// ListModels returns a list of available Claude models.
// Since Anthropic doesn't have a public models list API, we return a static list.
func (p *ClaudeProvider) ListModels(ctx context.Context) ([]usecases.ModelInfo, error) {
	// Static list of Claude models - Anthropic doesn't provide a models API
	models := []usecases.ModelInfo{
		{ID: "claude-sonnet-4-20250514", Name: "Claude Sonnet 4", Description: "Latest Claude Sonnet 4 model"},
		{ID: "claude-3-5-sonnet-20241022", Name: "Claude 3.5 Sonnet", Description: "Most intelligent Claude 3.5 model"},
		{ID: "claude-3-5-haiku-20241022", Name: "Claude 3.5 Haiku", Description: "Fastest Claude 3.5 model"},
		{ID: "claude-3-opus-20240229", Name: "Claude 3 Opus", Description: "Most powerful Claude 3 model"},
		{ID: "claude-3-sonnet-20240229", Name: "Claude 3 Sonnet", Description: "Balanced Claude 3 model"},
		{ID: "claude-3-haiku-20240307", Name: "Claude 3 Haiku", Description: "Fast Claude 3 model"},
	}
	return models, nil
}

// claudeStreamRequest represents a streaming message request to Claude API.
type claudeStreamRequest struct {
	Model     string               `json:"model"`
	MaxTokens int                  `json:"max_tokens"`
	Messages  []claudeMessage      `json:"messages"`
	Stream    bool                 `json:"stream"`
}

// claudeStreamEvent represents an SSE event from Claude.
type claudeStreamEvent struct {
	Type         string `json:"type"`
	Index        int    `json:"index,omitempty"`
	ContentBlock *struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content_block,omitempty"`
	Delta *struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"delta,omitempty"`
	Message *struct {
		Usage struct {
			InputTokens  int `json:"input_tokens"`
			OutputTokens int `json:"output_tokens"`
		} `json:"usage"`
	} `json:"message,omitempty"`
	Usage *struct {
		OutputTokens int `json:"output_tokens"`
	} `json:"usage,omitempty"`
}

// CompleteStream performs a streaming text completion request.
func (p *ClaudeProvider) CompleteStream(ctx context.Context, request *usecases.CompletionRequest, onChunk func(chunk string)) (*usecases.CompletionResponse, error) {
	maxTokens := request.MaxTokens
	if maxTokens <= 0 {
		maxTokens = 2048
	}

	// Use model from request if provided, otherwise use default
	model := request.Model
	if model == "" {
		model = p.model
	}

	reqBody := claudeStreamRequest{
		Model:     model,
		MaxTokens: maxTokens,
		Messages: []claudeMessage{
			{Role: "user", Content: request.Prompt},
		},
		Stream: true,
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
	req.Header.Set("Accept", "text/event-stream")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Claude API error: %s", string(body))
	}

	var fullContent string
	var tokensIn, tokensOut int

	// Read SSE stream
	buf := make([]byte, 4096)
	var lineBuffer string

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		n, err := resp.Body.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}

		lineBuffer += string(buf[:n])

		// Process complete lines
		for {
			idx := bytes.IndexByte([]byte(lineBuffer), '\n')
			if idx == -1 {
				break
			}

			line := lineBuffer[:idx]
			lineBuffer = lineBuffer[idx+1:]

			line = string(bytes.TrimSpace([]byte(line)))
			if line == "" {
				continue
			}

			// SSE format: "data: {...}"
			if !bytes.HasPrefix([]byte(line), []byte("data: ")) {
				continue
			}

			data := line[6:] // Remove "data: " prefix

			var event claudeStreamEvent
			if err := json.Unmarshal([]byte(data), &event); err != nil {
				continue
			}

			switch event.Type {
			case "content_block_delta":
				if event.Delta != nil && event.Delta.Text != "" {
					fullContent += event.Delta.Text
					onChunk(event.Delta.Text)
				}
			case "message_start":
				if event.Message != nil {
					tokensIn = event.Message.Usage.InputTokens
				}
			case "message_delta":
				if event.Usage != nil {
					tokensOut = event.Usage.OutputTokens
				}
			case "message_stop":
				// Stream complete
			}
		}
	}

	// Estimate tokens if not provided
	if tokensIn == 0 {
		tokensIn = len(request.Prompt) / 4
	}
	if tokensOut == 0 {
		tokensOut = len(fullContent) / 4
	}

	// Calculate cost
	var totalCost float64
	if p.pricingService != nil {
		cost, err := p.pricingService.CalculateCost(ctx, "claude", model, tokensIn, tokensOut)
		if err == nil {
			totalCost = cost
		}
	}
	if totalCost == 0 {
		inputCost := float64(tokensIn) * 0.00000025
		outputCost := float64(tokensOut) * 0.00000125
		totalCost = inputCost + outputCost
	}

	return &usecases.CompletionResponse{
		Content:   fullContent,
		Model:     model,
		TokensIn:  tokensIn,
		TokensOut: tokensOut,
		Cost:      totalCost,
	}, nil
}
