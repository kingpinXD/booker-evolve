package llm

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"booker/config"
	"booker/httpclient"
)

// newTestHTTPClient creates an httpclient.Client that talks to the given test server.
func newTestHTTPClient(srv *httptest.Server) *httpclient.Client {
	return httpclient.New(config.HTTPConfig{
		Timeout:         5 * time.Second,
		MaxIdleConns:    2,
		IdleConnTimeout: 10 * time.Second,
		MaxRetries:      1,
		RetryBaseDelay:  time.Millisecond,
	})
}

// successHandler returns JSON with a single choice containing the given content.
func successHandler(content string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		resp := chatResponse{
			Model: "test-model",
			Choices: []chatChoice{
				{Message: Message{Role: RoleAssistant, Content: content}},
			},
			Usage: chatUsage{PromptTokens: 10, CompletionTokens: 5, TotalTokens: 15},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}
}

func TestChatCompletion_Success(t *testing.T) {
	srv := httptest.NewServer(successHandler("hello world"))
	defer srv.Close()

	c := New(
		config.LLMConfig{
			Provider:   config.LLMOpenAI,
			APIKey:     "test-key",
			Model:      "gpt-test",
			BaseURL:    srv.URL,
			MaxTokens:  100,
			AuthHeader: "Authorization",
		},
		newTestHTTPClient(srv),
	)

	got, err := c.ChatCompletion(context.Background(), []Message{
		{Role: RoleUser, Content: "hi"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "hello world" {
		t.Errorf("got %q, want %q", got, "hello world")
	}
}

func TestChatCompletion_PrimaryFailFallbackSuccess(t *testing.T) {
	// Primary server returns an API error.
	primary := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := chatResponse{
			Error: &chatError{Message: "rate limited", Type: "rate_limit"},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer primary.Close()

	// Fallback server returns a valid response.
	fallback := httptest.NewServer(successHandler("fallback reply"))
	defer fallback.Close()

	c := New(
		config.LLMConfig{
			Provider:   config.LLMAnuma,
			APIKey:     "anuma_live_k1.secret",
			Model:      "auto",
			BaseURL:    primary.URL,
			MaxTokens:  100,
			AuthHeader: config.AnumaAuthHeader,
			Fallback: &config.LLMConfig{
				Provider:   config.LLMOpenAI,
				APIKey:     "sk-test",
				Model:      "gpt-test",
				BaseURL:    fallback.URL,
				MaxTokens:  100,
				AuthHeader: "Authorization",
			},
		},
		newTestHTTPClient(primary),
	)

	got, err := c.ChatCompletion(context.Background(), []Message{
		{Role: RoleUser, Content: "hi"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "fallback reply" {
		t.Errorf("got %q, want %q", got, "fallback reply")
	}
}

func TestChatCompletion_PrimaryFailNoFallback(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := chatResponse{
			Error: &chatError{Message: "bad request", Type: "invalid_request"},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	c := New(
		config.LLMConfig{
			Provider:   config.LLMOpenAI,
			APIKey:     "test-key",
			Model:      "gpt-test",
			BaseURL:    srv.URL,
			MaxTokens:  100,
			AuthHeader: "Authorization",
			// No fallback.
		},
		newTestHTTPClient(srv),
	)

	_, err := c.ChatCompletion(context.Background(), []Message{
		{Role: RoleUser, Content: "hi"},
	})
	if err == nil {
		t.Fatal("expected error when primary fails and no fallback configured")
	}
	if !strings.Contains(err.Error(), "bad request") {
		t.Errorf("expected 'bad request' in error, got: %v", err)
	}
}

func TestChatCompletion_APIErrorResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := chatResponse{
			Error: &chatError{Message: "model not found", Type: "not_found"},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	c := New(
		config.LLMConfig{
			Provider:   config.LLMOpenAI,
			APIKey:     "test-key",
			Model:      "nonexistent",
			BaseURL:    srv.URL,
			MaxTokens:  100,
			AuthHeader: "Authorization",
		},
		newTestHTTPClient(srv),
	)

	_, err := c.ChatCompletion(context.Background(), []Message{
		{Role: RoleUser, Content: "hi"},
	})
	if err == nil {
		t.Fatal("expected error for API error response")
	}
	if !strings.Contains(err.Error(), "model not found") {
		t.Errorf("expected 'model not found' in error, got: %v", err)
	}
	if !strings.Contains(err.Error(), "openai error") {
		t.Errorf("expected provider name in error, got: %v", err)
	}
}

func TestChatCompletion_EmptyChoices(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := chatResponse{
			Model:   "test-model",
			Choices: []chatChoice{},
			Usage:   chatUsage{PromptTokens: 5, CompletionTokens: 0, TotalTokens: 5},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	c := New(
		config.LLMConfig{
			Provider:   config.LLMOpenAI,
			APIKey:     "test-key",
			Model:      "gpt-test",
			BaseURL:    srv.URL,
			MaxTokens:  100,
			AuthHeader: "Authorization",
		},
		newTestHTTPClient(srv),
	)

	_, err := c.ChatCompletion(context.Background(), []Message{
		{Role: RoleUser, Content: "hi"},
	})
	if err == nil {
		t.Fatal("expected error for empty choices")
	}
	if !strings.Contains(err.Error(), "no choices") {
		t.Errorf("expected 'no choices' in error, got: %v", err)
	}
}

func TestChatCompletion_AnumaAuthHeader(t *testing.T) {
	var gotAuthHeader, gotAuthValue string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuthHeader = config.AnumaAuthHeader
		gotAuthValue = r.Header.Get(config.AnumaAuthHeader)

		resp := chatResponse{
			Model:   "auto",
			Choices: []chatChoice{{Message: Message{Role: RoleAssistant, Content: "anuma reply"}}},
			Usage:   chatUsage{PromptTokens: 8, CompletionTokens: 3, TotalTokens: 11},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	apiKey := "anuma_live_k1.secret123"
	c := New(
		config.LLMConfig{
			Provider:   config.LLMAnuma,
			APIKey:     apiKey,
			Model:      "auto",
			BaseURL:    srv.URL,
			MaxTokens:  100,
			AuthHeader: config.AnumaAuthHeader,
		},
		newTestHTTPClient(srv),
	)

	got, err := c.ChatCompletion(context.Background(), []Message{
		{Role: RoleUser, Content: "hi"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "anuma reply" {
		t.Errorf("got %q, want %q", got, "anuma reply")
	}

	// Anuma sends the raw API key (no "Bearer " prefix).
	if gotAuthHeader != config.AnumaAuthHeader {
		t.Errorf("auth header = %q, want %q", gotAuthHeader, config.AnumaAuthHeader)
	}
	if gotAuthValue != apiKey {
		t.Errorf("auth value = %q, want raw key %q (no Bearer prefix)", gotAuthValue, apiKey)
	}
}

func TestChatCompletion_BearerAuthHeader(t *testing.T) {
	var gotAuthValue string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuthValue = r.Header.Get("Authorization")

		resp := chatResponse{
			Model:   "gpt-test",
			Choices: []chatChoice{{Message: Message{Role: RoleAssistant, Content: "bearer reply"}}},
			Usage:   chatUsage{PromptTokens: 8, CompletionTokens: 3, TotalTokens: 11},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	apiKey := "sk-test-key-123"
	c := New(
		config.LLMConfig{
			Provider:   config.LLMOpenAI,
			APIKey:     apiKey,
			Model:      "gpt-test",
			BaseURL:    srv.URL,
			MaxTokens:  100,
			AuthHeader: "Authorization",
		},
		newTestHTTPClient(srv),
	)

	got, err := c.ChatCompletion(context.Background(), []Message{
		{Role: RoleUser, Content: "hi"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "bearer reply" {
		t.Errorf("got %q, want %q", got, "bearer reply")
	}

	// OpenAI-style sends "Bearer <key>".
	want := "Bearer " + apiKey
	if gotAuthValue != want {
		t.Errorf("auth value = %q, want %q", gotAuthValue, want)
	}
}

func TestChatCompletion_HTTPError(t *testing.T) {
	// Server that always returns 500.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("internal server error"))
	}))
	defer srv.Close()

	c := New(
		config.LLMConfig{
			Provider:   config.LLMOpenAI,
			APIKey:     "test-key",
			Model:      "gpt-test",
			BaseURL:    srv.URL,
			MaxTokens:  100,
			AuthHeader: "Authorization",
		},
		newTestHTTPClient(srv),
	)

	_, err := c.ChatCompletion(context.Background(), []Message{
		{Role: RoleUser, Content: "hi"},
	})
	if err == nil {
		t.Fatal("expected error for HTTP 500")
	}
	if !strings.Contains(err.Error(), "openai request") {
		t.Errorf("expected provider-prefixed error, got: %v", err)
	}
}

func TestChatCompletion_RequestBodySent(t *testing.T) {
	var gotReq chatRequest

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if ct := r.Header.Get("Content-Type"); ct != "application/json" {
			t.Errorf("Content-Type = %q, want application/json", ct)
		}

		_ = json.NewDecoder(r.Body).Decode(&gotReq)

		resp := chatResponse{
			Model:   "gpt-test",
			Choices: []chatChoice{{Message: Message{Role: RoleAssistant, Content: "ok"}}},
			Usage:   chatUsage{PromptTokens: 1, CompletionTokens: 1, TotalTokens: 2},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	c := New(
		config.LLMConfig{
			Provider:   config.LLMOpenAI,
			APIKey:     "test-key",
			Model:      "gpt-4o",
			BaseURL:    srv.URL,
			MaxTokens:  2048,
			AuthHeader: "Authorization",
		},
		newTestHTTPClient(srv),
	)

	msgs := []Message{
		{Role: RoleSystem, Content: "You are helpful."},
		{Role: RoleUser, Content: "Hello"},
	}
	_, err := c.ChatCompletion(context.Background(), msgs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotReq.Model != "gpt-4o" {
		t.Errorf("request model = %q, want %q", gotReq.Model, "gpt-4o")
	}
	if gotReq.MaxTokens != 2048 {
		t.Errorf("request max_tokens = %d, want %d", gotReq.MaxTokens, 2048)
	}
	if len(gotReq.Messages) != 2 {
		t.Fatalf("request messages len = %d, want 2", len(gotReq.Messages))
	}
	if gotReq.Messages[0].Role != RoleSystem || gotReq.Messages[0].Content != "You are helpful." {
		t.Errorf("message[0] = %+v, want system/You are helpful.", gotReq.Messages[0])
	}
}

func TestChatCompletion_URLPath(t *testing.T) {
	var gotPath string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path

		resp := chatResponse{
			Model:   "test",
			Choices: []chatChoice{{Message: Message{Role: RoleAssistant, Content: "ok"}}},
			Usage:   chatUsage{},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	c := New(
		config.LLMConfig{
			Provider:   config.LLMOpenAI,
			APIKey:     "k",
			Model:      "m",
			BaseURL:    srv.URL,
			MaxTokens:  1,
			AuthHeader: "Authorization",
		},
		newTestHTTPClient(srv),
	)

	_, _ = c.ChatCompletion(context.Background(), []Message{{Role: RoleUser, Content: "hi"}})

	if gotPath != "/chat/completions" {
		t.Errorf("request path = %q, want /chat/completions", gotPath)
	}
}

func TestNew(t *testing.T) {
	cfg := config.LLMConfig{
		Provider: config.LLMOpenAI,
		APIKey:   "key",
		Model:    "model",
	}
	httpCfg := config.HTTPConfig{
		Timeout:         time.Second,
		MaxRetries:      1,
		RetryBaseDelay:  time.Millisecond,
		MaxIdleConns:    1,
		IdleConnTimeout: time.Second,
	}
	hc := httpclient.New(httpCfg)
	c := New(cfg, hc)
	if c == nil {
		t.Fatal("New returned nil")
	}
}
