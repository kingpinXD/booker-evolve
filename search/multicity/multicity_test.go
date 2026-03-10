package multicity

import (
	"context"
	"log"
	"os"
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
	if os.Getenv("SERPAPI_KEY") == "" {
		t.Skip("skipping: SERPAPI_KEY not set (integration test)")
	}
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
	if len(results) == 0 {
		t.Fatal("expected itineraries but got 0")
	}

	// Print summary table.
	t.Logf("\n  %-3s  %-5s  %-11s  %-13s  %-35s  %-12s  %-7s  %-7s",
		"#", "Score", "Price (CAD)", "Route", "Airlines", "Stopover", "Leg1", "Leg2")
	t.Logf("  %s", "-------------------------------------------------------------------------------------------------------------")

	for i, itin := range results {
		cad, _ := currency.Convert(itin.TotalPrice, "CAD")

		// Route string.
		route := ""
		if len(itin.Legs) >= 2 {
			leg1 := itin.Legs[0].Flight.Outbound
			leg2 := itin.Legs[1].Flight.Outbound
			if len(leg1) > 0 && len(leg2) > 0 {
				route = leg1[0].Origin + "→" + leg2[0].Origin + "→" + leg2[len(leg2)-1].Destination
			}
		}

		// Airlines.
		airlines := ""
		for j, leg := range itin.Legs {
			if len(leg.Flight.Outbound) > 0 {
				if j > 0 {
					airlines += " + "
				}
				airlines += leg.Flight.Outbound[0].AirlineName
			}
		}

		// Stopover city.
		stopover := ""
		if itin.Legs[0].Stopover != nil {
			stopover = itin.Legs[0].Stopover.City
		}

		// Departure times for each leg.
		leg1Time, leg2Time := "", ""
		if len(itin.Legs) >= 1 && len(itin.Legs[0].Flight.Outbound) > 0 {
			leg1Time = itin.Legs[0].Flight.Outbound[0].DepartureTime.Format("15:04")
		}
		if len(itin.Legs) >= 2 && len(itin.Legs[1].Flight.Outbound) > 0 {
			leg2Time = itin.Legs[1].Flight.Outbound[0].DepartureTime.Format("15:04")
		}

		t.Logf("  %-3d  %-5.0f  C$%-9.2f  %-13s  %-35s  %-12s  %-7s  %-7s",
			i+1, itin.Score, cad.Amount, route, airlines, stopover, leg1Time, leg2Time)
	}
}
