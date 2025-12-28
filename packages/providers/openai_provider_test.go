package providers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/ingvar/aiaggregator/packages/domain"
	"github.com/ingvar/aiaggregator/packages/usecases"
)

func TestOpenAIProvider_Name(t *testing.T) {
	p := NewOpenAIProvider(OpenAIConfig{APIKey: "test-key"})
	if p.Name() != "openai" {
		t.Errorf("expected name 'openai', got %q", p.Name())
	}
}

func TestOpenAIProvider_Type(t *testing.T) {
	p := NewOpenAIProvider(OpenAIConfig{APIKey: "test-key"})
	if p.Type() != "openai" {
		t.Errorf("expected type 'openai', got %q", p.Type())
	}
}

func TestOpenAIProvider_IsAvailable(t *testing.T) {
	t.Run("available with API key", func(t *testing.T) {
		p := NewOpenAIProvider(OpenAIConfig{APIKey: "test-key"})
		if !p.IsAvailable(context.Background()) {
			t.Error("expected provider to be available")
		}
	})

	t.Run("not available without API key", func(t *testing.T) {
		p := NewOpenAIProvider(OpenAIConfig{APIKey: ""})
		if p.IsAvailable(context.Background()) {
			t.Error("expected provider to be unavailable")
		}
	})
}

func TestOpenAIProvider_Complete(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/chat/completions" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Error("missing or invalid authorization header")
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"id": "chatcmpl-123",
			"choices": [{
				"message": {"role": "assistant", "content": "Hello! How can I help you?"}
			}],
			"usage": {"prompt_tokens": 10, "completion_tokens": 20}
		}`))
	}))
	defer server.Close()

	p := NewOpenAIProvider(OpenAIConfig{
		APIKey:   "test-key",
		Endpoint: server.URL,
	})

	resp, err := p.Complete(context.Background(), &usecases.CompletionRequest{
		JobID:     uuid.New(),
		Prompt:    "Hello",
		MaxTokens: 100,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.Content != "Hello! How can I help you?" {
		t.Errorf("unexpected content: %q", resp.Content)
	}

	if resp.TokensIn != 10 {
		t.Errorf("expected 10 input tokens, got %d", resp.TokensIn)
	}

	if resp.TokensOut != 20 {
		t.Errorf("expected 20 output tokens, got %d", resp.TokensOut)
	}
}

func TestOpenAIProvider_Execute(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"id": "chatcmpl-123",
			"choices": [{
				"message": {"role": "assistant", "content": "Test response"}
			}],
			"usage": {"prompt_tokens": 5, "completion_tokens": 10}
		}`))
	}))
	defer server.Close()

	p := NewOpenAIProvider(OpenAIConfig{
		APIKey:   "test-key",
		Endpoint: server.URL,
	})

	job := domain.NewJob(uuid.New(), domain.JobTypeText, "Test input")

	result, err := p.Execute(context.Background(), job)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Result != "Test response" {
		t.Errorf("unexpected result: %q", result.Result)
	}

	if result.TokensIn != 5 {
		t.Errorf("expected 5 input tokens, got %d", result.TokensIn)
	}

	if result.TokensOut != 10 {
		t.Errorf("expected 10 output tokens, got %d", result.TokensOut)
	}
}

func TestOpenAIProvider_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error": {"message": "Invalid API key"}}`))
	}))
	defer server.Close()

	p := NewOpenAIProvider(OpenAIConfig{
		APIKey:   "invalid-key",
		Endpoint: server.URL,
	})

	_, err := p.Complete(context.Background(), &usecases.CompletionRequest{
		Prompt: "Hello",
	})

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
