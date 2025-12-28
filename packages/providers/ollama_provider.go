// Package providers contains AI provider implementations.
package providers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/ingvar/aiaggregator/packages/domain"
	"github.com/ingvar/aiaggregator/packages/usecases"
)

// OllamaProvider implements AI provider using local Ollama server.
type OllamaProvider struct {
	endpoint string
	model    string
	client   *http.Client
}

// OllamaConfig holds configuration for Ollama provider.
type OllamaConfig struct {
	Endpoint string
	Model    string
	Timeout  time.Duration
}

// NewOllamaProvider creates a new Ollama provider.
func NewOllamaProvider(cfg OllamaConfig) *OllamaProvider {
	if cfg.Endpoint == "" {
		cfg.Endpoint = "http://localhost:11434"
	}
	if cfg.Model == "" {
		cfg.Model = "llama3.2"
	}
	if cfg.Timeout == 0 {
		cfg.Timeout = 120 * time.Second
	}

	return &OllamaProvider{
		endpoint: cfg.Endpoint,
		model:    cfg.Model,
		client: &http.Client{
			Timeout: cfg.Timeout,
		},
	}
}

// Name returns the provider name.
func (p *OllamaProvider) Name() string {
	return "ollama"
}

// Type returns the provider type.
func (p *OllamaProvider) Type() string {
	return string(domain.ProviderTypeOllama)
}

// ollamaGenerateRequest represents Ollama generate API request.
type ollamaGenerateRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

// ollamaGenerateResponse represents Ollama generate API response.
type ollamaGenerateResponse struct {
	Model              string `json:"model"`
	Response           string `json:"response"`
	Done               bool   `json:"done"`
	TotalDuration      int64  `json:"total_duration"`
	LoadDuration       int64  `json:"load_duration"`
	PromptEvalCount    int    `json:"prompt_eval_count"`
	PromptEvalDuration int64  `json:"prompt_eval_duration"`
	EvalCount          int    `json:"eval_count"`
	EvalDuration       int64  `json:"eval_duration"`
}

// ollamaChatRequest represents Ollama chat API request.
type ollamaChatRequest struct {
	Model    string              `json:"model"`
	Messages []ollamaChatMessage `json:"messages"`
	Stream   bool                `json:"stream"`
}

type ollamaChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ollamaChatResponse struct {
	Model   string `json:"model"`
	Message struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	} `json:"message"`
	Done            bool  `json:"done"`
	PromptEvalCount int   `json:"prompt_eval_count"`
	EvalCount       int   `json:"eval_count"`
	TotalDuration   int64 `json:"total_duration"`
}

// Complete performs a text completion request.
func (p *OllamaProvider) Complete(ctx context.Context, request *usecases.CompletionRequest) (*usecases.CompletionResponse, error) {
	reqBody := ollamaChatRequest{
		Model: p.model,
		Messages: []ollamaChatMessage{
			{Role: "user", Content: request.Prompt},
		},
		Stream: false,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", p.endpoint+"/api/chat", bytes.NewReader(jsonBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Ollama API error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Ollama API error (%d): %s", resp.StatusCode, string(body))
	}

	var ollamaResp ollamaChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
		return nil, err
	}

	// Ollama is local and free, so cost is 0
	return &usecases.CompletionResponse{
		Content:   ollamaResp.Message.Content,
		Model:     ollamaResp.Model,
		TokensIn:  ollamaResp.PromptEvalCount,
		TokensOut: ollamaResp.EvalCount,
		Cost:      0, // Local models are free
	}, nil
}

// IsAvailable checks if the Ollama server is running.
func (p *OllamaProvider) IsAvailable(ctx context.Context) bool {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", p.endpoint+"/api/tags", nil)
	if err != nil {
		return false
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}

// Execute processes a job and returns the result.
func (p *OllamaProvider) Execute(ctx context.Context, job *domain.Job) (*usecases.ProviderResult, error) {
	switch job.Type {
	case domain.JobTypeText:
		resp, err := p.Complete(ctx, &usecases.CompletionRequest{
			JobID:  job.ID,
			Prompt: job.Input,
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
		return nil, fmt.Errorf("Ollama does not support image generation")

	default:
		return nil, fmt.Errorf("unsupported job type: %s", job.Type)
	}
}

// ListModels returns available models from Ollama.
func (p *OllamaProvider) ListModels(ctx context.Context) ([]string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", p.endpoint+"/api/tags", nil)
	if err != nil {
		return nil, err
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to list models: %d", resp.StatusCode)
	}

	var result struct {
		Models []struct {
			Name string `json:"name"`
		} `json:"models"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	models := make([]string, len(result.Models))
	for i, m := range result.Models {
		models[i] = m.Name
	}
	return models, nil
}

// Ensure OllamaProvider implements AIProvider
var _ usecases.AIProvider = (*OllamaProvider)(nil)
