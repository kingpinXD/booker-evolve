package config

import (
	"testing"
	"time"
)

func TestDefault_ProvidersPopulated(t *testing.T) {
	cfg := Default()

	if len(cfg.Providers) != 2 {
		t.Fatalf("got %d providers, want 2", len(cfg.Providers))
	}

	kiwi, ok := cfg.Providers[ProviderKiwi]
	if !ok {
		t.Fatal("missing kiwi provider config")
	}
	if !kiwi.Enabled {
		t.Error("kiwi should be enabled by default")
	}
	if kiwi.Timeout != 30*time.Second {
		t.Errorf("kiwi timeout = %v, want 30s", kiwi.Timeout)
	}
	if kiwi.BaseURL != KiwiBaseURL {
		t.Errorf("kiwi BaseURL = %q, want %q", kiwi.BaseURL, KiwiBaseURL)
	}

	serpapi, ok := cfg.Providers[ProviderSerpAPI]
	if !ok {
		t.Fatal("missing serpapi provider config")
	}
	if !serpapi.Enabled {
		t.Error("serpapi should be enabled by default")
	}
	if serpapi.BaseURL != SerpAPIBaseURL {
		t.Errorf("serpapi BaseURL = %q, want %q", serpapi.BaseURL, SerpAPIBaseURL)
	}
}

func TestDefault_LLMFallbackChain(t *testing.T) {
	cfg := Default()

	if cfg.LLM.Provider != LLMAnuma {
		t.Errorf("primary LLM provider = %q, want %q", cfg.LLM.Provider, LLMAnuma)
	}
	if cfg.LLM.BaseURL != AnumaBaseURL {
		t.Errorf("primary BaseURL = %q, want %q", cfg.LLM.BaseURL, AnumaBaseURL)
	}
	if cfg.LLM.AuthHeader != AnumaAuthHeader {
		t.Errorf("primary AuthHeader = %q, want %q", cfg.LLM.AuthHeader, AnumaAuthHeader)
	}

	fb := cfg.LLM.Fallback
	if fb == nil {
		t.Fatal("LLM fallback is nil, want OpenAI fallback")
	}
	if fb.Provider != LLMOpenAI {
		t.Errorf("fallback provider = %q, want %q", fb.Provider, LLMOpenAI)
	}
	if fb.BaseURL != OpenAIBaseURL {
		t.Errorf("fallback BaseURL = %q, want %q", fb.BaseURL, OpenAIBaseURL)
	}
	if fb.Fallback != nil {
		t.Error("fallback should not have a nested fallback")
	}
}

func TestDefault_HTTPConfig(t *testing.T) {
	cfg := Default()

	if cfg.HTTP.Timeout != 30*time.Second {
		t.Errorf("HTTP timeout = %v, want 30s", cfg.HTTP.Timeout)
	}
	if cfg.HTTP.MaxRetries != 3 {
		t.Errorf("HTTP MaxRetries = %d, want 3", cfg.HTTP.MaxRetries)
	}
	if cfg.HTTP.MaxIdleConns != 10 {
		t.Errorf("HTTP MaxIdleConns = %d, want 10", cfg.HTTP.MaxIdleConns)
	}
}
