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

func TestClaudeProvider_Name(t *testing.T) {
	p := NewClaudeProvider(ClaudeConfig{APIKey: "test-key"})
	if p.Name() != "claude" {
		t.Errorf("expected name 'claude', got %q", p.Name())
	}
}

func TestClaudeProvider_Type(t *testing.T) {
	p := NewClaudeProvider(ClaudeConfig{APIKey: "test-key"})
	if p.Type() != "claude" {
		t.Errorf("expected type 'claude', got %q", p.Type())
	}
}

func TestClaudeProvider_IsAvailable(t *testing.T) {
	t.Run("available with API key", func(t *testing.T) {
		p := NewClaudeProvider(ClaudeConfig{APIKey: "test-key"})
		if !p.IsAvailable(context.Background()) {
			t.Error("expected provider to be available")
		}
	})

	t.Run("not available without API key", func(t *testing.T) {
		p := NewClaudeProvider(ClaudeConfig{APIKey: ""})
		if p.IsAvailable(context.Background()) {
			t.Error("expected provider to be unavailable")
		}
	})
}

func TestClaudeProvider_Complete(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/messages" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Header.Get("x-api-key") != "test-key" {
			t.Error("missing or invalid x-api-key header")
		}
		if r.Header.Get("anthropic-version") != "2023-06-01" {
			t.Error("missing or invalid anthropic-version header")
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"content": [{
				"type": "text",
				"text": "Hello! I'm Claude, an AI assistant."
			}],
			"usage": {"input_tokens": 15, "output_tokens": 25}
		}`))
	}))
	defer server.Close()

	p := NewClaudeProvider(ClaudeConfig{
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

	if resp.Content != "Hello! I'm Claude, an AI assistant." {
		t.Errorf("unexpected content: %q", resp.Content)
	}

	if resp.TokensIn != 15 {
		t.Errorf("expected 15 input tokens, got %d", resp.TokensIn)
	}

	if resp.TokensOut != 25 {
		t.Errorf("expected 25 output tokens, got %d", resp.TokensOut)
	}
}

func TestClaudeProvider_Execute(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"content": [{"type": "text", "text": "Claude response"}],
			"usage": {"input_tokens": 8, "output_tokens": 12}
		}`))
	}))
	defer server.Close()

	p := NewClaudeProvider(ClaudeConfig{
		APIKey:   "test-key",
		Endpoint: server.URL,
	})

	job := domain.NewJob(uuid.New(), domain.JobTypeText, "Test input")

	result, err := p.Execute(context.Background(), job)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Result != "Claude response" {
		t.Errorf("unexpected result: %q", result.Result)
	}

	if result.TokensIn != 8 {
		t.Errorf("expected 8 input tokens, got %d", result.TokensIn)
	}

	if result.TokensOut != 12 {
		t.Errorf("expected 12 output tokens, got %d", result.TokensOut)
	}
}

func TestClaudeProvider_ImageNotSupported(t *testing.T) {
	p := NewClaudeProvider(ClaudeConfig{APIKey: "test-key"})

	job := domain.NewJob(uuid.New(), domain.JobTypeImage, "Generate an image")

	_, err := p.Execute(context.Background(), job)
	if err == nil {
		t.Fatal("expected error for image generation")
	}
}

func TestClaudeProvider_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error": {"message": "Invalid API key"}}`))
	}))
	defer server.Close()

	p := NewClaudeProvider(ClaudeConfig{
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
