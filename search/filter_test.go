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

func TestSortResults_ByScore(t *testing.T) {
	now := time.Date(2026, 3, 15, 10, 0, 0, 0, time.UTC)
	itins := []Itinerary{
		makeSortItinerary(400, 10*time.Hour, now),
		makeSortItinerary(600, 8*time.Hour, now),
		makeSortItinerary(800, 12*time.Hour, now),
	}
	itins[0].Score = 3.0
	itins[1].Score = 9.5
	itins[2].Score = 7.0
	SortResults(itins, "score")
	if itins[0].Score != 9.5 || itins[1].Score != 7.0 || itins[2].Score != 3.0 {
		t.Errorf("SortResults(score): got scores [%.1f, %.1f, %.1f], want [9.5, 7.0, 3.0]",
			itins[0].Score, itins[1].Score, itins[2].Score)
	}
}

func TestSortResults_ByScore_AllZero(t *testing.T) {
	now := time.Date(2026, 3, 15, 10, 0, 0, 0, time.UTC)
	itins := []Itinerary{
		makeSortItinerary(800, 10*time.Hour, now),
		makeSortItinerary(400, 14*time.Hour, now),
		makeSortItinerary(600, 8*time.Hour, now),
	}
	// All scores zero — should not panic and order is stable (all equal).
	SortResults(itins, "score")
	// Just verify it didn't panic — with all-zero scores, any order is valid.
}

// --- FilterByAvoidAirlines ---

func TestFilterByAvoidAirlines_MatchesAirline(t *testing.T) {
	ac := types.Flight{Price: types.Money{Amount: 500, Currency: "USD"}, Outbound: []types.Segment{{Airline: "AC"}}}
	ba := types.Flight{Price: types.Money{Amount: 600, Currency: "USD"}, Outbound: []types.Segment{{Airline: "BA"}}}
	result := FilterByAvoidAirlines([]types.Flight{ac, ba}, "BA")
	if len(result) != 1 || result[0].Price.Amount != 500 {
		t.Errorf("FilterByAvoidAirlines(BA): got %d flights, want 1 (AC only)", len(result))
	}
}

func TestFilterByAvoidAirlines_MatchesOperatingCarrier(t *testing.T) {
	f := types.Flight{Price: types.Money{Amount: 500, Currency: "USD"}, Outbound: []types.Segment{{Airline: "AC", OperatingCarrier: "UA"}}}
	result := FilterByAvoidAirlines([]types.Flight{f}, "UA")
	if len(result) != 0 {
		t.Errorf("FilterByAvoidAirlines(UA operating): got %d, want 0", len(result))
	}
}

func TestFilterByAvoidAirlines_NoMatch(t *testing.T) {
	ac := types.Flight{Price: types.Money{Amount: 500, Currency: "USD"}, Outbound: []types.Segment{{Airline: "AC"}}}
	ba := types.Flight{Price: types.Money{Amount: 600, Currency: "USD"}, Outbound: []types.Segment{{Airline: "BA"}}}
	result := FilterByAvoidAirlines([]types.Flight{ac, ba}, "LH")
	if len(result) != 2 {
		t.Errorf("FilterByAvoidAirlines(LH): got %d, want 2 (no match)", len(result))
	}
}

func TestFilterByAvoidAirlines_EmptyString(t *testing.T) {
	flights := []types.Flight{{Price: types.Money{Amount: 500, Currency: "USD"}, Outbound: []types.Segment{{Airline: "AC"}}}}
	result := FilterByAvoidAirlines(flights, "")
	if len(result) != 1 {
		t.Errorf("FilterByAvoidAirlines(empty): got %d, want 1 (no filter)", len(result))
	}
}

func TestFilterByAvoidAirlines_MultipleAirlines(t *testing.T) {
	ac := types.Flight{Price: types.Money{Amount: 500, Currency: "USD"}, Outbound: []types.Segment{{Airline: "AC"}}}
	ba := types.Flight{Price: types.Money{Amount: 600, Currency: "USD"}, Outbound: []types.Segment{{Airline: "BA"}}}
	lh := types.Flight{Price: types.Money{Amount: 700, Currency: "USD"}, Outbound: []types.Segment{{Airline: "LH"}}}
	result := FilterByAvoidAirlines([]types.Flight{ac, ba, lh}, "BA,LH")
	if len(result) != 1 || result[0].Price.Amount != 500 {
		t.Errorf("FilterByAvoidAirlines(BA,LH): got %d, want 1 (AC only)", len(result))
	}
}

// --- FilterByArrivalTime ---

func TestFilterByArrivalTime_BeforeEvening(t *testing.T) {
	earlyArr := types.Flight{
		Price:    types.Money{Amount: 500, Currency: "USD"},
		Outbound: []types.Segment{{ArrivalTime: time.Date(2026, 3, 15, 14, 0, 0, 0, time.UTC)}},
	}
	lateArr := types.Flight{
		Price:    types.Money{Amount: 600, Currency: "USD"},
		Outbound: []types.Segment{{ArrivalTime: time.Date(2026, 3, 15, 22, 0, 0, 0, time.UTC)}},
	}
	flights := []types.Flight{earlyArr, lateArr}

	result := FilterByArrivalTime(flights, "", "18:00")
	if len(result) != 1 {
		t.Fatalf("FilterByArrivalTime(before 18:00): got %d, want 1", len(result))
	}
	if result[0].Price.Amount != 500 {
		t.Errorf("FilterByArrivalTime: kept wrong flight, price=%v", result[0].Price.Amount)
	}
}

func TestFilterByArrivalTime_AfterNoon(t *testing.T) {
	morningArr := types.Flight{
		Price:    types.Money{Amount: 400, Currency: "USD"},
		Outbound: []types.Segment{{ArrivalTime: time.Date(2026, 3, 15, 9, 0, 0, 0, time.UTC)}},
	}
	afternoonArr := types.Flight{
		Price:    types.Money{Amount: 600, Currency: "USD"},
		Outbound: []types.Segment{{ArrivalTime: time.Date(2026, 3, 15, 15, 0, 0, 0, time.UTC)}},
	}
	flights := []types.Flight{morningArr, afternoonArr}

	result := FilterByArrivalTime(flights, "12:00", "")
	if len(result) != 1 {
		t.Fatalf("FilterByArrivalTime(after 12:00): got %d, want 1", len(result))
	}
	if result[0].Price.Amount != 600 {
		t.Errorf("FilterByArrivalTime: kept wrong flight, price=%v", result[0].Price.Amount)
	}
}

func TestFilterByArrivalTime_EmptyBounds(t *testing.T) {
	flights := []types.Flight{
		{Price: types.Money{Amount: 500, Currency: "USD"}, Outbound: []types.Segment{{ArrivalTime: time.Date(2026, 3, 15, 3, 0, 0, 0, time.UTC)}}},
		{Price: types.Money{Amount: 600, Currency: "USD"}, Outbound: []types.Segment{{ArrivalTime: time.Date(2026, 3, 15, 20, 0, 0, 0, time.UTC)}}},
	}

	result := FilterByArrivalTime(flights, "", "")
	if len(result) != 2 {
		t.Fatalf("FilterByArrivalTime(empty bounds): got %d, want 2", len(result))
	}
}

func TestFilterByArrivalTime_InvalidFormat(t *testing.T) {
	flights := []types.Flight{
		{Price: types.Money{Amount: 500, Currency: "USD"}, Outbound: []types.Segment{{ArrivalTime: time.Date(2026, 3, 15, 8, 0, 0, 0, time.UTC)}}},
	}

	result := FilterByArrivalTime(flights, "invalid", "")
	if len(result) != 1 {
		t.Fatalf("FilterByArrivalTime(invalid format): got %d, want 1", len(result))
	}
}

func TestFilterByArrivalTime_MultiSegment(t *testing.T) {
	// Arrival time is based on the LAST segment's arrival.
	flight := types.Flight{
		Price: types.Money{Amount: 500, Currency: "USD"},
		Outbound: []types.Segment{
			{ArrivalTime: time.Date(2026, 3, 15, 10, 0, 0, 0, time.UTC)},
			{ArrivalTime: time.Date(2026, 3, 15, 16, 0, 0, 0, time.UTC)},
		},
	}

	// Last segment arrives at 16:00 — should be kept with before=18:00.
	result := FilterByArrivalTime([]types.Flight{flight}, "", "18:00")
	if len(result) != 1 {
		t.Fatalf("FilterByArrivalTime(multi-seg): got %d, want 1", len(result))
	}

	// Last segment arrives at 16:00 — should be dropped with before=14:00.
	result = FilterByArrivalTime([]types.Flight{flight}, "", "14:00")
	if len(result) != 0 {
		t.Fatalf("FilterByArrivalTime(multi-seg excluded): got %d, want 0", len(result))
	}
}

// --- FilterByMaxDuration ---

func TestFilterByMaxDuration_ExcludesLong(t *testing.T) {
	short := types.Flight{
		Price:         types.Money{Amount: 500, Currency: "USD"},
		TotalDuration: 8 * time.Hour,
	}
	long := types.Flight{
		Price:         types.Money{Amount: 400, Currency: "USD"},
		TotalDuration: 20 * time.Hour,
	}
	flights := []types.Flight{short, long}

	result := FilterByMaxDuration(flights, 12*time.Hour)
	if len(result) != 1 {
		t.Fatalf("FilterByMaxDuration: got %d, want 1", len(result))
	}
	if result[0].Price.Amount != 500 {
		t.Errorf("FilterByMaxDuration: kept wrong flight")
	}
}

func TestFilterByMaxDuration_ZeroMeansNoLimit(t *testing.T) {
	flights := []types.Flight{
		{Price: types.Money{Amount: 500, Currency: "USD"}, TotalDuration: 30 * time.Hour},
	}

	result := FilterByMaxDuration(flights, 0)
	if len(result) != 1 {
		t.Fatalf("FilterByMaxDuration(zero=no limit): got %d, want 1", len(result))
	}
}

func TestFilterByMaxDuration_ExactBoundary(t *testing.T) {
	exact := types.Flight{
		Price:         types.Money{Amount: 500, Currency: "USD"},
		TotalDuration: 12 * time.Hour,
	}
	result := FilterByMaxDuration([]types.Flight{exact}, 12*time.Hour)
	if len(result) != 1 {
		t.Fatalf("FilterByMaxDuration(exact boundary): got %d, want 1 (should include equal)", len(result))
	}
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

// --- FilterByPreferredAirlines ---

func TestFilterByPreferredAirlines_EmptyKeepsAll(t *testing.T) {
	flights := []types.Flight{
		{Outbound: []types.Segment{{Airline: "AC"}}, Price: types.Money{Amount: 500, Currency: "USD"}},
		{Outbound: []types.Segment{{Airline: "BA"}}, Price: types.Money{Amount: 600, Currency: "USD"}},
	}
	result := FilterByPreferredAirlines(flights, "")
	if len(result) != 2 {
		t.Fatalf("empty preferred should keep all: got %d, want 2", len(result))
	}
}

func TestFilterByPreferredAirlines_SingleCode(t *testing.T) {
	flights := []types.Flight{
		{Outbound: []types.Segment{{Airline: "AC"}}, Price: types.Money{Amount: 500, Currency: "USD"}},
		{Outbound: []types.Segment{{Airline: "BA"}}, Price: types.Money{Amount: 600, Currency: "USD"}},
		{Outbound: []types.Segment{{Airline: "LH"}}, Price: types.Money{Amount: 700, Currency: "USD"}},
	}
	result := FilterByPreferredAirlines(flights, "AC")
	if len(result) != 1 {
		t.Fatalf("single code: got %d, want 1", len(result))
	}
	if result[0].Outbound[0].Airline != "AC" {
		t.Errorf("kept wrong airline: %s", result[0].Outbound[0].Airline)
	}
}

func TestFilterByPreferredAirlines_MultipleCodes(t *testing.T) {
	flights := []types.Flight{
		{Outbound: []types.Segment{{Airline: "AC"}}, Price: types.Money{Amount: 500, Currency: "USD"}},
		{Outbound: []types.Segment{{Airline: "BA"}}, Price: types.Money{Amount: 600, Currency: "USD"}},
		{Outbound: []types.Segment{{Airline: "LH"}}, Price: types.Money{Amount: 700, Currency: "USD"}},
	}
	result := FilterByPreferredAirlines(flights, "AC,LH")
	if len(result) != 2 {
		t.Fatalf("multiple codes: got %d, want 2", len(result))
	}
}

func TestFilterByPreferredAirlines_OperatingCarrierMatch(t *testing.T) {
	flights := []types.Flight{
		{Outbound: []types.Segment{{Airline: "AC", OperatingCarrier: "UA"}}, Price: types.Money{Amount: 500, Currency: "USD"}},
		{Outbound: []types.Segment{{Airline: "BA"}}, Price: types.Money{Amount: 600, Currency: "USD"}},
	}
	// Prefer UA -- should match via OperatingCarrier.
	result := FilterByPreferredAirlines(flights, "UA")
	if len(result) != 1 {
		t.Fatalf("op carrier match: got %d, want 1", len(result))
	}
}

// --- Single-flight predicates ---

func TestFlightPassesBlocked(t *testing.T) {
	clean := types.Flight{Outbound: []types.Segment{{Airline: "AC", Origin: "YYZ", Destination: "LHR"}}}
	blocked := types.Flight{Outbound: []types.Segment{{Airline: "EK", Origin: "DXB", Destination: "LHR"}}}
	if !FlightPassesBlocked(clean) {
		t.Error("clean flight should pass")
	}
	if FlightPassesBlocked(blocked) {
		t.Error("blocked flight should not pass")
	}
}

func TestFlightPassesAlliance(t *testing.T) {
	star := types.Flight{Outbound: []types.Segment{{Airline: "AC"}}}
	ow := types.Flight{Outbound: []types.Segment{{Airline: "AA"}}}

	if !FlightPassesAlliance(star, "") {
		t.Error("empty alliance should pass all")
	}
	if !FlightPassesAlliance(star, "Star Alliance") {
		t.Error("AC should match Star Alliance")
	}
	if FlightPassesAlliance(ow, "Star Alliance") {
		t.Error("AA should not match Star Alliance")
	}
}

func TestFlightPassesDepartureTime(t *testing.T) {
	morning := types.Flight{Outbound: []types.Segment{{DepartureTime: time.Date(2026, 3, 15, 8, 30, 0, 0, time.UTC)}}}
	evening := types.Flight{Outbound: []types.Segment{{DepartureTime: time.Date(2026, 3, 15, 20, 0, 0, 0, time.UTC)}}}
	noSegs := types.Flight{}

	if !FlightPassesDepartureTime(morning, "06:00", "12:00") {
		t.Error("8:30 should pass 06:00-12:00")
	}
	if FlightPassesDepartureTime(evening, "06:00", "12:00") {
		t.Error("20:00 should not pass 06:00-12:00")
	}
	if !FlightPassesDepartureTime(morning, "", "") {
		t.Error("empty bounds should pass all")
	}
	if FlightPassesDepartureTime(noSegs, "06:00", "12:00") {
		t.Error("no segments should not pass")
	}
}

func TestFlightPassesArrivalTime(t *testing.T) {
	early := types.Flight{Outbound: []types.Segment{{ArrivalTime: time.Date(2026, 3, 15, 14, 0, 0, 0, time.UTC)}}}
	late := types.Flight{Outbound: []types.Segment{{ArrivalTime: time.Date(2026, 3, 15, 22, 0, 0, 0, time.UTC)}}}

	if !FlightPassesArrivalTime(early, "", "18:00") {
		t.Error("14:00 should pass before 18:00")
	}
	if FlightPassesArrivalTime(late, "", "18:00") {
		t.Error("22:00 should not pass before 18:00")
	}
}

func TestFlightPassesMaxDuration(t *testing.T) {
	short := types.Flight{TotalDuration: 8 * time.Hour}
	long := types.Flight{TotalDuration: 20 * time.Hour}

	if !FlightPassesMaxDuration(short, 12*time.Hour) {
		t.Error("8h should pass 12h max")
	}
	if FlightPassesMaxDuration(long, 12*time.Hour) {
		t.Error("20h should not pass 12h max")
	}
	if !FlightPassesMaxDuration(long, 0) {
		t.Error("zero max should pass all")
	}
}

func TestFlightPassesAvoidAirlines(t *testing.T) {
	ac := types.Flight{Outbound: []types.Segment{{Airline: "AC"}}}
	ba := types.Flight{Outbound: []types.Segment{{Airline: "BA"}}}

	if !FlightPassesAvoidAirlines(ac, "BA") {
		t.Error("AC should pass when avoiding BA")
	}
	if FlightPassesAvoidAirlines(ba, "BA") {
		t.Error("BA should not pass when avoiding BA")
	}
	if !FlightPassesAvoidAirlines(ba, "") {
		t.Error("empty avoid should pass all")
	}
}

func TestFlightPassesPreferredAirlines(t *testing.T) {
	ac := types.Flight{Outbound: []types.Segment{{Airline: "AC"}}}
	ba := types.Flight{Outbound: []types.Segment{{Airline: "BA"}}}

	if !FlightPassesPreferredAirlines(ac, "AC,LH") {
		t.Error("AC should pass when preferred AC,LH")
	}
	if FlightPassesPreferredAirlines(ba, "AC,LH") {
		t.Error("BA should not pass when preferred AC,LH")
	}
	if !FlightPassesPreferredAirlines(ba, "") {
		t.Error("empty preferred should pass all")
	}
}

// --- DiversifyResults ---

func makeDiverseItinerary(price float64, travel time.Duration, airline string, score float64) Itinerary {
	return Itinerary{
		TotalPrice:  types.Money{Amount: price, Currency: "USD"},
		TotalTravel: travel,
		Score:       score,
		Legs: []Leg{{
			Flight: types.Flight{
				Price:    types.Money{Amount: price, Currency: "USD"},
				Outbound: []types.Segment{{Airline: airline}},
			},
		}},
	}
}

func TestDiversifyResults_NearIdentical(t *testing.T) {
	// 6 nearly identical Air Canada flights — diversification should
	// still pick cheapest and fastest, then fill with diversity.
	itins := []Itinerary{
		makeDiverseItinerary(500, 10*time.Hour, "AC", 0),
		makeDiverseItinerary(510, 10*time.Hour, "AC", 0),
		makeDiverseItinerary(490, 11*time.Hour, "AC", 0),
		makeDiverseItinerary(505, 9*time.Hour, "AC", 0),
		makeDiverseItinerary(520, 10*time.Hour, "BA", 0),
		makeDiverseItinerary(530, 10*time.Hour, "LH", 0),
	}
	result := DiversifyResults(itins, 4)
	if len(result) != 4 {
		t.Fatalf("DiversifyResults: got %d, want 4", len(result))
	}
	// Should include: cheapest ($490), fastest (9h), BA, LH for diversity.
	airlines := make(map[string]bool)
	for _, r := range result {
		if len(r.Legs) > 0 && len(r.Legs[0].Flight.Outbound) > 0 {
			airlines[r.Legs[0].Flight.Outbound[0].Airline] = true
		}
	}
	if !airlines["BA"] || !airlines["LH"] {
		t.Errorf("DiversifyResults: expected BA and LH for diversity, got airlines: %v", airlines)
	}
}

func TestDiversifyResults_WithScores(t *testing.T) {
	itins := []Itinerary{
		makeDiverseItinerary(500, 12*time.Hour, "AC", 70),
		makeDiverseItinerary(800, 8*time.Hour, "BA", 95),
		makeDiverseItinerary(400, 14*time.Hour, "LH", 50),
		makeDiverseItinerary(600, 10*time.Hour, "UA", 85),
	}
	result := DiversifyResults(itins, 3)
	if len(result) != 3 {
		t.Fatalf("DiversifyResults: got %d, want 3", len(result))
	}
	// Should include: cheapest ($400 LH), fastest (8h BA), best-scored (95 BA).
	// Since BA is both fastest and best-scored, slot is reused; fill 3rd from diversity.
	var prices []float64
	for _, r := range result {
		prices = append(prices, r.TotalPrice.Amount)
	}
	// Cheapest ($400) must be included.
	found400 := false
	for _, p := range prices {
		if p == 400 {
			found400 = true
		}
	}
	if !found400 {
		t.Errorf("DiversifyResults: expected cheapest ($400) in results, got prices: %v", prices)
	}
}

func TestDiversifyResults_FewerThanN(t *testing.T) {
	itins := []Itinerary{
		makeDiverseItinerary(500, 10*time.Hour, "AC", 0),
		makeDiverseItinerary(600, 12*time.Hour, "BA", 0),
	}
	result := DiversifyResults(itins, 5)
	if len(result) != 2 {
		t.Fatalf("DiversifyResults(fewer than N): got %d, want 2", len(result))
	}
}

func TestDiversifyResults_Empty(t *testing.T) {
	result := DiversifyResults(nil, 5)
	if len(result) != 0 {
		t.Fatalf("DiversifyResults(nil): got %d, want 0", len(result))
	}
}

func TestDiversifyResults_ZeroMaxReturnsAll(t *testing.T) {
	itins := []Itinerary{
		makeDiverseItinerary(500, 10*time.Hour, "AC", 0),
	}
	result := DiversifyResults(itins, 0)
	if len(result) != 1 {
		t.Fatalf("DiversifyResults(max=0): got %d, want 1", len(result))
	}
}

// --- firstDeparture edge cases ---

func TestFirstDeparture_ZeroLegs(t *testing.T) {
	itin := Itinerary{Legs: nil}
	got := firstDeparture(itin)
	if !got.IsZero() {
		t.Errorf("firstDeparture(zero legs) = %v, want zero time", got)
	}
}

func TestFirstDeparture_EmptyOutbound(t *testing.T) {
	itin := Itinerary{Legs: []Leg{{Flight: types.Flight{Outbound: nil}}}}
	got := firstDeparture(itin)
	if !got.IsZero() {
		t.Errorf("firstDeparture(empty outbound) = %v, want zero time", got)
	}
}

// --- flightPassesTimeOfDay edge cases ---

func TestFlightPassesDepartureTime_EmptyOutbound(t *testing.T) {
	noSegs := types.Flight{Outbound: nil}
	if FlightPassesDepartureTime(noSegs, "06:00", "12:00") {
		t.Error("empty outbound should return false when time constraints set")
	}
}

func TestFlightPassesDepartureTime_InvalidFormat(t *testing.T) {
	morning := types.Flight{Outbound: []types.Segment{{DepartureTime: time.Date(2026, 3, 15, 8, 0, 0, 0, time.UTC)}}}
	// Invalid "after" format should gracefully return true (no constraint).
	if !FlightPassesDepartureTime(morning, "not-a-time", "12:00") {
		t.Error("invalid after format should degrade to true (no constraint)")
	}
	// Invalid "before" format should also degrade.
	if !FlightPassesDepartureTime(morning, "06:00", "garbage") {
		t.Error("invalid before format should degrade to true (no constraint)")
	}
}

func TestFlightPassesArrivalTime_EmptyOutbound(t *testing.T) {
	noSegs := types.Flight{Outbound: nil}
	if FlightPassesArrivalTime(noSegs, "10:00", "18:00") {
		t.Error("empty outbound should return false when time constraints set")
	}
}
