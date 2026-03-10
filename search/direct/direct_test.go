package direct

import (
	"context"
	"testing"
	"time"

	"booker/config"
	"booker/provider"
	"booker/search"
	"booker/types"
)

// fakeProvider returns a fixed set of flights.
type fakeProvider struct {
	name    config.ProviderName
	flights []types.Flight
}

func (f *fakeProvider) Name() config.ProviderName          { return f.name }
func (f *fakeProvider) Search(_ context.Context, _ types.SearchRequest) ([]types.Flight, error) {
	return f.flights, nil
}

func newRegistry(flights []types.Flight) *provider.Registry {
	r := provider.NewRegistry()
	_ = r.Register(&fakeProvider{name: "fake", flights: flights})
	return r
}

func TestDirectSearch_BasicFlow(t *testing.T) {
	flights := []types.Flight{
		{
			Price:         types.Money{Amount: 500, Currency: "USD"},
			TotalDuration: 10 * time.Hour,
			Outbound: []types.Segment{
				{
					Airline:       "AC",
					AirlineName:   "Air Canada",
					FlightNumber:  "AC100",
					Origin:        "DEL",
					Destination:   "LHR",
					DepartureTime: time.Date(2026, 3, 24, 8, 0, 0, 0, time.UTC),
					ArrivalTime:   time.Date(2026, 3, 24, 18, 0, 0, 0, time.UTC),
					Duration:      10 * time.Hour,
				},
			},
		},
		{
			Price:         types.Money{Amount: 450, Currency: "USD"},
			TotalDuration: 11 * time.Hour,
			Outbound: []types.Segment{
				{
					Airline:       "BA",
					AirlineName:   "British Airways",
					FlightNumber:  "BA142",
					Origin:        "DEL",
					Destination:   "LHR",
					DepartureTime: time.Date(2026, 3, 24, 10, 0, 0, 0, time.UTC),
					ArrivalTime:   time.Date(2026, 3, 24, 21, 0, 0, 0, time.UTC),
					Duration:      11 * time.Hour,
				},
			},
		},
	}

	s := NewSearcher(newRegistry(flights), nil)
	results, err := s.Search(context.Background(), search.Request{
		Origin:        "DEL",
		Destination:   "LHR",
		DepartureDate: "2026-03-24",
		Passengers:    1,
		CabinClass:    types.CabinEconomy,
		MaxStops:      -1,
		MaxResults:    5,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}

	// Should be price-sorted: BA ($450) first, AC ($500) second.
	if results[0].TotalPrice.Amount != 450 {
		t.Errorf("first result price = %.0f, want 450", results[0].TotalPrice.Amount)
	}
	if results[1].TotalPrice.Amount != 500 {
		t.Errorf("second result price = %.0f, want 500", results[1].TotalPrice.Amount)
	}

	// Each result should have exactly 1 leg, no stopover.
	for i, r := range results {
		if len(r.Legs) != 1 {
			t.Errorf("result[%d] has %d legs, want 1", i, len(r.Legs))
		}
		if r.Legs[0].Stopover != nil {
			t.Errorf("result[%d] has a stopover, want nil", i)
		}
	}
}

func TestDirectSearch_FiltersBlockedAirlines(t *testing.T) {
	flights := []types.Flight{
		{
			Price:         types.Money{Amount: 300, Currency: "USD"},
			TotalDuration: 8 * time.Hour,
			Outbound: []types.Segment{
				{
					Airline:       "EK", // Emirates — blocked
					AirlineName:   "Emirates",
					FlightNumber:  "EK500",
					Origin:        "DEL",
					Destination:   "LHR",
					DepartureTime: time.Date(2026, 3, 24, 6, 0, 0, 0, time.UTC),
					ArrivalTime:   time.Date(2026, 3, 24, 14, 0, 0, 0, time.UTC),
					Duration:      8 * time.Hour,
				},
			},
		},
		{
			Price:         types.Money{Amount: 500, Currency: "USD"},
			TotalDuration: 9 * time.Hour,
			Outbound: []types.Segment{
				{
					Airline:       "AI",
					AirlineName:   "Air India",
					FlightNumber:  "AI111",
					Origin:        "DEL",
					Destination:   "LHR",
					DepartureTime: time.Date(2026, 3, 24, 9, 0, 0, 0, time.UTC),
					ArrivalTime:   time.Date(2026, 3, 24, 18, 0, 0, 0, time.UTC),
					Duration:      9 * time.Hour,
				},
			},
		},
	}

	s := NewSearcher(newRegistry(flights), nil)
	results, err := s.Search(context.Background(), search.Request{
		Origin:        "DEL",
		Destination:   "LHR",
		DepartureDate: "2026-03-24",
		Passengers:    1,
		CabinClass:    types.CabinEconomy,
		MaxStops:      -1,
		MaxResults:    5,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result (Emirates filtered), got %d", len(results))
	}
	if results[0].Legs[0].Flight.Outbound[0].Airline != "AI" {
		t.Errorf("expected Air India, got %s", results[0].Legs[0].Flight.Outbound[0].Airline)
	}
}
