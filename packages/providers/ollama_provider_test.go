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

func TestOllamaProvider_Name(t *testing.T) {
	p := NewOllamaProvider(OllamaConfig{})
	if p.Name() != "ollama" {
		t.Errorf("expected name 'ollama', got %q", p.Name())
	}
}

func TestOllamaProvider_Type(t *testing.T) {
	p := NewOllamaProvider(OllamaConfig{})
	if p.Type() != "ollama" {
		t.Errorf("expected type 'ollama', got %q", p.Type())
	}
}

func TestOllamaProvider_IsAvailable(t *testing.T) {
	t.Run("available when server responds", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/api/tags" {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"models": []}`))
			}
		}))
		defer server.Close()

		p := NewOllamaProvider(OllamaConfig{Endpoint: server.URL})
		if !p.IsAvailable(context.Background()) {
			t.Error("expected provider to be available")
		}
	})

	t.Run("not available when server is down", func(t *testing.T) {
		p := NewOllamaProvider(OllamaConfig{Endpoint: "http://localhost:99999"})
		if p.IsAvailable(context.Background()) {
			t.Error("expected provider to be unavailable")
		}
	})
}

func TestOllamaProvider_Complete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/chat" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"model": "llama3.2",
			"message": {"role": "assistant", "content": "Hello from Ollama!"},
			"done": true,
			"prompt_eval_count": 12,
			"eval_count": 18,
			"total_duration": 1000000000
		}`))
	}))
	defer server.Close()

	p := NewOllamaProvider(OllamaConfig{
		Endpoint: server.URL,
		Model:    "llama3.2",
	})

	resp, err := p.Complete(context.Background(), &usecases.CompletionRequest{
		JobID:  uuid.New(),
		Prompt: "Hello",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.Content != "Hello from Ollama!" {
		t.Errorf("unexpected content: %q", resp.Content)
	}

	if resp.TokensIn != 12 {
		t.Errorf("expected 12 input tokens, got %d", resp.TokensIn)
	}

	if resp.TokensOut != 18 {
		t.Errorf("expected 18 output tokens, got %d", resp.TokensOut)
	}

	if resp.Cost != 0 {
		t.Errorf("expected 0 cost for local model, got %f", resp.Cost)
	}
}

func TestOllamaProvider_Execute(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"model": "llama3.2",
			"message": {"role": "assistant", "content": "Ollama response"},
			"done": true,
			"prompt_eval_count": 5,
			"eval_count": 10
		}`))
	}))
	defer server.Close()

	p := NewOllamaProvider(OllamaConfig{
		Endpoint: server.URL,
	})

	job := domain.NewJob(uuid.New(), domain.JobTypeText, "Test input")

	result, err := p.Execute(context.Background(), job)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Result != "Ollama response" {
		t.Errorf("unexpected result: %q", result.Result)
	}

	if result.Cost != 0 {
		t.Errorf("expected 0 cost, got %f", result.Cost)
	}
}

func TestOllamaProvider_ListModels(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/tags" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"models": [
					{"name": "llama3.2"},
					{"name": "codellama"},
					{"name": "mistral"}
				]
			}`))
		}
	}))
	defer server.Close()

	p := NewOllamaProvider(OllamaConfig{Endpoint: server.URL})

	models, err := p.ListModels(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(models) != 3 {
		t.Errorf("expected 3 models, got %d", len(models))
	}

	expected := []string{"llama3.2", "codellama", "mistral"}
	for i, m := range expected {
		if models[i].ID != m {
			t.Errorf("expected model %q at index %d, got %q", m, i, models[i].ID)
		}
	}
}

func TestOllamaProvider_ImageNotSupported(t *testing.T) {
	p := NewOllamaProvider(OllamaConfig{})

	job := domain.NewJob(uuid.New(), domain.JobTypeImage, "Generate an image")

	_, err := p.Execute(context.Background(), job)
	if err == nil {
		t.Fatal("expected error for image generation")
	}
}
