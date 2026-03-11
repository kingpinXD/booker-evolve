package search

import (
	"context"
	"errors"
	"strings"
	"testing"

	"booker/llm"
	"booker/types"
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

	got, reason, err := p.Pick(context.Background(), Request{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Name() != "direct" {
		t.Errorf("got strategy %q, want %q", got.Name(), "direct")
	}
	if reason == "" {
		t.Error("expected non-empty reason for single strategy")
	}
}

func TestPicker_FallbackNoContext(t *testing.T) {
	direct := &fakeStrategy{name: "direct", desc: "Direct flights"}
	mc := &fakeStrategy{name: "multicity", desc: "Multi-city with stopover"}
	p := NewPicker(nil, direct, mc)

	// No context provided — should fall back to "direct".
	got, reason, err := p.Pick(context.Background(), Request{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Name() != "direct" {
		t.Errorf("got strategy %q, want %q", got.Name(), "direct")
	}
	if reason == "" {
		t.Error("expected non-empty reason for fallback")
	}
}

func TestPicker_FallbackWhenNoDirectExists(t *testing.T) {
	mc := &fakeStrategy{name: "multicity", desc: "Multi-city with stopover"}
	p := NewPicker(nil, mc)

	got, reason, err := p.Pick(context.Background(), Request{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Name() != "multicity" {
		t.Errorf("got strategy %q, want %q", got.Name(), "multicity")
	}
	if reason == "" {
		t.Error("expected non-empty reason")
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

	got, reason, err := p.Pick(context.Background(), Request{Context: "I want a stopover in Istanbul"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Name() != "multicity" {
		t.Errorf("got strategy %q, want %q", got.Name(), "multicity")
	}
	if reason != "user wants stopover" {
		t.Errorf("reason = %q, want %q", reason, "user wants stopover")
	}
}

func TestPicker_LLMReturnsJSONInMarkdownFences(t *testing.T) {
	direct := &fakeStrategy{name: "direct", desc: "Direct flights"}
	mc := &fakeStrategy{name: "multicity", desc: "Multi-city with stopover"}
	mock := &mockLLM{response: "```json\n{\"strategy\": \"direct\", \"reason\": \"simple route\"}\n```"}
	p := NewPicker(mock, direct, mc)

	got, reason, err := p.Pick(context.Background(), Request{Context: "JFK to LAX"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Name() != "direct" {
		t.Errorf("got strategy %q, want %q", got.Name(), "direct")
	}
	if reason != "simple route" {
		t.Errorf("reason = %q, want %q", reason, "simple route")
	}
}

func TestPicker_LLMReturnsUnknownStrategy(t *testing.T) {
	direct := &fakeStrategy{name: "direct", desc: "Direct flights"}
	mc := &fakeStrategy{name: "multicity", desc: "Multi-city with stopover"}
	mock := &mockLLM{response: `{"strategy": "nonexistent", "reason": "oops"}`}
	p := NewPicker(mock, direct, mc)

	got, reason, err := p.Pick(context.Background(), Request{Context: "some context"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should fall back to "direct" since LLM returned unknown strategy.
	if got.Name() != "direct" {
		t.Errorf("got strategy %q, want %q (fallback)", got.Name(), "direct")
	}
	if reason == "" {
		t.Error("expected non-empty reason on LLM fallback")
	}
}

func TestPicker_LLMReturnsInvalidJSON(t *testing.T) {
	direct := &fakeStrategy{name: "direct", desc: "Direct flights"}
	mc := &fakeStrategy{name: "multicity", desc: "Multi-city with stopover"}
	mock := &mockLLM{response: "not json at all"}
	p := NewPicker(mock, direct, mc)

	got, reason, err := p.Pick(context.Background(), Request{Context: "some context"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Name() != "direct" {
		t.Errorf("got strategy %q, want %q (fallback)", got.Name(), "direct")
	}
	if reason == "" {
		t.Error("expected non-empty reason on LLM fallback")
	}
}

func TestPicker_LLMReturnsError(t *testing.T) {
	direct := &fakeStrategy{name: "direct", desc: "Direct flights"}
	mc := &fakeStrategy{name: "multicity", desc: "Multi-city with stopover"}
	mock := &mockLLM{err: errors.New("LLM unavailable")}
	p := NewPicker(mock, direct, mc)

	got, reason, err := p.Pick(context.Background(), Request{Context: "some context"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Name() != "direct" {
		t.Errorf("got strategy %q, want %q (fallback)", got.Name(), "direct")
	}
	if reason != "LLM unavailable, using default" {
		t.Errorf("reason = %q, want %q", reason, "LLM unavailable, using default")
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
	if !strings.Contains(prompt, "both") {
		t.Error("system prompt missing 'both' option")
	}
}

func TestPicker_LLMReturnsBoth(t *testing.T) {
	direct := &fakeStrategy{name: "direct", desc: "Direct flights"}
	mc := &fakeStrategy{name: "multicity", desc: "Multi-city with stopover"}
	mock := &mockLLM{response: `{"strategy": "both", "reason": "compare options"}`}
	p := NewPicker(mock, direct, mc)

	got, reason, err := p.Pick(context.Background(), Request{Context: "not sure which is better"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Name() != "composite" {
		t.Errorf("got strategy %q, want %q", got.Name(), "composite")
	}
	if reason != "compare options" {
		t.Errorf("reason = %q, want %q", reason, "compare options")
	}
}

func TestPicker_LLMReturnsBoth_SingleStrategy(t *testing.T) {
	direct := &fakeStrategy{name: "direct", desc: "Direct flights"}
	mock := &mockLLM{response: `{"strategy": "both", "reason": "compare"}`}
	// With only one strategy registered, "both" should return it directly.
	p := NewPicker(mock, direct)

	got, reason, err := p.Pick(context.Background(), Request{Context: "compare"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Name() != "direct" {
		t.Errorf("got strategy %q, want %q", got.Name(), "direct")
	}
	if reason == "" {
		t.Error("expected non-empty reason")
	}
}

// fakeRanker assigns incrementing scores to itineraries.
type fakeRanker struct{ called bool }

func (r *fakeRanker) Rank(_ context.Context, itins []Itinerary) ([]Itinerary, error) {
	r.called = true
	for i := range itins {
		itins[i].Score = float64(len(itins) - i)
	}
	return itins, nil
}

// fakeSearchStrategy returns canned itineraries.
type fakeSearchStrategy struct {
	name  string
	itins []Itinerary
}

func (f *fakeSearchStrategy) Name() string        { return f.name }
func (f *fakeSearchStrategy) Description() string { return f.name + " strategy" }
func (f *fakeSearchStrategy) Search(_ context.Context, _ Request) ([]Itinerary, error) {
	return f.itins, nil
}

func TestPicker_BothPassesRankerToComposite(t *testing.T) {
	s1 := &fakeSearchStrategy{
		name: "direct",
		itins: []Itinerary{
			{TotalPrice: types.Money{Amount: 500, Currency: "USD"}},
		},
	}
	s2 := &fakeSearchStrategy{
		name: "multicity",
		itins: []Itinerary{
			{TotalPrice: types.Money{Amount: 700, Currency: "USD"}},
		},
	}
	ranker := &fakeRanker{}
	mock := &mockLLM{response: `{"strategy": "both", "reason": "compare"}`}
	p := NewPicker(mock, s1, s2)
	p.SetRanker(ranker)

	strategy, _, err := p.Pick(context.Background(), Request{Context: "compare options"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strategy.Name() != "composite" {
		t.Fatalf("got strategy %q, want composite", strategy.Name())
	}

	results, err := strategy.Search(context.Background(), Request{})
	if err != nil {
		t.Fatalf("search error: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("got %d results, want 2", len(results))
	}
	if !ranker.called {
		t.Error("ranker was not called — composite strategy should use picker's ranker")
	}
	for _, r := range results {
		if r.Score == 0 {
			t.Error("expected non-zero score after ranking")
		}
	}
}
