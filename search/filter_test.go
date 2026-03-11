package search

import (
	"testing"
	"time"

	"booker/types"
)

func TestIsAirlineBlocked(t *testing.T) {
	tests := []struct {
		code string
		want bool
	}{
		{"EK", true},  // Emirates
		{"QR", true},  // Qatar Airways
		{"LY", true},  // El Al
		{"AA", false}, // American Airlines
		{"AC", false}, // Air Canada
		{"", false},
	}
	for _, tt := range tests {
		if got := IsAirlineBlocked(tt.code); got != tt.want {
			t.Errorf("IsAirlineBlocked(%q) = %v, want %v", tt.code, got, tt.want)
		}
	}
}

func TestIsHubBlocked(t *testing.T) {
	tests := []struct {
		code string
		want bool
	}{
		{"DXB", true},  // Dubai
		{"DOH", true},  // Doha
		{"TLV", true},  // Tel Aviv
		{"YYZ", false}, // Toronto
		{"DEL", false}, // Delhi
		{"", false},
	}
	for _, tt := range tests {
		if got := IsHubBlocked(tt.code); got != tt.want {
			t.Errorf("IsHubBlocked(%q) = %v, want %v", tt.code, got, tt.want)
		}
	}
}

func TestFilterFlights(t *testing.T) {
	clean := types.Flight{
		Price:    types.Money{Amount: 500, Currency: "USD"},
		Outbound: []types.Segment{{Airline: "AC", Origin: "YYZ", Destination: "DEL"}},
	}
	blockedAirline := types.Flight{
		Price:    types.Money{Amount: 400, Currency: "USD"},
		Outbound: []types.Segment{{Airline: "EK", Origin: "YYZ", Destination: "DXB"}},
	}
	blockedHub := types.Flight{
		Price:    types.Money{Amount: 450, Currency: "USD"},
		Outbound: []types.Segment{{Airline: "AC", Origin: "YYZ", Destination: "DOH"}},
	}
	blockedOperating := types.Flight{
		Price:    types.Money{Amount: 460, Currency: "USD"},
		Outbound: []types.Segment{{Airline: "AC", OperatingCarrier: "QR", Origin: "YYZ", Destination: "LHR"}},
	}
	blockedReturn := types.Flight{
		Price:  types.Money{Amount: 470, Currency: "USD"},
		Return: []types.Segment{{Airline: "EK", Origin: "DXB", Destination: "YYZ"}},
	}

	result := FilterFlights([]types.Flight{clean, blockedAirline, blockedHub, blockedOperating, blockedReturn})
	if len(result) != 1 {
		t.Fatalf("FilterFlights: got %d flights, want 1", len(result))
	}
	if result[0].Price.Amount != 500 {
		t.Errorf("FilterFlights: kept wrong flight, price=%v", result[0].Price.Amount)
	}
}

func TestFilterFlights_Empty(t *testing.T) {
	result := FilterFlights(nil)
	if len(result) != 0 {
		t.Errorf("FilterFlights(nil) = %d flights, want 0", len(result))
	}
}

func TestFilterByDateRange(t *testing.T) {
	base := time.Date(2026, 3, 15, 10, 0, 0, 0, time.UTC)
	earliest := time.Date(2026, 3, 14, 0, 0, 0, 0, time.UTC)
	latest := time.Date(2026, 3, 16, 23, 59, 59, 0, time.UTC)

	inRange := types.Flight{
		Price:    types.Money{Amount: 500, Currency: "USD"},
		Outbound: []types.Segment{{DepartureTime: base}},
	}
	tooEarly := types.Flight{
		Price:    types.Money{Amount: 400, Currency: "USD"},
		Outbound: []types.Segment{{DepartureTime: base.AddDate(0, 0, -5)}},
	}
	tooLate := types.Flight{
		Price:    types.Money{Amount: 450, Currency: "USD"},
		Outbound: []types.Segment{{DepartureTime: base.AddDate(0, 0, 5)}},
	}
	noSegments := types.Flight{
		Price: types.Money{Amount: 300, Currency: "USD"},
	}
	onBoundary := types.Flight{
		Price:    types.Money{Amount: 350, Currency: "USD"},
		Outbound: []types.Segment{{DepartureTime: earliest}},
	}

	result := FilterByDateRange([]types.Flight{inRange, tooEarly, tooLate, noSegments, onBoundary}, earliest, latest)
	if len(result) != 2 {
		t.Fatalf("FilterByDateRange: got %d flights, want 2", len(result))
	}
}

func TestFilterByMaxStops(t *testing.T) {
	direct := types.Flight{
		Outbound: []types.Segment{{Origin: "A", Destination: "B"}},
	}
	oneStop := types.Flight{
		Outbound: []types.Segment{
			{Origin: "A", Destination: "C"},
			{Origin: "C", Destination: "B"},
		},
	}
	twoStops := types.Flight{
		Outbound: []types.Segment{
			{Origin: "A", Destination: "C"},
			{Origin: "C", Destination: "D"},
			{Origin: "D", Destination: "B"},
		},
	}
	flights := []types.Flight{direct, oneStop, twoStops}

	// Max 0 stops: only direct
	result := FilterByMaxStops(flights, 0)
	if len(result) != 1 {
		t.Errorf("FilterByMaxStops(0): got %d, want 1", len(result))
	}

	// Max 1 stop: direct + one-stop
	result = FilterByMaxStops(flights, 1)
	if len(result) != 2 {
		t.Errorf("FilterByMaxStops(1): got %d, want 2", len(result))
	}

	// Negative means no limit
	result = FilterByMaxStops(flights, -1)
	if len(result) != 3 {
		t.Errorf("FilterByMaxStops(-1): got %d, want 3", len(result))
	}
}

func TestFilterByMaxPrice(t *testing.T) {
	cheap := types.Flight{Price: types.Money{Amount: 400, Currency: "USD"}}
	mid := types.Flight{Price: types.Money{Amount: 800, Currency: "USD"}}
	expensive := types.Flight{Price: types.Money{Amount: 1500, Currency: "USD"}}
	flights := []types.Flight{cheap, mid, expensive}

	// Zero means no limit.
	result := FilterByMaxPrice(flights, 0)
	if len(result) != 3 {
		t.Errorf("FilterByMaxPrice(0): got %d, want 3 (no limit)", len(result))
	}

	// Cap at 800: keeps cheap and mid.
	result = FilterByMaxPrice(flights, 800)
	if len(result) != 2 {
		t.Errorf("FilterByMaxPrice(800): got %d, want 2", len(result))
	}

	// Cap at 399: removes all.
	result = FilterByMaxPrice(flights, 399)
	if len(result) != 0 {
		t.Errorf("FilterByMaxPrice(399): got %d, want 0", len(result))
	}
}

func TestFilterByAlliance(t *testing.T) {
	starFlight := types.Flight{
		Price:    types.Money{Amount: 500, Currency: "USD"},
		Outbound: []types.Segment{{Airline: "AC"}}, // Air Canada = Star Alliance
	}
	oneWorldFlight := types.Flight{
		Price:    types.Money{Amount: 600, Currency: "USD"},
		Outbound: []types.Segment{{Airline: "AA"}}, // American Airlines = OneWorld
	}
	unknownFlight := types.Flight{
		Price:    types.Money{Amount: 400, Currency: "USD"},
		Outbound: []types.Segment{{Airline: "WN"}}, // Southwest = no alliance
	}
	operatingMatch := types.Flight{
		Price:    types.Money{Amount: 700, Currency: "USD"},
		Outbound: []types.Segment{{Airline: "WN", OperatingCarrier: "LH"}}, // Lufthansa operating = Star Alliance
	}
	flights := []types.Flight{starFlight, oneWorldFlight, unknownFlight, operatingMatch}

	// Empty preference keeps all flights.
	result := FilterByAlliance(flights, "")
	if len(result) != 4 {
		t.Errorf("FilterByAlliance(empty): got %d, want 4", len(result))
	}

	// Star Alliance filter: keeps AC and LH-operated.
	result = FilterByAlliance(flights, "Star Alliance")
	if len(result) != 2 {
		t.Errorf("FilterByAlliance(Star Alliance): got %d, want 2", len(result))
	}
	if result[0].Price.Amount != 500 || result[1].Price.Amount != 700 {
		t.Errorf("FilterByAlliance(Star Alliance): wrong flights kept")
	}

	// OneWorld filter: keeps only AA.
	result = FilterByAlliance(flights, "OneWorld")
	if len(result) != 1 {
		t.Errorf("FilterByAlliance(OneWorld): got %d, want 1", len(result))
	}

	// SkyTeam filter: keeps none of these.
	result = FilterByAlliance(flights, "SkyTeam")
	if len(result) != 0 {
		t.Errorf("FilterByAlliance(SkyTeam): got %d, want 0", len(result))
	}
}

func TestFilterByDepartureTime_MorningOnly(t *testing.T) {
	morning := types.Flight{
		Price:    types.Money{Amount: 500, Currency: "USD"},
		Outbound: []types.Segment{{DepartureTime: time.Date(2026, 3, 15, 8, 30, 0, 0, time.UTC)}},
	}
	afternoon := types.Flight{
		Price:    types.Money{Amount: 600, Currency: "USD"},
		Outbound: []types.Segment{{DepartureTime: time.Date(2026, 3, 15, 14, 0, 0, 0, time.UTC)}},
	}
	evening := types.Flight{
		Price:    types.Money{Amount: 400, Currency: "USD"},
		Outbound: []types.Segment{{DepartureTime: time.Date(2026, 3, 15, 20, 0, 0, 0, time.UTC)}},
	}
	flights := []types.Flight{morning, afternoon, evening}

	result := FilterByDepartureTime(flights, "06:00", "12:00")
	if len(result) != 1 {
		t.Fatalf("FilterByDepartureTime(morning only): got %d, want 1", len(result))
	}
	if result[0].Price.Amount != 500 {
		t.Errorf("FilterByDepartureTime: kept wrong flight, price=%v", result[0].Price.Amount)
	}
}

func TestFilterByDepartureTime_NoRedEyes(t *testing.T) {
	redEye := types.Flight{
		Price:    types.Money{Amount: 300, Currency: "USD"},
		Outbound: []types.Segment{{DepartureTime: time.Date(2026, 3, 15, 2, 0, 0, 0, time.UTC)}},
	}
	morning := types.Flight{
		Price:    types.Money{Amount: 500, Currency: "USD"},
		Outbound: []types.Segment{{DepartureTime: time.Date(2026, 3, 15, 7, 0, 0, 0, time.UTC)}},
	}
	evening := types.Flight{
		Price:    types.Money{Amount: 600, Currency: "USD"},
		Outbound: []types.Segment{{DepartureTime: time.Date(2026, 3, 15, 21, 0, 0, 0, time.UTC)}},
	}
	flights := []types.Flight{redEye, morning, evening}

	// after="05:00", before="" — keep flights departing at/after 5am, no upper bound.
	result := FilterByDepartureTime(flights, "05:00", "")
	if len(result) != 2 {
		t.Fatalf("FilterByDepartureTime(no red-eyes): got %d, want 2", len(result))
	}
}

func TestFilterByDepartureTime_EmptyBounds(t *testing.T) {
	flights := []types.Flight{
		{Price: types.Money{Amount: 500, Currency: "USD"}, Outbound: []types.Segment{{DepartureTime: time.Date(2026, 3, 15, 3, 0, 0, 0, time.UTC)}}},
		{Price: types.Money{Amount: 600, Currency: "USD"}, Outbound: []types.Segment{{DepartureTime: time.Date(2026, 3, 15, 12, 0, 0, 0, time.UTC)}}},
		{Price: types.Money{Amount: 700, Currency: "USD"}, Outbound: []types.Segment{{DepartureTime: time.Date(2026, 3, 15, 23, 0, 0, 0, time.UTC)}}},
	}

	result := FilterByDepartureTime(flights, "", "")
	if len(result) != 3 {
		t.Fatalf("FilterByDepartureTime(empty bounds): got %d, want 3", len(result))
	}
}

func TestFilterByDepartureTime_InvalidFormat(t *testing.T) {
	flights := []types.Flight{
		{Price: types.Money{Amount: 500, Currency: "USD"}, Outbound: []types.Segment{{DepartureTime: time.Date(2026, 3, 15, 8, 0, 0, 0, time.UTC)}}},
		{Price: types.Money{Amount: 600, Currency: "USD"}, Outbound: []types.Segment{{DepartureTime: time.Date(2026, 3, 15, 14, 0, 0, 0, time.UTC)}}},
	}

	// Invalid format should gracefully return all flights.
	result := FilterByDepartureTime(flights, "invalid", "12:00")
	if len(result) != 2 {
		t.Fatalf("FilterByDepartureTime(invalid after): got %d, want 2", len(result))
	}

	result = FilterByDepartureTime(flights, "06:00", "not-a-time")
	if len(result) != 2 {
		t.Fatalf("FilterByDepartureTime(invalid before): got %d, want 2", len(result))
	}
}

// --- SortResults ---

func makeSortItinerary(price float64, travel time.Duration, depTime time.Time) Itinerary {
	return Itinerary{
		TotalPrice:  types.Money{Amount: price, Currency: "USD"},
		TotalTravel: travel,
		Legs: []Leg{{
			Flight: types.Flight{
				Price:    types.Money{Amount: price, Currency: "USD"},
				Outbound: []types.Segment{{DepartureTime: depTime}},
			},
		}},
	}
}

func TestSortResults_ByPrice(t *testing.T) {
	now := time.Date(2026, 3, 15, 10, 0, 0, 0, time.UTC)
	itins := []Itinerary{
		makeSortItinerary(800, 10*time.Hour, now),
		makeSortItinerary(400, 14*time.Hour, now.Add(2*time.Hour)),
		makeSortItinerary(600, 8*time.Hour, now.Add(-time.Hour)),
	}
	SortResults(itins, "price")
	if itins[0].TotalPrice.Amount != 400 || itins[1].TotalPrice.Amount != 600 || itins[2].TotalPrice.Amount != 800 {
		t.Errorf("SortResults(price): got prices [%.0f, %.0f, %.0f], want [400, 600, 800]",
			itins[0].TotalPrice.Amount, itins[1].TotalPrice.Amount, itins[2].TotalPrice.Amount)
	}
}

func TestSortResults_ByDuration(t *testing.T) {
	now := time.Date(2026, 3, 15, 10, 0, 0, 0, time.UTC)
	itins := []Itinerary{
		makeSortItinerary(400, 14*time.Hour, now),
		makeSortItinerary(800, 8*time.Hour, now),
		makeSortItinerary(600, 10*time.Hour, now),
	}
	SortResults(itins, "duration")
	if itins[0].TotalTravel != 8*time.Hour || itins[1].TotalTravel != 10*time.Hour || itins[2].TotalTravel != 14*time.Hour {
		t.Errorf("SortResults(duration): got durations [%v, %v, %v], want [8h, 10h, 14h]",
			itins[0].TotalTravel, itins[1].TotalTravel, itins[2].TotalTravel)
	}
}

func TestSortResults_ByDeparture(t *testing.T) {
	now := time.Date(2026, 3, 15, 10, 0, 0, 0, time.UTC)
	itins := []Itinerary{
		makeSortItinerary(400, 10*time.Hour, now.Add(5*time.Hour)),
		makeSortItinerary(800, 10*time.Hour, now),
		makeSortItinerary(600, 10*time.Hour, now.Add(2*time.Hour)),
	}
	SortResults(itins, "departure")
	wantTimes := []time.Time{now, now.Add(2 * time.Hour), now.Add(5 * time.Hour)}
	for i, want := range wantTimes {
		got := itins[i].Legs[0].Flight.Outbound[0].DepartureTime
		if !got.Equal(want) {
			t.Errorf("SortResults(departure)[%d]: got %v, want %v", i, got, want)
		}
	}
}

func TestSortResults_UnknownDefaultsToPrice(t *testing.T) {
	now := time.Date(2026, 3, 15, 10, 0, 0, 0, time.UTC)
	itins := []Itinerary{
		makeSortItinerary(800, 10*time.Hour, now),
		makeSortItinerary(400, 14*time.Hour, now),
	}
	SortResults(itins, "unknown")
	if itins[0].TotalPrice.Amount != 400 {
		t.Errorf("SortResults(unknown): expected price sort, got first price=%.0f", itins[0].TotalPrice.Amount)
	}
}

func TestSortResults_Empty(t *testing.T) {
	// Should not panic on empty/nil.
	SortResults(nil, "price")
	SortResults([]Itinerary{}, "duration")
}

func TestFilterZeroPrices(t *testing.T) {
	zero := types.Flight{Price: types.Money{Amount: 0, Currency: "USD"}}
	valid := types.Flight{Price: types.Money{Amount: 100, Currency: "USD"}}
	negative := types.Flight{Price: types.Money{Amount: -5, Currency: "USD"}}

	result := FilterZeroPrices([]types.Flight{zero, valid, negative})
	if len(result) != 1 {
		t.Fatalf("FilterZeroPrices: got %d, want 1", len(result))
	}
	if result[0].Price.Amount != 100 {
		t.Errorf("FilterZeroPrices: kept wrong flight")
	}
}
