package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"booker/config"
	"booker/httpclient"
	"booker/llm"
	"booker/provider"
	"booker/provider/cache"
	"booker/provider/serpapi"
	"booker/search"
	"booker/search/multicity"
	"booker/types"
)

func main() {
	cfg := config.Default()
	httpClient := httpclient.New(cfg.HTTP)

	// Register providers.
	registry := provider.NewRegistry()
	serpCfg := cfg.Providers[config.ProviderSerpAPI]
	raw := serpapi.New(serpCfg, httpClient)
	cached := cache.Wrap(raw, ".cache/flights", 0)
	if err := registry.Register(cached); err != nil {
		log.Fatalf("registering serpapi: %v", err)
	}

	// Create LLM client.
	llmClient := llm.New(cfg.LLM, httpClient)

	// Create multi-city searcher.
	searcher := multicity.NewSearcher(registry, llmClient, multicity.WeightsBudget)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Only search cities we have cached data for.
	cachedCities := []multicity.StopoverCity{
		{City: "Hong Kong", Airport: "HKG", Region: "east_asia",
			MinStay: types.DefaultMinStopover, MaxStay: types.DefaultMaxStopover},
		{City: "Bangkok", Airport: "BKK", Region: "southeast_asia",
			MinStay: types.DefaultMinStopover, MaxStay: types.DefaultMaxStopover},
		{City: "Istanbul", Airport: "IST", Region: "europe",
			MinStay: types.DefaultMinStopover, MaxStay: types.DefaultMaxStopover},
		{City: "Tokyo", Airport: "NRT", Region: "east_asia",
			MinStay: types.DefaultMinStopover, MaxStay: types.DefaultMaxStopover},
	}

	params := multicity.SearchParams{
		Origin:        "DEL",
		Destination:   "YYZ",
		DepartureDate: "2026-03-24",
		Leg2Date:      "2026-03-30",
		Passengers:    1,
		CabinClass:    types.CabinEconomy,
		Stopovers:     cachedCities,
		FlexDays:      0,
		MaxResults:    5,
	}

	log.Println("=== Booker: Multi-City Halt Search ===")
	log.Printf("Route: %s → %s", params.Origin, params.Destination)
	log.Printf("Target date: %s (±%d days)", params.DepartureDate, params.FlexDays)

	results, err := searcher.Search(ctx, params)
	if err != nil {
		log.Fatalf("search failed: %v", err)
	}

	if len(results) == 0 {
		log.Println("No itineraries found.")
		return
	}

	printResults(results)
}

func printResults(itineraries []search.Itinerary) {
	fmt.Println()
	fmt.Println("========================================")
	fmt.Println("  TOP ITINERARIES")
	fmt.Println("========================================")

	for i, itin := range itineraries {
		fmt.Printf("\n--- #%d | $%.2f | Score: %.0f ---\n", i+1, itin.TotalPrice.Amount, itin.Score)
		if itin.Reasoning != "" {
			fmt.Printf("    LLM: %s\n", itin.Reasoning)
		}
		fmt.Printf("    Total in-air: %s | Total trip: %s\n",
			formatDuration(itin.TotalTravel), formatDuration(itin.TotalTrip))

		for j, leg := range itin.Legs {
			fmt.Printf("\n    LEG %d ($%.2f):\n", j+1, leg.Flight.Price.Amount)
			for _, seg := range leg.Flight.Outbound {
				fmt.Printf("      %s  %s (%s) → %s (%s)\n",
					seg.FlightNumber, seg.OriginCity, seg.Origin,
					seg.DestinationCity, seg.Destination)
				fmt.Printf("        %s → %s  [%s]  %s\n",
					seg.DepartureTime.Format("Jan 02 15:04"),
					seg.ArrivalTime.Format("Jan 02 15:04"),
					formatDuration(seg.Duration),
					seg.AirlineName)
				if seg.LayoverDuration > 0 {
					fmt.Printf("        ↳ layover: %s\n", formatDuration(seg.LayoverDuration))
				}
			}
			if leg.Stopover != nil {
				fmt.Printf("\n    ✈ STOPOVER: %s (%s) — %s\n",
					leg.Stopover.City, leg.Stopover.Airport,
					formatDuration(leg.Stopover.Duration))
			}
		}
		fmt.Println()
	}
}

func formatDuration(d time.Duration) string {
	hours := int(d.Hours())
	mins := int(d.Minutes()) % 60
	if hours >= 24 {
		days := hours / 24
		hours = hours % 24
		return fmt.Sprintf("%dd %dh %dm", days, hours, mins)
	}
	return fmt.Sprintf("%dh %dm", hours, mins)
}
