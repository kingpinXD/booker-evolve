package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"booker/currency"
	"booker/search"
	"booker/search/multicity"
	"booker/types"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Flag/config key constants.
const (
	keyDate          = "date"
	keyReturnDate    = "return-date"
	keyLeg2Date      = "leg2-date"
	keyPassengers    = "passengers"
	keyCabin         = "cabin"
	keyFlexDays      = "flex-days"
	keyMaxStops      = "max-stops"
	keyMaxPrice      = "max-price"
	keyMaxResults    = "max-results"
	keyProfile       = "profile"
	keyCurrency      = "currency"
	keyContext       = "context"
	keySortBy        = "sort-by"
	keyArrivalAfter  = "arrival-after"
	keyArrivalBefore = "arrival-before"
	keyMaxDuration   = "max-duration"
	keyAvoidAirlines = "avoid-airlines"
	keyFormat        = "format"
	keyVerbose       = "verbose"
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
	f.String(keySortBy, "price", "sort results by: price, duration, or departure")
	f.String(keyArrivalAfter, "", "earliest acceptable arrival time (HH:MM)")
	f.String(keyArrivalBefore, "", "latest acceptable arrival time (HH:MM)")
	f.Duration(keyMaxDuration, 0, "max flight duration (e.g. 12h, 8h30m); 0 = no limit")
	f.String(keyAvoidAirlines, "", "comma-separated IATA codes to exclude (e.g. BA,LH)")
	f.String(keyFormat, "table", "output format: table or json")
	f.BoolP(keyVerbose, "v", false, "show debug output from providers, cache, and LLM")

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

	// Suppress logs unless --verbose is set.
	if !viper.GetBool(keyVerbose) {
		log.SetOutput(io.Discard)
	}

	picker, _, rawProvider, err := buildPicker(weights, viper.GetString(keyLeg2Date))
	if err != nil {
		return err
	}

	// Build common request.
	req := search.Request{
		Origin:        origin,
		Destination:   destination,
		DepartureDate: viper.GetString(keyDate),
		ReturnDate:    viper.GetString(keyReturnDate),
		Passengers:    viper.GetInt(keyPassengers),
		CabinClass:    cabin,
		FlexDays:      viper.GetInt(keyFlexDays),
		MaxStops:      viper.GetInt(keyMaxStops),
		MaxPrice:      viper.GetFloat64(keyMaxPrice),
		ArrivalAfter:  viper.GetString(keyArrivalAfter),
		ArrivalBefore: viper.GetString(keyArrivalBefore),
		MaxDuration:   viper.GetDuration(keyMaxDuration),
		SortBy:        viper.GetString(keySortBy),
		AvoidAirlines: viper.GetString(keyAvoidAirlines),
		MaxResults:    viper.GetInt(keyMaxResults),
		Context:       viper.GetString(keyContext),
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

// isMultiLeg returns true if any itinerary has more than one leg.
func isMultiLeg(itineraries []search.Itinerary) bool {
	for _, itin := range itineraries {
		if len(itin.Legs) > 1 {
			return true
		}
	}
	return false
}

// hasScores returns true if any itinerary has a non-zero score.
func hasScores(itineraries []search.Itinerary) bool {
	for _, itin := range itineraries {
		if itin.Score != 0 {
			return true
		}
	}
	return false
}

func printTable(w io.Writer, itineraries []search.Itinerary, cur string) {
	t := table.NewWriter()
	t.SetOutputMirror(w)
	t.SetStyle(table.StyleRounded)

	multiLeg := isMultiLeg(itineraries)
	scored := hasScores(itineraries)

	// Build header dynamically based on layout and whether scores exist.
	var header table.Row
	if multiLeg {
		header = table.Row{"#"}
		if scored {
			header = append(header, "Score")
		}
		header = append(header, "Price", "Route",
			"Leg 1 Airlines", "Leg 2 Airlines", "Cabin",
			"Leg 1 Departure", "Leg 1 Arrival",
			"Leg 2 Departure", "Leg 2 Arrival",
			"Stopover", "Stops", "Duration", "Leg 1 CO2", "Leg 2 CO2")
		if scored {
			header = append(header, "Reason")
		}
		header = append(header, "Book")
	} else {
		header = table.Row{"#"}
		if scored {
			header = append(header, "Score")
		}
		header = append(header, "Price", "Route",
			"Airlines", "Cabin", "Departure", "Arrival", "Stops", "Duration", "CO2")
		if scored {
			header = append(header, "Reason")
		}
		header = append(header, "Book")
	}
	t.AppendHeader(header)

	for i, itin := range itineraries {
		converted, _ := currency.Convert(itin.TotalPrice, cur)
		dur := formatDuration(itin.TotalTravel)
		stops := formatStops(itin)

		var row table.Row
		if multiLeg {
			row = table.Row{i + 1}
			if scored {
				row = append(row, fmt.Sprintf("%.0f", itin.Score))
			}
			row = append(row,
				fmt.Sprintf("%s%.0f", currencySymbol(cur), converted.Amount),
				routeString(itin),
				legAirlines(itin, 0),
				legAirlines(itin, 1),
				legCabin(itin, 0),
				legDeparture(itin, 0),
				legArrival(itin, 0),
				legDeparture(itin, 1),
				legArrival(itin, 1),
				stopoverString(itin),
				stops,
				dur,
				legCarbon(itin, 0),
				legCarbon(itin, 1))
			if scored {
				row = append(row, itin.Reasoning)
			}
			row = append(row, legBookingURL(itin, 0))
		} else {
			row = table.Row{i + 1}
			if scored {
				row = append(row, fmt.Sprintf("%.0f", itin.Score))
			}
			row = append(row,
				fmt.Sprintf("%s%.0f", currencySymbol(cur), converted.Amount),
				routeString(itin),
				legAirlines(itin, 0),
				legCabin(itin, 0),
				legDeparture(itin, 0),
				legArrival(itin, 0),
				stops,
				dur,
				legCarbon(itin, 0))
			if scored {
				row = append(row, itin.Reasoning)
			}
			row = append(row, legBookingURL(itin, 0))
		}
		t.AppendRow(row)
	}

	_, _ = fmt.Fprintln(w)
	t.Render()
	if s := priceSummary(itineraries, cur); s != "" {
		_, _ = fmt.Fprintln(w, s)
	}
	_, _ = fmt.Fprintln(w)
}

// priceSummary returns a one-line summary of price range and result count.
func priceSummary(itineraries []search.Itinerary, cur string) string {
	if len(itineraries) == 0 {
		return ""
	}
	sym := currencySymbol(cur)
	minPrice, maxPrice := itineraries[0].TotalPrice, itineraries[0].TotalPrice
	for _, itin := range itineraries[1:] {
		if itin.TotalPrice.Amount < minPrice.Amount {
			minPrice = itin.TotalPrice
		}
		if itin.TotalPrice.Amount > maxPrice.Amount {
			maxPrice = itin.TotalPrice
		}
	}
	minC, _ := currency.Convert(minPrice, cur)
	maxC, _ := currency.Convert(maxPrice, cur)

	noun := "results"
	if len(itineraries) == 1 {
		noun = "result"
	}
	if minC.Amount == maxC.Amount {
		return fmt.Sprintf("%d %s | %s%.0f", len(itineraries), noun, sym, minC.Amount)
	}
	return fmt.Sprintf("%d %s | %s%.0f - %s%.0f", len(itineraries), noun, sym, minC.Amount, sym, maxC.Amount)
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

func legCabin(itin search.Itinerary, legIdx int) string {
	if legIdx >= len(itin.Legs) {
		return ""
	}
	segs := itin.Legs[legIdx].Flight.Outbound
	if len(segs) == 0 {
		return ""
	}
	return string(segs[0].CabinClass)
}

// legAircraft returns the aircraft type from the first segment of the given leg.
func legAircraft(itin search.Itinerary, legIdx int) string {
	if legIdx >= len(itin.Legs) {
		return ""
	}
	segs := itin.Legs[legIdx].Flight.Outbound
	if len(segs) == 0 {
		return ""
	}
	return segs[0].Aircraft
}

// legLegroom returns the legroom from the first segment of the given leg.
func legLegroom(itin search.Itinerary, legIdx int) string {
	if legIdx >= len(itin.Legs) {
		return ""
	}
	segs := itin.Legs[legIdx].Flight.Outbound
	if len(segs) == 0 {
		return ""
	}
	return segs[0].Legroom
}

// legSeatsLeft returns the minimum SeatsLeft across all segments of the given leg.
// Returns 0 if no segment has seat data.
func legSeatsLeft(itin search.Itinerary, legIdx int) int {
	if legIdx >= len(itin.Legs) {
		return 0
	}
	minSeats := 0
	for _, seg := range itin.Legs[legIdx].Flight.Outbound {
		if seg.SeatsLeft > 0 {
			if minSeats == 0 || seg.SeatsLeft < minSeats {
				minSeats = seg.SeatsLeft
			}
		}
	}
	return minSeats
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

// isNextDay returns true if the arrival date is after the departure date.
func isNextDay(dep, arr time.Time) bool {
	return arr.YearDay() != dep.YearDay() || arr.Year() != dep.Year()
}

func legArrival(itin search.Itinerary, legIdx int) string {
	if legIdx >= len(itin.Legs) {
		return ""
	}
	segs := itin.Legs[legIdx].Flight.Outbound
	if len(segs) == 0 {
		return ""
	}
	dep := segs[0].DepartureTime
	arr := segs[len(segs)-1].ArrivalTime
	s := arr.Format(outputDateTimeFmt)
	if isNextDay(dep, arr) {
		days := (arr.YearDay() - dep.YearDay())
		if arr.Year() != dep.Year() {
			days += 365 // approximate; good enough for display
		}
		s += fmt.Sprintf(" (+%d)", days)
	}
	return s
}

func legCarbon(itin search.Itinerary, legIdx int) string {
	if legIdx >= len(itin.Legs) || itin.Legs[legIdx].Flight.CarbonKg == 0 {
		return ""
	}
	return fmt.Sprintf("%dkg", itin.Legs[legIdx].Flight.CarbonKg)
}

func legBookingURL(itin search.Itinerary, legIdx int) string {
	if legIdx >= len(itin.Legs) {
		return ""
	}
	return itin.Legs[legIdx].Flight.BookingURL
}

func stopoverString(itin search.Itinerary) string {
	if len(itin.Legs) == 0 || itin.Legs[0].Stopover == nil {
		return ""
	}
	s := itin.Legs[0].Stopover
	return fmt.Sprintf("%s (%s)", s.City, formatDuration(s.Duration))
}

// itineraryStops returns the total number of connections across all legs.
func itineraryStops(itin search.Itinerary) int {
	n := 0
	for _, leg := range itin.Legs {
		n += leg.Flight.Stops()
	}
	return n
}

// formatStops returns a display string for total stops across all legs.
// If stops > 0 and layover data is available, it includes total layover time
// (e.g. "1 (2h 30m)"). Otherwise returns just the count (e.g. "0" or "1").
func formatStops(itin search.Itinerary) string {
	stops := itineraryStops(itin)
	if stops == 0 {
		return "0"
	}
	var totalLayover time.Duration
	for _, leg := range itin.Legs {
		for _, seg := range leg.Flight.Outbound {
			totalLayover += seg.LayoverDuration
		}
	}
	if totalLayover == 0 {
		return fmt.Sprintf("%d", stops)
	}
	return fmt.Sprintf("%d (%s)", stops, formatDuration(totalLayover))
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

// jsonLeg is the JSON representation of a single flight leg.
type jsonLeg struct {
	Airlines        string `json:"airlines"`
	AirlineCode     string `json:"airline_code,omitempty"`
	CabinClass      string `json:"cabin_class,omitempty"`
	Origin          string `json:"origin"`
	OriginCity      string `json:"origin_city,omitempty"`
	OriginName      string `json:"origin_name,omitempty"`
	Dest            string `json:"destination"`
	DestinationCity string `json:"destination_city,omitempty"`
	DestinationName string `json:"destination_name,omitempty"`
	Departure       string `json:"departure"`
	Arrival         string `json:"arrival,omitempty"`
	Duration        string `json:"duration"`
	Stops           int    `json:"stops"`
	CarbonKg        int    `json:"carbon_kg,omitempty"`
	TypicalCarbonKg int    `json:"typical_carbon_kg,omitempty"`
	CarbonDiffPct   int    `json:"carbon_diff_percent,omitempty"`
	BookingURL      string `json:"booking_url,omitempty"`
	Aircraft        string `json:"aircraft,omitempty"`
	Legroom         string `json:"legroom,omitempty"`
	SeatsLeft       int    `json:"seats_left,omitempty"`
	ArrivalNextDay  bool   `json:"arrival_next_day,omitempty"`
}

// jsonItinerary is the JSON representation of a search result.
type jsonItinerary struct {
	Rank      int       `json:"rank"`
	Score     float64   `json:"score,omitempty"`
	Reasoning string    `json:"reasoning,omitempty"`
	Price     float64   `json:"price"`
	Currency  string    `json:"currency"`
	Route     string    `json:"route"`
	Duration  string    `json:"duration"`
	Legs      []jsonLeg `json:"legs"`
	Stopover  string    `json:"stopover,omitempty"`
}

// formatPriceInsights returns a one-line summary of price insights, or empty
// if no meaningful data is available.
func formatPriceInsights(pi search.PriceInsights) string {
	if pi.PriceLevel == "" {
		return ""
	}
	low, high := pi.TypicalPriceRange[0], pi.TypicalPriceRange[1]
	if low == 0 && high == 0 {
		return fmt.Sprintf("Price level: %s", pi.PriceLevel)
	}
	return fmt.Sprintf("Price level: %s | Typical: $%.0f - $%.0f", pi.PriceLevel, low, high)
}

// jsonPriceInsights is the JSON representation of price insights.
type jsonPriceInsights struct {
	PriceLevel        string     `json:"price_level,omitempty"`
	LowestPrice       float64    `json:"lowest_price,omitempty"`
	TypicalPriceRange [2]float64 `json:"typical_price_range,omitempty"`
}

func printJSONWithInsights(w io.Writer, itineraries []search.Itinerary, cur string, pi search.PriceInsights) error {
	type jsonOutput struct {
		Results       []jsonItinerary    `json:"results"`
		PriceInsights *jsonPriceInsights `json:"price_insights,omitempty"`
	}

	results := buildJSONItineraries(itineraries, cur)
	out := jsonOutput{Results: results}
	if pi.PriceLevel != "" {
		out.PriceInsights = &jsonPriceInsights{
			PriceLevel:        pi.PriceLevel,
			LowestPrice:       pi.LowestPrice,
			TypicalPriceRange: pi.TypicalPriceRange,
		}
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(out)
}

func buildJSONItineraries(itineraries []search.Itinerary, cur string) []jsonItinerary {
	out := make([]jsonItinerary, len(itineraries))
	for i, itin := range itineraries {
		converted, _ := currency.Convert(itin.TotalPrice, cur)

		var legs []jsonLeg
		for idx, leg := range itin.Legs {
			segs := leg.Flight.Outbound
			origin, dest, dep, arr := "", "", "", ""
			airlineCode, originCity, destCity, originName, destName := "", "", "", "", ""
			if len(segs) > 0 {
				origin = segs[0].Origin
				dest = segs[len(segs)-1].Destination
				dep = segs[0].DepartureTime.Format(time.RFC3339)
				arr = segs[len(segs)-1].ArrivalTime.Format(time.RFC3339)
				airlineCode = segs[0].Airline
				originCity = segs[0].OriginCity
				originName = segs[0].OriginName
				destCity = segs[len(segs)-1].DestinationCity
				destName = segs[len(segs)-1].DestinationName
			}
			nextDay := false
			if len(segs) > 0 {
				nextDay = isNextDay(segs[0].DepartureTime, segs[len(segs)-1].ArrivalTime)
			}
			legs = append(legs, jsonLeg{
				Airlines:        legAirlines(itin, idx),
				AirlineCode:     airlineCode,
				CabinClass:      legCabin(itin, idx),
				Origin:          origin,
				OriginCity:      originCity,
				OriginName:      originName,
				Dest:            dest,
				DestinationCity: destCity,
				DestinationName: destName,
				Departure:       dep,
				Arrival:         arr,
				Duration:        formatDuration(leg.Flight.TotalDuration),
				Stops:           leg.Flight.Stops(),
				CarbonKg:        leg.Flight.CarbonKg,
				TypicalCarbonKg: leg.Flight.TypicalCarbonKg,
				CarbonDiffPct:   leg.Flight.CarbonDiffPct,
				BookingURL:      leg.Flight.BookingURL,
				Aircraft:        legAircraft(itin, idx),
				Legroom:         legLegroom(itin, idx),
				SeatsLeft:       legSeatsLeft(itin, idx),
				ArrivalNextDay:  nextDay,
			})
		}

		out[i] = jsonItinerary{
			Rank:      i + 1,
			Score:     itin.Score,
			Reasoning: itin.Reasoning,
			Price:     converted.Amount,
			Currency:  cur,
			Route:     routeString(itin),
			Duration:  formatDuration(itin.TotalTravel),
			Legs:      legs,
			Stopover:  stopoverString(itin),
		}
	}
	return out
}

func printJSON(w io.Writer, itineraries []search.Itinerary, cur string) error {
	out := buildJSONItineraries(itineraries, cur)
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(out)
}
