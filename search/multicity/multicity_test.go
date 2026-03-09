package multicity

import (
	"context"
	"log"
	"testing"
	"time"

	"booker/config"
	"booker/currency"
	"booker/httpclient"
	"booker/llm"
	"booker/provider"
	"booker/provider/cache"
	"booker/provider/serpapi"
	"booker/types"
)

// TestDELToYYZ_March24 exercises the full multi-city pipeline using
// cached SerpAPI data: DEL → (stopover) → YYZ, departing March 24,
// leg2 on March 30, with 4 stopover cities.
func TestDELToYYZ_March24(t *testing.T) {
	cfg := config.Default()
	httpClient := httpclient.New(cfg.HTTP)

	registry := provider.NewRegistry()
	serpCfg := cfg.Providers[config.ProviderSerpAPI]
	raw := serpapi.New(serpCfg, httpClient)
	cached := cache.Wrap(raw, "../../.cache/flights", 0)
	if err := registry.Register(cached); err != nil {
		t.Fatalf("registering serpapi: %v", err)
	}

	llmClient := llm.New(cfg.LLM, httpClient)
	searcher := NewSearcher(registry, llmClient, WeightsBudget)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Only search the 4 cities we have cached data for.
	cachedCities := []StopoverCity{
		{City: "Hong Kong", Airport: "HKG", Region: "east_asia",
			MinStay: types.DefaultMinStopover, MaxStay: types.DefaultMaxStopover},
		{City: "Bangkok", Airport: "BKK", Region: "southeast_asia",
			MinStay: types.DefaultMinStopover, MaxStay: types.DefaultMaxStopover},
		{City: "Istanbul", Airport: "IST", Region: "europe",
			MinStay: types.DefaultMinStopover, MaxStay: types.DefaultMaxStopover},
		{City: "Tokyo", Airport: "NRT", Region: "east_asia",
			MinStay: types.DefaultMinStopover, MaxStay: types.DefaultMaxStopover},
	}

	params := SearchParams{
		Origin:          "DEL",
		Destination:     "YYZ",
		DepartureDate:   "2026-03-24",
		Leg2Date:        "2026-03-30",
		Passengers:      1,
		CabinClass:      types.CabinEconomy,
		Stopovers:       cachedCities,
		MaxLayoversLeg1: 1,
		MaxLayoversLeg2: 1,
		FlexDays:        0,
		MaxResults:      5,
	}

	log.Printf("=== Test: DEL → YYZ, leg1: 2026-03-24, leg2: 2026-03-30 ===")

	results, err := searcher.Search(ctx, params)
	if err != nil {
		t.Fatalf("search failed: %v", err)
	}

	t.Logf("Total itineraries found: %d", len(results))

	for i, itin := range results {
		cad, _ := currency.Convert(itin.TotalPrice, "CAD")
		t.Logf("\n--- #%d | C$%.2f (US$%.2f) | Score: %.0f ---",
			i+1, cad.Amount, itin.TotalPrice.Amount, itin.Score)
		if itin.Reasoning != "" {
			t.Logf("  LLM: %s", itin.Reasoning)
		}
		for j, leg := range itin.Legs {
			legCAD, _ := currency.Convert(leg.Flight.Price, "CAD")
			t.Logf("  LEG %d (C$%.2f):", j+1, legCAD.Amount)
			for _, seg := range leg.Flight.Outbound {
				t.Logf("    %s  %s(%s) → %s(%s)  %s → %s  [%s] %s",
					seg.FlightNumber,
					seg.OriginCity, seg.Origin,
					seg.DestinationCity, seg.Destination,
					seg.DepartureTime.Format("Jan 02 15:04"),
					seg.ArrivalTime.Format("Jan 02 15:04"),
					seg.Duration,
					seg.AirlineName)
			}
			if leg.Stopover != nil {
				t.Logf("  STOPOVER: %s (%s) — %s",
					leg.Stopover.City, leg.Stopover.Airport, leg.Stopover.Duration)
			}
		}
	}

	if len(results) == 0 {
		t.Fatal("expected itineraries but got 0")
	}
}
