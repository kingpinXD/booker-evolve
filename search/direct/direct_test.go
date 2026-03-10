package direct

import (
	"context"
	"fmt"
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

func (f *fakeProvider) Name() config.ProviderName { return f.name }
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

func TestDirectSearch_InvalidDate(t *testing.T) {
	s := NewSearcher(newRegistry(nil), nil)
	_, err := s.Search(context.Background(), search.Request{
		Origin:        "DEL",
		Destination:   "LHR",
		DepartureDate: "not-a-date",
		MaxStops:      -1,
	})
	if err == nil {
		t.Fatal("expected error for invalid date")
	}
}

func TestDirectSearch_ZeroPriceFiltered(t *testing.T) {
	flights := []types.Flight{
		{
			Price:         types.Money{Amount: 0, Currency: "USD"},
			TotalDuration: 5 * time.Hour,
			Outbound: []types.Segment{{
				Airline: "AI", AirlineName: "Air India", FlightNumber: "AI100",
				Origin: "DEL", Destination: "LHR",
				DepartureTime: time.Date(2026, 3, 24, 8, 0, 0, 0, time.UTC),
				ArrivalTime:   time.Date(2026, 3, 24, 13, 0, 0, 0, time.UTC),
				Duration:      5 * time.Hour,
			}},
		},
		{
			Price:         types.Money{Amount: 400, Currency: "USD"},
			TotalDuration: 6 * time.Hour,
			Outbound: []types.Segment{{
				Airline: "BA", AirlineName: "British Airways", FlightNumber: "BA200",
				Origin: "DEL", Destination: "LHR",
				DepartureTime: time.Date(2026, 3, 24, 10, 0, 0, 0, time.UTC),
				ArrivalTime:   time.Date(2026, 3, 24, 16, 0, 0, 0, time.UTC),
				Duration:      6 * time.Hour,
			}},
		},
	}

	s := NewSearcher(newRegistry(flights), nil)
	results, err := s.Search(context.Background(), search.Request{
		Origin: "DEL", Destination: "LHR", DepartureDate: "2026-03-24",
		Passengers: 1, CabinClass: types.CabinEconomy, MaxStops: -1,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result (zero-price filtered), got %d", len(results))
	}
	if results[0].TotalPrice.Amount != 400 {
		t.Errorf("expected $400, got $%.0f", results[0].TotalPrice.Amount)
	}
}

func TestDirectSearch_MaxStopsFiltering(t *testing.T) {
	dep := time.Date(2026, 3, 24, 8, 0, 0, 0, time.UTC)
	flights := []types.Flight{
		{
			Price:         types.Money{Amount: 300, Currency: "USD"},
			TotalDuration: 5 * time.Hour,
			Outbound: []types.Segment{{
				Airline: "AI", FlightNumber: "AI100", Origin: "DEL", Destination: "LHR",
				DepartureTime: dep, ArrivalTime: dep.Add(5 * time.Hour), Duration: 5 * time.Hour,
			}},
		},
		{
			Price:         types.Money{Amount: 250, Currency: "USD"},
			TotalDuration: 10 * time.Hour,
			Outbound: []types.Segment{
				{
					Airline: "TG", FlightNumber: "TG300", Origin: "DEL", Destination: "BKK",
					DepartureTime: dep, ArrivalTime: dep.Add(4 * time.Hour), Duration: 4 * time.Hour,
				},
				{
					Airline: "TG", FlightNumber: "TG400", Origin: "BKK", Destination: "LHR",
					DepartureTime: dep.Add(6 * time.Hour), ArrivalTime: dep.Add(10 * time.Hour), Duration: 4 * time.Hour,
				},
			},
		},
	}

	s := NewSearcher(newRegistry(flights), nil)
	results, err := s.Search(context.Background(), search.Request{
		Origin: "DEL", Destination: "LHR", DepartureDate: "2026-03-24",
		Passengers: 1, CabinClass: types.CabinEconomy, MaxStops: 0, // direct only
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 direct flight, got %d", len(results))
	}
	if results[0].TotalPrice.Amount != 300 {
		t.Errorf("expected $300, got $%.0f", results[0].TotalPrice.Amount)
	}
}

func TestDirectSearch_EmptyProviderResults(t *testing.T) {
	s := NewSearcher(newRegistry(nil), nil)
	results, err := s.Search(context.Background(), search.Request{
		Origin: "DEL", Destination: "LHR", DepartureDate: "2026-03-24",
		MaxStops: -1,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 0 {
		t.Fatalf("expected 0 results, got %d", len(results))
	}
}

// mockRanker implements search.Ranker for testing.
type mockRanker struct {
	err error
}

func (m *mockRanker) Rank(_ context.Context, itins []search.Itinerary) ([]search.Itinerary, error) {
	if m.err != nil {
		return nil, m.err
	}
	// Reverse the order to prove ranker was called.
	reversed := make([]search.Itinerary, len(itins))
	for i, it := range itins {
		reversed[len(itins)-1-i] = it
	}
	return reversed, nil
}

func TestDirectSearch_RankerSuccess(t *testing.T) {
	dep := time.Date(2026, 3, 24, 8, 0, 0, 0, time.UTC)
	flights := []types.Flight{
		{
			Price:         types.Money{Amount: 300, Currency: "USD"},
			TotalDuration: 5 * time.Hour,
			Outbound: []types.Segment{{
				Airline: "AI", FlightNumber: "AI100", Origin: "DEL", Destination: "LHR",
				DepartureTime: dep, ArrivalTime: dep.Add(5 * time.Hour), Duration: 5 * time.Hour,
			}},
		},
		{
			Price:         types.Money{Amount: 500, Currency: "USD"},
			TotalDuration: 6 * time.Hour,
			Outbound: []types.Segment{{
				Airline: "BA", FlightNumber: "BA200", Origin: "DEL", Destination: "LHR",
				DepartureTime: dep, ArrivalTime: dep.Add(6 * time.Hour), Duration: 6 * time.Hour,
			}},
		},
	}

	s := NewSearcher(newRegistry(flights), &mockRanker{})
	results, err := s.Search(context.Background(), search.Request{
		Origin: "DEL", Destination: "LHR", DepartureDate: "2026-03-24",
		Passengers: 1, CabinClass: types.CabinEconomy, MaxStops: -1,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// mockRanker reverses: price-sorted [300, 500] becomes [500, 300].
	if results[0].TotalPrice.Amount != 500 {
		t.Errorf("expected ranker-reversed order: first=$500, got $%.0f", results[0].TotalPrice.Amount)
	}
}

func TestDirectSearch_RankerFailureFallsBack(t *testing.T) {
	dep := time.Date(2026, 3, 24, 8, 0, 0, 0, time.UTC)
	flights := []types.Flight{
		{
			Price:         types.Money{Amount: 300, Currency: "USD"},
			TotalDuration: 5 * time.Hour,
			Outbound: []types.Segment{{
				Airline: "AI", FlightNumber: "AI100", Origin: "DEL", Destination: "LHR",
				DepartureTime: dep, ArrivalTime: dep.Add(5 * time.Hour), Duration: 5 * time.Hour,
			}},
		},
		{
			Price:         types.Money{Amount: 500, Currency: "USD"},
			TotalDuration: 6 * time.Hour,
			Outbound: []types.Segment{{
				Airline: "BA", FlightNumber: "BA200", Origin: "DEL", Destination: "LHR",
				DepartureTime: dep, ArrivalTime: dep.Add(6 * time.Hour), Duration: 6 * time.Hour,
			}},
		},
	}

	s := NewSearcher(newRegistry(flights), &mockRanker{err: fmt.Errorf("LLM down")})
	results, err := s.Search(context.Background(), search.Request{
		Origin: "DEL", Destination: "LHR", DepartureDate: "2026-03-24",
		Passengers: 1, CabinClass: types.CabinEconomy, MaxStops: -1,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should fall back to price-sorted order.
	if results[0].TotalPrice.Amount != 300 {
		t.Errorf("expected price-sorted fallback: first=$300, got $%.0f", results[0].TotalPrice.Amount)
	}
}

func TestFlightToItinerary_TotalTrip(t *testing.T) {
	dep := time.Date(2026, 3, 24, 8, 0, 0, 0, time.UTC)
	f := types.Flight{
		Price:         types.Money{Amount: 400, Currency: "USD"},
		TotalDuration: 12 * time.Hour,
		Outbound: []types.Segment{
			{
				DepartureTime: dep,
				ArrivalTime:   dep.Add(5 * time.Hour),
			},
			{
				DepartureTime: dep.Add(7 * time.Hour),
				ArrivalTime:   dep.Add(12 * time.Hour),
			},
		},
	}

	itin := flightToItinerary(f)
	// TotalTrip = last arrival - first departure = 12h.
	if itin.TotalTrip != 12*time.Hour {
		t.Errorf("TotalTrip = %v, want 12h", itin.TotalTrip)
	}
	if itin.TotalTravel != 12*time.Hour {
		t.Errorf("TotalTravel = %v, want 12h", itin.TotalTravel)
	}
	if itin.TotalPrice.Amount != 400 {
		t.Errorf("TotalPrice = %.0f, want 400", itin.TotalPrice.Amount)
	}
}

// errorProvider always returns an error from Search.
type errorProvider struct {
	name config.ProviderName
}

func (e *errorProvider) Name() config.ProviderName { return e.name }
func (e *errorProvider) Search(_ context.Context, _ types.SearchRequest) ([]types.Flight, error) {
	return nil, fmt.Errorf("provider down")
}

func TestDirectSearch_ProviderErrorSkipped(t *testing.T) {
	r := provider.NewRegistry()
	_ = r.Register(&errorProvider{name: "broken"})
	dep := time.Date(2026, 3, 24, 8, 0, 0, 0, time.UTC)
	_ = r.Register(&fakeProvider{name: "working", flights: []types.Flight{
		{
			Price:         types.Money{Amount: 300, Currency: "USD"},
			TotalDuration: 5 * time.Hour,
			Outbound: []types.Segment{{
				Airline: "AI", FlightNumber: "AI100", Origin: "DEL", Destination: "LHR",
				DepartureTime: dep, ArrivalTime: dep.Add(5 * time.Hour), Duration: 5 * time.Hour,
			}},
		},
	}})

	s := NewSearcher(r, nil)
	results, err := s.Search(context.Background(), search.Request{
		Origin: "DEL", Destination: "LHR", DepartureDate: "2026-03-24",
		MaxStops: -1,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result from working provider, got %d", len(results))
	}
}

func TestDirectSearch_MaxResults(t *testing.T) {
	dep := time.Date(2026, 3, 24, 8, 0, 0, 0, time.UTC)
	flights := make([]types.Flight, 5)
	for i := range flights {
		flights[i] = types.Flight{
			Price:         types.Money{Amount: float64(100 + i*50), Currency: "USD"},
			TotalDuration: 5 * time.Hour,
			Outbound: []types.Segment{{
				Airline: "AI", FlightNumber: fmt.Sprintf("AI%d", i), Origin: "DEL", Destination: "LHR",
				DepartureTime: dep, ArrivalTime: dep.Add(5 * time.Hour), Duration: 5 * time.Hour,
			}},
		}
	}

	s := NewSearcher(newRegistry(flights), nil)
	results, err := s.Search(context.Background(), search.Request{
		Origin: "DEL", Destination: "LHR", DepartureDate: "2026-03-24",
		MaxStops: -1, MaxResults: 2,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results (MaxResults cap), got %d", len(results))
	}
}

func TestDirectSearch_FlexDaysFiltering(t *testing.T) {
	// Base date: March 24. Flex ±2 days means 5 searches (March 22-26).
	// Each date returns one flight. Date-range filter keeps all 5.
	dp := &dateTrackingProvider{
		name: "tracking",
		flightsByDate: map[string][]types.Flight{
			"2026-03-22": {{
				Price: types.Money{Amount: 300, Currency: "USD"}, TotalDuration: 5 * time.Hour,
				Outbound: []types.Segment{{
					Airline: "AI", FlightNumber: "AI100", Origin: "DEL", Destination: "LHR",
					DepartureTime: time.Date(2026, 3, 22, 8, 0, 0, 0, time.UTC),
					ArrivalTime:   time.Date(2026, 3, 22, 13, 0, 0, 0, time.UTC), Duration: 5 * time.Hour,
				}},
			}},
			"2026-03-23": {{
				Price: types.Money{Amount: 280, Currency: "USD"}, TotalDuration: 5 * time.Hour,
				Outbound: []types.Segment{{
					Airline: "BA", FlightNumber: "BA150", Origin: "DEL", Destination: "LHR",
					DepartureTime: time.Date(2026, 3, 23, 9, 0, 0, 0, time.UTC),
					ArrivalTime:   time.Date(2026, 3, 23, 14, 0, 0, 0, time.UTC), Duration: 5 * time.Hour,
				}},
			}},
			"2026-03-24": {{
				Price: types.Money{Amount: 500, Currency: "USD"}, TotalDuration: 6 * time.Hour,
				Outbound: []types.Segment{{
					Airline: "AI", FlightNumber: "AI200", Origin: "DEL", Destination: "LHR",
					DepartureTime: time.Date(2026, 3, 24, 10, 0, 0, 0, time.UTC),
					ArrivalTime:   time.Date(2026, 3, 24, 16, 0, 0, 0, time.UTC), Duration: 6 * time.Hour,
				}},
			}},
			"2026-03-25": {{
				Price: types.Money{Amount: 350, Currency: "USD"}, TotalDuration: 5 * time.Hour,
				Outbound: []types.Segment{{
					Airline: "AC", FlightNumber: "AC300", Origin: "DEL", Destination: "LHR",
					DepartureTime: time.Date(2026, 3, 25, 8, 0, 0, 0, time.UTC),
					ArrivalTime:   time.Date(2026, 3, 25, 13, 0, 0, 0, time.UTC), Duration: 5 * time.Hour,
				}},
			}},
			"2026-03-26": {{
				Price: types.Money{Amount: 400, Currency: "USD"}, TotalDuration: 7 * time.Hour,
				Outbound: []types.Segment{{
					Airline: "AI", FlightNumber: "AI300", Origin: "DEL", Destination: "LHR",
					DepartureTime: time.Date(2026, 3, 26, 14, 0, 0, 0, time.UTC),
					ArrivalTime:   time.Date(2026, 3, 26, 21, 0, 0, 0, time.UTC), Duration: 7 * time.Hour,
				}},
			}},
		},
	}

	r := provider.NewRegistry()
	_ = r.Register(dp)

	s := NewSearcher(r, nil)
	results, err := s.Search(context.Background(), search.Request{
		Origin: "DEL", Destination: "LHR", DepartureDate: "2026-03-24",
		Passengers: 1, CabinClass: types.CabinEconomy, MaxStops: -1,
		FlexDays: 2,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 5 dates searched, 1 flight per date, all within range.
	if len(results) != 5 {
		t.Fatalf("expected 5 results with FlexDays=2, got %d", len(results))
	}
	// Price sorted: $280, $300, $350, $400, $500.
	if results[0].TotalPrice.Amount != 280 {
		t.Errorf("first result price = %.0f, want 280", results[0].TotalPrice.Amount)
	}
	if results[4].TotalPrice.Amount != 500 {
		t.Errorf("last result price = %.0f, want 500", results[4].TotalPrice.Amount)
	}
}

func TestDirectSearch_FlexDaysZeroMeansExactDate(t *testing.T) {
	dep := time.Date(2026, 3, 24, 8, 0, 0, 0, time.UTC)
	flights := []types.Flight{
		{
			Price:         types.Money{Amount: 300, Currency: "USD"},
			TotalDuration: 5 * time.Hour,
			Outbound: []types.Segment{{
				Airline: "AI", FlightNumber: "AI100", Origin: "DEL", Destination: "LHR",
				DepartureTime: dep, ArrivalTime: dep.Add(5 * time.Hour), Duration: 5 * time.Hour,
			}},
		},
		{
			Price:         types.Money{Amount: 250, Currency: "USD"},
			TotalDuration: 6 * time.Hour,
			Outbound: []types.Segment{{
				Airline: "BA", FlightNumber: "BA200", Origin: "DEL", Destination: "LHR",
				DepartureTime: time.Date(2026, 3, 25, 10, 0, 0, 0, time.UTC), // next day
				ArrivalTime:   time.Date(2026, 3, 25, 16, 0, 0, 0, time.UTC),
				Duration:      6 * time.Hour,
			}},
		},
	}

	s := NewSearcher(newRegistry(flights), nil)

	// FlexDays=0: no date range filter applied, all flights returned.
	results, err := s.Search(context.Background(), search.Request{
		Origin: "DEL", Destination: "LHR", DepartureDate: "2026-03-24",
		Passengers: 1, CabinClass: types.CabinEconomy, MaxStops: -1,
		FlexDays: 0,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// With FlexDays=0, no filtering is applied (provider already handles the exact date).
	if len(results) != 2 {
		t.Fatalf("expected 2 results with FlexDays=0, got %d", len(results))
	}
}

// dateTrackingProvider records searched dates and returns per-date flights.
type dateTrackingProvider struct {
	name          config.ProviderName
	dates         []string // dates searched (YYYY-MM-DD)
	flightsByDate map[string][]types.Flight
}

func (d *dateTrackingProvider) Name() config.ProviderName { return d.name }
func (d *dateTrackingProvider) Search(_ context.Context, req types.SearchRequest) ([]types.Flight, error) {
	dateStr := req.DepartureDate.Format(DateLayout)
	d.dates = append(d.dates, dateStr)
	return d.flightsByDate[dateStr], nil
}

func TestSearch_FlexDaysMultiDate(t *testing.T) {
	base := time.Date(2026, 3, 24, 0, 0, 0, 0, time.UTC)

	makeFlights := func(date time.Time, price float64, airline, fn string) []types.Flight {
		dep := date.Add(8 * time.Hour)
		return []types.Flight{{
			Price:         types.Money{Amount: price, Currency: "USD"},
			TotalDuration: 5 * time.Hour,
			Outbound: []types.Segment{{
				Airline: airline, FlightNumber: fn, Origin: "DEL", Destination: "LHR",
				DepartureTime: dep, ArrivalTime: dep.Add(5 * time.Hour), Duration: 5 * time.Hour,
			}},
		}}
	}

	dp := &dateTrackingProvider{
		name: "tracking",
		flightsByDate: map[string][]types.Flight{
			"2026-03-23": makeFlights(base.AddDate(0, 0, -1), 350, "AI", "AI101"),
			"2026-03-24": makeFlights(base, 500, "BA", "BA200"),
			"2026-03-25": makeFlights(base.AddDate(0, 0, 1), 400, "AC", "AC300"),
		},
	}

	r := provider.NewRegistry()
	_ = r.Register(dp)

	s := NewSearcher(r, nil)
	results, err := s.Search(context.Background(), search.Request{
		Origin: "DEL", Destination: "LHR", DepartureDate: "2026-03-24",
		Passengers: 1, CabinClass: types.CabinEconomy, MaxStops: -1,
		FlexDays: 1,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Provider should have been called 3 times (day-1, day, day+1).
	if len(dp.dates) != 3 {
		t.Fatalf("expected 3 provider calls, got %d: %v", len(dp.dates), dp.dates)
	}

	// Verify all 3 dates were searched.
	wantDates := map[string]bool{"2026-03-23": true, "2026-03-24": true, "2026-03-25": true}
	for _, d := range dp.dates {
		if !wantDates[d] {
			t.Errorf("unexpected date searched: %s", d)
		}
		delete(wantDates, d)
	}
	if len(wantDates) > 0 {
		t.Errorf("dates not searched: %v", wantDates)
	}

	// All 3 flights should be merged (one per date).
	if len(results) != 3 {
		t.Fatalf("expected 3 merged results, got %d", len(results))
	}

	// Price-sorted: $350, $400, $500.
	if results[0].TotalPrice.Amount != 350 {
		t.Errorf("first result price = %.0f, want 350", results[0].TotalPrice.Amount)
	}
	if results[1].TotalPrice.Amount != 400 {
		t.Errorf("second result price = %.0f, want 400", results[1].TotalPrice.Amount)
	}
	if results[2].TotalPrice.Amount != 500 {
		t.Errorf("third result price = %.0f, want 500", results[2].TotalPrice.Amount)
	}
}

func TestFlightToItinerary_EmptyOutbound(t *testing.T) {
	f := types.Flight{
		Price:         types.Money{Amount: 100, Currency: "USD"},
		TotalDuration: 5 * time.Hour,
	}
	itin := flightToItinerary(f)
	if itin.TotalTrip != 0 {
		t.Errorf("TotalTrip = %v, want 0 for empty outbound", itin.TotalTrip)
	}
}
