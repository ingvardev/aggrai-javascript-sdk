// Package handlers contains HTTP handlers for the API.
package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	appMiddleware "github.com/ingvar/aiaggregator/apps/api/internal/middleware"
	"github.com/ingvar/aiaggregator/packages/domain"
	"github.com/ingvar/aiaggregator/packages/providers"
	"github.com/ingvar/aiaggregator/packages/usecases"
)

// ChatCompletionRequest represents a completion request.
type ChatCompletionRequest struct {
	Prompt     string                   `json:"prompt"`
	Messages   []ChatCompletionMessage  `json:"messages,omitempty"`
	Provider   string                   `json:"provider,omitempty"`
	Model      string                   `json:"model,omitempty"`
	MaxTokens  int                      `json:"maxTokens,omitempty"`
	Tools      []usecases.Tool          `json:"tools,omitempty"`
	ToolChoice interface{}              `json:"toolChoice,omitempty"`
}

// ChatCompletionMessage represents a chat message.
type ChatCompletionMessage struct {
	Role       string                    `json:"role"`
	Content    string                    `json:"content"`
	ToolCalls  []ChatCompletionToolCall  `json:"toolCalls,omitempty"`
	ToolCallID string                    `json:"toolCallId,omitempty"`
}

// ChatCompletionToolCall represents a tool/function call.
type ChatCompletionToolCall struct {
	ID       string                      `json:"id"`
	Type     string                      `json:"type"`
	Function ChatCompletionFunctionCall  `json:"function"`
}

// ChatCompletionFunctionCall represents a function call.
type ChatCompletionFunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// ChatCompletionResponse represents a completion response.
type ChatCompletionResponse struct {
	Content      string                   `json:"content"`
	ToolCalls    []ChatCompletionToolCall `json:"toolCalls,omitempty"`
	FinishReason string                   `json:"finishReason"`
	TokensIn     int                      `json:"tokensIn"`
	TokensOut    int                      `json:"tokensOut"`
	Cost         float64                  `json:"cost"`
	Provider     string                   `json:"provider"`
	Model        string                   `json:"model"`
}

// CompletionsHandler handles synchronous completion requests.
type CompletionsHandler struct {
	registry    *providers.ProviderRegistry
	authService *usecases.AuthService
}

// NewCompletionsHandler creates a new completions handler.
func NewCompletionsHandler(registry *providers.ProviderRegistry, authService *usecases.AuthService) *CompletionsHandler {
	return &CompletionsHandler{
		registry:    registry,
		authService: authService,
	}
}

// ServeHTTP handles the completion request.
func (h *CompletionsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Only POST method
	if r.Method != http.MethodPost {
		writeCompletionError(w, http.StatusMethodNotAllowed, "method_not_allowed", "Only POST method is allowed")
		return
	}

	// Get tenant from context (set by auth middleware)
	tenant := appMiddleware.TenantFromContext(r.Context())
	if tenant == nil {
		writeCompletionError(w, http.StatusUnauthorized, "unauthorized", "API key required")
		return
	}

	// Parse request
	var req ChatCompletionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeCompletionError(w, http.StatusBadRequest, "invalid_request", "Invalid JSON body")
		return
	}

	// Validate: need either prompt or messages
	if req.Prompt == "" && len(req.Messages) == 0 {
		writeCompletionError(w, http.StatusBadRequest, "validation_error", "Either 'prompt' or 'messages' is required")
		return
	}

	// Get provider
	providerName := req.Provider
	if providerName == "" {
		providerName = tenant.DefaultProvider
	}
	if providerName == "" {
		providerName = "openai" // Default
	}

	provider, found := h.registry.Get(providerName)
	if !found {
		writeCompletionError(w, http.StatusBadRequest, "provider_error", fmt.Sprintf("Provider '%s' not available", providerName))
		return
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Minute)
	defer cancel()

	// Prepare request
	maxTokens := req.MaxTokens
	if maxTokens <= 0 {
		maxTokens = 2048
	}

	// Build completion request
	completionReq := &usecases.CompletionRequest{
		JobID:     uuid.New(),
		Prompt:    req.Prompt,
		Model:     req.Model,
		MaxTokens: maxTokens,
		Tools:     req.Tools,
	}

	// Convert messages if provided
	if len(req.Messages) > 0 {
		completionReq.Messages = make([]usecases.ChatMessage, len(req.Messages))
		for i, m := range req.Messages {
			msg := usecases.ChatMessage{
				Role:       m.Role,
				Content:    m.Content,
				ToolCallID: m.ToolCallID,
			}
			// Convert tool calls
			if len(m.ToolCalls) > 0 {
				msg.ToolCalls = make([]usecases.ToolCall, len(m.ToolCalls))
				for j, tc := range m.ToolCalls {
					msg.ToolCalls[j] = usecases.ToolCall{
						ID:   tc.ID,
						Type: tc.Type,
						Function: usecases.ToolCallFunction{
							Name:      tc.Function.Name,
							Arguments: tc.Function.Arguments,
						},
					}
				}
			}
			completionReq.Messages[i] = msg
		}
	}

	// Handle tool choice
	if req.ToolChoice != nil {
		switch v := req.ToolChoice.(type) {
		case string:
			completionReq.ToolChoice = v
		case map[string]interface{}:
			// Specific function choice - convert to JSON string
			if data, err := json.Marshal(v); err == nil {
				completionReq.ToolChoice = string(data)
			}
		}
	}

	// Try to use ToolsProvider for tools support, otherwise use StreamingProvider
	var resp *usecases.CompletionResponse
	var err error

	if toolsProvider, ok := provider.(usecases.ToolsProvider); ok && len(req.Tools) > 0 {
		// Use Complete() for tools
		resp, err = toolsProvider.Complete(ctx, completionReq)
	} else if streamingProvider, ok := provider.(usecases.StreamingProvider); ok {
		// Use CompleteStream but collect full response
		var content string
		resp, err = streamingProvider.CompleteStream(ctx, completionReq, func(chunk string) {
			content += chunk
		})
		if resp != nil {
			resp.Content = content
		}
	} else {
		writeCompletionError(w, http.StatusBadRequest, "provider_error", fmt.Sprintf("Provider '%s' does not support completions", providerName))
		return
	}

	if err != nil {
		writeCompletionError(w, http.StatusInternalServerError, "completion_error", err.Error())
		return
	}

	// Log request activity
	h.authService.LogRequestActivity(ctx, &usecases.RequestLogParams{
		TenantID:  tenant.ID,
		APIUserID: appMiddleware.APIUserIDFromContext(r.Context()),
		Action:    domain.AuditActionCompletion,
		Provider:  providerName,
		Model:     req.Model,
		TokensIn:  resp.TokensIn,
		TokensOut: resp.TokensOut,
		Cost:      resp.Cost,
		ClientIP:  extractClientIP(r),
		UserAgent: r.UserAgent(),
	})

	// Build response
	response := ChatCompletionResponse{
		Content:      resp.Content,
		FinishReason: resp.FinishReason,
		TokensIn:     resp.TokensIn,
		TokensOut:    resp.TokensOut,
		Cost:         resp.Cost,
		Provider:     providerName,
		Model:        req.Model,
	}

	// Convert tool calls if present
	if len(resp.ToolCalls) > 0 {
		response.ToolCalls = make([]ChatCompletionToolCall, len(resp.ToolCalls))
		for i, tc := range resp.ToolCalls {
			response.ToolCalls[i] = ChatCompletionToolCall{
				ID:   tc.ID,
				Type: tc.Type,
				Function: ChatCompletionFunctionCall{
					Name:      tc.Function.Name,
					Arguments: tc.Function.Arguments,
				},
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func writeCompletionError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{
		"error":   code,
		"message": message,
	})
}

// extractClientIP extracts the client IP from the request.
func extractClientIP(r *http.Request) string {
	// Check X-Forwarded-For header
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return xff
	}
	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}
	// Fallback to RemoteAddr
	return r.RemoteAddr
}
