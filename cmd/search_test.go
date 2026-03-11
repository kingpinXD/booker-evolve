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

// --- priceSummary ---

func TestPriceSummary_MultipleResults(t *testing.T) {
	itins := []search.Itinerary{
		makeItin(makeLeg("BA", "JFK", "LHR", basetime, 7*time.Hour, 450, nil)),
		makeItin(makeLeg("AA", "JFK", "LHR", basetime, 8*time.Hour, 800, nil)),
		makeItin(makeLeg("VS", "JFK", "LHR", basetime, 7*time.Hour, 600, nil)),
	}
	got := priceSummary(itins, "USD")
	want := "3 results | $450 - $800"
	if got != want {
		t.Errorf("priceSummary = %q, want %q", got, want)
	}
}

func TestPriceSummary_SingleResult(t *testing.T) {
	itins := []search.Itinerary{
		makeItin(makeLeg("BA", "JFK", "LHR", basetime, 7*time.Hour, 450, nil)),
	}
	got := priceSummary(itins, "USD")
	want := "1 result | $450"
	if got != want {
		t.Errorf("priceSummary = %q, want %q", got, want)
	}
}

func TestPriceSummary_Empty(t *testing.T) {
	got := priceSummary(nil, "USD")
	if got != "" {
		t.Errorf("priceSummary = %q, want empty", got)
	}
}

// --- printJSON ---

func TestPrintJSON_BookingURL(t *testing.T) {
	leg := makeLeg("BA", "JFK", "LHR", basetime, 7*time.Hour, 450, nil)
	leg.Flight.BookingURL = "https://book.example.com/abc123"
	itins := []search.Itinerary{makeItin(leg)}

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
	if len(results[0].Legs) != 1 {
		t.Fatalf("legs = %d, want 1", len(results[0].Legs))
	}
	if results[0].Legs[0].BookingURL != "https://book.example.com/abc123" {
		t.Errorf("booking_url = %q, want %q", results[0].Legs[0].BookingURL, "https://book.example.com/abc123")
	}
}

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

// --- printTable ---

func capturePrintTable(itins []search.Itinerary, cur string) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	printTable(itins, cur)

	_ = w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	return buf.String()
}

func TestPrintTable_SingleLeg(t *testing.T) {
	itins := []search.Itinerary{
		makeItin(makeLeg("BA", "JFK", "LHR", basetime, 7*time.Hour, 450, nil)),
	}
	out := capturePrintTable(itins, "USD")

	// Single-leg headers should be present (go-pretty uppercases them).
	for _, want := range []string{"SCORE", "PRICE", "ROUTE", "AIRLINES", "DEPARTURE", "DURATION"} {
		if !bytes.Contains([]byte(out), []byte(want)) {
			t.Errorf("table output missing header %q", want)
		}
	}
	// Multi-leg headers should NOT be present.
	for _, absent := range []string{"LEG 1 AIRLINES", "LEG 2 AIRLINES", "STOPOVER"} {
		if bytes.Contains([]byte(out), []byte(absent)) {
			t.Errorf("single-leg table should not contain %q", absent)
		}
	}
	// Data should be present.
	if !bytes.Contains([]byte(out), []byte("JFK")) {
		t.Error("table output missing route JFK")
	}
	// Price summary footer should be present.
	if !bytes.Contains([]byte(out), []byte("1 result")) {
		t.Error("table output missing price summary footer")
	}
}

func TestPrintTable_MultiLeg(t *testing.T) {
	itins := []search.Itinerary{
		makeItin(
			makeLeg("CX", "DEL", "HKG", basetime, 8*time.Hour, 300,
				&search.Stopover{City: "Hong Kong", Airport: "HKG", Duration: 72 * time.Hour}),
			makeLeg("AC", "HKG", "YYZ", basetime.Add(72*time.Hour), 16*time.Hour, 500, nil),
		),
	}
	out := capturePrintTable(itins, "CAD")

	// Multi-leg headers should be present (go-pretty uppercases them).
	for _, want := range []string{"LEG 1 AIRLINES", "LEG 2 AIRLINES", "STOPOVER"} {
		if !bytes.Contains([]byte(out), []byte(want)) {
			t.Errorf("multi-leg table output missing header %q", want)
		}
	}
	// Data should be present.
	if !bytes.Contains([]byte(out), []byte("Hong Kong")) {
		t.Error("table output missing stopover city Hong Kong")
	}
	// Price summary footer should be present.
	if !bytes.Contains([]byte(out), []byte("1 result")) {
		t.Error("table output missing price summary footer")
	}
}

// --- printTable reasoning ---

func TestPrintTable_ReasoningColumn(t *testing.T) {
	itin := makeItin(makeLeg("BA", "JFK", "LHR", basetime, 7*time.Hour, 450, nil))
	itin.Reasoning = "cheap and fast"
	out := capturePrintTable([]search.Itinerary{itin}, "USD")

	if !bytes.Contains([]byte(out), []byte("REASON")) {
		t.Error("table output missing REASON header")
	}
	if !bytes.Contains([]byte(out), []byte("cheap and fast")) {
		t.Error("table output missing reasoning text")
	}
}

func TestPrintTable_NoReasoning(t *testing.T) {
	itin := makeItin(makeLeg("BA", "JFK", "LHR", basetime, 7*time.Hour, 450, nil))
	itin.Score = 0
	itin.Reasoning = ""
	out := capturePrintTable([]search.Itinerary{itin}, "USD")

	if !bytes.Contains([]byte(out), []byte("REASON")) {
		t.Error("table output missing REASON header even when reasoning is empty")
	}
}

// --- buildJSONItineraries reasoning ---

func TestBuildJSONItineraries_Reasoning(t *testing.T) {
	itin := makeItin(makeLeg("BA", "JFK", "LHR", basetime, 7*time.Hour, 450, nil))
	itin.Reasoning = "good schedule"
	results := buildJSONItineraries([]search.Itinerary{itin}, "USD")

	if len(results) != 1 {
		t.Fatalf("results = %d, want 1", len(results))
	}
	if results[0].Reasoning != "good schedule" {
		t.Errorf("reasoning = %q, want %q", results[0].Reasoning, "good schedule")
	}
}

func TestBuildJSONItineraries_NoReasoning(t *testing.T) {
	itin := makeItin(makeLeg("BA", "JFK", "LHR", basetime, 7*time.Hour, 450, nil))
	results := buildJSONItineraries([]search.Itinerary{itin}, "USD")

	if len(results) != 1 {
		t.Fatalf("results = %d, want 1", len(results))
	}
	if results[0].Reasoning != "" {
		t.Errorf("reasoning = %q, want empty", results[0].Reasoning)
	}

	// Verify omitempty: marshal and check "reasoning" is absent.
	data, err := json.Marshal(results[0])
	if err != nil {
		t.Fatalf("json.Marshal error: %v", err)
	}
	if bytes.Contains(data, []byte(`"reasoning"`)) {
		t.Error("empty reasoning should be omitted from JSON with omitempty")
	}
}

// --- formatPriceInsights ---

func TestFormatPriceInsights_WithData(t *testing.T) {
	insights := search.PriceInsights{
		LowestPrice:       450,
		PriceLevel:        "low",
		TypicalPriceRange: [2]float64{800, 1200},
	}
	got := formatPriceInsights(insights)
	if got == "" {
		t.Fatal("formatPriceInsights returned empty for valid insights")
	}
	for _, want := range []string{"low", "800", "1200"} {
		if !bytes.Contains([]byte(got), []byte(want)) {
			t.Errorf("formatPriceInsights missing %q in %q", want, got)
		}
	}
}

func TestFormatPriceInsights_Empty(t *testing.T) {
	got := formatPriceInsights(search.PriceInsights{})
	if got != "" {
		t.Errorf("formatPriceInsights(empty) = %q, want empty", got)
	}
}
