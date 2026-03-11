package cmd

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"booker/search"
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
		Results       []jsonItinerary `json:"results"`
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
