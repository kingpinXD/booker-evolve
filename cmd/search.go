package cmd

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"booker/config"
	"booker/currency"
	"booker/httpclient"
	"booker/llm"
	"booker/provider"
	"booker/provider/cache"
	"booker/provider/serpapi"
	"booker/search"
	"booker/search/direct"
	"booker/search/multicity"
	"booker/types"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Flag/config key constants.
const (
	keyDate       = "date"
	keyLeg2Date   = "leg2-date"
	keyPassengers = "passengers"
	keyCabin      = "cabin"
	keyFlexDays   = "flex-days"
	keyMaxStops   = "max-stops"
	keyMaxResults = "max-results"
	keyProfile    = "profile"
	keyCurrency   = "currency"
	keyContext    = "context"
)

// Default values.
const (
	defaultPassengers = 1
	defaultCabin      = string(types.CabinEconomy)
	defaultFlexDays   = 3
	defaultMaxStops   = -1
	defaultMaxResults = 5
	defaultProfile    = "budget"
	defaultCurrency   = "CAD"
	defaultTimeout    = 5 * time.Minute
)

// Date/time formats for output.
const (
	outputDateTimeFmt = "Jan 02 15:04"
)

var searchCmd = &cobra.Command{
	Use:   "search <origin> <destination>",
	Short: "Search for flights (direct or multi-city with stopover)",
	Long:  `Search for flights from origin to destination. Uses an LLM to pick the best strategy (direct or multi-city) based on context.`,
	Args:  cobra.ExactArgs(2),
	RunE:  runSearch,
}

func init() {
	rootCmd.AddCommand(searchCmd)

	f := searchCmd.Flags()
	f.String(keyDate, "", "departure date for leg 1 (YYYY-MM-DD) [required]")
	f.String(keyLeg2Date, "", "departure date for leg 2 (YYYY-MM-DD) [required]")
	f.Int(keyPassengers, defaultPassengers, "number of passengers")
	f.String(keyCabin, defaultCabin, "cabin class (economy, premium_economy, business, first)")
	f.Int(keyFlexDays, defaultFlexDays, "date flexibility ± days")
	f.Int(keyMaxStops, defaultMaxStops, "max layovers per leg (-1 = no limit, 0 = direct)")
	f.Int(keyMaxResults, defaultMaxResults, "number of results to show")
	f.String(keyProfile, defaultProfile, "ranking profile (budget, comfort, balanced)")
	f.String(keyCurrency, defaultCurrency, "display currency (e.g. CAD, USD, EUR)")
	f.String(keyContext, "", "search context/preferences (e.g. 'cheapest option' or 'want to explore a city on the way')")

	_ = searchCmd.MarkFlagRequired(keyDate)

	_ = viper.BindPFlags(f)
}

var profiles = map[string]multicity.RankingWeights{
	"budget":   multicity.WeightsBudget,
	"comfort":  multicity.WeightsComfort,
	"balanced": multicity.WeightsBalanced,
}

func runSearch(cmd *cobra.Command, args []string) error {
	origin := args[0]
	destination := args[1]

	weights, ok := profiles[viper.GetString(keyProfile)]
	if !ok {
		return fmt.Errorf("unknown profile %q, choose: budget, comfort, balanced", viper.GetString(keyProfile))
	}

	cabin := types.CabinClass(viper.GetString(keyCabin))

	// Suppress verbose logs from providers/cache/llm; only print the table.
	log.SetOutput(io.Discard)

	cfg := config.Default()
	httpClient := httpclient.New(cfg.HTTP)

	registry := provider.NewRegistry()
	raw := serpapi.New(cfg.Providers[config.ProviderSerpAPI], httpClient)
	cached := cache.Wrap(raw, ".cache/flights", 0)
	if err := registry.Register(cached); err != nil {
		return fmt.Errorf("registering serpapi: %w", err)
	}

	llmClient := llm.New(cfg.LLM, httpClient)
	ranker := multicity.NewRanker(llmClient, weights)

	// Build common request.
	req := search.Request{
		Origin:        origin,
		Destination:   destination,
		DepartureDate: viper.GetString(keyDate),
		Passengers:    viper.GetInt(keyPassengers),
		CabinClass:    cabin,
		FlexDays:      viper.GetInt(keyFlexDays),
		MaxStops:      viper.GetInt(keyMaxStops),
		MaxResults:    viper.GetInt(keyMaxResults),
		Context:       viper.GetString(keyContext),
	}

	// Create strategies.
	directStrategy := direct.NewSearcher(registry, ranker)
	mcSearcher := multicity.NewSearcher(registry, llmClient, weights)
	mcStrategy := multicity.NewStrategy(mcSearcher, viper.GetString(keyLeg2Date))

	// Pick strategy.
	picker := search.NewPicker(llmClient, directStrategy, mcStrategy)

	ctx, cancel := context.WithTimeout(cmd.Context(), defaultTimeout)
	defer cancel()

	strategy, err := picker.Pick(ctx, req)
	if err != nil {
		return fmt.Errorf("picking strategy: %w", err)
	}

	// Multicity requires leg2-date.
	if strategy.Name() == "multicity" && viper.GetString(keyLeg2Date) == "" {
		return fmt.Errorf("multicity strategy requires --leg2-date flag")
	}

	fmt.Printf("Strategy: %s\n", strategy.Name())

	results, err := strategy.Search(ctx, req)
	if err != nil {
		return fmt.Errorf("search failed: %w", err)
	}

	if len(results) == 0 {
		fmt.Println("No itineraries found.")
		return nil
	}

	printTable(results, viper.GetString(keyCurrency))
	return nil
}

func printTable(itineraries []search.Itinerary, cur string) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.SetStyle(table.StyleRounded)

	t.AppendHeader(table.Row{
		"#", "Score", "Price", "Route",
		"Leg 1 Airlines", "Leg 2 Airlines",
		"Leg 1 Departure", "Leg 2 Departure",
		"Stopover",
	})

	for i, itin := range itineraries {
		converted, _ := currency.Convert(itin.TotalPrice, cur)
		t.AppendRow(table.Row{
			i + 1,
			fmt.Sprintf("%.0f", itin.Score),
			fmt.Sprintf("%s%.0f", currencySymbol(cur), converted.Amount),
			routeString(itin),
			legAirlines(itin, 0),
			legAirlines(itin, 1),
			legDeparture(itin, 0),
			legDeparture(itin, 1),
			stopoverString(itin),
		})
	}

	fmt.Println()
	t.Render()
	fmt.Println()
}

func routeString(itin search.Itinerary) string {
	if len(itin.Legs) == 0 {
		return ""
	}
	route := ""
	for _, leg := range itin.Legs {
		for _, seg := range leg.Flight.Outbound {
			if route != "" {
				route += "→"
			}
			route += seg.Origin
		}
	}
	// Append final destination from last segment.
	lastLeg := itin.Legs[len(itin.Legs)-1].Flight.Outbound
	if len(lastLeg) > 0 {
		route += "→" + lastLeg[len(lastLeg)-1].Destination
	}
	return route
}

func legAirlines(itin search.Itinerary, legIdx int) string {
	if legIdx >= len(itin.Legs) {
		return ""
	}
	seen := map[string]bool{}
	result := ""
	for _, seg := range itin.Legs[legIdx].Flight.Outbound {
		name := seg.AirlineName
		if name == "" {
			name = seg.Airline
		}
		if seen[name] {
			continue
		}
		seen[name] = true
		if result != "" {
			result += ", "
		}
		result += name
	}
	return result
}

func legDeparture(itin search.Itinerary, legIdx int) string {
	if legIdx >= len(itin.Legs) {
		return ""
	}
	segs := itin.Legs[legIdx].Flight.Outbound
	if len(segs) == 0 {
		return ""
	}
	return segs[0].DepartureTime.Format(outputDateTimeFmt)
}

func stopoverString(itin search.Itinerary) string {
	if len(itin.Legs) == 0 || itin.Legs[0].Stopover == nil {
		return ""
	}
	s := itin.Legs[0].Stopover
	return fmt.Sprintf("%s (%s)", s.City, formatDuration(s.Duration))
}

func currencySymbol(cur string) string {
	switch cur {
	case "CAD":
		return "C$"
	case "USD":
		return "$"
	case "EUR":
		return "€"
	case "GBP":
		return "£"
	case "INR":
		return "₹"
	default:
		return cur + " "
	}
}

func formatDuration(d time.Duration) string {
	hours := int(d.Hours())
	mins := int(d.Minutes()) % 60
	switch {
	case hours >= 24:
		return fmt.Sprintf("%dd %dh", hours/24, hours%24)
	case mins == 0:
		return fmt.Sprintf("%dh", hours)
	default:
		return fmt.Sprintf("%dh %dm", hours, mins)
	}
}
