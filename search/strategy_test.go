package search

import (
	"context"
	"testing"

	"booker/types"
)

// stubStrategy verifies that Strategy can be implemented.
type stubStrategy struct{}

func (s stubStrategy) Name() string        { return "stub" }
func (s stubStrategy) Description() string { return "a test stub" }
func (s stubStrategy) Search(_ context.Context, _ Request) ([]Itinerary, error) {
	return nil, nil
}

// stubRanker verifies that Ranker can be implemented.
type stubRanker struct{}

func (r stubRanker) Rank(_ context.Context, itineraries []Itinerary) ([]Itinerary, error) {
	return itineraries, nil
}

func TestStrategyInterface(t *testing.T) {
	var s Strategy = stubStrategy{}
	if s.Name() != "stub" {
		t.Errorf("Name() = %q, want %q", s.Name(), "stub")
	}
	if s.Description() != "a test stub" {
		t.Errorf("Description() = %q, want %q", s.Description(), "a test stub")
	}
	results, err := s.Search(context.Background(), Request{
		Origin:      "DEL",
		Destination: "YYZ",
		CabinClass:  types.CabinEconomy,
	})
	if err != nil {
		t.Errorf("Search() error: %v", err)
	}
	if results != nil {
		t.Errorf("Search() = %v, want nil", results)
	}
}

func TestRankerInterface(t *testing.T) {
	var r Ranker = stubRanker{}
	input := []Itinerary{{Score: 42}}
	out, err := r.Rank(context.Background(), input)
	if err != nil {
		t.Errorf("Rank() error: %v", err)
	}
	if len(out) != 1 || out[0].Score != 42 {
		t.Errorf("Rank() = %v, want %v", out, input)
	}
}
