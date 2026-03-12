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

// --- printJSONWithInsights ---

func TestPrintJSONWithInsights_WithPriceInsights(t *testing.T) {
	itin := makeItin(makeLeg("BA", "JFK", "LHR", basetime, 7*time.Hour, 450, nil))
	pi := search.PriceInsights{
		PriceLevel:        "low",
		LowestPrice:       380,
		TypicalPriceRange: [2]float64{400, 600},
	}

	var buf bytes.Buffer
	if err := printJSONWithInsights(&buf, []search.Itinerary{itin}, "USD", pi); err != nil {
		t.Fatalf("printJSONWithInsights error: %v", err)
	}

	var out struct {
		Results       []jsonItinerary    `json:"results"`
		PriceInsights *jsonPriceInsights `json:"price_insights"`
	}
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if out.PriceInsights == nil {
		t.Fatal("price_insights should be present")
	}
	if out.PriceInsights.PriceLevel != "low" {
		t.Errorf("price_level = %q, want %q", out.PriceInsights.PriceLevel, "low")
	}
	if out.PriceInsights.LowestPrice != 380 {
		t.Errorf("lowest_price = %v, want 380", out.PriceInsights.LowestPrice)
	}
	if out.PriceInsights.TypicalPriceRange != [2]float64{400, 600} {
		t.Errorf("typical_price_range = %v, want [400 600]", out.PriceInsights.TypicalPriceRange)
	}
}

func TestPrintJSONWithInsights_EmptyInsights(t *testing.T) {
	itin := makeItin(makeLeg("BA", "JFK", "LHR", basetime, 7*time.Hour, 450, nil))

	var buf bytes.Buffer
	if err := printJSONWithInsights(&buf, []search.Itinerary{itin}, "USD", search.PriceInsights{}); err != nil {
		t.Fatalf("printJSONWithInsights error: %v", err)
	}

	raw := buf.String()
	if strings.Contains(raw, "price_insights") {
		t.Errorf("price_insights key should be omitted when empty, got:\n%s", raw)
	}
}

func TestPrintJSONWithInsights_ResultsCount(t *testing.T) {
	itins := []search.Itinerary{
		makeItin(makeLeg("BA", "JFK", "LHR", basetime, 7*time.Hour, 450, nil)),
		makeItin(makeLeg("AC", "YYZ", "LHR", basetime, 8*time.Hour, 550, nil)),
		makeItin(makeLeg("LH", "FRA", "JFK", basetime, 9*time.Hour, 620, nil)),
	}

	var buf bytes.Buffer
	if err := printJSONWithInsights(&buf, itins, "USD", search.PriceInsights{}); err != nil {
		t.Fatalf("printJSONWithInsights error: %v", err)
	}

	var out struct {
		Results []jsonItinerary `json:"results"`
	}
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if len(out.Results) != 3 {
		t.Errorf("results count = %d, want 3", len(out.Results))
	}
	for i, r := range out.Results {
		if r.Rank != i+1 {
			t.Errorf("results[%d].rank = %d, want %d", i, r.Rank, i+1)
		}
	}
}

func TestPrintJSONWithInsights_EmptyResults(t *testing.T) {
	var buf bytes.Buffer
	if err := printJSONWithInsights(&buf, nil, "USD", search.PriceInsights{}); err != nil {
		t.Fatalf("printJSONWithInsights error: %v", err)
	}

	var out struct {
		Results []jsonItinerary `json:"results"`
	}
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if len(out.Results) != 0 {
		t.Errorf("results count = %d, want 0", len(out.Results))
	}
}

func TestPrintJSONWithInsights_InsightsFieldValues(t *testing.T) {
	itin := makeItin(makeLeg("BA", "JFK", "LHR", basetime, 7*time.Hour, 450, nil))
	pi := search.PriceInsights{
		PriceLevel:        "high",
		LowestPrice:       700,
		TypicalPriceRange: [2]float64{500, 900},
	}

	var buf bytes.Buffer
	if err := printJSONWithInsights(&buf, []search.Itinerary{itin}, "USD", pi); err != nil {
		t.Fatalf("printJSONWithInsights error: %v", err)
	}

	// Verify the raw JSON contains expected price_insights fields.
	raw := buf.String()
	for _, want := range []string{`"price_level"`, `"lowest_price"`, `"typical_price_range"`} {
		if !strings.Contains(raw, want) {
			t.Errorf("JSON missing %s field, got:\n%s", want, raw)
		}
	}
}

// --- segments in JSON output ---

func TestBuildJSONItineraries_MultiSegmentLeg(t *testing.T) {
	dep1 := basetime
	arr1 := dep1.Add(4 * time.Hour)
	dep2 := arr1.Add(2 * time.Hour) // 2h layover
	arr2 := dep2.Add(5 * time.Hour)

	itin := search.Itinerary{
		Legs: []search.Leg{{
			Flight: types.Flight{
				Outbound: []types.Segment{
					{
						Airline:         "BA",
						FlightNumber:    "BA117",
						Origin:          "JFK",
						Destination:     "LHR",
						DepartureTime:   dep1,
						ArrivalTime:     arr1,
						Duration:        4 * time.Hour,
						Aircraft:        "Boeing 777",
						Legroom:         "32 in",
						CabinClass:      types.CabinBusiness,
						LayoverDuration: 2 * time.Hour,
						Overnight:       false,
					},
					{
						Airline:       "BA",
						FlightNumber:  "BA223",
						Origin:        "LHR",
						Destination:   "CDG",
						DepartureTime: dep2,
						ArrivalTime:   arr2,
						Duration:      5 * time.Hour,
						Aircraft:      "Airbus A320",
						Legroom:       "30 in",
						CabinClass:    types.CabinBusiness,
						Overnight:     true,
					},
				},
				TotalDuration: 11 * time.Hour,
				Price:         types.Money{Amount: 800, Currency: "USD"},
			},
		}},
		TotalPrice:  types.Money{Amount: 800, Currency: "USD"},
		TotalTravel: 11 * time.Hour,
	}

	results := buildJSONItineraries([]search.Itinerary{itin}, "USD")
	if len(results) != 1 {
		t.Fatalf("results count = %d, want 1", len(results))
	}

	leg := results[0].Legs[0]
	if len(leg.Segments) != 2 {
		t.Fatalf("segments count = %d, want 2", len(leg.Segments))
	}

	// Verify first segment fields.
	seg0 := leg.Segments[0]
	if seg0.Airline != "BA" {
		t.Errorf("seg[0].airline = %q, want %q", seg0.Airline, "BA")
	}
	if seg0.FlightNumber != "BA117" {
		t.Errorf("seg[0].flight_number = %q, want %q", seg0.FlightNumber, "BA117")
	}
	if seg0.Origin != "JFK" {
		t.Errorf("seg[0].origin = %q, want %q", seg0.Origin, "JFK")
	}
	if seg0.Destination != "LHR" {
		t.Errorf("seg[0].destination = %q, want %q", seg0.Destination, "LHR")
	}
	if seg0.Departure != dep1.Format(time.RFC3339) {
		t.Errorf("seg[0].departure = %q, want %q", seg0.Departure, dep1.Format(time.RFC3339))
	}
	if seg0.Arrival != arr1.Format(time.RFC3339) {
		t.Errorf("seg[0].arrival = %q, want %q", seg0.Arrival, arr1.Format(time.RFC3339))
	}
	if seg0.Duration != "4h" {
		t.Errorf("seg[0].duration = %q, want %q", seg0.Duration, "4h")
	}
	if seg0.Aircraft != "Boeing 777" {
		t.Errorf("seg[0].aircraft = %q, want %q", seg0.Aircraft, "Boeing 777")
	}
	if seg0.Legroom != "32 in" {
		t.Errorf("seg[0].legroom = %q, want %q", seg0.Legroom, "32 in")
	}
	if seg0.LayoverDuration != "2h" {
		t.Errorf("seg[0].layover_duration = %q, want %q", seg0.LayoverDuration, "2h")
	}
	if seg0.Overnight {
		t.Error("seg[0].overnight should be false")
	}

	// Verify second segment.
	seg1 := leg.Segments[1]
	if seg1.FlightNumber != "BA223" {
		t.Errorf("seg[1].flight_number = %q, want %q", seg1.FlightNumber, "BA223")
	}
	if seg1.Origin != "LHR" {
		t.Errorf("seg[1].origin = %q, want %q", seg1.Origin, "LHR")
	}
	if seg1.Destination != "CDG" {
		t.Errorf("seg[1].destination = %q, want %q", seg1.Destination, "CDG")
	}
	if !seg1.Overnight {
		t.Error("seg[1].overnight should be true")
	}
	if seg1.LayoverDuration != "" {
		t.Errorf("seg[1].layover_duration = %q, want empty (last segment)", seg1.LayoverDuration)
	}

	// Existing leg-level fields should still be populated.
	if leg.Origin != "JFK" {
		t.Errorf("leg.origin = %q, want %q", leg.Origin, "JFK")
	}
	if leg.Dest != "CDG" {
		t.Errorf("leg.destination = %q, want %q", leg.Dest, "CDG")
	}
}

func TestBuildJSONItineraries_SingleSegmentLeg(t *testing.T) {
	dep := basetime
	arr := dep.Add(7 * time.Hour)

	itin := search.Itinerary{
		Legs: []search.Leg{{
			Flight: types.Flight{
				Outbound: []types.Segment{{
					Airline:       "AC",
					FlightNumber:  "AC850",
					Origin:        "YYZ",
					Destination:   "LHR",
					DepartureTime: dep,
					ArrivalTime:   arr,
					Duration:      7 * time.Hour,
				}},
				TotalDuration: 7 * time.Hour,
				Price:         types.Money{Amount: 500, Currency: "USD"},
			},
		}},
		TotalPrice:  types.Money{Amount: 500, Currency: "USD"},
		TotalTravel: 7 * time.Hour,
	}

	results := buildJSONItineraries([]search.Itinerary{itin}, "USD")
	leg := results[0].Legs[0]

	if len(leg.Segments) != 1 {
		t.Fatalf("segments count = %d, want 1", len(leg.Segments))
	}
	if leg.Segments[0].Airline != "AC" {
		t.Errorf("seg[0].airline = %q, want %q", leg.Segments[0].Airline, "AC")
	}
	if leg.Segments[0].Origin != "YYZ" {
		t.Errorf("seg[0].origin = %q, want %q", leg.Segments[0].Origin, "YYZ")
	}
}

// --- legCarbon ---

func carbonItin(carbonKg, diffPct int) search.Itinerary {
	return search.Itinerary{
		Legs: []search.Leg{{
			Flight: types.Flight{
				Outbound: []types.Segment{{
					Airline:       "BA",
					Origin:        "JFK",
					Destination:   "LHR",
					DepartureTime: basetime,
					ArrivalTime:   basetime.Add(7 * time.Hour),
				}},
				TotalDuration: 7 * time.Hour,
				Price:         types.Money{Amount: 450, Currency: "USD"},
				CarbonKg:      carbonKg,
				CarbonDiffPct: diffPct,
			},
		}},
		TotalPrice:  types.Money{Amount: 450, Currency: "USD"},
		TotalTravel: 7 * time.Hour,
	}
}

func TestLegCarbon_WithPositiveDiff(t *testing.T) {
	got := legCarbon(carbonItin(150, 5), 0)
	if got != "150kg (+5%)" {
		t.Errorf("legCarbon = %q, want %q", got, "150kg (+5%)")
	}
}

func TestLegCarbon_WithNegativeDiff(t *testing.T) {
	got := legCarbon(carbonItin(120, -12), 0)
	if got != "120kg (-12%)" {
		t.Errorf("legCarbon = %q, want %q", got, "120kg (-12%)")
	}
}

func TestLegCarbon_WithZeroDiff(t *testing.T) {
	got := legCarbon(carbonItin(130, 0), 0)
	if got != "130kg" {
		t.Errorf("legCarbon = %q, want %q", got, "130kg")
	}
}

func TestLegCarbon_NoCarbonData(t *testing.T) {
	got := legCarbon(carbonItin(0, 0), 0)
	if got != "" {
		t.Errorf("legCarbon = %q, want empty", got)
	}
}

func TestBuildJSONItineraries_CarbonDiffPct(t *testing.T) {
	itin := carbonItin(150, -8)
	results := buildJSONItineraries([]search.Itinerary{itin}, "USD")
	if len(results) != 1 {
		t.Fatalf("results count = %d, want 1", len(results))
	}
	leg := results[0].Legs[0]
	if leg.CarbonKg != 150 {
		t.Errorf("carbon_kg = %d, want 150", leg.CarbonKg)
	}
	if leg.CarbonDiffPct != -8 {
		t.Errorf("carbon_diff_percent = %d, want -8", leg.CarbonDiffPct)
	}
}

func TestBuildJSONItineraries_SegmentsOmittedWhenEmpty(t *testing.T) {
	// A leg with no segments should produce no segments key in JSON.
	itin := search.Itinerary{
		Legs: []search.Leg{{
			Flight: types.Flight{
				Outbound:      nil,
				TotalDuration: 0,
				Price:         types.Money{Amount: 100, Currency: "USD"},
			},
		}},
		TotalPrice:  types.Money{Amount: 100, Currency: "USD"},
		TotalTravel: 0,
	}

	var buf bytes.Buffer
	if err := printJSON(&buf, []search.Itinerary{itin}, "USD"); err != nil {
		t.Fatalf("printJSON error: %v", err)
	}
	if strings.Contains(buf.String(), `"segments"`) {
		t.Error("segments key should be omitted when no segments exist")
	}
}

// --- truncateText ---

func TestTruncateText(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		maxLen int
		want   string
	}{
		{"under limit", "short text", 50, "short text"},
		{"at limit", strings.Repeat("a", 50), 50, strings.Repeat("a", 50)},
		{"over limit", strings.Repeat("a", 51), 50, strings.Repeat("a", 47) + "..."},
		{"empty", "", 50, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := truncateText(tt.input, tt.maxLen)
			if got != tt.want {
				t.Errorf("truncateText(%d chars, %d) = %q (len %d), want %q (len %d)",
					len(tt.input), tt.maxLen, got, len(got), tt.want, len(tt.want))
			}
		})
	}
}

func TestPrintTable_ReasonTruncated(t *testing.T) {
	itin := makeItin(makeLeg("BA", "JFK", "LHR", basetime, 7*time.Hour, 450, nil))
	itin.Score = 85
	itin.Reasoning = strings.Repeat("x", 80) // 80 chars should be truncated

	var buf bytes.Buffer
	printTable(&buf, []search.Itinerary{itin}, "USD")
	out := buf.String()

	// The full 80-char string should not appear.
	if strings.Contains(out, strings.Repeat("x", 80)) {
		t.Error("expected Reason to be truncated in table output")
	}
	// But a truncated version with "..." should.
	if !strings.Contains(out, "...") {
		t.Error("expected '...' in truncated Reason")
	}
}

// --- printBulletResults ---

func TestPrintBulletResults_SingleLeg(t *testing.T) {
	itin := makeItin(makeLeg("BA", "JFK", "LHR", basetime, 7*time.Hour, 450, nil))
	itin.Score = 85
	itin.Reasoning = "good value"

	var buf bytes.Buffer
	printBulletResults(&buf, []search.Itinerary{itin}, "USD")
	out := buf.String()

	if !strings.Contains(out, "1.") {
		t.Error("expected numbered bullet")
	}
	if !strings.Contains(out, "BA") {
		t.Error("expected airline")
	}
	if !strings.Contains(out, "JFK") || !strings.Contains(out, "LHR") {
		t.Error("expected route airports")
	}
	if !strings.Contains(out, "$450") {
		t.Error("expected price")
	}
	if !strings.Contains(out, "Score: 85") {
		t.Error("expected score when scored")
	}
}

func TestPrintBulletResults_MultiLeg(t *testing.T) {
	leg1 := makeLeg("CX", "DEL", "HKG", basetime, 8*time.Hour, 300,
		&search.Stopover{City: "Hong Kong", Airport: "HKG", Duration: 72 * time.Hour})
	leg2 := makeLeg("AC", "HKG", "YYZ", basetime.Add(72*time.Hour), 16*time.Hour, 500, nil)
	itin := makeItin(leg1, leg2)

	var buf bytes.Buffer
	printBulletResults(&buf, []search.Itinerary{itin}, "USD")
	out := buf.String()

	if !strings.Contains(out, "1.") {
		t.Error("expected numbered bullet")
	}
	// Should show per-leg sub-bullets.
	if !strings.Contains(out, "Leg 1") {
		t.Error("expected Leg 1 sub-bullet")
	}
	if !strings.Contains(out, "Leg 2") {
		t.Error("expected Leg 2 sub-bullet")
	}
}

func TestPrintBulletResults_NoScore(t *testing.T) {
	itin := makeItin(makeLeg("BA", "JFK", "LHR", basetime, 7*time.Hour, 450, nil))
	itin.Score = 0

	var buf bytes.Buffer
	printBulletResults(&buf, []search.Itinerary{itin}, "USD")
	out := buf.String()

	if strings.Contains(out, "Score:") {
		t.Error("should not show score when zero")
	}
}

func TestPrintBulletResults_Empty(t *testing.T) {
	var buf bytes.Buffer
	printBulletResults(&buf, nil, "USD")
	if buf.Len() != 0 {
		t.Errorf("expected empty output for nil itineraries, got %q", buf.String())
	}
}
