// Package usecases contains application business logic and use case implementations.
package usecases

import (
	"context"

	"github.com/google/uuid"
	"github.com/ingvar/aiaggregator/packages/domain"
)

// AIProvider defines the interface for AI provider drivers.
// This is the simplified interface used for job processing.
type AIProvider interface {
	// Name returns the provider name.
	Name() string
	// Type returns the provider type.
	Type() string
	// Execute processes a job and returns the result.
	Execute(ctx context.Context, job *domain.Job) (*ProviderResult, error)
	// IsAvailable checks if the provider is currently available.
	IsAvailable(ctx context.Context) bool
}

// ProviderResult represents the result from an AI provider.
type ProviderResult struct {
	Result    string
	TokensIn  int
	TokensOut int
	Cost      float64
	Model     string
}

// CompletionRequest represents a text completion request.
type CompletionRequest struct {
	JobID     uuid.UUID
	Prompt    string
	Model     string
	MaxTokens int
	Options   map[string]interface{}

	// Tools/Functions support
	Messages     []ChatMessage  `json:"messages,omitempty"`
	Tools        []Tool         `json:"tools,omitempty"`
	ToolChoice   string         `json:"tool_choice,omitempty"` // "auto", "none", "required", or specific tool
}

// ChatMessage represents a message in a conversation.
type ChatMessage struct {
	Role       string     `json:"role"`                  // "system", "user", "assistant", "tool"
	Content    string     `json:"content,omitempty"`
	Name       string     `json:"name,omitempty"`        // For tool messages
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"`  // For assistant messages with tool calls
	ToolCallID string     `json:"tool_call_id,omitempty"` // For tool response messages
}

// Tool represents a tool/function that can be called by the AI.
type Tool struct {
	Type     string       `json:"type"` // "function"
	Function ToolFunction `json:"function"`
}

// ToolFunction describes a function that can be called.
type ToolFunction struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	Parameters  map[string]interface{} `json:"parameters,omitempty"` // JSON Schema
}

// ToolCall represents a tool call made by the assistant.
type ToolCall struct {
	ID       string           `json:"id"`
	Type     string           `json:"type"` // "function"
	Function ToolCallFunction `json:"function"`
}

// ToolCallFunction contains the function call details.
type ToolCallFunction struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"` // JSON string
}

// CompletionResponse represents a text completion response.
type CompletionResponse struct {
	Content   string
	Model     string
	TokensIn  int
	TokensOut int
	Cost      float64

	// Tools/Functions support
	ToolCalls    []ToolCall `json:"tool_calls,omitempty"`
	FinishReason string     `json:"finish_reason,omitempty"` // "stop", "tool_calls", "length"
}

// ImageRequest represents an image generation request.
type ImageRequest struct {
	JobID   uuid.UUID
	Prompt  string
	Model   string
	Size    string
	Options map[string]interface{}
}

// ImageResponse represents an image generation response.
type ImageResponse struct {
	URL   string
	Model string
	Cost  float64
}

// StreamingProvider interface for providers that support streaming responses.
type StreamingProvider interface {
	AIProvider
	// CompleteStream performs a streaming text completion request.
	// Calls onChunk for each received text chunk.
	// Returns final token counts and cost when stream completes.
	CompleteStream(ctx context.Context, request *CompletionRequest, onChunk func(chunk string)) (*CompletionResponse, error)
}

// StreamChunk represents a chunk of streaming response.
type StreamChunk struct {
	Content      string     `json:"content"`
	Done         bool       `json:"done"`
	TokensIn     int        `json:"tokensIn,omitempty"`
	TokensOut    int        `json:"tokensOut,omitempty"`
	Cost         float64    `json:"cost,omitempty"`
	Error        string     `json:"error,omitempty"`

	// Tools/Functions support
	ToolCalls    []ToolCall `json:"tool_calls,omitempty"`
	FinishReason string     `json:"finish_reason,omitempty"`
}

// ModelInfo represents information about an available model.
type ModelInfo struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	MaxTokens   int    `json:"maxTokens,omitempty"`
}

// ModelListProvider interface for providers that can list available models.
type ModelListProvider interface {
	AIProvider
	// ListModels returns a list of available models for this provider.
	ListModels(ctx context.Context) ([]ModelInfo, error)
}

// ToolsProvider interface for providers that support function/tool calling.
// Both OpenAI and Claude implement this through their Complete/CompleteStream methods.
type ToolsProvider interface {
	StreamingProvider
	// Complete performs a completion with optional tools support.
	Complete(ctx context.Context, request *CompletionRequest) (*CompletionResponse, error)
}

// ProviderSelector selects the best available provider for a request.
type ProviderSelector interface {
	SelectProvider(ctx context.Context, jobType domain.JobType) (AIProvider, error)
}
