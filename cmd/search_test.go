package cmd

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"
	"time"

	"booker/search"
	"booker/types"
)

var basetime = time.Date(2026, 4, 10, 8, 0, 0, 0, time.UTC)

func makeItin(legs ...search.Leg) search.Itinerary {
	var total time.Duration
	for _, l := range legs {
		total += l.Flight.TotalDuration
	}
	price := 0.0
	for _, l := range legs {
		price += l.Flight.Price.Amount
	}
	return search.Itinerary{
		Legs:        legs,
		TotalPrice:  types.Money{Amount: price, Currency: "USD"},
		TotalTravel: total,
		Score:       85,
	}
}

func makeLeg(airline, origin, dest string, dep time.Time, dur time.Duration, price float64, stopover *search.Stopover) search.Leg {
	return search.Leg{
		Flight: types.Flight{
			Outbound: []types.Segment{{
				Airline:       airline,
				Origin:        origin,
				Destination:   dest,
				DepartureTime: dep,
				ArrivalTime:   dep.Add(dur),
			}},
			TotalDuration: dur,
			Price:         types.Money{Amount: price, Currency: "USD"},
		},
		Stopover: stopover,
	}
}

// --- routeString ---

func TestRouteString_SingleLeg(t *testing.T) {
	itin := makeItin(makeLeg("BA", "JFK", "LHR", basetime, 7*time.Hour, 450, nil))
	got := routeString(itin)
	if got != "JFK→LHR" {
		t.Errorf("routeString = %q, want %q", got, "JFK→LHR")
	}
}

func TestRouteString_TwoLegs(t *testing.T) {
	itin := makeItin(
		makeLeg("CX", "DEL", "HKG", basetime, 8*time.Hour, 300, &search.Stopover{City: "Hong Kong", Airport: "HKG", Duration: 72 * time.Hour}),
		makeLeg("AC", "HKG", "YYZ", basetime.Add(72*time.Hour), 16*time.Hour, 500, nil),
	)
	got := routeString(itin)
	if got != "DEL→HKG→YYZ" {
		t.Errorf("routeString = %q, want %q", got, "DEL→HKG→YYZ")
	}
}

func TestRouteString_Empty(t *testing.T) {
	itin := search.Itinerary{}
	got := routeString(itin)
	if got != "" {
		t.Errorf("routeString = %q, want empty", got)
	}
}

// --- legAirlines ---

func TestLegAirlines_Single(t *testing.T) {
	itin := makeItin(makeLeg("BA", "JFK", "LHR", basetime, 7*time.Hour, 450, nil))
	got := legAirlines(itin, 0)
	if got != "BA" {
		t.Errorf("legAirlines = %q, want %q", got, "BA")
	}
}

func TestLegAirlines_WithName(t *testing.T) {
	leg := makeLeg("BA", "JFK", "LHR", basetime, 7*time.Hour, 450, nil)
	leg.Flight.Outbound[0].AirlineName = "British Airways"
	itin := makeItin(leg)
	got := legAirlines(itin, 0)
	if got != "British Airways" {
		t.Errorf("legAirlines = %q, want %q", got, "British Airways")
	}
}

func TestLegAirlines_OutOfBounds(t *testing.T) {
	itin := makeItin(makeLeg("BA", "JFK", "LHR", basetime, 7*time.Hour, 450, nil))
	got := legAirlines(itin, 5)
	if got != "" {
		t.Errorf("legAirlines = %q, want empty for out-of-bounds index", got)
	}
}

// --- legDeparture ---

func TestLegDeparture_Valid(t *testing.T) {
	itin := makeItin(makeLeg("BA", "JFK", "LHR", basetime, 7*time.Hour, 450, nil))
	got := legDeparture(itin, 0)
	want := basetime.Format(outputDateTimeFmt)
	if got != want {
		t.Errorf("legDeparture = %q, want %q", got, want)
	}
}

func TestLegDeparture_OutOfBounds(t *testing.T) {
	itin := makeItin(makeLeg("BA", "JFK", "LHR", basetime, 7*time.Hour, 450, nil))
	got := legDeparture(itin, 3)
	if got != "" {
		t.Errorf("legDeparture = %q, want empty for out-of-bounds index", got)
	}
}

// --- stopoverString ---

func TestStopoverString_WithStopover(t *testing.T) {
	itin := makeItin(
		makeLeg("CX", "DEL", "HKG", basetime, 8*time.Hour, 300,
			&search.Stopover{City: "Hong Kong", Airport: "HKG", Duration: 72 * time.Hour}),
		makeLeg("AC", "HKG", "YYZ", basetime.Add(72*time.Hour), 16*time.Hour, 500, nil),
	)
	got := stopoverString(itin)
	if got != "Hong Kong (3d 0h)" {
		t.Errorf("stopoverString = %q, want %q", got, "Hong Kong (3d 0h)")
	}
}

func TestStopoverString_NoStopover(t *testing.T) {
	itin := makeItin(makeLeg("BA", "JFK", "LHR", basetime, 7*time.Hour, 450, nil))
	got := stopoverString(itin)
	if got != "" {
		t.Errorf("stopoverString = %q, want empty", got)
	}
}

// --- currencySymbol ---

func TestCurrencySymbol(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"CAD", "C$"},
		{"USD", "$"},
		{"EUR", "€"},
		{"GBP", "£"},
		{"INR", "₹"},
		{"JPY", "JPY "},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := currencySymbol(tt.input)
			if got != tt.want {
				t.Errorf("currencySymbol(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// --- formatDuration ---

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		input time.Duration
		want  string
	}{
		{7 * time.Hour, "7h"},
		{7*time.Hour + 30*time.Minute, "7h 30m"},
		{25 * time.Hour, "1d 1h"},
		{48 * time.Hour, "2d 0h"},
		{0, "0h"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := formatDuration(tt.input)
			if got != tt.want {
				t.Errorf("formatDuration(%v) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// --- isMultiLeg ---

func TestIsMultiLeg_SingleLeg(t *testing.T) {
	itins := []search.Itinerary{
		makeItin(makeLeg("BA", "JFK", "LHR", basetime, 7*time.Hour, 450, nil)),
	}
	if isMultiLeg(itins) {
		t.Error("isMultiLeg should be false for single-leg itineraries")
	}
}

func TestIsMultiLeg_TwoLegs(t *testing.T) {
	itins := []search.Itinerary{
		makeItin(
			makeLeg("CX", "DEL", "HKG", basetime, 8*time.Hour, 300, nil),
			makeLeg("AC", "HKG", "YYZ", basetime.Add(72*time.Hour), 16*time.Hour, 500, nil),
		),
	}
	if !isMultiLeg(itins) {
		t.Error("isMultiLeg should be true for multi-leg itineraries")
	}
}

// --- printJSON ---

func TestPrintJSON_SingleLeg(t *testing.T) {
	itins := []search.Itinerary{
		makeItin(makeLeg("BA", "JFK", "LHR", basetime, 7*time.Hour, 450, nil)),
	}

	// Capture stdout.
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := printJSON(itins, "USD")

	_ = w.Close()
	os.Stdout = old

	if err != nil {
		t.Fatalf("printJSON error: %v", err)
	}

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)

	var results []jsonItinerary
	if err := json.Unmarshal(buf.Bytes(), &results); err != nil {
		t.Fatalf("invalid JSON output: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("results = %d, want 1", len(results))
	}
	if results[0].Route != "JFK→LHR" {
		t.Errorf("route = %q, want %q", results[0].Route, "JFK→LHR")
	}
	if results[0].Currency != "USD" {
		t.Errorf("currency = %q, want USD", results[0].Currency)
	}
	if len(results[0].Legs) != 1 {
		t.Errorf("legs = %d, want 1", len(results[0].Legs))
	}
}
