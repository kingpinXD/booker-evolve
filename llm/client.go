package llm

import (
	"context"
	"fmt"
	"log"
	"strings"

	"booker/config"
	"booker/httpclient"
)

// Role constants for chat messages.
const (
	RoleSystem    = "system"
	RoleUser      = "user"
	RoleAssistant = "assistant"
)

// StripCodeFences removes markdown code fences (```json and ```) that LLMs
// sometimes wrap around JSON responses, and trims surrounding whitespace.
func StripCodeFences(s string) string {
	s = strings.TrimPrefix(s, "```json")
	s = strings.TrimPrefix(s, "```")
	s = strings.TrimSuffix(s, "```")
	return strings.TrimSpace(s)
}

// Message is a single chat message.
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Client talks to an LLM backend. Tries primary config first,
// falls back to Fallback config on failure.
type Client struct {
	cfg  config.LLMConfig
	http *httpclient.Client
}

// New creates an LLM client.
func New(cfg config.LLMConfig, http *httpclient.Client) *Client {
	return &Client{cfg: cfg, http: http}
}

// ChatCompletion sends messages and returns the assistant's response text.
// Tries the primary provider first, falls back if configured.
func (c *Client) ChatCompletion(ctx context.Context, messages []Message) (string, error) {
	resp, err := c.chatCompletion(ctx, c.cfg, messages)
	if err == nil {
		return resp, nil
	}

	if c.cfg.Fallback == nil {
		return "", err
	}

	log.Printf("[llm] %s failed (%v), falling back to %s", c.cfg.Provider, err, c.cfg.Fallback.Provider)
	return c.chatCompletion(ctx, *c.cfg.Fallback, messages)
}

// chatCompletion makes a chat completion request to a specific provider config.
// Both Anuma and OpenAI use the same request/response format.
func (c *Client) chatCompletion(ctx context.Context, cfg config.LLMConfig, messages []Message) (string, error) {
	reqBody := chatRequest{
		Model:     cfg.Model,
		Messages:  messages,
		MaxTokens: cfg.MaxTokens,
	}

	url := cfg.BaseURL + "/chat/completions"

	// Build auth header based on provider.
	var authValue string
	switch cfg.AuthHeader {
	case config.AnumaAuthHeader:
		authValue = cfg.APIKey
	default:
		authValue = "Bearer " + cfg.APIKey
	}

	headers := map[string]string{
		config.HeaderContentType: config.ContentTypeJSON,
		cfg.AuthHeader:           authValue,
	}

	var resp chatResponse
	if err := c.http.PostJSON(ctx, url, reqBody, headers, &resp); err != nil {
		return "", fmt.Errorf("%s request: %w", cfg.Provider, err)
	}

	if resp.Error != nil {
		return "", fmt.Errorf("%s error: %s", cfg.Provider, resp.Error.Message)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("%s returned no choices", cfg.Provider)
	}

	log.Printf("[llm] %s (%s): %d prompt + %d completion tokens",
		cfg.Provider, resp.Model, resp.Usage.PromptTokens, resp.Usage.CompletionTokens)

	return resp.Choices[0].Message.Content, nil
}

// chatRequest is the OpenAI-compatible request body.
type chatRequest struct {
	Model     string    `json:"model"`
	Messages  []Message `json:"messages"`
	MaxTokens int       `json:"max_tokens"`
}

// chatResponse is the OpenAI-compatible response.
type chatResponse struct {
	Model   string       `json:"model"`
	Choices []chatChoice `json:"choices"`
	Usage   chatUsage    `json:"usage"`
	Error   *chatError   `json:"error"`
}

type chatChoice struct {
	Message Message `json:"message"`
}

type chatUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type chatError struct {
	Message string `json:"message"`
	Type    string `json:"type"`
}
