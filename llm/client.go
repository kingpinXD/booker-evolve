package llm

import (
	"context"
	"fmt"

	"booker/config"
	"booker/httpclient"
)

// Role constants for chat messages.
const (
	RoleSystem    = "system"
	RoleUser      = "user"
	RoleAssistant = "assistant"
)

// Message is a single chat message.
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Client talks to an LLM backend.
type Client struct {
	cfg  config.LLMConfig
	http *httpclient.Client
}

// New creates an LLM client.
func New(cfg config.LLMConfig, http *httpclient.Client) *Client {
	return &Client{cfg: cfg, http: http}
}

// ChatCompletion sends messages and returns the assistant's response text.
func (c *Client) ChatCompletion(ctx context.Context, messages []Message) (string, error) {
	switch c.cfg.Provider {
	case config.LLMOpenAI:
		return c.openAICompletion(ctx, messages)
	default:
		return "", fmt.Errorf("unsupported LLM provider: %s", c.cfg.Provider)
	}
}

// openAIRequest is the request body for OpenAI chat completions.
type openAIRequest struct {
	Model     string    `json:"model"`
	Messages  []Message `json:"messages"`
	MaxTokens int       `json:"max_tokens"`
}

// openAIResponse is the response from OpenAI chat completions.
type openAIResponse struct {
	Choices []openAIChoice `json:"choices"`
	Error   *openAIError   `json:"error"`
}

type openAIChoice struct {
	Message Message `json:"message"`
}

type openAIError struct {
	Message string `json:"message"`
	Type    string `json:"type"`
}

func (c *Client) openAICompletion(ctx context.Context, messages []Message) (string, error) {
	reqBody := openAIRequest{
		Model:     c.cfg.Model,
		Messages:  messages,
		MaxTokens: c.cfg.MaxTokens,
	}

	url := c.cfg.BaseURL + config.OpenAIChatCompletions
	headers := map[string]string{
		config.HeaderContentType: config.ContentTypeJSON,
		"Authorization":          "Bearer " + c.cfg.APIKey,
	}

	var resp openAIResponse
	if err := c.http.PostJSON(ctx, url, reqBody, headers, &resp); err != nil {
		return "", fmt.Errorf("openai request: %w", err)
	}

	if resp.Error != nil {
		return "", fmt.Errorf("openai error: %s", resp.Error.Message)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("openai returned no choices")
	}

	return resp.Choices[0].Message.Content, nil
}
