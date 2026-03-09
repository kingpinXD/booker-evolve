package config

import (
	"os"
	"time"
)

// ProviderName identifies a flight data provider.
type ProviderName string

const (
	ProviderKiwi    ProviderName = "kiwi"
	ProviderSerpAPI ProviderName = "serpapi"
)

// LLMProvider identifies which LLM backend to use.
type LLMProvider string

const (
	LLMOpenAI LLMProvider = "openai"
)

// Config is the top-level application configuration.
type Config struct {
	Providers map[ProviderName]ProviderConfig
	LLM       LLMConfig
	HTTP      HTTPConfig
}

// LLMConfig holds settings for the LLM backend.
type LLMConfig struct {
	Provider LLMProvider
	APIKey   string
	Model    string
	BaseURL  string
	MaxTokens int
}

// ProviderConfig holds credentials and settings for a single provider.
type ProviderConfig struct {
	APIKey  string
	APIHost string // RapidAPI host header value
	BaseURL string
	Enabled bool
	Timeout time.Duration
}

// HTTPConfig holds shared HTTP client settings.
type HTTPConfig struct {
	Timeout         time.Duration
	MaxIdleConns    int
	IdleConnTimeout time.Duration
	MaxRetries      int
	RetryBaseDelay  time.Duration
}

// Environment variable names for API keys.
const (
	EnvKiwiAPIKey    = "BOOKER_KIWI_API_KEY"
	EnvSerpAPIKey    = "BOOKER_SERPAPI_KEY"
	EnvOpenAIAPIKey  = "BOOKER_OPENAI_API_KEY"
)

// Default returns the default configuration.
// API keys are read from environment variables.
func Default() Config {
	return Config{
		Providers: map[ProviderName]ProviderConfig{
			ProviderKiwi: {
				APIKey:  os.Getenv(EnvKiwiAPIKey),
				APIHost: KiwiRapidAPIHost,
				BaseURL: KiwiBaseURL,
				Enabled: true,
				Timeout: 30 * time.Second,
			},
			ProviderSerpAPI: {
				APIKey:  os.Getenv(EnvSerpAPIKey),
				BaseURL: SerpAPIBaseURL,
				Enabled: true,
				Timeout: 30 * time.Second,
			},
		},
		LLM: LLMConfig{
			Provider:  LLMOpenAI,
			APIKey:    os.Getenv(EnvOpenAIAPIKey),
			Model:     OpenAIModelDefault,
			BaseURL:   OpenAIBaseURL,
			MaxTokens: OpenAIMaxTokensDefault,
		},
		HTTP: HTTPConfig{
			Timeout:         30 * time.Second,
			MaxIdleConns:    10,
			IdleConnTimeout: 90 * time.Second,
			MaxRetries:      3,
			RetryBaseDelay:  500 * time.Millisecond,
		},
	}
}
