package httpclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"booker/config"
)

// Client wraps http.Client with retry logic and common helpers
// shared across all providers.
type Client struct {
	http       *http.Client
	maxRetries int
	retryDelay time.Duration
}

// New creates a Client from the shared HTTP config.
func New(cfg config.HTTPConfig) *Client {
	return &Client{
		http: &http.Client{
			Timeout: cfg.Timeout,
			Transport: &http.Transport{
				MaxIdleConns:    cfg.MaxIdleConns,
				IdleConnTimeout: cfg.IdleConnTimeout,
			},
		},
		maxRetries: cfg.MaxRetries,
		retryDelay: cfg.RetryBaseDelay,
	}
}

// GetJSON performs a GET request and decodes the JSON response into dest.
func (c *Client) GetJSON(ctx context.Context, rawURL string, headers map[string]string, dest any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return fmt.Errorf("building request: %w", err)
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	return c.doJSON(req, dest)
}


// PostJSON performs a POST with a JSON body and decodes the JSON response into dest.
func (c *Client) PostJSON(ctx context.Context, rawURL string, body any, headers map[string]string, dest any) error {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("marshaling request body: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, rawURL, bytes.NewReader(jsonBody))
	if err != nil {
		return fmt.Errorf("building request: %w", err)
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	return c.doJSON(req, dest)
}

// doJSON executes the request with retries and decodes the response.
func (c *Client) doJSON(req *http.Request, dest any) error {
	var lastErr error
	for attempt := range c.maxRetries {
		if attempt > 0 {
			time.Sleep(c.retryDelay * time.Duration(1<<(attempt-1)))
		}

		resp, err := c.http.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("request failed: %w", err)
			continue
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			lastErr = fmt.Errorf("reading response: %w", err)
			continue
		}

		if resp.StatusCode >= 500 {
			lastErr = fmt.Errorf("server error: %d %s", resp.StatusCode, string(body))
			continue
		}

		if resp.StatusCode >= 400 {
			return fmt.Errorf("client error: %d %s", resp.StatusCode, string(body))
		}

		if dest != nil {
			if err := json.Unmarshal(body, dest); err != nil {
				return fmt.Errorf("decoding response: %w", err)
			}
		}
		return nil
	}
	return fmt.Errorf("all %d attempts failed, last error: %w", c.maxRetries, lastErr)
}

// BuildURL constructs a full URL from base, path, and query parameters.
func BuildURL(base, path string, params map[string]string) (string, error) {
	u, err := url.Parse(base + path)
	if err != nil {
		return "", fmt.Errorf("parsing url: %w", err)
	}
	q := u.Query()
	for k, v := range params {
		q.Set(k, v)
	}
	u.RawQuery = q.Encode()
	return u.String(), nil
}
