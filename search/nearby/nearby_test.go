package nearby

import (
	"context"
	"sort"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"booker/search"
	"booker/types"
)

// mockStrategy records every request it receives and returns canned results.
// Thread-safe for concurrent use by the nearby searcher.
type mockStrategy struct {
	name     string
	mu       sync.Mutex
	calls    []search.Request
	resultFn func(req search.Request) []search.Itinerary
	err      error
}

func (m *mockStrategy) Name() string        { return m.name }
func (m *mockStrategy) Description() string { return "mock" }
func (m *mockStrategy) Search(_ context.Context, req search.Request) ([]search.Itinerary, error) {
	m.mu.Lock()
	m.calls = append(m.calls, req)
	m.mu.Unlock()
	if m.err != nil {
		return nil, m.err
	}
	if m.resultFn != nil {
		return m.resultFn(req), nil
	}
	return nil, nil
}

func (m *mockStrategy) getCalls() []search.Request {
	m.mu.Lock()
	defer m.mu.Unlock()
	cp := make([]search.Request, len(m.calls))
	copy(cp, m.calls)
	return cp
}

func makeItin(origin, dest string, price float64) search.Itinerary {
	return search.Itinerary{
		Legs: []search.Leg{{
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

func TestSearcher_NameAndDescription(t *testing.T) {
	s := NewSearcher(&mockStrategy{name: "direct"})
	if s.Name() != "nearby" {
		t.Errorf("Name() = %q, want %q", s.Name(), "nearby")
	}
	if s.Description() == "" {
		t.Error("Description() should not be empty")
	}
}

func TestSearcher_FanOutToClusterAirports(t *testing.T) {
	// JFK is in the New York cluster with EWR and LGA.
	// YYZ is in the Toronto cluster with YTZ.
	// We expect calls for all combinations:
	//   JFK->YYZ, JFK->YTZ, EWR->YYZ, EWR->YTZ, LGA->YYZ, LGA->YTZ
	delegate := &mockStrategy{
		name: "direct",
		resultFn: func(req search.Request) []search.Itinerary {
			return []search.Itinerary{makeItin(req.Origin, req.Destination, 500)}
		},
	}

	s := NewSearcher(delegate)
	results, err := s.Search(context.Background(), search.Request{
		Origin:      "JFK",
		Destination: "YYZ",
		MaxResults:  20,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify all 6 combinations were searched.
	type pair struct{ origin, dest string }
	called := make(map[pair]bool)
	for _, c := range delegate.getCalls() {
		called[pair{c.Origin, c.Destination}] = true
	}

	expected := []pair{
		{"JFK", "YYZ"}, {"JFK", "YTZ"},
		{"EWR", "YYZ"}, {"EWR", "YTZ"},
		{"LGA", "YYZ"}, {"LGA", "YTZ"},
	}
	for _, p := range expected {
		if !called[p] {
			t.Errorf("expected call with origin=%s dest=%s, not found", p.origin, p.dest)
		}
	}

	// All 6 should produce unique results (different routes).
	if len(results) != 6 {
		t.Errorf("got %d results, want 6", len(results))
	}
}

func TestSearcher_Deduplication(t *testing.T) {
	// Delegate returns the same itinerary for two different origin/dest calls.
	// The deduplicate logic uses route+price, so identical itineraries should merge.
	delegate := &mockStrategy{
		name: "direct",
		resultFn: func(_ search.Request) []search.Itinerary {
			// Always return DEL->YYZ at $800, regardless of request.
			return []search.Itinerary{makeItin("DEL", "YYZ", 800)}
		},
	}

	// DEL has no cluster, but YYZ has YTZ. So two calls: DEL->YYZ, DEL->YTZ.
	// Both return identical DEL->YYZ $800 itinerary => deduplicated to 1.
	s := NewSearcher(delegate)
	results, err := s.Search(context.Background(), search.Request{
		Origin:      "DEL",
		Destination: "YYZ",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("got %d results, want 1 (deduplicated)", len(results))
	}
}

func TestSearcher_MaxResultsCap(t *testing.T) {
	delegate := &mockStrategy{
		name: "direct",
		resultFn: func(req search.Request) []search.Itinerary {
			return []search.Itinerary{
				makeItin(req.Origin, req.Destination, 500),
				makeItin(req.Origin, req.Destination, 600),
				makeItin(req.Origin, req.Destination, 700),
			}
		},
	}

	// JFK cluster has 3 airports, YYZ has 2 => 6 combos * 3 results = 18 unique.
	// Cap at 5.
	s := NewSearcher(delegate)
	results, err := s.Search(context.Background(), search.Request{
		Origin:      "JFK",
		Destination: "YYZ",
		MaxResults:  5,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 5 {
		t.Errorf("got %d results, want 5 (capped)", len(results))
	}
}

func TestSearcher_NoClusterFallback(t *testing.T) {
	// DEL and BOM are not in any cluster. Should just delegate as-is.
	delegate := &mockStrategy{
		name: "direct",
		resultFn: func(req search.Request) []search.Itinerary {
			return []search.Itinerary{makeItin(req.Origin, req.Destination, 900)}
		},
	}

	s := NewSearcher(delegate)
	results, err := s.Search(context.Background(), search.Request{
		Origin:      "DEL",
		Destination: "BOM",
		MaxResults:  10,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	calls := delegate.getCalls()
	if len(calls) != 1 {
		t.Errorf("delegate called %d times, want 1 (no cluster expansion)", len(calls))
	}
	if calls[0].Origin != "DEL" || calls[0].Destination != "BOM" {
		t.Errorf("unexpected request: origin=%s dest=%s", calls[0].Origin, calls[0].Destination)
	}
	if len(results) != 1 {
		t.Errorf("got %d results, want 1", len(results))
	}
}

func TestSearcher_SortsByPrice(t *testing.T) {
	delegate := &mockStrategy{
		name: "direct",
		resultFn: func(req search.Request) []search.Itinerary {
			// Return different prices per origin.
			prices := map[string]float64{
				"JFK": 900, "EWR": 400, "LGA": 700,
			}
			p := prices[req.Origin]
			return []search.Itinerary{makeItin(req.Origin, "DEL", p)}
		},
	}

	s := NewSearcher(delegate)
	results, err := s.Search(context.Background(), search.Request{
		Origin:      "JFK",
		Destination: "DEL",
		MaxResults:  10,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) < 2 {
		t.Fatalf("expected at least 2 results, got %d", len(results))
	}
	if !sort.SliceIsSorted(results, func(i, j int) bool {
		return results[i].TotalPrice.Amount < results[j].TotalPrice.Amount
	}) {
		t.Error("results not sorted by price")
	}
}

func TestSearcher_PreservesRequestFields(t *testing.T) {
	// Verify that expanded requests preserve all fields except Origin/Destination.
	delegate := &mockStrategy{name: "direct"}

	s := NewSearcher(delegate)
	req := search.Request{
		Origin:        "JFK",
		Destination:   "YYZ",
		DepartureDate: "2025-06-15",
		ReturnDate:    "2025-06-25",
		Passengers:    2,
		CabinClass:    types.CabinBusiness,
		FlexDays:      3,
		MaxStops:      1,
		MaxResults:    10,
		Context:       "prefer morning flights",
	}
	_, _ = s.Search(context.Background(), req)

	for _, c := range delegate.getCalls() {
		if c.DepartureDate != req.DepartureDate {
			t.Errorf("DepartureDate = %q, want %q", c.DepartureDate, req.DepartureDate)
		}
		if c.ReturnDate != req.ReturnDate {
			t.Errorf("ReturnDate = %q, want %q", c.ReturnDate, req.ReturnDate)
		}
		if c.Passengers != req.Passengers {
			t.Errorf("Passengers = %d, want %d", c.Passengers, req.Passengers)
		}
		if c.CabinClass != req.CabinClass {
			t.Errorf("CabinClass = %q, want %q", c.CabinClass, req.CabinClass)
		}
		if c.FlexDays != req.FlexDays {
			t.Errorf("FlexDays = %d, want %d", c.FlexDays, req.FlexDays)
		}
		if c.MaxStops != req.MaxStops {
			t.Errorf("MaxStops = %d, want %d", c.MaxStops, req.MaxStops)
		}
		if c.Context != req.Context {
			t.Errorf("Context = %q, want %q", c.Context, req.Context)
		}
	}
}

func TestSearcher_DelegateError(t *testing.T) {
	// When the only delegate call fails, the searcher should return the error.
	delegate := &mockStrategy{
		name: "direct",
		err:  context.DeadlineExceeded,
	}

	s := NewSearcher(delegate)
	_, err := s.Search(context.Background(), search.Request{
		Origin:      "DEL",
		Destination: "BOM",
	})
	if err == nil {
		t.Fatal("expected error from delegate, got nil")
	}
}

func TestSearcher_PartialDelegateErrors(t *testing.T) {
	// When some expanded pairs fail but others succeed, return the successes.
	var callCount atomic.Int32
	delegate := &mockStrategy{
		name: "direct",
		resultFn: func(req search.Request) []search.Itinerary {
			n := callCount.Add(1)
			// Fail on first call (return nil), succeed on rest.
			if n == 1 {
				return nil
			}
			return []search.Itinerary{makeItin(req.Origin, req.Destination, 500)}
		},
	}

	s := NewSearcher(delegate)
	results, err := s.Search(context.Background(), search.Request{
		Origin:      "YYZ", // cluster: YYZ, YTZ
		Destination: "DEL",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) == 0 {
		t.Error("expected some results from non-failing pairs")
	}
}
