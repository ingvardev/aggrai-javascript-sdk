// Package providers contains AI provider implementations.
package providers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/ingvar/aiaggregator/packages/domain"
	"github.com/ingvar/aiaggregator/packages/usecases"
)

// OpenAIProvider implements AI provider using OpenAI API.
type OpenAIProvider struct {
	apiKey         string
	model          string
	endpoint       string
	client         *http.Client
	pricingService *usecases.PricingService
}

// OpenAIConfig holds configuration for OpenAI provider.
type OpenAIConfig struct {
	APIKey         string
	Model          string
	Endpoint       string
	PricingService *usecases.PricingService
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
		apiKey:         cfg.APIKey,
		model:          cfg.Model,
		endpoint:       cfg.Endpoint,
		client:         &http.Client{},
		pricingService: cfg.PricingService,
	}
}

// Name returns the provider name.
func (p *OpenAIProvider) Name() string {
	return "openai"
}

// Type returns the provider type.
func (p *OpenAIProvider) Type() string {
	return string(domain.ProviderTypeOpenAI)
}

// openAIChatRequest represents OpenAI chat completion request.
type openAIChatRequest struct {
	Model               string              `json:"model"`
	Messages            []openAIChatMessage `json:"messages"`
	MaxTokens           int                 `json:"max_tokens,omitempty"`
	MaxCompletionTokens int                 `json:"max_completion_tokens,omitempty"`
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
	// Use model from request if provided, otherwise use default
	model := request.Model
	if model == "" {
		model = p.model
	}

	reqBody := openAIChatRequest{
		Model: model,
		Messages: []openAIChatMessage{
			{Role: "user", Content: request.Prompt},
		},
	}

	// o1/o3 models use max_completion_tokens instead of max_tokens
	if isReasoningModel(model) {
		if request.MaxTokens > 0 {
			reqBody.MaxCompletionTokens = request.MaxTokens
		} else {
			reqBody.MaxCompletionTokens = 4096
		}
	} else {
		reqBody.MaxTokens = request.MaxTokens
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

	// Calculate cost using pricing service or fallback to defaults
	var totalCost float64
	if p.pricingService != nil {
		cost, err := p.pricingService.CalculateCost(ctx, "openai", model, openAIResp.Usage.PromptTokens, openAIResp.Usage.CompletionTokens)
		if err == nil {
			totalCost = cost
		}
	}
	if totalCost == 0 {
		// Fallback to default pricing based on model
		totalCost = calculateOpenAIFallbackCost(model, openAIResp.Usage.PromptTokens, openAIResp.Usage.CompletionTokens)
	}

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

	// Calculate image cost using pricing service or fallback to default
	var imageCost float64
	if p.pricingService != nil {
		cost, err := p.pricingService.CalculateImageCost(ctx, "openai", "dall-e-3")
		if err == nil && cost > 0 {
			imageCost = cost
		}
	}
	if imageCost == 0 {
		imageCost = 0.04 // Fallback to DALL-E 3 standard pricing
	}

	return &usecases.ImageResponse{
		URL:   result.Data[0].URL,
		Model: "dall-e-3",
		Cost:  imageCost,
	}, nil
}

// IsAvailable checks if the provider is available.
func (p *OpenAIProvider) IsAvailable(ctx context.Context) bool {
	return p.apiKey != ""
}

// Execute processes a job and returns the result.
// This is the main interface method for AIProvider.
func (p *OpenAIProvider) Execute(ctx context.Context, job *domain.Job) (*usecases.ProviderResult, error) {
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
		resp, err := p.GenerateImage(ctx, &usecases.ImageRequest{
			JobID:  job.ID,
			Prompt: job.Input,
			Size:   "1024x1024",
		})
		if err != nil {
			return nil, err
		}
		return &usecases.ProviderResult{
			Result: resp.URL,
			Cost:   resp.Cost,
			Model:  resp.Model,
		}, nil

	default:
		return nil, fmt.Errorf("unsupported job type: %s", job.Type)
	}
}

// Ensure OpenAIProvider implements AIProvider
var _ usecases.AIProvider = (*OpenAIProvider)(nil)

// Ensure OpenAIProvider implements StreamingProvider
var _ usecases.StreamingProvider = (*OpenAIProvider)(nil)

// Ensure OpenAIProvider implements ModelListProvider
var _ usecases.ModelListProvider = (*OpenAIProvider)(nil)

// openAIModelsResponse represents the response from /v1/models endpoint.
type openAIModelsResponse struct {
	Data []struct {
		ID      string `json:"id"`
		Object  string `json:"object"`
		Created int64  `json:"created"`
		OwnedBy string `json:"owned_by"`
	} `json:"data"`
}

// ListModels returns a list of available models from OpenAI.
func (p *OpenAIProvider) ListModels(ctx context.Context) ([]usecases.ModelInfo, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", p.endpoint+"/models", nil)
	if err != nil {
		return nil, err
	}

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

	var modelsResp openAIModelsResponse
	if err := json.NewDecoder(resp.Body).Decode(&modelsResp); err != nil {
		return nil, err
	}

	// Filter to only chat/completion models (gpt-*, o1-*)
	var models []usecases.ModelInfo
	for _, m := range modelsResp.Data {
		// Filter by common prefixes for chat models
		if isOpenAIChatModel(m.ID) {
			models = append(models, usecases.ModelInfo{
				ID:   m.ID,
				Name: formatOpenAIModelName(m.ID),
			})
		}
	}

	return models, nil
}

// isOpenAIChatModel checks if a model ID is a chat/completion model.
func isOpenAIChatModel(id string) bool {
	prefixes := []string{"gpt-4", "gpt-3.5", "o1", "o3", "chatgpt"}
	for _, prefix := range prefixes {
		if len(id) >= len(prefix) && id[:len(prefix)] == prefix {
			return true
		}
	}
	return false
}

// calculateOpenAIFallbackCost calculates cost using default pricing for different models.
func calculateOpenAIFallbackCost(model string, tokensIn, tokensOut int) float64 {
	// Pricing per 1M tokens (as of Dec 2024)
	var inputPer1M, outputPer1M float64

	switch {
	case strings.HasPrefix(model, "o1-mini"):
		inputPer1M, outputPer1M = 3.0, 12.0
	case strings.HasPrefix(model, "o1-preview"), strings.HasPrefix(model, "o1"):
		inputPer1M, outputPer1M = 15.0, 60.0
	case strings.HasPrefix(model, "o3-mini"):
		inputPer1M, outputPer1M = 1.1, 4.4
	case strings.HasPrefix(model, "gpt-4o-mini"):
		inputPer1M, outputPer1M = 0.15, 0.60
	case strings.HasPrefix(model, "gpt-4o"):
		inputPer1M, outputPer1M = 2.5, 10.0
	case strings.HasPrefix(model, "gpt-4-turbo"):
		inputPer1M, outputPer1M = 10.0, 30.0
	case strings.HasPrefix(model, "gpt-4"):
		inputPer1M, outputPer1M = 30.0, 60.0
	case strings.HasPrefix(model, "gpt-3.5"):
		inputPer1M, outputPer1M = 0.5, 1.5
	default:
		// Default to gpt-4o-mini pricing
		inputPer1M, outputPer1M = 0.15, 0.60
	}

	inputCost := float64(tokensIn) * inputPer1M / 1_000_000
	outputCost := float64(tokensOut) * outputPer1M / 1_000_000
	return inputCost + outputCost
}

// formatOpenAIModelName creates a human-readable name from model ID.
func formatOpenAIModelName(id string) string {
	// Simple formatting - could be enhanced
	return id
}

// openAIStreamRequest represents OpenAI streaming chat completion request.
type openAIStreamRequest struct {
	Model               string              `json:"model"`
	Messages            []openAIChatMessage `json:"messages"`
	MaxTokens           int                 `json:"max_tokens,omitempty"`
	MaxCompletionTokens int                 `json:"max_completion_tokens,omitempty"`
	Stream              bool                `json:"stream"`
	StreamOptions       *streamOptions      `json:"stream_options,omitempty"`
}

type streamOptions struct {
	IncludeUsage bool `json:"include_usage"`
}

// isReasoningModel checks if the model is an o1/o3 reasoning model.
func isReasoningModel(model string) bool {
	prefixes := []string{"o1", "o3"}
	for _, prefix := range prefixes {
		if len(model) >= len(prefix) && model[:len(prefix)] == prefix {
			return true
		}
	}
	return false
}

// openAIStreamChunk represents a streaming chunk from OpenAI.
type openAIStreamChunk struct {
	ID      string `json:"id"`
	Choices []struct {
		Delta struct {
			Content string `json:"content"`
		} `json:"delta"`
		FinishReason *string `json:"finish_reason"`
	} `json:"choices"`
	Usage *struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
	} `json:"usage"`
}

// CompleteStream performs a streaming text completion request.
func (p *OpenAIProvider) CompleteStream(ctx context.Context, request *usecases.CompletionRequest, onChunk func(chunk string)) (*usecases.CompletionResponse, error) {
	// Use model from request if provided, otherwise use default
	model := request.Model
	if model == "" {
		model = p.model
	}

	reqBody := openAIStreamRequest{
		Model: model,
		Messages: []openAIChatMessage{
			{Role: "user", Content: request.Prompt},
		},
		Stream: true,
		StreamOptions: &streamOptions{
			IncludeUsage: true,
		},
	}

	// o1/o3 models use max_completion_tokens instead of max_tokens
	if isReasoningModel(model) {
		if request.MaxTokens > 0 {
			reqBody.MaxCompletionTokens = request.MaxTokens
		} else {
			reqBody.MaxCompletionTokens = 4096
		}
	} else {
		reqBody.MaxTokens = request.MaxTokens
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
	req.Header.Set("Accept", "text/event-stream")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("OpenAI API error: %s", string(body))
	}

	var fullContent string
	var tokensIn, tokensOut int

	// Read SSE stream
	reader := resp.Body
	buf := make([]byte, 4096)
	var lineBuffer string

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		n, err := reader.Read(buf)
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
			if data == "[DONE]" {
				break
			}

			var chunk openAIStreamChunk
			if err := json.Unmarshal([]byte(data), &chunk); err != nil {
				continue
			}

			if len(chunk.Choices) > 0 && chunk.Choices[0].Delta.Content != "" {
				content := chunk.Choices[0].Delta.Content
				fullContent += content
				onChunk(content)
			}

			// OpenAI returns usage in the final chunk with stream_options
			if chunk.Usage != nil {
				tokensIn = chunk.Usage.PromptTokens
				tokensOut = chunk.Usage.CompletionTokens
			}
		}
	}

	// Estimate tokens if not provided
	if tokensIn == 0 {
		tokensIn = len(request.Prompt) / 4 // Rough estimate
	}
	if tokensOut == 0 {
		tokensOut = len(fullContent) / 4 // Rough estimate
	}

	// Calculate cost
	var totalCost float64
	if p.pricingService != nil {
		cost, err := p.pricingService.CalculateCost(ctx, "openai", model, tokensIn, tokensOut)
		if err == nil {
			totalCost = cost
		}
	}
	if totalCost == 0 {
		inputCost := float64(tokensIn) * 0.00000015
		outputCost := float64(tokensOut) * 0.0000006
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
