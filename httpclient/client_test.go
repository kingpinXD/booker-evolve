package httpclient

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

// testClient creates a Client that routes requests through the given test server.
func testClient(srv *httptest.Server, maxRetries int) *Client {
	return &Client{
		http:       srv.Client(),
		maxRetries: maxRetries,
		retryDelay: time.Millisecond,
	}
}

// ---------- GetJSON ----------

func TestGetJSON_Success(t *testing.T) {
	type resp struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if got := r.Header.Get("X-Custom"); got != "value" {
			t.Errorf("expected header X-Custom=value, got %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp{Name: "alice", Age: 30})
	}))
	defer srv.Close()

	c := testClient(srv, 1)
	var got resp
	err := c.GetJSON(context.Background(), srv.URL, map[string]string{"X-Custom": "value"}, &got)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Name != "alice" || got.Age != 30 {
		t.Errorf("got %+v, want {alice 30}", got)
	}
}

func TestGetJSON_NilDest(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ignored":true}`))
	}))
	defer srv.Close()

	c := testClient(srv, 1)
	if err := c.GetJSON(context.Background(), srv.URL, nil, nil); err != nil {
		t.Fatalf("unexpected error with nil dest: %v", err)
	}
}

func TestGetJSON_InvalidURL(t *testing.T) {
	c := &Client{http: http.DefaultClient, maxRetries: 1, retryDelay: time.Millisecond}
	err := c.GetJSON(context.Background(), "://bad", nil, nil)
	if err == nil || !strings.Contains(err.Error(), "building request") {
		t.Fatalf("expected building request error, got: %v", err)
	}
}

// ---------- PostJSON ----------

func TestPostJSON_Success(t *testing.T) {
	type reqBody struct {
		Query string `json:"query"`
	}
	type respBody struct {
		Result string `json:"result"`
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		var rb reqBody
		if err := json.NewDecoder(r.Body).Decode(&rb); err != nil {
			t.Fatalf("decoding request body: %v", err)
		}
		if rb.Query != "flights" {
			t.Errorf("expected query=flights, got %q", rb.Query)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(respBody{Result: "ok"})
	}))
	defer srv.Close()

	c := testClient(srv, 1)
	var got respBody
	err := c.PostJSON(context.Background(), srv.URL, reqBody{Query: "flights"}, nil, &got)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Result != "ok" {
		t.Errorf("got %+v, want {ok}", got)
	}
}

func TestPostJSON_MarshalError(t *testing.T) {
	c := &Client{http: http.DefaultClient, maxRetries: 1, retryDelay: time.Millisecond}
	// channels cannot be marshaled to JSON
	err := c.PostJSON(context.Background(), "http://localhost", make(chan int), nil, nil)
	if err == nil || !strings.Contains(err.Error(), "marshaling request body") {
		t.Fatalf("expected marshal error, got: %v", err)
	}
}

// ---------- Retry logic ----------

func TestDoJSON_RetriesOn500(t *testing.T) {
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		n := calls.Add(1)
		if n < 3 {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("temporary failure"))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "recovered"})
	}))
	defer srv.Close()

	c := testClient(srv, 3)
	var got map[string]string
	err := c.GetJSON(context.Background(), srv.URL, nil, &got)
	if err != nil {
		t.Fatalf("expected success after retries, got: %v", err)
	}
	if got["status"] != "recovered" {
		t.Errorf("got %v, want recovered", got)
	}
	if n := calls.Load(); n != 3 {
		t.Errorf("expected 3 calls, got %d", n)
	}
}

func TestDoJSON_AllRetriesExhausted(t *testing.T) {
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("always failing"))
	}))
	defer srv.Close()

	c := testClient(srv, 3)
	err := c.GetJSON(context.Background(), srv.URL, nil, nil)
	if err == nil {
		t.Fatal("expected error after exhausting retries")
	}
	if !strings.Contains(err.Error(), "all 3 attempts failed") {
		t.Errorf("unexpected error message: %v", err)
	}
	if !strings.Contains(err.Error(), "server error: 500") {
		t.Errorf("expected server error in message: %v", err)
	}
	if n := calls.Load(); n != 3 {
		t.Errorf("expected 3 calls, got %d", n)
	}
}

// ---------- No retry on 4xx ----------

func TestDoJSON_NoRetryOn400(t *testing.T) {
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("bad request"))
	}))
	defer srv.Close()

	c := testClient(srv, 3)
	err := c.GetJSON(context.Background(), srv.URL, nil, nil)
	if err == nil {
		t.Fatal("expected error for 400 response")
	}
	if !strings.Contains(err.Error(), "client error: 400") {
		t.Errorf("unexpected error: %v", err)
	}
	if n := calls.Load(); n != 1 {
		t.Errorf("expected exactly 1 call (no retry), got %d", n)
	}
}

func TestDoJSON_NoRetryOn404(t *testing.T) {
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte("not found"))
	}))
	defer srv.Close()

	c := testClient(srv, 3)
	err := c.GetJSON(context.Background(), srv.URL, nil, nil)
	if err == nil {
		t.Fatal("expected error for 404 response")
	}
	if !strings.Contains(err.Error(), "client error: 404") {
		t.Errorf("unexpected error: %v", err)
	}
	if n := calls.Load(); n != 1 {
		t.Errorf("expected exactly 1 call (no retry), got %d", n)
	}
}

// ---------- JSON decode errors ----------

func TestDoJSON_InvalidJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("not json at all"))
	}))
	defer srv.Close()

	c := testClient(srv, 1)
	var dest map[string]string
	err := c.GetJSON(context.Background(), srv.URL, nil, &dest)
	if err == nil {
		t.Fatal("expected decode error for invalid JSON")
	}
	if !strings.Contains(err.Error(), "decoding response") {
		t.Errorf("unexpected error: %v", err)
	}
}

// ---------- BuildURL ----------

func TestBuildURL(t *testing.T) {
	tests := []struct {
		name   string
		base   string
		path   string
		params map[string]string
		want   string
	}{
		{
			name:   "base with path and params",
			base:   "https://api.example.com",
			path:   "/v1/search",
			params: map[string]string{"q": "flights", "limit": "10"},
			want:   "https://api.example.com/v1/search",
		},
		{
			name:   "no params",
			base:   "https://api.example.com",
			path:   "/health",
			params: nil,
			want:   "https://api.example.com/health",
		},
		{
			name:   "single param",
			base:   "https://api.example.com",
			path:   "/search",
			params: map[string]string{"key": "abc123"},
			want:   "https://api.example.com/search?key=abc123",
		},
		{
			name:   "empty base and path",
			base:   "",
			path:   "",
			params: nil,
			want:   "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := BuildURL(tt.base, tt.path, tt.params)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			// For cases with multiple params, verify each param is present
			// since map iteration order is non-deterministic.
			if len(tt.params) > 1 {
				for k, v := range tt.params {
					if !strings.Contains(got, k+"="+v) {
						t.Errorf("BuildURL() = %q, missing param %s=%s", got, k, v)
					}
				}
				return
			}
			if got != tt.want {
				t.Errorf("BuildURL() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestBuildURL_InvalidBase(t *testing.T) {
	_, err := BuildURL("://invalid", "/path", nil)
	if err == nil {
		t.Fatal("expected error for invalid base URL")
	}
	if !strings.Contains(err.Error(), "parsing url") {
		t.Errorf("unexpected error: %v", err)
	}
}

// ---------- Connection error retries ----------

func TestDoJSON_RetriesOnConnectionError(t *testing.T) {
	// Create a server and immediately close it to force connection errors.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {}))
	closedURL := srv.URL
	srv.Close()

	c := &Client{
		http:       &http.Client{Timeout: 50 * time.Millisecond},
		maxRetries: 2,
		retryDelay: time.Millisecond,
	}
	err := c.GetJSON(context.Background(), closedURL, nil, nil)
	if err == nil {
		t.Fatal("expected error for closed server")
	}
	if !strings.Contains(err.Error(), "all 2 attempts failed") {
		t.Errorf("unexpected error: %v", err)
	}
}
