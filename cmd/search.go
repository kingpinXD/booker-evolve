package cmd

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"booker/search"
	"booker/search/multicity"
	"booker/types"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Flag/config key constants.
const (
	keyDate              = "date"
	keyReturnDate        = "return-date"
	keyLeg2Date          = "leg2-date"
	keyPassengers        = "passengers"
	keyCabin             = "cabin"
	keyFlexDays          = "flex-days"
	keyMaxStops          = "max-stops"
	keyMaxPrice          = "max-price"
	keyMaxResults        = "max-results"
	keyProfile           = "profile"
	keyCurrency          = "currency"
	keyContext           = "context"
	keySortBy            = "sort-by"
	keyDepartureAfter    = "departure-after"
	keyDepartureBefore   = "departure-before"
	keyArrivalAfter      = "arrival-after"
	keyArrivalBefore     = "arrival-before"
	keyMaxDuration       = "max-duration"
	keyAvoidAirlines     = "avoid-airlines"
	keyPreferredAirlines = "preferred-airlines"
	keyMinStopover       = "min-stopover"
	keyMaxStopover       = "max-stopover"
	keyFormat            = "format"
	keyVerbose           = "verbose"
)

// Default values.
const (
	defaultPassengers  = 1
	defaultCabin       = string(types.CabinEconomy)
	defaultFlexDays    = 3
	defaultMaxStops    = -1
	defaultMaxResults  = 5
	defaultProfile     = "budget"
	defaultCurrency    = "CAD"
	defaultTimeout     = 5 * time.Minute
	maxHistoryMessages = 20
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
	f.String(keyReturnDate, "", "return date for round-trip (YYYY-MM-DD)")
	f.String(keyLeg2Date, "", "departure date for leg 2 (YYYY-MM-DD) [required]")
	f.Int(keyPassengers, defaultPassengers, "number of passengers")
	f.String(keyCabin, defaultCabin, "cabin class (economy, premium_economy, business, first)")
	f.Int(keyFlexDays, defaultFlexDays, "date flexibility ± days")
	f.Int(keyMaxStops, defaultMaxStops, "max layovers per leg (-1 = no limit, 0 = direct)")
	f.Float64(keyMaxPrice, 0, "max price per flight in USD (0 = no limit)")
	f.Int(keyMaxResults, defaultMaxResults, "number of results to show")
	f.String(keyProfile, defaultProfile, "ranking profile (budget, comfort, balanced)")
	f.String(keyCurrency, defaultCurrency, "display currency (e.g. CAD, USD, EUR)")
	f.String(keyContext, "", "search context/preferences (e.g. 'cheapest option' or 'want to explore a city on the way')")
	f.String(keySortBy, "price", "sort results by: price, duration, departure, or score")
	f.String(keyDepartureAfter, "", "earliest acceptable departure time (HH:MM)")
	f.String(keyDepartureBefore, "", "latest acceptable departure time (HH:MM)")
	f.String(keyArrivalAfter, "", "earliest acceptable arrival time (HH:MM)")
	f.String(keyArrivalBefore, "", "latest acceptable arrival time (HH:MM)")
	f.Duration(keyMaxDuration, 0, "max flight duration (e.g. 12h, 8h30m); 0 = no limit")
	f.String(keyAvoidAirlines, "", "comma-separated IATA codes to exclude (e.g. BA,LH)")
	f.String(keyPreferredAirlines, "", "comma-separated IATA codes to keep (e.g. AC,UA)")
	f.Duration(keyMinStopover, 0, "minimum stopover duration for multi-city (e.g. 24h, 48h); 0 = default")
	f.Duration(keyMaxStopover, 0, "maximum stopover duration for multi-city (e.g. 96h, 168h); 0 = default")
	f.String(keyFormat, "table", "output format: table or json")
	f.BoolP(keyVerbose, "v", false, "show debug output from providers, cache, and LLM")

	_ = searchCmd.MarkFlagRequired(keyDate)

	_ = viper.BindPFlags(f)
}

var profiles = map[string]multicity.RankingWeights{
	"budget":   multicity.WeightsBudget,
	"comfort":  multicity.WeightsComfort,
	"balanced": multicity.WeightsBalanced,
	"eco":      multicity.WeightsEco,
}

func runSearch(cmd *cobra.Command, args []string) error {
	origin := args[0]
	destination := args[1]

	weights, ok := profiles[viper.GetString(keyProfile)]
	if !ok {
		return fmt.Errorf("unknown profile %q, choose: budget, comfort, balanced, eco", viper.GetString(keyProfile))
	}

	cabin := types.CabinClass(viper.GetString(keyCabin))

	// Suppress logs unless --verbose is set.
	if !viper.GetBool(keyVerbose) {
		log.SetOutput(io.Discard)
	}

	picker, _, rawProvider, _, err := buildPicker(weights, viper.GetString(keyLeg2Date))
	if err != nil {
		return err
	}

	// Build common request.
	req := search.Request{
		Origin:            origin,
		Destination:       destination,
		DepartureDate:     viper.GetString(keyDate),
		ReturnDate:        viper.GetString(keyReturnDate),
		Passengers:        viper.GetInt(keyPassengers),
		CabinClass:        cabin,
		FlexDays:          viper.GetInt(keyFlexDays),
		MaxStops:          viper.GetInt(keyMaxStops),
		MaxPrice:          viper.GetFloat64(keyMaxPrice),
		DepartureAfter:    viper.GetString(keyDepartureAfter),
		DepartureBefore:   viper.GetString(keyDepartureBefore),
		ArrivalAfter:      viper.GetString(keyArrivalAfter),
		ArrivalBefore:     viper.GetString(keyArrivalBefore),
		MaxDuration:       viper.GetDuration(keyMaxDuration),
		SortBy:            viper.GetString(keySortBy),
		AvoidAirlines:     viper.GetString(keyAvoidAirlines),
		PreferredAirlines: viper.GetString(keyPreferredAirlines),
		MinStopover:       viper.GetDuration(keyMinStopover),
		MaxStopover:       viper.GetDuration(keyMaxStopover),
		MaxResults:        viper.GetInt(keyMaxResults),
		Context:           viper.GetString(keyContext),
	}

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

	cur := viper.GetString(keyCurrency)
	insights := rawProvider.LastPriceInsights()
	switch viper.GetString(keyFormat) {
	case "json":
		return printJSONWithInsights(os.Stdout, results, cur, insights)
	default:
		printTable(os.Stdout, results, cur)
		if s := formatPriceInsights(insights); s != "" {
			fmt.Println(s)
		}
		return nil
	}
}
