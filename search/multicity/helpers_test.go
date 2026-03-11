package multicity

import (
	"context"
	"fmt"
	"testing"
	"time"

	"booker/config"
	"booker/provider"
	"booker/types"
)

// --- deduplicateFlights ---

func TestDeduplicateFlights_ByBookingURL(t *testing.T) {
	a := types.Flight{BookingURL: "https://example.com/1", Price: types.Money{Amount: 100}}
	b := types.Flight{BookingURL: "https://example.com/2", Price: types.Money{Amount: 200}}
	aDup := types.Flight{BookingURL: "https://example.com/1", Price: types.Money{Amount: 999}}

	tests := []struct {
		name string
		sets [][]types.Flight
		want int
	}{
		{"two sets with overlap", [][]types.Flight{{a, b}, {aDup, b}}, 2},
		{"no overlap", [][]types.Flight{{a}, {b}}, 2},
		{"all duplicates", [][]types.Flight{{a}, {a, a}}, 1},
		{"empty sets", [][]types.Flight{{}, {}}, 0},
		{"single set", [][]types.Flight{{a, b}}, 2},
		{"nil set", nil, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := deduplicateFlights(tt.sets...)
			if len(got) != tt.want {
				t.Errorf("deduplicateFlights() returned %d flights, want %d", len(got), tt.want)
			}
		})
	}
}

func TestDeduplicateFlights_FallbackKey(t *testing.T) {
	dep := basetime
	// Flights without BookingURL use FlightNumber+DepartureTime as key.
	f1 := types.Flight{
		Outbound: []types.Segment{{
			FlightNumber:  "CX123",
			DepartureTime: dep,
		}},
		Price: types.Money{Amount: 100},
	}
	f1Dup := types.Flight{
		Outbound: []types.Segment{{
			FlightNumber:  "CX123",
			DepartureTime: dep,
		}},
		Price: types.Money{Amount: 999},
	}
	f2 := types.Flight{
		Outbound: []types.Segment{{
			FlightNumber:  "AI456",
			DepartureTime: dep,
		}},
		Price: types.Money{Amount: 200},
	}

	got := deduplicateFlights([]types.Flight{f1, f2}, []types.Flight{f1Dup})
	if len(got) != 2 {
		t.Errorf("expected 2 unique flights, got %d", len(got))
	}
}

func TestDeduplicateFlights_EmptyOutbound(t *testing.T) {
	// No BookingURL, no outbound segments => empty key. Each gets treated as unique
	// because the first one sets seen[""] = true, subsequent ones are skipped.
	f := types.Flight{Price: types.Money{Amount: 100}}
	got := deduplicateFlights([]types.Flight{f, f, f})
	if len(got) != 1 {
		t.Errorf("expected 1 (empty key dedup), got %d", len(got))
	}
}

func TestDeduplicateFlights_PreservesOrder(t *testing.T) {
	// First occurrence should be kept, not the duplicate.
	f1 := types.Flight{BookingURL: "url1", Price: types.Money{Amount: 100}}
	f2 := types.Flight{BookingURL: "url2", Price: types.Money{Amount: 200}}
	f1Expensive := types.Flight{BookingURL: "url1", Price: types.Money{Amount: 999}}

	got := deduplicateFlights([]types.Flight{f1, f2}, []types.Flight{f1Expensive})
	if len(got) != 2 {
		t.Fatalf("expected 2 flights, got %d", len(got))
	}
	if got[0].Price.Amount != 100 {
		t.Errorf("first flight price = %.0f, want 100 (first occurrence kept)", got[0].Price.Amount)
	}
}

// --- buildMultiCityItinerary ---

func TestBuildMultiCityItinerary(t *testing.T) {
	leg1Dep := basetime
	leg1Arr := basetime.Add(8 * time.Hour)
	leg2Dep := basetime.Add(72 * time.Hour)
	leg2Arr := basetime.Add(88 * time.Hour)

	r := provider.MultiCityResult{
		Leg1: types.Flight{
			Outbound: []types.Segment{
				{DepartureTime: leg1Dep, ArrivalTime: leg1Arr, Airline: "CX"},
			},
			TotalDuration: 8 * time.Hour,
		},
		Leg2: types.Flight{
			Outbound: []types.Segment{
				{DepartureTime: leg2Dep, ArrivalTime: leg2Arr, Airline: "AC"},
			},
			TotalDuration: 16 * time.Hour,
		},
		Price: types.Money{Amount: 750, Currency: "USD"},
	}

	stop := StopoverCity{City: "Hong Kong", Airport: "HKG"}
	itin := buildMultiCityItinerary(r, stop)

	// Check legs count.
	if len(itin.Legs) != 2 {
		t.Fatalf("expected 2 legs, got %d", len(itin.Legs))
	}

	// First leg should have the stopover.
	if itin.Legs[0].Stopover == nil {
		t.Fatal("first leg stopover is nil")
	}
	if itin.Legs[0].Stopover.City != "Hong Kong" {
		t.Errorf("stopover city = %q, want %q", itin.Legs[0].Stopover.City, "Hong Kong")
	}
	if itin.Legs[0].Stopover.Airport != "HKG" {
		t.Errorf("stopover airport = %q, want %q", itin.Legs[0].Stopover.Airport, "HKG")
	}

	// Stopover duration = leg2Dep - leg1Arr = 64h.
	wantStopover := 64 * time.Hour
	if itin.Legs[0].Stopover.Duration != wantStopover {
		t.Errorf("stopover duration = %v, want %v", itin.Legs[0].Stopover.Duration, wantStopover)
	}

	// Second leg should NOT have a stopover.
	if itin.Legs[1].Stopover != nil {
		t.Error("second leg should not have a stopover")
	}

	// Price comes from the combined result.
	if itin.TotalPrice.Amount != 750 {
		t.Errorf("TotalPrice = %.0f, want 750", itin.TotalPrice.Amount)
	}

	// TotalTravel = sum of leg durations.
	if itin.TotalTravel != 24*time.Hour {
		t.Errorf("TotalTravel = %v, want %v", itin.TotalTravel, 24*time.Hour)
	}

	// TotalTrip = lastArr - firstDep = 88h.
	wantTrip := 88 * time.Hour
	if itin.TotalTrip != wantTrip {
		t.Errorf("TotalTrip = %v, want %v", itin.TotalTrip, wantTrip)
	}
}

func TestBuildMultiCityItinerary_EmptyOutbound(t *testing.T) {
	// When outbound segments are empty, durations should be zero.
	r := provider.MultiCityResult{
		Leg1:  types.Flight{TotalDuration: 5 * time.Hour},
		Leg2:  types.Flight{TotalDuration: 10 * time.Hour},
		Price: types.Money{Amount: 500, Currency: "USD"},
	}

	stop := StopoverCity{City: "Tokyo", Airport: "NRT"}
	itin := buildMultiCityItinerary(r, stop)

	if itin.Legs[0].Stopover.Duration != 0 {
		t.Errorf("stopover duration = %v, want 0 for empty outbound", itin.Legs[0].Stopover.Duration)
	}
	if itin.TotalTrip != 0 {
		t.Errorf("TotalTrip = %v, want 0 for empty outbound", itin.TotalTrip)
	}
	if itin.TotalTravel != 15*time.Hour {
		t.Errorf("TotalTravel = %v, want %v", itin.TotalTravel, 15*time.Hour)
	}
}

func TestBuildMultiCityItinerary_MultiSegment(t *testing.T) {
	// Leg1 has two segments (connecting flight): DEL -> BKK -> HKG
	// Leg2 has one segment: HKG -> YYZ
	leg1Seg1Dep := basetime
	leg1Seg1Arr := basetime.Add(4 * time.Hour)
	leg1Seg2Dep := basetime.Add(6 * time.Hour)
	leg1Seg2Arr := basetime.Add(9 * time.Hour)
	leg2Dep := basetime.Add(72 * time.Hour)
	leg2Arr := basetime.Add(88 * time.Hour)

	r := provider.MultiCityResult{
		Leg1: types.Flight{
			Outbound: []types.Segment{
				{DepartureTime: leg1Seg1Dep, ArrivalTime: leg1Seg1Arr},
				{DepartureTime: leg1Seg2Dep, ArrivalTime: leg1Seg2Arr},
			},
			TotalDuration: 9 * time.Hour,
		},
		Leg2: types.Flight{
			Outbound: []types.Segment{
				{DepartureTime: leg2Dep, ArrivalTime: leg2Arr},
			},
			TotalDuration: 16 * time.Hour,
		},
		Price: types.Money{Amount: 600, Currency: "USD"},
	}

	stop := StopoverCity{City: "Hong Kong", Airport: "HKG"}
	itin := buildMultiCityItinerary(r, stop)

	// Stopover = leg2Dep - leg1 last segment arrival = 72h - 9h = 63h.
	wantStopover := leg2Dep.Sub(leg1Seg2Arr)
	if itin.Legs[0].Stopover.Duration != wantStopover {
		t.Errorf("stopover duration = %v, want %v", itin.Legs[0].Stopover.Duration, wantStopover)
	}

	// TotalTrip = leg2Arr - leg1 first segment departure = 88h.
	wantTrip := leg2Arr.Sub(leg1Seg1Dep)
	if itin.TotalTrip != wantTrip {
		t.Errorf("TotalTrip = %v, want %v", itin.TotalTrip, wantTrip)
	}
}

// --- fetchFromAllProviders ---

// mockProvider implements provider.Provider for testing.
type mockProvider struct {
	name    config.ProviderName
	flights []types.Flight
	err     error
}

func (m *mockProvider) Name() config.ProviderName { return m.name }

func (m *mockProvider) Search(_ context.Context, _ types.SearchRequest) ([]types.Flight, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.flights, nil
}

func TestFetchFromAllProviders_Success(t *testing.T) {
	flights := []types.Flight{
		{BookingURL: "url1", Price: types.Money{Amount: 100}},
		{BookingURL: "url2", Price: types.Money{Amount: 200}},
	}

	reg := provider.NewRegistry()
	if err := reg.Register(&mockProvider{name: "mock", flights: flights}); err != nil {
		t.Fatal(err)
	}

	s := &Searcher{registry: reg}
	got := s.fetchFromAllProviders(context.Background(), types.SearchRequest{})
	if len(got) != 2 {
		t.Errorf("fetchFromAllProviders() returned %d flights, want 2", len(got))
	}
}

func TestFetchFromAllProviders_Error(t *testing.T) {
	reg := provider.NewRegistry()
	if err := reg.Register(&mockProvider{name: "broken", err: fmt.Errorf("api down")}); err != nil {
		t.Fatal(err)
	}

	s := &Searcher{registry: reg}
	got := s.fetchFromAllProviders(context.Background(), types.SearchRequest{})
	if len(got) != 0 {
		t.Errorf("fetchFromAllProviders() returned %d flights on error, want 0", len(got))
	}
}

func TestFetchFromAllProviders_MixedProviders(t *testing.T) {
	// One provider succeeds, one fails. We should get results from the good one.
	good := &mockProvider{
		name:    "good",
		flights: []types.Flight{{BookingURL: "url1", Price: types.Money{Amount: 100}}},
	}
	bad := &mockProvider{
		name: "bad",
		err:  fmt.Errorf("timeout"),
	}

	reg := provider.NewRegistry()
	if err := reg.Register(good); err != nil {
		t.Fatal(err)
	}
	if err := reg.Register(bad); err != nil {
		t.Fatal(err)
	}

	s := &Searcher{registry: reg}
	got := s.fetchFromAllProviders(context.Background(), types.SearchRequest{})
	if len(got) != 1 {
		t.Errorf("fetchFromAllProviders() returned %d flights, want 1 (from good provider)", len(got))
	}
}

func TestFetchFromAllProviders_EmptyRegistry(t *testing.T) {
	reg := provider.NewRegistry()
	s := &Searcher{registry: reg}
	got := s.fetchFromAllProviders(context.Background(), types.SearchRequest{})
	if len(got) != 0 {
		t.Errorf("fetchFromAllProviders() returned %d flights, want 0 for empty registry", len(got))
	}
}

// --- fetchWithDualSort ---

// sortCapturingProvider records which SortBy values it was called with.
type sortCapturingProvider struct {
	name   config.ProviderName
	result map[string][]types.Flight
}

func (p *sortCapturingProvider) Name() config.ProviderName { return p.name }

func (p *sortCapturingProvider) Search(_ context.Context, req types.SearchRequest) ([]types.Flight, error) {
	return p.result[req.SortBy], nil
}

func TestFetchWithDualSort(t *testing.T) {
	// QUALITY sort returns flight A, PRICE sort returns flight B.
	// Both should appear in the merged result after dedup.
	fQuality := types.Flight{BookingURL: "quality-url", Price: types.Money{Amount: 300}}
	fPrice := types.Flight{BookingURL: "price-url", Price: types.Money{Amount: 100}}

	p := &sortCapturingProvider{
		name: "test",
		result: map[string][]types.Flight{
			"QUALITY": {fQuality},
			"PRICE":   {fPrice},
		},
	}

	reg := provider.NewRegistry()
	if err := reg.Register(p); err != nil {
		t.Fatal(err)
	}

	s := &Searcher{registry: reg}
	got := s.fetchWithDualSort(context.Background(), types.SearchRequest{})
	if len(got) != 2 {
		t.Errorf("fetchWithDualSort() returned %d flights, want 2", len(got))
	}
}

func TestFetchWithDualSort_Deduplicates(t *testing.T) {
	// Both sort orders return the same flight. Should deduplicate to 1.
	f := types.Flight{BookingURL: "same-url", Price: types.Money{Amount: 200}}

	p := &sortCapturingProvider{
		name: "test",
		result: map[string][]types.Flight{
			"QUALITY": {f},
			"PRICE":   {f},
		},
	}

	reg := provider.NewRegistry()
	if err := reg.Register(p); err != nil {
		t.Fatal(err)
	}

	s := &Searcher{registry: reg}
	got := s.fetchWithDualSort(context.Background(), types.SearchRequest{})
	if len(got) != 1 {
		t.Errorf("fetchWithDualSort() returned %d flights, want 1 (deduped)", len(got))
	}
}

// --- NewSearcher ---

func TestNewSearcher(t *testing.T) {
	reg := provider.NewRegistry()
	s := NewSearcher(reg, nil, WeightsBudget)
	if s == nil {
		t.Fatal("NewSearcher returned nil")
	}
	if s.registry != reg {
		t.Error("registry not set correctly")
	}
	if s.ranker == nil {
		t.Error("ranker should not be nil")
	}
}

// --- StopoversForRoute ---

func TestStopoversForRoute(t *testing.T) {
	tests := []struct {
		name        string
		origin      string
		destination string
		wantMin     int
	}{
		{"DEL to YYZ returns known stopovers", "DEL", "YYZ", 5},
		{"BOM to YYZ returns known stopovers", "BOM", "YYZ", 5},
		{"DEL to YVR returns known stopovers", "DEL", "YVR", 5},
		{"unknown route returns nil", "JFK", "LHR", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := StopoversForRoute(tt.origin, tt.destination)
			if len(got) < tt.wantMin {
				t.Errorf("StopoversForRoute(%s, %s) returned %d cities, want >= %d",
					tt.origin, tt.destination, len(got), tt.wantMin)
			}
		})
	}
}

// --- passesAllFilters ---

func TestPassesAllFilters(t *testing.T) {
	good := makeFlight("AC", "DEL", "HKG", basetime, basetime.Add(8*time.Hour), 300)

	tests := []struct {
		name   string
		flight types.Flight
		params SearchParams
		want   bool
	}{
		{
			name:   "valid flight with no constraints",
			flight: good,
			params: SearchParams{},
			want:   true,
		},
		{
			name:   "blocked airline rejected",
			flight: makeFlight("EK", "DEL", "HKG", basetime, basetime.Add(8*time.Hour), 300),
			params: SearchParams{},
			want:   false,
		},
		{
			name:   "zero price rejected",
			flight: makeFlight("AC", "DEL", "HKG", basetime, basetime.Add(8*time.Hour), 0),
			params: SearchParams{},
			want:   false,
		},
		{
			name:   "wrong alliance rejected",
			flight: makeFlight("AA", "DEL", "HKG", basetime, basetime.Add(8*time.Hour), 300),
			params: SearchParams{PreferredAlliance: "Star Alliance"},
			want:   false,
		},
		{
			name:   "correct alliance passes",
			flight: good, // AC is Star Alliance
			params: SearchParams{PreferredAlliance: "Star Alliance"},
			want:   true,
		},
		{
			name:   "departure too early rejected",
			flight: makeFlight("AC", "DEL", "HKG", time.Date(2026, 3, 24, 5, 0, 0, 0, time.UTC), time.Date(2026, 3, 24, 13, 0, 0, 0, time.UTC), 300),
			params: SearchParams{DepartureAfter: "06:00"},
			want:   false,
		},
		{
			name:   "departure within window passes",
			flight: good, // departs 10:00
			params: SearchParams{DepartureAfter: "06:00", DepartureBefore: "12:00"},
			want:   true,
		},
		{
			name:   "arrival too late rejected",
			flight: makeFlight("AC", "DEL", "HKG", basetime, time.Date(2026, 3, 24, 23, 0, 0, 0, time.UTC), 300),
			params: SearchParams{ArrivalBefore: "20:00"},
			want:   false,
		},
		{
			name:   "exceeds max duration rejected",
			flight: makeFlight("AC", "DEL", "HKG", basetime, basetime.Add(20*time.Hour), 300),
			params: SearchParams{MaxDuration: 10 * time.Hour},
			want:   false,
		},
		{
			name:   "within max duration passes",
			flight: good, // 8h
			params: SearchParams{MaxDuration: 10 * time.Hour},
			want:   true,
		},
		{
			name:   "avoided airline rejected",
			flight: good, // AC
			params: SearchParams{AvoidAirlines: "AC"},
			want:   false,
		},
		{
			name:   "non-avoided airline passes",
			flight: good, // AC
			params: SearchParams{AvoidAirlines: "UA,DL"},
			want:   true,
		},
		{
			name:   "preferred airline matches",
			flight: good, // AC
			params: SearchParams{PreferredAirlines: "AC,UA"},
			want:   true,
		},
		{
			name:   "preferred airline mismatch rejected",
			flight: good, // AC
			params: SearchParams{PreferredAirlines: "UA,DL"},
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := passesAllFilters(tt.flight, tt.params)
			if got != tt.want {
				t.Errorf("passesAllFilters() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStopoversForRoute_DELYYZContainsExpectedCities(t *testing.T) {
	stopovers := StopoversForRoute("DEL", "YYZ")
	expected := map[string]bool{"HKG": false, "BKK": false, "IST": false, "NRT": false}
	for _, s := range stopovers {
		if _, ok := expected[s.Airport]; ok {
			expected[s.Airport] = true
		}
	}
	for airport, found := range expected {
		if !found {
			t.Errorf("expected stopover airport %s not found in DEL-YYZ route", airport)
		}
	}
}
