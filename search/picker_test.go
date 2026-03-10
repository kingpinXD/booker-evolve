package search

import (
	"context"
	"errors"
	"strings"
	"testing"

	"booker/llm"
)

// fakeStrategy implements Strategy for testing.
type fakeStrategy struct {
	name string
	desc string
}

func (f *fakeStrategy) Name() string        { return f.name }
func (f *fakeStrategy) Description() string { return f.desc }
func (f *fakeStrategy) Search(_ context.Context, _ Request) ([]Itinerary, error) {
	return nil, nil
}

func TestPicker_SingleStrategy(t *testing.T) {
	direct := &fakeStrategy{name: "direct", desc: "Direct flights"}
	p := NewPicker(nil, direct)

	got, err := p.Pick(context.Background(), Request{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Name() != "direct" {
		t.Errorf("got strategy %q, want %q", got.Name(), "direct")
	}
}

func TestPicker_FallbackNoContext(t *testing.T) {
	direct := &fakeStrategy{name: "direct", desc: "Direct flights"}
	mc := &fakeStrategy{name: "multicity", desc: "Multi-city with stopover"}
	p := NewPicker(nil, direct, mc)

	// No context provided — should fall back to "direct".
	got, err := p.Pick(context.Background(), Request{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Name() != "direct" {
		t.Errorf("got strategy %q, want %q", got.Name(), "direct")
	}
}

func TestPicker_FallbackWhenNoDirectExists(t *testing.T) {
	mc := &fakeStrategy{name: "multicity", desc: "Multi-city with stopover"}
	p := NewPicker(nil, mc)

	got, err := p.Pick(context.Background(), Request{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Name() != "multicity" {
		t.Errorf("got strategy %q, want %q", got.Name(), "multicity")
	}
}

// mockLLM implements ChatCompleter for testing.
type mockLLM struct {
	response string
	err      error
}

func (m *mockLLM) ChatCompletion(_ context.Context, _ []llm.Message) (string, error) {
	return m.response, m.err
}

func TestPicker_LLMReturnsValidJSON(t *testing.T) {
	direct := &fakeStrategy{name: "direct", desc: "Direct flights"}
	mc := &fakeStrategy{name: "multicity", desc: "Multi-city with stopover"}
	mock := &mockLLM{response: `{"strategy": "multicity", "reason": "user wants stopover"}`}
	p := NewPicker(mock, direct, mc)

	got, err := p.Pick(context.Background(), Request{Context: "I want a stopover in Istanbul"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Name() != "multicity" {
		t.Errorf("got strategy %q, want %q", got.Name(), "multicity")
	}
}

func TestPicker_LLMReturnsJSONInMarkdownFences(t *testing.T) {
	direct := &fakeStrategy{name: "direct", desc: "Direct flights"}
	mc := &fakeStrategy{name: "multicity", desc: "Multi-city with stopover"}
	mock := &mockLLM{response: "```json\n{\"strategy\": \"direct\", \"reason\": \"simple route\"}\n```"}
	p := NewPicker(mock, direct, mc)

	got, err := p.Pick(context.Background(), Request{Context: "JFK to LAX"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Name() != "direct" {
		t.Errorf("got strategy %q, want %q", got.Name(), "direct")
	}
}

func TestPicker_LLMReturnsUnknownStrategy(t *testing.T) {
	direct := &fakeStrategy{name: "direct", desc: "Direct flights"}
	mc := &fakeStrategy{name: "multicity", desc: "Multi-city with stopover"}
	mock := &mockLLM{response: `{"strategy": "nonexistent", "reason": "oops"}`}
	p := NewPicker(mock, direct, mc)

	got, err := p.Pick(context.Background(), Request{Context: "some context"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should fall back to "direct" since LLM returned unknown strategy.
	if got.Name() != "direct" {
		t.Errorf("got strategy %q, want %q (fallback)", got.Name(), "direct")
	}
}

func TestPicker_LLMReturnsInvalidJSON(t *testing.T) {
	direct := &fakeStrategy{name: "direct", desc: "Direct flights"}
	mc := &fakeStrategy{name: "multicity", desc: "Multi-city with stopover"}
	mock := &mockLLM{response: "not json at all"}
	p := NewPicker(mock, direct, mc)

	got, err := p.Pick(context.Background(), Request{Context: "some context"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Name() != "direct" {
		t.Errorf("got strategy %q, want %q (fallback)", got.Name(), "direct")
	}
}

func TestPicker_LLMReturnsError(t *testing.T) {
	direct := &fakeStrategy{name: "direct", desc: "Direct flights"}
	mc := &fakeStrategy{name: "multicity", desc: "Multi-city with stopover"}
	mock := &mockLLM{err: errors.New("LLM unavailable")}
	p := NewPicker(mock, direct, mc)

	got, err := p.Pick(context.Background(), Request{Context: "some context"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Name() != "direct" {
		t.Errorf("got strategy %q, want %q (fallback)", got.Name(), "direct")
	}
}

func TestPicker_BuildSystemPrompt(t *testing.T) {
	direct := &fakeStrategy{name: "direct", desc: "Direct flights"}
	mc := &fakeStrategy{name: "multicity", desc: "Multi-city with stopover"}
	p := NewPicker(nil, direct, mc)

	prompt := p.buildSystemPrompt()
	if !strings.Contains(prompt, "direct") {
		t.Error("system prompt missing strategy name 'direct'")
	}
	if !strings.Contains(prompt, "Direct flights") {
		t.Error("system prompt missing strategy description 'Direct flights'")
	}
	if !strings.Contains(prompt, "multicity") {
		t.Error("system prompt missing strategy name 'multicity'")
	}
	if !strings.Contains(prompt, "Multi-city with stopover") {
		t.Error("system prompt missing strategy description 'Multi-city with stopover'")
	}
	if !strings.Contains(prompt, `"strategy"`) {
		t.Error("system prompt missing JSON format instruction")
	}
}
