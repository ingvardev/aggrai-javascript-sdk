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
	Model      string                 `json:"model"`
	MaxTokens  int                    `json:"max_tokens"`
	Messages   []claudeMessage        `json:"messages"`
	Tools      []claudeTool           `json:"tools,omitempty"`
	ToolChoice *claudeToolChoice      `json:"tool_choice,omitempty"`
}

// claudeMessage represents a message in Claude API format.
type claudeMessage struct {
	Role    string               `json:"role"`
	Content interface{}          `json:"content"` // string or []claudeContentBlock
}

// claudeContentBlock represents a content block in Claude messages.
type claudeContentBlock struct {
	Type       string `json:"type"` // "text", "tool_use", "tool_result"
	Text       string `json:"text,omitempty"`
	ID         string `json:"id,omitempty"`
	Name       string `json:"name,omitempty"`
	Input      interface{} `json:"input,omitempty"`
	ToolUseID  string `json:"tool_use_id,omitempty"`
	Content    string `json:"content,omitempty"`
}

// claudeTool represents a tool definition for Claude.
type claudeTool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	InputSchema map[string]interface{} `json:"input_schema"`
}

// claudeToolChoice specifies tool selection behavior.
type claudeToolChoice struct {
	Type string `json:"type"` // "auto", "any", "tool"
	Name string `json:"name,omitempty"` // For "tool" type
}

type claudeResponse struct {
	Content []struct {
		Type  string `json:"type"`
		Text  string `json:"text,omitempty"`
		ID    string `json:"id,omitempty"`
		Name  string `json:"name,omitempty"`
		Input interface{} `json:"input,omitempty"`
	} `json:"content"`
	StopReason string `json:"stop_reason"`
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

	// Use model from request if provided
	model := request.Model
	if model == "" {
		model = p.model
	}

	// Build messages from request
	var messages []claudeMessage
	if len(request.Messages) > 0 {
		for _, msg := range request.Messages {
			claudeMsg := p.convertToClaudeMessage(msg)
			messages = append(messages, claudeMsg)
		}
	} else {
		messages = []claudeMessage{
			{Role: "user", Content: request.Prompt},
		}
	}

	reqBody := claudeRequest{
		Model:     model,
		MaxTokens: maxTokens,
		Messages:  messages,
	}

	// Add tools if provided
	if len(request.Tools) > 0 {
		for _, tool := range request.Tools {
			reqBody.Tools = append(reqBody.Tools, claudeTool{
				Name:        tool.Function.Name,
				Description: tool.Function.Description,
				InputSchema: tool.Function.Parameters,
			})
		}
		// Set tool_choice if specified
		if request.ToolChoice != "" {
			switch request.ToolChoice {
			case "auto":
				reqBody.ToolChoice = &claudeToolChoice{Type: "auto"}
			case "none":
				// Claude doesn't have "none" - just don't send tools
				reqBody.Tools = nil
			case "required":
				reqBody.ToolChoice = &claudeToolChoice{Type: "any"}
			default:
				// Specific tool by name
				reqBody.ToolChoice = &claudeToolChoice{Type: "tool", Name: request.ToolChoice}
			}
		}
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
		cost, err := p.pricingService.CalculateCost(ctx, "claude", model, claudeResp.Usage.InputTokens, claudeResp.Usage.OutputTokens)
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

	// Build response
	response := &usecases.CompletionResponse{
		Model:        model,
		TokensIn:     claudeResp.Usage.InputTokens,
		TokensOut:    claudeResp.Usage.OutputTokens,
		Cost:         totalCost,
		FinishReason: claudeResp.StopReason,
	}

	// Extract text content and tool calls
	for _, block := range claudeResp.Content {
		switch block.Type {
		case "text":
			response.Content += block.Text
		case "tool_use":
			// Convert input to JSON string
			inputJSON, _ := json.Marshal(block.Input)
			response.ToolCalls = append(response.ToolCalls, usecases.ToolCall{
				ID:   block.ID,
				Type: "function",
				Function: usecases.ToolCallFunction{
					Name:      block.Name,
					Arguments: string(inputJSON),
				},
			})
		}
	}

	return response, nil
}

// convertToClaudeMessage converts a usecases.ChatMessage to Claude format.
func (p *ClaudeProvider) convertToClaudeMessage(msg usecases.ChatMessage) claudeMessage {
	// Handle tool response messages
	if msg.Role == "tool" {
		return claudeMessage{
			Role: "user",
			Content: []claudeContentBlock{
				{
					Type:      "tool_result",
					ToolUseID: msg.ToolCallID,
					Content:   msg.Content,
				},
			},
		}
	}

	// Handle assistant messages with tool calls
	if msg.Role == "assistant" && len(msg.ToolCalls) > 0 {
		var blocks []claudeContentBlock
		if msg.Content != "" {
			blocks = append(blocks, claudeContentBlock{
				Type: "text",
				Text: msg.Content,
			})
		}
		for _, tc := range msg.ToolCalls {
			var input interface{}
			json.Unmarshal([]byte(tc.Function.Arguments), &input)
			blocks = append(blocks, claudeContentBlock{
				Type:  "tool_use",
				ID:    tc.ID,
				Name:  tc.Function.Name,
				Input: input,
			})
		}
		return claudeMessage{
			Role:    "assistant",
			Content: blocks,
		}
	}

	// Simple text message
	return claudeMessage{
		Role:    msg.Role,
		Content: msg.Content,
	}
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
	Model      string               `json:"model"`
	MaxTokens  int                  `json:"max_tokens"`
	Messages   []claudeMessage      `json:"messages"`
	Stream     bool                 `json:"stream"`
	Tools      []claudeTool         `json:"tools,omitempty"`
	ToolChoice *claudeToolChoice    `json:"tool_choice,omitempty"`
}

// claudeStreamEvent represents an SSE event from Claude.
type claudeStreamEvent struct {
	Type         string `json:"type"`
	Index        int    `json:"index,omitempty"`
	ContentBlock *struct {
		Type  string      `json:"type"`
		Text  string      `json:"text,omitempty"`
		ID    string      `json:"id,omitempty"`
		Name  string      `json:"name,omitempty"`
		Input interface{} `json:"input,omitempty"`
	} `json:"content_block,omitempty"`
	Delta *struct {
		Type        string `json:"type"`
		Text        string `json:"text,omitempty"`
		PartialJSON string `json:"partial_json,omitempty"`
		StopReason  string `json:"stop_reason,omitempty"`
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

	// Build messages from request
	var messages []claudeMessage
	if len(request.Messages) > 0 {
		for _, msg := range request.Messages {
			messages = append(messages, p.convertToClaudeMessage(msg))
		}
	} else {
		messages = []claudeMessage{
			{Role: "user", Content: request.Prompt},
		}
	}

	reqBody := claudeStreamRequest{
		Model:     model,
		MaxTokens: maxTokens,
		Messages:  messages,
		Stream:    true,
	}

	// Add tools if provided
	if len(request.Tools) > 0 {
		for _, tool := range request.Tools {
			reqBody.Tools = append(reqBody.Tools, claudeTool{
				Name:        tool.Function.Name,
				Description: tool.Function.Description,
				InputSchema: tool.Function.Parameters,
			})
		}
		if request.ToolChoice != "" {
			switch request.ToolChoice {
			case "auto":
				reqBody.ToolChoice = &claudeToolChoice{Type: "auto"}
			case "none":
				reqBody.Tools = nil
			case "required":
				reqBody.ToolChoice = &claudeToolChoice{Type: "any"}
			default:
				reqBody.ToolChoice = &claudeToolChoice{Type: "tool", Name: request.ToolChoice}
			}
		}
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
	var finishReason string
	toolCalls := make(map[int]*usecases.ToolCall) // indexed by content block index
	toolInputBuffers := make(map[int]string)       // buffer for streaming JSON input

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
			case "content_block_start":
				if event.ContentBlock != nil && event.ContentBlock.Type == "tool_use" {
					toolCalls[event.Index] = &usecases.ToolCall{
						ID:   event.ContentBlock.ID,
						Type: "function",
						Function: usecases.ToolCallFunction{
							Name: event.ContentBlock.Name,
						},
					}
					toolInputBuffers[event.Index] = ""
				}
			case "content_block_delta":
				if event.Delta != nil {
					if event.Delta.Text != "" {
						fullContent += event.Delta.Text
						onChunk(event.Delta.Text)
					}
					if event.Delta.PartialJSON != "" {
						toolInputBuffers[event.Index] += event.Delta.PartialJSON
					}
				}
			case "content_block_stop":
				// Finalize tool call input
				if tc, ok := toolCalls[event.Index]; ok {
					tc.Function.Arguments = toolInputBuffers[event.Index]
				}
			case "message_start":
				if event.Message != nil {
					tokensIn = event.Message.Usage.InputTokens
				}
			case "message_delta":
				if event.Usage != nil {
					tokensOut = event.Usage.OutputTokens
				}
				if event.Delta != nil && event.Delta.StopReason != "" {
					finishReason = event.Delta.StopReason
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

	// Build response
	response := &usecases.CompletionResponse{
		Content:      fullContent,
		Model:        model,
		TokensIn:     tokensIn,
		TokensOut:    tokensOut,
		Cost:         totalCost,
		FinishReason: finishReason,
	}

	// Convert tool calls map to slice (ordered by index)
	for i := 0; i < len(toolCalls); i++ {
		if tc, ok := toolCalls[i]; ok {
			response.ToolCalls = append(response.ToolCalls, *tc)
		}
	}

	return response, nil
}
