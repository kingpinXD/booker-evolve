package cmd

import (
	"bytes"
	"encoding/json"
	"strings"
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

func TestLegAirlines_Codeshare(t *testing.T) {
	leg := makeLeg("AC", "YYZ", "LHR", basetime, 7*time.Hour, 600, nil)
	leg.Flight.Outbound[0].AirlineName = "Air Canada"
	leg.Flight.Outbound[0].OperatingCarrier = "UA"
	itin := makeItin(leg)
	got := legAirlines(itin, 0)
	if got != "Air Canada (op. UA)" {
		t.Errorf("legAirlines codeshare = %q, want %q", got, "Air Canada (op. UA)")
	}
}

func TestLegAirlines_NoCodeshare_SameCarrier(t *testing.T) {
	leg := makeLeg("AC", "YYZ", "LHR", basetime, 7*time.Hour, 600, nil)
	leg.Flight.Outbound[0].AirlineName = "Air Canada"
	leg.Flight.Outbound[0].OperatingCarrier = "AC"
	itin := makeItin(leg)
	got := legAirlines(itin, 0)
	if got != "Air Canada" {
		t.Errorf("legAirlines same carrier = %q, want %q", got, "Air Canada")
	}
}

func TestLegAirlines_NoCodeshare_EmptyCarrier(t *testing.T) {
	leg := makeLeg("AC", "YYZ", "LHR", basetime, 7*time.Hour, 600, nil)
	leg.Flight.Outbound[0].AirlineName = "Air Canada"
	// OperatingCarrier empty = no codeshare annotation.
	itin := makeItin(leg)
	got := legAirlines(itin, 0)
	if got != "Air Canada" {
		t.Errorf("legAirlines empty carrier = %q, want %q", got, "Air Canada")
	}
}

// --- JSON operating_carrier ---

func TestBuildJSONItineraries_OperatingCarrier(t *testing.T) {
	leg := makeLeg("AC", "YYZ", "LHR", basetime, 7*time.Hour, 600, nil)
	leg.Flight.Outbound[0].OperatingCarrier = "UA"
	results := buildJSONItineraries([]search.Itinerary{makeItin(leg)}, "USD")

	if len(results) != 1 || len(results[0].Legs) != 1 {
		t.Fatal("unexpected results structure")
	}
	if results[0].Legs[0].OperatingCarrier != "UA" {
		t.Errorf("operating_carrier = %q, want %q", results[0].Legs[0].OperatingCarrier, "UA")
	}
}

func TestBuildJSONItineraries_OperatingCarrier_OmitEmpty(t *testing.T) {
	itin := makeItin(makeLeg("BA", "JFK", "LHR", basetime, 7*time.Hour, 450, nil))
	results := buildJSONItineraries([]search.Itinerary{itin}, "USD")

	data, err := json.Marshal(results[0].Legs[0])
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), "operating_carrier") {
		t.Error("operating_carrier should be omitted when empty")
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

// --- legArrival ---

func TestLegArrival_Valid(t *testing.T) {
	itin := makeItin(makeLeg("BA", "JFK", "LHR", basetime, 7*time.Hour, 450, nil))
	got := legArrival(itin, 0)
	want := basetime.Add(7 * time.Hour).Format(outputDateTimeFmt)
	if got != want {
		t.Errorf("legArrival = %q, want %q", got, want)
	}
}

func TestLegArrival_OutOfBounds(t *testing.T) {
	itin := makeItin(makeLeg("BA", "JFK", "LHR", basetime, 7*time.Hour, 450, nil))
	got := legArrival(itin, 3)
	if got != "" {
		t.Errorf("legArrival = %q, want empty for out-of-bounds index", got)
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

func TestPriceSummary_MultiLeg_WithTripDuration(t *testing.T) {
	leg1 := makeLeg("CX", "DEL", "HKG", basetime, 8*time.Hour, 300,
		&search.Stopover{City: "Hong Kong", Airport: "HKG", Duration: 72 * time.Hour})
	leg2 := makeLeg("AC", "HKG", "YYZ", basetime.Add(72*time.Hour), 16*time.Hour, 500, nil)
	itin1 := makeItin(leg1, leg2)
	itin1.TotalTrip = 96 * time.Hour // 4 days

	leg1b := makeLeg("TG", "DEL", "BKK", basetime, 5*time.Hour, 200,
		&search.Stopover{City: "Bangkok", Airport: "BKK", Duration: 48 * time.Hour})
	leg2b := makeLeg("AC", "BKK", "YYZ", basetime.Add(48*time.Hour), 18*time.Hour, 600, nil)
	itin2 := makeItin(leg1b, leg2b)
	itin2.TotalTrip = 71 * time.Hour // 2d 23h

	got := priceSummary([]search.Itinerary{itin1, itin2}, "USD")
	// Should include trip duration range.
	if !strings.Contains(got, "2d 23h") {
		t.Errorf("priceSummary missing min trip duration, got: %q", got)
	}
	if !strings.Contains(got, "4d 0h") {
		t.Errorf("priceSummary missing max trip duration, got: %q", got)
	}
}

func TestPriceSummary_MultiLeg_SingleResult(t *testing.T) {
	leg1 := makeLeg("CX", "DEL", "HKG", basetime, 8*time.Hour, 300,
		&search.Stopover{City: "Hong Kong", Airport: "HKG", Duration: 72 * time.Hour})
	leg2 := makeLeg("AC", "HKG", "YYZ", basetime.Add(72*time.Hour), 16*time.Hour, 500, nil)
	itin := makeItin(leg1, leg2)
	itin.TotalTrip = 96 * time.Hour

	got := priceSummary([]search.Itinerary{itin}, "USD")
	// Single result should show duration without range.
	if !strings.Contains(got, "4d 0h") {
		t.Errorf("priceSummary missing trip duration, got: %q", got)
	}
}

func TestPriceSummary_SingleLeg_Unchanged(t *testing.T) {
	itins := []search.Itinerary{
		makeItin(makeLeg("BA", "JFK", "LHR", basetime, 7*time.Hour, 450, nil)),
		makeItin(makeLeg("AA", "JFK", "LHR", basetime, 8*time.Hour, 800, nil)),
	}
	got := priceSummary(itins, "USD")
	// Single-leg should not include trip duration.
	want := "2 results | $450 - $800"
	if got != want {
		t.Errorf("priceSummary = %q, want %q", got, want)
	}
}

// --- printJSON ---

func TestPrintJSON_BookingURL(t *testing.T) {
	leg := makeLeg("BA", "JFK", "LHR", basetime, 7*time.Hour, 450, nil)
	leg.Flight.BookingURL = "https://book.example.com/abc123"
	itins := []search.Itinerary{makeItin(leg)}

	var buf bytes.Buffer
	err := printJSON(&buf, itins, "USD")
	if err != nil {
		t.Fatalf("printJSON error: %v", err)
	}

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

	var buf bytes.Buffer
	err := printJSON(&buf, itins, "USD")
	if err != nil {
		t.Fatalf("printJSON error: %v", err)
	}

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
	var buf bytes.Buffer
	printTable(&buf, itins, cur)
	return buf.String()
}

func TestPrintTable_SingleLeg(t *testing.T) {
	itins := []search.Itinerary{
		makeItin(makeLeg("BA", "JFK", "LHR", basetime, 7*time.Hour, 450, nil)),
	}
	out := capturePrintTable(itins, "USD")

	// Single-leg headers should be present (go-pretty uppercases them).
	for _, want := range []string{"SCORE", "PRICE", "ROUTE", "AIRLINES", "DEPARTURE", "ARRIVAL", "DURATION"} {
		if !bytes.Contains([]byte(out), []byte(want)) {
			t.Errorf("table output missing header %q", want)
		}
	}
	// Multi-leg headers should NOT be present.
	for _, absent := range []string{"L1 AIRLINES", "L2 AIRLINES", "STOPOVER"} {
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
	for _, want := range []string{"L1 AIRLINES", "L2 AIRLINES", "L1 ARRIVE", "L2 ARRIVE", "STOPOVER"} {
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

	// When all scores are 0, Score and Reason columns should be hidden.
	if strings.Contains(out, "REASON") {
		t.Error("table output should hide REASON header when all scores are 0")
	}
	if strings.Contains(out, "SCORE") {
		t.Error("table output should hide SCORE header when all scores are 0")
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

// --- printTable booking URL removed (issue #9) ---

func TestPrintTable_NoBookColumn(t *testing.T) {
	leg := makeLeg("BA", "JFK", "LHR", basetime, 7*time.Hour, 450, nil)
	leg.Flight.BookingURL = "https://book.example.com/abc123"
	itins := []search.Itinerary{makeItin(leg)}
	out := capturePrintTable(itins, "USD")

	if bytes.Contains([]byte(out), []byte("BOOK")) {
		t.Error("table output should not have BOOK column")
	}
	if bytes.Contains([]byte(out), []byte("https://book.example.com")) {
		t.Error("table output should not contain booking URL")
	}
}

func TestPrintTable_MultiLeg_NoBookColumn(t *testing.T) {
	leg1 := makeLeg("CX", "DEL", "HKG", basetime, 8*time.Hour, 300,
		&search.Stopover{City: "Hong Kong", Airport: "HKG", Duration: 72 * time.Hour})
	leg1.Flight.BookingURL = "https://book.example.com/leg1"
	leg2 := makeLeg("AC", "HKG", "YYZ", basetime.Add(72*time.Hour), 16*time.Hour, 500, nil)
	leg2.Flight.BookingURL = "https://book.example.com/leg2"
	itins := []search.Itinerary{makeItin(leg1, leg2)}
	out := capturePrintTable(itins, "CAD")

	if bytes.Contains([]byte(out), []byte("BOOK")) {
		t.Error("multi-leg table output should not have BOOK column")
	}
}

// --- printTable stops ---

func TestPrintTable_StopsColumn(t *testing.T) {
	// Direct flight (0 stops): 1 segment.
	directLeg := makeLeg("AC", "DEL", "YYZ", basetime, 14*time.Hour, 800, nil)

	// Connecting flight (1 stop): 2 segments.
	connectingLeg := search.Leg{
		Flight: types.Flight{
			Outbound: []types.Segment{
				{Airline: "LH", Origin: "DEL", Destination: "FRA", DepartureTime: basetime, ArrivalTime: basetime.Add(8 * time.Hour)},
				{Airline: "LH", Origin: "FRA", Destination: "YYZ", DepartureTime: basetime.Add(10 * time.Hour), ArrivalTime: basetime.Add(20 * time.Hour)},
			},
			TotalDuration: 20 * time.Hour,
			Price:         types.Money{Amount: 600, Currency: "USD"},
		},
	}

	itins := []search.Itinerary{
		makeItin(directLeg),
		makeItin(connectingLeg),
	}
	out := capturePrintTable(itins, "USD")

	if !bytes.Contains([]byte(out), []byte("STOPS")) {
		t.Error("table output missing STOPS header")
	}
	// The table should contain "0" for the direct flight and "1" for the connecting flight.
	// We check for both values in the output.
	if !strings.Contains(out, " 0 ") && !strings.Contains(out, "│ 0") {
		t.Errorf("table output missing stops=0 for direct flight, got:\n%s", out)
	}
	if !strings.Contains(out, " 1 ") && !strings.Contains(out, "│ 1") {
		t.Errorf("table output missing stops=1 for connecting flight, got:\n%s", out)
	}
}

// --- formatStops ---

func TestFormatStops_Direct(t *testing.T) {
	// 0 stops, no layover
	itin := makeItin(makeLeg("AC", "DEL", "YYZ", basetime, 14*time.Hour, 800, nil))
	got := formatStops(itin)
	if got != "0" {
		t.Errorf("formatStops = %q, want %q", got, "0")
	}
}

func TestFormatStops_WithLayover(t *testing.T) {
	// 1 stop with 2h30m layover
	leg := search.Leg{
		Flight: types.Flight{
			Outbound: []types.Segment{
				{Airline: "LH", Origin: "DEL", Destination: "FRA", DepartureTime: basetime, ArrivalTime: basetime.Add(8 * time.Hour), LayoverDuration: 2*time.Hour + 30*time.Minute},
				{Airline: "LH", Origin: "FRA", Destination: "YYZ", DepartureTime: basetime.Add(10*time.Hour + 30*time.Minute), ArrivalTime: basetime.Add(20 * time.Hour)},
			},
			TotalDuration: 20 * time.Hour,
			Price:         types.Money{Amount: 600, Currency: "USD"},
		},
	}
	itin := makeItin(leg)
	got := formatStops(itin)
	if got != "1 (2h 30m)" {
		t.Errorf("formatStops = %q, want %q", got, "1 (2h 30m)")
	}
}

func TestFormatStops_NoLayoverData(t *testing.T) {
	// 1 stop but LayoverDuration=0 (no data)
	leg := search.Leg{
		Flight: types.Flight{
			Outbound: []types.Segment{
				{Airline: "LH", Origin: "DEL", Destination: "FRA", DepartureTime: basetime, ArrivalTime: basetime.Add(8 * time.Hour)},
				{Airline: "LH", Origin: "FRA", Destination: "YYZ", DepartureTime: basetime.Add(10 * time.Hour), ArrivalTime: basetime.Add(20 * time.Hour)},
			},
			TotalDuration: 20 * time.Hour,
			Price:         types.Money{Amount: 600, Currency: "USD"},
		},
	}
	itin := makeItin(leg)
	got := formatStops(itin)
	if got != "1" {
		t.Errorf("formatStops = %q, want %q", got, "1")
	}
}

// --- legCabin ---

func TestLegCabin_ReturnsCabinFromFirstSegment(t *testing.T) {
	leg := makeLeg("BA", "JFK", "LHR", basetime, 7*time.Hour, 450, nil)
	leg.Flight.Outbound[0].CabinClass = types.CabinBusiness
	itin := makeItin(leg)
	got := legCabin(itin, 0)
	if got != "business" {
		t.Errorf("legCabin = %q, want %q", got, "business")
	}
}

func TestLegCabin_OutOfBounds(t *testing.T) {
	itin := makeItin(makeLeg("BA", "JFK", "LHR", basetime, 7*time.Hour, 450, nil))
	got := legCabin(itin, 5)
	if got != "" {
		t.Errorf("legCabin = %q, want empty for out-of-bounds index", got)
	}
}

// --- printTable cabin column ---

func TestPrintTable_CabinColumn(t *testing.T) {
	leg := makeLeg("BA", "JFK", "LHR", basetime, 7*time.Hour, 450, nil)
	leg.Flight.Outbound[0].CabinClass = types.CabinBusiness
	itins := []search.Itinerary{makeItin(leg)}
	out := capturePrintTable(itins, "USD")

	if !strings.Contains(out, "CABIN") {
		t.Error("table output missing CABIN header")
	}
	if !strings.Contains(out, "business") {
		t.Error("table output missing cabin class value")
	}
}

// --- JSON cabin_class ---

func TestBuildJSONItineraries_CabinClass(t *testing.T) {
	leg := makeLeg("BA", "JFK", "LHR", basetime, 7*time.Hour, 450, nil)
	leg.Flight.Outbound[0].CabinClass = types.CabinPremiumEconomy
	results := buildJSONItineraries([]search.Itinerary{makeItin(leg)}, "USD")

	if len(results) != 1 || len(results[0].Legs) != 1 {
		t.Fatal("unexpected results structure")
	}
	if results[0].Legs[0].CabinClass != "premium_economy" {
		t.Errorf("cabin_class = %q, want %q", results[0].Legs[0].CabinClass, "premium_economy")
	}
}

// --- JSON aircraft ---

func TestBuildJSONItineraries_Aircraft(t *testing.T) {
	leg := makeLeg("BA", "JFK", "LHR", basetime, 7*time.Hour, 450, nil)
	leg.Flight.Outbound[0].Aircraft = "Boeing 787-9"
	results := buildJSONItineraries([]search.Itinerary{makeItin(leg)}, "USD")

	if len(results) != 1 || len(results[0].Legs) != 1 {
		t.Fatal("unexpected results structure")
	}
	if results[0].Legs[0].Aircraft != "Boeing 787-9" {
		t.Errorf("aircraft = %q, want %q", results[0].Legs[0].Aircraft, "Boeing 787-9")
	}
}

func TestBuildJSONItineraries_Aircraft_OmitEmpty(t *testing.T) {
	itin := makeItin(makeLeg("BA", "JFK", "LHR", basetime, 7*time.Hour, 450, nil))
	results := buildJSONItineraries([]search.Itinerary{itin}, "USD")

	data, err := json.Marshal(results[0].Legs[0])
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), "aircraft") {
		t.Error("aircraft should be omitted when empty")
	}
}

// --- JSON legroom ---

func TestBuildJSONItineraries_Legroom(t *testing.T) {
	leg := makeLeg("BA", "JFK", "LHR", basetime, 7*time.Hour, 450, nil)
	leg.Flight.Outbound[0].Legroom = "32 in"
	results := buildJSONItineraries([]search.Itinerary{makeItin(leg)}, "USD")

	if len(results) != 1 || len(results[0].Legs) != 1 {
		t.Fatal("unexpected results structure")
	}
	if results[0].Legs[0].Legroom != "32 in" {
		t.Errorf("legroom = %q, want %q", results[0].Legs[0].Legroom, "32 in")
	}
}

func TestBuildJSONItineraries_Legroom_OmitEmpty(t *testing.T) {
	itin := makeItin(makeLeg("BA", "JFK", "LHR", basetime, 7*time.Hour, 450, nil))
	results := buildJSONItineraries([]search.Itinerary{itin}, "USD")

	data, err := json.Marshal(results[0].Legs[0])
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), "legroom") {
		t.Error("legroom should be omitted when empty")
	}
}

// --- JSON flight_number ---

func TestBuildJSONItineraries_FlightNumber(t *testing.T) {
	leg := makeLeg("BA", "JFK", "LHR", basetime, 7*time.Hour, 450, nil)
	leg.Flight.Outbound[0].FlightNumber = "BA117"
	results := buildJSONItineraries([]search.Itinerary{makeItin(leg)}, "USD")

	if len(results) != 1 || len(results[0].Legs) != 1 {
		t.Fatal("unexpected results structure")
	}
	if results[0].Legs[0].FlightNumber != "BA117" {
		t.Errorf("flight_number = %q, want %q", results[0].Legs[0].FlightNumber, "BA117")
	}
}

func TestBuildJSONItineraries_FlightNumber_OmitEmpty(t *testing.T) {
	itin := makeItin(makeLeg("BA", "JFK", "LHR", basetime, 7*time.Hour, 450, nil))
	results := buildJSONItineraries([]search.Itinerary{itin}, "USD")

	data, err := json.Marshal(results[0].Legs[0])
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), "flight_number") {
		t.Error("flight_number should be omitted when empty")
	}
}

// --- legCarbon ---

func TestLegCarbon_WithEmissions(t *testing.T) {
	leg := makeLeg("AC", "DEL", "YYZ", basetime, 14*time.Hour, 850, nil)
	leg.Flight.CarbonKg = 1106
	itin := makeItin(leg)
	got := legCarbon(itin, 0)
	if got != "1106kg" {
		t.Errorf("legCarbon = %q, want %q", got, "1106kg")
	}
}

func TestLegCarbon_Zero(t *testing.T) {
	itin := makeItin(makeLeg("BA", "JFK", "LHR", basetime, 7*time.Hour, 450, nil))
	got := legCarbon(itin, 0)
	if got != "" {
		t.Errorf("legCarbon = %q, want empty for zero emissions", got)
	}
}

func TestLegCarbon_OutOfBounds(t *testing.T) {
	itin := makeItin(makeLeg("BA", "JFK", "LHR", basetime, 7*time.Hour, 450, nil))
	got := legCarbon(itin, 5)
	if got != "" {
		t.Errorf("legCarbon = %q, want empty for out-of-bounds index", got)
	}
}

func TestPrintTable_CO2Column(t *testing.T) {
	leg := makeLeg("AC", "DEL", "YYZ", basetime, 14*time.Hour, 850, nil)
	leg.Flight.CarbonKg = 1106
	itins := []search.Itinerary{makeItin(leg)}
	out := capturePrintTable(itins, "USD")

	if !strings.Contains(out, "CO2") {
		t.Error("table output missing CO2 header")
	}
	if !strings.Contains(out, "1106kg") {
		t.Errorf("table output missing carbon value, got:\n%s", out)
	}
}

func TestBuildJSONItineraries_CarbonKg(t *testing.T) {
	leg := makeLeg("AC", "DEL", "YYZ", basetime, 14*time.Hour, 850, nil)
	leg.Flight.CarbonKg = 1106
	itin := makeItin(leg)
	results := buildJSONItineraries([]search.Itinerary{itin}, "USD")

	if len(results) != 1 || len(results[0].Legs) != 1 {
		t.Fatalf("unexpected results shape")
	}
	if results[0].Legs[0].CarbonKg != 1106 {
		t.Errorf("CarbonKg = %d, want 1106", results[0].Legs[0].CarbonKg)
	}
}

func TestBuildJSONItineraries_CarbonKg_OmitEmpty(t *testing.T) {
	itin := makeItin(makeLeg("BA", "JFK", "LHR", basetime, 7*time.Hour, 450, nil))
	results := buildJSONItineraries([]search.Itinerary{itin}, "USD")

	data, err := json.Marshal(results[0].Legs[0])
	if err != nil {
		t.Fatalf("json.Marshal error: %v", err)
	}
	if bytes.Contains(data, []byte(`"carbon_kg"`)) {
		t.Error("zero carbon_kg should be omitted from JSON with omitempty")
	}
}

func TestBuildJSONItineraries_CarbonBenchmark(t *testing.T) {
	leg := makeLeg("AC", "DEL", "YYZ", basetime, 14*time.Hour, 850, nil)
	leg.Flight.CarbonKg = 1106
	leg.Flight.TypicalCarbonKg = 949
	leg.Flight.CarbonDiffPct = 17
	itin := makeItin(leg)
	results := buildJSONItineraries([]search.Itinerary{itin}, "USD")

	if len(results) != 1 || len(results[0].Legs) != 1 {
		t.Fatalf("unexpected results shape")
	}
	jl := results[0].Legs[0]
	if jl.TypicalCarbonKg != 949 {
		t.Errorf("TypicalCarbonKg = %d, want 949", jl.TypicalCarbonKg)
	}
	if jl.CarbonDiffPct != 17 {
		t.Errorf("CarbonDiffPct = %d, want 17", jl.CarbonDiffPct)
	}
}

func TestBuildJSONItineraries_CarbonBenchmark_OmitEmpty(t *testing.T) {
	itin := makeItin(makeLeg("BA", "JFK", "LHR", basetime, 7*time.Hour, 450, nil))
	results := buildJSONItineraries([]search.Itinerary{itin}, "USD")

	data, err := json.Marshal(results[0].Legs[0])
	if err != nil {
		t.Fatalf("json.Marshal error: %v", err)
	}
	if bytes.Contains(data, []byte(`"typical_carbon_kg"`)) {
		t.Error("zero typical_carbon_kg should be omitted from JSON")
	}
	if bytes.Contains(data, []byte(`"carbon_diff_percent"`)) {
		t.Error("zero carbon_diff_percent should be omitted from JSON")
	}
}

// --- hasScores ---

func TestHasScores_AllZero(t *testing.T) {
	itin := makeItin(makeLeg("BA", "JFK", "LHR", basetime, 7*time.Hour, 450, nil))
	itin.Score = 0
	if hasScores([]search.Itinerary{itin}) {
		t.Error("hasScores should return false when all scores are 0")
	}
}

func TestHasScores_SomeNonZero(t *testing.T) {
	itin1 := makeItin(makeLeg("BA", "JFK", "LHR", basetime, 7*time.Hour, 450, nil))
	itin1.Score = 0
	itin2 := makeItin(makeLeg("AA", "JFK", "LHR", basetime, 8*time.Hour, 500, nil))
	itin2.Score = 72
	if !hasScores([]search.Itinerary{itin1, itin2}) {
		t.Error("hasScores should return true when any score is non-zero")
	}
}

// --- conditional Score/Reason columns in table ---

func TestPrintTable_HidesScoreReasonWhenNoScores(t *testing.T) {
	itin := makeItin(makeLeg("BA", "JFK", "LHR", basetime, 7*time.Hour, 450, nil))
	itin.Score = 0
	itin.Reasoning = ""
	out := capturePrintTable([]search.Itinerary{itin}, "USD")

	if strings.Contains(out, "SCORE") {
		t.Errorf("table should not contain SCORE header when all scores are 0, got:\n%s", out)
	}
	if strings.Contains(out, "REASON") {
		t.Errorf("table should not contain REASON header when all scores are 0, got:\n%s", out)
	}
}

func TestPrintTable_ShowsScoreReasonWhenScored(t *testing.T) {
	itin := makeItin(makeLeg("BA", "JFK", "LHR", basetime, 7*time.Hour, 450, nil))
	itin.Score = 85
	itin.Reasoning = "good price"
	out := capturePrintTable([]search.Itinerary{itin}, "USD")

	if !strings.Contains(out, "SCORE") {
		t.Errorf("table should contain SCORE header when scores exist, got:\n%s", out)
	}
	if !strings.Contains(out, "REASON") {
		t.Errorf("table should contain REASON header when scores exist, got:\n%s", out)
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

// --- JSON arrival and stops ---

func TestBuildJSONItineraries_ArrivalAndStops(t *testing.T) {
	// Single segment: arrival = segment arrival time, stops = 0.
	leg := makeLeg("BA", "JFK", "LHR", basetime, 7*time.Hour, 450, nil)
	results := buildJSONItineraries([]search.Itinerary{makeItin(leg)}, "USD")

	if len(results) != 1 || len(results[0].Legs) != 1 {
		t.Fatal("unexpected results structure")
	}
	wantArrival := basetime.Add(7 * time.Hour).Format(time.RFC3339)
	if results[0].Legs[0].Arrival != wantArrival {
		t.Errorf("arrival = %q, want %q", results[0].Legs[0].Arrival, wantArrival)
	}
	if results[0].Legs[0].Stops != 0 {
		t.Errorf("stops = %d, want 0", results[0].Legs[0].Stops)
	}
}

func TestBuildJSONItineraries_StopsWithConnection(t *testing.T) {
	// Two segments = 1 stop.
	leg := search.Leg{
		Flight: types.Flight{
			Outbound: []types.Segment{
				{Airline: "LH", Origin: "DEL", Destination: "FRA", DepartureTime: basetime, ArrivalTime: basetime.Add(8 * time.Hour)},
				{Airline: "LH", Origin: "FRA", Destination: "YYZ", DepartureTime: basetime.Add(10 * time.Hour), ArrivalTime: basetime.Add(20 * time.Hour)},
			},
			TotalDuration: 20 * time.Hour,
			Price:         types.Money{Amount: 600, Currency: "USD"},
		},
	}
	results := buildJSONItineraries([]search.Itinerary{makeItin(leg)}, "USD")

	if results[0].Legs[0].Stops != 1 {
		t.Errorf("stops = %d, want 1", results[0].Legs[0].Stops)
	}
	// Arrival should be last segment's arrival.
	wantArrival := basetime.Add(20 * time.Hour).Format(time.RFC3339)
	if results[0].Legs[0].Arrival != wantArrival {
		t.Errorf("arrival = %q, want %q", results[0].Legs[0].Arrival, wantArrival)
	}
}

// --- JSON score omitempty ---

func TestBuildJSONItineraries_ScoreOmitEmpty(t *testing.T) {
	itin := makeItin(makeLeg("BA", "JFK", "LHR", basetime, 7*time.Hour, 450, nil))
	itin.Score = 0
	results := buildJSONItineraries([]search.Itinerary{itin}, "USD")

	data, err := json.Marshal(results[0])
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(data, []byte(`"score"`)) {
		t.Error("zero score should be omitted from JSON with omitempty")
	}
}

func TestBuildJSONItineraries_ScorePresent(t *testing.T) {
	itin := makeItin(makeLeg("BA", "JFK", "LHR", basetime, 7*time.Hour, 450, nil))
	itin.Score = 85
	results := buildJSONItineraries([]search.Itinerary{itin}, "USD")

	if results[0].Score != 85 {
		t.Errorf("score = %f, want 85", results[0].Score)
	}
}

// --- isNextDay ---

func TestIsNextDay_SameDay(t *testing.T) {
	dep := time.Date(2026, 4, 10, 8, 0, 0, 0, time.UTC)
	arr := time.Date(2026, 4, 10, 15, 0, 0, 0, time.UTC)
	if isNextDay(dep, arr) {
		t.Error("isNextDay should be false for same-day arrival")
	}
}

func TestIsNextDay_NextDay(t *testing.T) {
	dep := time.Date(2026, 4, 10, 22, 0, 0, 0, time.UTC)
	arr := time.Date(2026, 4, 11, 6, 0, 0, 0, time.UTC)
	if !isNextDay(dep, arr) {
		t.Error("isNextDay should be true for next-day arrival")
	}
}

func TestIsNextDay_TwoDaysLater(t *testing.T) {
	dep := time.Date(2026, 4, 10, 8, 0, 0, 0, time.UTC)
	arr := time.Date(2026, 4, 12, 14, 0, 0, 0, time.UTC)
	if !isNextDay(dep, arr) {
		t.Error("isNextDay should be true for multi-day arrival")
	}
}

// --- legArrival next-day marker ---

func TestLegArrival_NextDayMarker(t *testing.T) {
	// Departure at 22:00, arrival 14h later = next day 12:00.
	itin := makeItin(makeLeg("BA", "JFK", "LHR", basetime.Add(14*time.Hour), 14*time.Hour, 450, nil))
	// dep = Apr 10 22:00, arr = Apr 11 12:00
	got := legArrival(itin, 0)
	if !strings.Contains(got, "(+1)") {
		t.Errorf("legArrival should contain (+1) for next-day arrival, got %q", got)
	}
}

func TestLegArrival_SameDayNoMarker(t *testing.T) {
	// Departure at 08:00, arrival 7h later = same day 15:00.
	itin := makeItin(makeLeg("BA", "JFK", "LHR", basetime, 7*time.Hour, 450, nil))
	got := legArrival(itin, 0)
	if strings.Contains(got, "(+1)") {
		t.Errorf("legArrival should NOT contain (+1) for same-day arrival, got %q", got)
	}
}

// --- JSON arrival_next_day ---

func TestBuildJSONItineraries_ArrivalNextDay(t *testing.T) {
	// Next-day arrival: depart 22:00, arrive 14h later.
	leg := makeLeg("BA", "JFK", "LHR", basetime.Add(14*time.Hour), 14*time.Hour, 450, nil)
	results := buildJSONItineraries([]search.Itinerary{makeItin(leg)}, "USD")

	if len(results) != 1 || len(results[0].Legs) != 1 {
		t.Fatal("unexpected results structure")
	}
	if !results[0].Legs[0].ArrivalNextDay {
		t.Error("arrival_next_day should be true for next-day arrival")
	}
}

func TestBuildJSONItineraries_ArrivalNextDay_SameDay(t *testing.T) {
	leg := makeLeg("BA", "JFK", "LHR", basetime, 7*time.Hour, 450, nil)
	results := buildJSONItineraries([]search.Itinerary{makeItin(leg)}, "USD")

	if results[0].Legs[0].ArrivalNextDay {
		t.Error("arrival_next_day should be false for same-day arrival")
	}
	// Also verify omitempty: false should be omitted.
	data, _ := json.Marshal(results[0].Legs[0])
	if strings.Contains(string(data), "arrival_next_day") {
		t.Error("arrival_next_day should be omitted when false")
	}
}

// --- legSeatsLeft ---

func TestLegSeatsLeft_MinAcrossSegments(t *testing.T) {
	leg := search.Leg{
		Flight: types.Flight{
			Outbound: []types.Segment{
				{Airline: "LH", Origin: "DEL", Destination: "FRA", DepartureTime: basetime, ArrivalTime: basetime.Add(8 * time.Hour), SeatsLeft: 5},
				{Airline: "LH", Origin: "FRA", Destination: "YYZ", DepartureTime: basetime.Add(10 * time.Hour), ArrivalTime: basetime.Add(20 * time.Hour), SeatsLeft: 3},
			},
			TotalDuration: 20 * time.Hour,
			Price:         types.Money{Amount: 600, Currency: "USD"},
		},
	}
	itin := makeItin(leg)
	got := legSeatsLeft(itin, 0)
	if got != 3 {
		t.Errorf("legSeatsLeft = %d, want 3 (min across segments)", got)
	}
}

func TestLegSeatsLeft_NoData(t *testing.T) {
	itin := makeItin(makeLeg("BA", "JFK", "LHR", basetime, 7*time.Hour, 450, nil))
	got := legSeatsLeft(itin, 0)
	if got != 0 {
		t.Errorf("legSeatsLeft = %d, want 0 when no segment has SeatsLeft", got)
	}
}

// --- JSON seats_left ---

func TestBuildJSONItineraries_SeatsLeft(t *testing.T) {
	leg := makeLeg("BA", "JFK", "LHR", basetime, 7*time.Hour, 450, nil)
	leg.Flight.Outbound[0].SeatsLeft = 3
	results := buildJSONItineraries([]search.Itinerary{makeItin(leg)}, "USD")

	if len(results) != 1 || len(results[0].Legs) != 1 {
		t.Fatal("unexpected results structure")
	}
	if results[0].Legs[0].SeatsLeft != 3 {
		t.Errorf("seats_left = %d, want 3", results[0].Legs[0].SeatsLeft)
	}
}

func TestBuildJSONItineraries_SeatsLeft_OmitEmpty(t *testing.T) {
	itin := makeItin(makeLeg("BA", "JFK", "LHR", basetime, 7*time.Hour, 450, nil))
	results := buildJSONItineraries([]search.Itinerary{itin}, "USD")

	data, err := json.Marshal(results[0].Legs[0])
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), "seats_left") {
		t.Error("seats_left should be omitted when 0")
	}
}

// --- JSON airline_code, origin_city, destination_city, origin_name, destination_name ---

func TestBuildJSONItineraries_AirlineCodeAndCityNames(t *testing.T) {
	leg := makeLeg("BA", "JFK", "LHR", basetime, 7*time.Hour, 450, nil)
	leg.Flight.Outbound[0].OriginCity = "New York"
	leg.Flight.Outbound[0].DestinationCity = "London"
	leg.Flight.Outbound[0].OriginName = "John F. Kennedy International Airport"
	leg.Flight.Outbound[0].DestinationName = "Heathrow Airport"
	results := buildJSONItineraries([]search.Itinerary{makeItin(leg)}, "USD")

	if len(results) != 1 || len(results[0].Legs) != 1 {
		t.Fatal("unexpected results structure")
	}
	jl := results[0].Legs[0]
	if jl.AirlineCode != "BA" {
		t.Errorf("airline_code = %q, want %q", jl.AirlineCode, "BA")
	}
	if jl.OriginCity != "New York" {
		t.Errorf("origin_city = %q, want %q", jl.OriginCity, "New York")
	}
	if jl.DestinationCity != "London" {
		t.Errorf("destination_city = %q, want %q", jl.DestinationCity, "London")
	}
	if jl.OriginName != "John F. Kennedy International Airport" {
		t.Errorf("origin_name = %q, want %q", jl.OriginName, "John F. Kennedy International Airport")
	}
	if jl.DestinationName != "Heathrow Airport" {
		t.Errorf("destination_name = %q, want %q", jl.DestinationName, "Heathrow Airport")
	}
}

func TestBuildJSONItineraries_AirlineCodeAndCityNames_OmitEmpty(t *testing.T) {
	// makeLeg always sets Airline, so airline_code will be present.
	// City/airport name fields are empty by default and should be omitted.
	itin := makeItin(makeLeg("BA", "JFK", "LHR", basetime, 7*time.Hour, 450, nil))
	results := buildJSONItineraries([]search.Itinerary{itin}, "USD")

	data, err := json.Marshal(results[0].Legs[0])
	if err != nil {
		t.Fatal(err)
	}
	s := string(data)
	for _, field := range []string{"origin_city", "destination_city", "origin_name", "destination_name"} {
		if strings.Contains(s, field) {
			t.Errorf("%s should be omitted when empty", field)
		}
	}
	// airline_code should be present since makeLeg sets Airline="BA".
	if !strings.Contains(s, "airline_code") {
		t.Error("airline_code should be present when Airline is set")
	}
}

func TestBuildJSONItineraries_CityNames_MultiSegment(t *testing.T) {
	// Two segments: origin_city from first, destination_city from last.
	leg := search.Leg{
		Flight: types.Flight{
			Outbound: []types.Segment{
				{Airline: "LH", Origin: "DEL", Destination: "FRA", DepartureTime: basetime, ArrivalTime: basetime.Add(8 * time.Hour), OriginCity: "Delhi", DestinationCity: "Frankfurt", OriginName: "Indira Gandhi Airport", DestinationName: "Frankfurt Airport"},
				{Airline: "LH", Origin: "FRA", Destination: "YYZ", DepartureTime: basetime.Add(10 * time.Hour), ArrivalTime: basetime.Add(20 * time.Hour), OriginCity: "Frankfurt", DestinationCity: "Toronto", OriginName: "Frankfurt Airport", DestinationName: "Toronto Pearson"},
			},
			TotalDuration: 20 * time.Hour,
			Price:         types.Money{Amount: 600, Currency: "USD"},
		},
	}
	results := buildJSONItineraries([]search.Itinerary{makeItin(leg)}, "USD")

	jl := results[0].Legs[0]
	if jl.AirlineCode != "LH" {
		t.Errorf("airline_code = %q, want %q", jl.AirlineCode, "LH")
	}
	if jl.OriginCity != "Delhi" {
		t.Errorf("origin_city = %q, want %q (from first segment)", jl.OriginCity, "Delhi")
	}
	if jl.DestinationCity != "Toronto" {
		t.Errorf("destination_city = %q, want %q (from last segment)", jl.DestinationCity, "Toronto")
	}
	if jl.OriginName != "Indira Gandhi Airport" {
		t.Errorf("origin_name = %q, want %q (from first segment)", jl.OriginName, "Indira Gandhi Airport")
	}
	if jl.DestinationName != "Toronto Pearson" {
		t.Errorf("destination_name = %q, want %q (from last segment)", jl.DestinationName, "Toronto Pearson")
	}
}

// --- edge cases: empty segments ---

func makeEmptySegLeg() search.Leg {
	return search.Leg{
		Flight: types.Flight{
			Outbound:      nil,
			TotalDuration: 0,
			Price:         types.Money{Amount: 100, Currency: "USD"},
		},
	}
}

func TestLegAircraft_OutOfBounds(t *testing.T) {
	itin := makeItin(makeLeg("BA", "JFK", "LHR", basetime, 7*time.Hour, 450, nil))
	if got := legAircraft(itin, 5); got != "" {
		t.Errorf("legAircraft(OOB) = %q, want empty", got)
	}
}

func TestLegAircraft_EmptySegments(t *testing.T) {
	itin := makeItin(makeEmptySegLeg())
	if got := legAircraft(itin, 0); got != "" {
		t.Errorf("legAircraft(empty segs) = %q, want empty", got)
	}
}

func TestLegLegroom_OutOfBounds(t *testing.T) {
	itin := makeItin(makeLeg("BA", "JFK", "LHR", basetime, 7*time.Hour, 450, nil))
	if got := legLegroom(itin, 5); got != "" {
		t.Errorf("legLegroom(OOB) = %q, want empty", got)
	}
}

func TestLegLegroom_EmptySegments(t *testing.T) {
	itin := makeItin(makeEmptySegLeg())
	if got := legLegroom(itin, 0); got != "" {
		t.Errorf("legLegroom(empty segs) = %q, want empty", got)
	}
}

func TestLegBookingURL_OutOfBounds(t *testing.T) {
	itin := makeItin(makeLeg("BA", "JFK", "LHR", basetime, 7*time.Hour, 450, nil))
	if got := legBookingURL(itin, 5); got != "" {
		t.Errorf("legBookingURL(OOB) = %q, want empty", got)
	}
}

func TestLegCabin_EmptySegments(t *testing.T) {
	itin := makeItin(makeEmptySegLeg())
	if got := legCabin(itin, 0); got != "" {
		t.Errorf("legCabin(empty segs) = %q, want empty", got)
	}
}

func TestLegDeparture_EmptySegments(t *testing.T) {
	itin := makeItin(makeEmptySegLeg())
	if got := legDeparture(itin, 0); got != "" {
		t.Errorf("legDeparture(empty segs) = %q, want empty", got)
	}
}

func TestLegArrival_EmptySegments(t *testing.T) {
	itin := makeItin(makeEmptySegLeg())
	if got := legArrival(itin, 0); got != "" {
		t.Errorf("legArrival(empty segs) = %q, want empty", got)
	}
}

func TestLegSeatsLeft_OutOfBounds(t *testing.T) {
	itin := makeItin(makeLeg("BA", "JFK", "LHR", basetime, 7*time.Hour, 450, nil))
	if got := legSeatsLeft(itin, 5); got != 0 {
		t.Errorf("legSeatsLeft(OOB) = %d, want 0", got)
	}
}

func TestFormatPriceInsights_PriceLevelOnly(t *testing.T) {
	// PriceLevel set but range is zero — should show only the level.
	pi := search.PriceInsights{PriceLevel: "high"}
	got := formatPriceInsights(pi)
	if got != "Price level: high" {
		t.Errorf("formatPriceInsights = %q, want %q", got, "Price level: high")
	}
}

// --- multi-leg CO2 columns ---

func TestPrintTable_MultiLeg_CO2BothLegs(t *testing.T) {
	leg1 := makeLeg("CX", "DEL", "HKG", basetime, 8*time.Hour, 300,
		&search.Stopover{City: "Hong Kong", Airport: "HKG", Duration: 72 * time.Hour})
	leg1.Flight.CarbonKg = 450
	leg2 := makeLeg("AC", "HKG", "YYZ", basetime.Add(72*time.Hour), 16*time.Hour, 500, nil)
	leg2.Flight.CarbonKg = 890
	itins := []search.Itinerary{makeItin(leg1, leg2)}
	out := capturePrintTable(itins, "CAD")

	// Multi-leg should have separate CO2 columns for each leg.
	if !strings.Contains(out, "L1 CO2") {
		t.Errorf("multi-leg table missing L1 CO2 header, got:\n%s", out)
	}
	if !strings.Contains(out, "L2 CO2") {
		t.Errorf("multi-leg table missing L2 CO2 header, got:\n%s", out)
	}
	if !strings.Contains(out, "450kg") {
		t.Errorf("multi-leg table missing leg 1 carbon value 450kg, got:\n%s", out)
	}
	if !strings.Contains(out, "890kg") {
		t.Errorf("multi-leg table missing leg 2 carbon value 890kg, got:\n%s", out)
	}
}

// --- multi-leg cabin columns ---

// --- total_trip in JSON ---

func TestBuildJSONItineraries_TotalTrip(t *testing.T) {
	leg1 := makeLeg("CX", "DEL", "HKG", basetime, 8*time.Hour, 300,
		&search.Stopover{City: "Hong Kong", Airport: "HKG", Duration: 72 * time.Hour})
	leg2 := makeLeg("AC", "HKG", "YYZ", basetime.Add(80*time.Hour), 16*time.Hour, 500, nil)
	itin := makeItin(leg1, leg2)
	itin.TotalTrip = 96 * time.Hour // 4 days
	results := buildJSONItineraries([]search.Itinerary{itin}, "USD")

	if results[0].TotalTrip != "4d 0h" {
		t.Errorf("total_trip = %q, want %q", results[0].TotalTrip, "4d 0h")
	}
}

func TestBuildJSONItineraries_TotalTrip_OmitEmpty(t *testing.T) {
	itin := makeItin(makeLeg("BA", "JFK", "LHR", basetime, 7*time.Hour, 450, nil))
	results := buildJSONItineraries([]search.Itinerary{itin}, "USD")

	// Marshal to JSON and check total_trip is not present when zero.
	data, _ := json.Marshal(results[0])
	if strings.Contains(string(data), "total_trip") {
		t.Errorf("total_trip should be omitted when zero, got: %s", data)
	}
}

// --- formatFareTrend ---

func TestFormatFareTrend_MultiDate(t *testing.T) {
	ft := search.FareTrend{
		CheapestDate: "2026-03-16",
		PriciestDate: "2026-03-18",
		MinPrice:     300,
		MaxPrice:     800,
	}
	got := formatFareTrend(ft)
	if !strings.Contains(got, "Mar 16") || !strings.Contains(got, "cheapest") {
		t.Errorf("formatFareTrend missing cheapest date: %q", got)
	}
	if !strings.Contains(got, "Mar 18") || !strings.Contains(got, "expensive") {
		t.Errorf("formatFareTrend missing priciest date: %q", got)
	}
	if !strings.Contains(got, "$500") {
		t.Errorf("formatFareTrend missing difference: %q", got)
	}
}

func TestFormatFareTrend_SameDate(t *testing.T) {
	ft := search.FareTrend{
		CheapestDate: "2026-03-16",
		PriciestDate: "2026-03-16",
		MinPrice:     300,
		MaxPrice:     300,
	}
	got := formatFareTrend(ft)
	if got != "" {
		t.Errorf("formatFareTrend(same date) = %q, want empty", got)
	}
}

func TestFormatFareTrend_Empty(t *testing.T) {
	got := formatFareTrend(search.FareTrend{})
	if got != "" {
		t.Errorf("formatFareTrend(empty) = %q, want empty", got)
	}
}

func TestPrintTable_MultiLeg_CabinBothLegs(t *testing.T) {
	leg1 := makeLeg("CX", "DEL", "HKG", basetime, 8*time.Hour, 300,
		&search.Stopover{City: "Hong Kong", Airport: "HKG", Duration: 72 * time.Hour})
	leg1.Flight.Outbound[0].CabinClass = types.CabinBusiness
	leg2 := makeLeg("AC", "HKG", "YYZ", basetime.Add(72*time.Hour), 16*time.Hour, 500, nil)
	leg2.Flight.Outbound[0].CabinClass = types.CabinEconomy
	itins := []search.Itinerary{makeItin(leg1, leg2)}
	out := capturePrintTable(itins, "CAD")

	// Multi-leg should have separate cabin columns for each leg.
	if !strings.Contains(out, "L1 CABIN") {
		t.Errorf("multi-leg table missing L1 CABIN header, got:\n%s", out)
	}
	if !strings.Contains(out, "L2 CABIN") {
		t.Errorf("multi-leg table missing L2 CABIN header, got:\n%s", out)
	}
	if !strings.Contains(out, "business") {
		t.Errorf("multi-leg table missing leg 1 cabin value business, got:\n%s", out)
	}
	if !strings.Contains(out, "economy") {
		t.Errorf("multi-leg table missing leg 2 cabin value economy, got:\n%s", out)
	}
}
