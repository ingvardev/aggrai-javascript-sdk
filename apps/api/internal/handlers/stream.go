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

// StreamRequest represents a streaming completion request.
type StreamRequest struct {
	Prompt    string `json:"prompt"`
	Provider  string `json:"provider,omitempty"`
	Model     string `json:"model,omitempty"`
	MaxTokens int    `json:"maxTokens,omitempty"`
}

// StreamEvent represents an SSE event sent to the client.
type StreamEvent struct {
	Type      string  `json:"type"` // "chunk", "done", "error"
	Content   string  `json:"content,omitempty"`
	TokensIn  int     `json:"tokensIn,omitempty"`
	TokensOut int     `json:"tokensOut,omitempty"`
	Cost      float64 `json:"cost,omitempty"`
	Error     string  `json:"error,omitempty"`
}

// StreamHandler handles streaming completion requests.
type StreamHandler struct {
	registry *providers.ProviderRegistry
	authService *usecases.AuthService
}

// NewStreamHandler creates a new streaming handler.
func NewStreamHandler(registry *providers.ProviderRegistry, authService *usecases.AuthService) *StreamHandler {
	return &StreamHandler{
		registry:    registry,
		authService: authService,
	}
}

// ServeHTTP handles the streaming request.
func (h *StreamHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Only POST method
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get tenant from context (set by auth middleware)
	tenant := appMiddleware.TenantFromContext(r.Context())
	if tenant == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse request
	var req StreamRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Prompt == "" {
		http.Error(w, "Prompt is required", http.StatusBadRequest)
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
		http.Error(w, fmt.Sprintf("Provider '%s' not available", providerName), http.StatusBadRequest)
		return
	}

	// Check if provider supports streaming
	streamingProvider, ok := provider.(usecases.StreamingProvider)
	if !ok {
		http.Error(w, fmt.Sprintf("Provider '%s' does not support streaming", providerName), http.StatusBadRequest)
		return
	}

	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("X-Accel-Buffering", "no") // Disable nginx buffering

	// Flush headers
	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Minute)
	defer cancel()

	// Prepare request
	maxTokens := req.MaxTokens
	if maxTokens <= 0 {
		maxTokens = 2048
	}

	completionReq := &usecases.CompletionRequest{
		JobID:     uuid.New(),
		Prompt:    req.Prompt,
		Model:     req.Model,
		MaxTokens: maxTokens,
	}

	// Stream response
	sendEvent := func(event StreamEvent) {
		data, _ := json.Marshal(event)
		fmt.Fprintf(w, "data: %s\n\n", data)
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}
	}

	// Call streaming provider
	resp, err := streamingProvider.CompleteStream(ctx, completionReq, func(chunk string) {
		sendEvent(StreamEvent{
			Type:    "chunk",
			Content: chunk,
		})
	})

	if err != nil {
		sendEvent(StreamEvent{
			Type:  "error",
			Error: err.Error(),
		})
		return
	}

	// Log streaming request activity
	h.authService.LogRequestActivity(ctx, &usecases.RequestLogParams{
		TenantID:  tenant.ID,
		APIUserID: appMiddleware.APIUserIDFromContext(r.Context()),
		Action:    domain.AuditActionStreaming,
		Provider:  providerName,
		Model:     req.Model,
		TokensIn:  resp.TokensIn,
		TokensOut: resp.TokensOut,
		Cost:      resp.Cost,
		ClientIP:  extractClientIP(r),
		UserAgent: r.UserAgent(),
	})

	// Send final stats
	sendEvent(StreamEvent{
		Type:      "done",
		TokensIn:  resp.TokensIn,
		TokensOut: resp.TokensOut,
		Cost:      resp.Cost,
	})
}
