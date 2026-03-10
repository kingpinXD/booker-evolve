package search

import (
	"context"
	"errors"
	"testing"
	"time"

	"booker/types"
)

// mockStrategy is a test double for Strategy.
type mockStrategy struct {
	name        string
	desc        string
	results     []Itinerary
	err         error
	searchCount int
}

func (m *mockStrategy) Name() string        { return m.name }
func (m *mockStrategy) Description() string { return m.desc }
func (m *mockStrategy) Search(ctx context.Context, req Request) ([]Itinerary, error) {
	m.searchCount++
	return m.results, m.err
}

// mockRanker records calls and returns itineraries with scores set.
type mockRanker struct {
	called bool
}

func (m *mockRanker) Rank(_ context.Context, itins []Itinerary) ([]Itinerary, error) {
	m.called = true
	for i := range itins {
		itins[i].Score = float64(100 - i*10)
	}
	return itins, nil
}

func makeItin(origin, dest string, price float64) Itinerary {
	return Itinerary{
		Legs: []Leg{{
			Flight: types.Flight{
				Outbound: []types.Segment{
					{Origin: origin, Destination: dest},
				},
				Price: types.Money{Amount: price, Currency: "USD"},
			},
		}},
		TotalPrice:  types.Money{Amount: price, Currency: "USD"},
		TotalTravel: 10 * time.Hour,
	}
}

func TestCompositeStrategy_Name(t *testing.T) {
	cs := NewCompositeStrategy(nil, nil)
	if cs.Name() != "composite" {
		t.Errorf("Name() = %q, want %q", cs.Name(), "composite")
	}
}

func TestCompositeStrategy_Description(t *testing.T) {
	cs := NewCompositeStrategy(nil, nil)
	if cs.Description() == "" {
		t.Error("Description() should not be empty")
	}
}

func TestCompositeStrategy_MergesResults(t *testing.T) {
	s1 := &mockStrategy{
		name:    "direct",
		results: []Itinerary{makeItin("DEL", "YYZ", 800)},
	}
	s2 := &mockStrategy{
		name:    "multicity",
		results: []Itinerary{makeItin("DEL", "YYZ", 600)},
	}

	cs := NewCompositeStrategy(nil, s1, s2)
	results, err := cs.Search(context.Background(), Request{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("got %d results, want 2", len(results))
	}
}

func TestCompositeStrategy_DeduplicatesByRouteAndPrice(t *testing.T) {
	itin := makeItin("DEL", "YYZ", 800)
	s1 := &mockStrategy{name: "a", results: []Itinerary{itin}}
	s2 := &mockStrategy{name: "b", results: []Itinerary{itin}}

	cs := NewCompositeStrategy(nil, s1, s2)
	results, err := cs.Search(context.Background(), Request{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("got %d results, want 1 (deduplicated)", len(results))
	}
}

func TestCompositeStrategy_UsesRanker(t *testing.T) {
	s1 := &mockStrategy{
		name:    "a",
		results: []Itinerary{makeItin("DEL", "YYZ", 800)},
	}
	s2 := &mockStrategy{
		name:    "b",
		results: []Itinerary{makeItin("BOM", "YYZ", 600)},
	}
	ranker := &mockRanker{}

	cs := NewCompositeStrategy(ranker, s1, s2)
	results, err := cs.Search(context.Background(), Request{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ranker.called {
		t.Error("ranker was not called")
	}
	if results[0].Score != 100 {
		t.Errorf("first result score = %.0f, want 100", results[0].Score)
	}
}

func TestCompositeStrategy_PartialFailure(t *testing.T) {
	s1 := &mockStrategy{
		name:    "good",
		results: []Itinerary{makeItin("DEL", "YYZ", 800)},
	}
	s2 := &mockStrategy{
		name: "bad",
		err:  errors.New("provider down"),
	}

	cs := NewCompositeStrategy(nil, s1, s2)
	results, err := cs.Search(context.Background(), Request{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("got %d results, want 1 (partial success)", len(results))
	}
}

func TestCompositeStrategy_AllFail(t *testing.T) {
	s1 := &mockStrategy{name: "a", err: errors.New("fail 1")}
	s2 := &mockStrategy{name: "b", err: errors.New("fail 2")}

	cs := NewCompositeStrategy(nil, s1, s2)
	_, err := cs.Search(context.Background(), Request{})
	if err == nil {
		t.Fatal("expected error when all strategies fail")
	}
}

func TestCompositeStrategy_NoStrategies(t *testing.T) {
	cs := NewCompositeStrategy(nil)
	results, err := cs.Search(context.Background(), Request{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 0 {
		t.Fatalf("got %d results, want 0", len(results))
	}
}

func TestCompositeStrategy_RespectsMaxResults(t *testing.T) {
	s1 := &mockStrategy{
		name: "a",
		results: []Itinerary{
			makeItin("DEL", "YYZ", 800),
			makeItin("DEL", "YYZ", 900),
			makeItin("DEL", "YYZ", 1000),
		},
	}
	s2 := &mockStrategy{
		name: "b",
		results: []Itinerary{
			makeItin("BOM", "YYZ", 600),
			makeItin("BOM", "YYZ", 700),
		},
	}

	cs := NewCompositeStrategy(nil, s1, s2)
	results, err := cs.Search(context.Background(), Request{MaxResults: 3})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 3 {
		t.Fatalf("got %d results, want 3 (capped by MaxResults)", len(results))
	}
}

func TestCompositeStrategy_CallsAllStrategies(t *testing.T) {
	s1 := &mockStrategy{name: "a", results: []Itinerary{makeItin("DEL", "YYZ", 800)}}
	s2 := &mockStrategy{name: "b", results: []Itinerary{makeItin("BOM", "YYZ", 600)}}
	s3 := &mockStrategy{name: "c", results: []Itinerary{makeItin("DEL", "YVR", 700)}}

	cs := NewCompositeStrategy(nil, s1, s2, s3)
	_, err := cs.Search(context.Background(), Request{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, s := range []*mockStrategy{s1, s2, s3} {
		if s.searchCount != 1 {
			t.Errorf("strategy %q called %d times, want 1", s.name, s.searchCount)
		}
	}
}
