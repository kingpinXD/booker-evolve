package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"strings"
	"time"

	"booker/llm"
	"booker/search"
	"booker/search/multicity"
	"booker/types"
)

// chatSystemPrompt returns the system prompt for the chat conversation.
// The provided time is injected so the LLM knows "today" for relative dates.
func chatSystemPrompt(now time.Time) string {
	return fmt.Sprintf("Today's date is %s.\n\n", now.Format("2006-01-02")) +
		`You are a proactive travel planning agent, not a search form. Your goal is to find the best flights for the traveler by understanding their trip and making smart recommendations.

Gather the following through natural conversation:

Required:
- origin: departure airport IATA code (e.g. "DEL", "JFK")
- destination: arrival airport IATA code (e.g. "YYZ", "LHR")
- departure_date: in YYYY-MM-DD format

Optional:
- return_date: in YYYY-MM-DD format (for round trips)
- leg2_date: in YYYY-MM-DD format (for multi-city trips — when the traveler leaves the stopover city)
- passengers: number of travelers (default: 1)
- cabin: economy, premium_economy, business, or first (default: economy)
- max_price: maximum budget per flight in USD (e.g. 1200)
- direct_only: true to show only non-stop flights
- flex_days: search ± N days around departure date (default: 3)
- profile: ranking profile — "budget" (cheapest), "comfort" (best schedule/airline), "balanced", or "eco" (lowest carbon emissions) (default: budget)
- preferred_alliance: "Star Alliance", "OneWorld", or "SkyTeam" — filter to this alliance only
- departure_after: earliest acceptable departure time (HH:MM, e.g. "06:00")
- departure_before: latest acceptable departure time (HH:MM, e.g. "22:00")
- arrival_after: earliest acceptable arrival time (HH:MM, e.g. "08:00")
- arrival_before: latest acceptable arrival time (HH:MM, e.g. "18:00")
- max_duration_hours: maximum flight duration in hours (e.g. 12)
- sort_by: sort results by "price" (default), "duration", "departure", or "score" (by ranker score, highest first)
- avoid_airlines: comma-separated IATA codes to exclude (e.g. "BA,LH")
- preferred_airlines: comma-separated IATA codes to keep only (e.g. "AC,UA")
- min_stopover_hours: minimum city stopover duration in hours for multi-city (default: 48)
- max_stopover_hours: maximum city stopover duration in hours for multi-city (default: 144)
- context: any preferences like "cheapest option" or "prefer direct flights"
- clear_fields: array of field names to reset (e.g. ["direct_only", "max_price"]) — use when the user wants to remove a previously set filter

Be a helpful travel advisor:
- When the user is flexible on routing, suggest stopover cities that could save money (e.g. "Flying Delhi to Toronto via Bangkok often saves $300-400 and you get a 2-day city break").
- If the user mentions a city served by multiple airports, suggest nearby alternatives (e.g., for New York: JFK, EWR, LGA). Searching nearby airports can reveal cheaper fares.
- Explain tradeoffs briefly when relevant (e.g. "A 3-hour longer layover saves $200" or "Business class on this route is only $400 more than premium economy").
- Ask about flexibility on dates and routing — small changes often unlock much better prices.

When you have at least the origin, destination, and departure_date, output the parameters as a JSON object on its own line. You may include optional fields if the user mentioned them. Example:
{"origin":"DEL","destination":"YYZ","departure_date":"2025-06-15","passengers":2,"cabin":"economy","context":"budget trip"}

After outputting the JSON, briefly explain what you're searching for and why you chose that approach.

Multi-city flow: When you suggest a stopover city and the user shows interest (e.g. "yes", "sure", "sounds good", "how do I do that"), ask them what date they want to leave the stopover city, then emit a JSON with the original route plus leg2_date set to that date. Do not ask the user to type raw parameter names — guide them conversationally.`
}

// nearbyAirportHint returns a message mentioning nearby airports for the
// origin and/or destination, or empty string if neither has alternatives.
func nearbyAirportHint(origin, destination string) string {
	originNearby := search.NearbyAirports(origin)
	destNearby := search.NearbyAirports(destination)
	if len(originNearby) == 0 && len(destNearby) == 0 {
		return ""
	}
	var parts []string
	if len(originNearby) > 0 {
		parts = append(parts, fmt.Sprintf("Nearby origin airports: %s", strings.Join(originNearby, ", ")))
	}
	if len(destNearby) > 0 {
		parts = append(parts, fmt.Sprintf("Nearby destination airports: %s", strings.Join(destNearby, ", ")))
	}
	return strings.Join(parts, ". ") + "."
}

// parseTripParams extracts trip parameters from an LLM response.
// Returns the params and true if a valid JSON block with required fields was found.
func parseTripParams(response string) (tripParams, bool) {
	// Try each line for a JSON object with required fields.
	for _, line := range strings.Split(response, "\n") {
		line = llm.StripCodeFences(line)

		if !strings.HasPrefix(line, "{") {
			continue
		}

		var p tripParams
		if err := json.Unmarshal([]byte(line), &p); err != nil {
			continue
		}
		if p.Origin == "" || p.Destination == "" || p.DepartureDate == "" {
			continue
		}
		return p, true
	}
	return tripParams{}, false
}

// buildRequestFromParams converts extracted trip params to a search.Request,
// applying defaults for unset fields.
func buildRequestFromParams(p tripParams) search.Request {
	passengers := p.Passengers
	if passengers == 0 {
		passengers = defaultPassengers
	}
	cabin := p.Cabin
	if cabin == "" {
		cabin = defaultCabin
	}
	maxStops := defaultMaxStops
	if p.DirectOnly {
		maxStops = 0
	}
	flexDays := p.FlexDays
	if flexDays == 0 {
		flexDays = defaultFlexDays
	}
	return search.Request{
		Origin:            p.Origin,
		Destination:       p.Destination,
		DepartureDate:     p.DepartureDate,
		ReturnDate:        p.ReturnDate,
		Leg2Date:          p.Leg2Date,
		Passengers:        passengers,
		CabinClass:        types.CabinClass(cabin),
		FlexDays:          flexDays,
		MaxStops:          maxStops,
		MaxPrice:          p.MaxPrice,
		PreferredAlliance: p.PreferredAlliance,
		DepartureAfter:    p.DepartureAfter,
		DepartureBefore:   p.DepartureBefore,
		ArrivalAfter:      p.ArrivalAfter,
		ArrivalBefore:     p.ArrivalBefore,
		MaxDuration:       time.Duration(p.MaxDurationHours) * time.Hour,
		SortBy:            p.SortBy,
		AvoidAirlines:     p.AvoidAirlines,
		PreferredAirlines: p.PreferredAirlines,
		MinStopover:       time.Duration(p.MinStopoverHours) * time.Hour,
		MaxStopover:       time.Duration(p.MaxStopoverHours) * time.Hour,
		MaxResults:        defaultMaxResults,
		Context:           p.Context,
	}
}

// searchTimeout is the per-search context timeout for chatSearch.
const searchTimeout = 2 * time.Minute

// chatSearch builds a request from params, picks a strategy, and executes the search.
// Status messages and tips are written to out during execution.
// A per-search timeout prevents individual searches from hanging the chat session.
func chatSearch(ctx context.Context, params tripParams, picker *search.Picker, out io.Writer) ([]search.Itinerary, error) {
	ctx, cancel := context.WithTimeout(ctx, searchTimeout)
	defer cancel()

	req := buildRequestFromParams(params)
	_, _ = fmt.Fprintln(out, formatSearchParams(params)+"...")
	if hint := nearbyAirportHint(req.Origin, req.Destination); hint != "" {
		_, _ = fmt.Fprintf(out, "Tip: %s\n", hint)
	}

	strategy, reason, err := picker.Pick(ctx, req)
	if err != nil {
		return nil, wrapTimeoutError(err)
	}
	_, _ = fmt.Fprintf(out, "Using %s strategy (%s)\n", strategy.Name(), reason)

	// Show stopover cities for multicity searches so user knows what's happening.
	if strategy.Name() == "multicity" {
		if msg := stopoverProgressMessage(req.Origin, req.Destination); msg != "" {
			_, _ = fmt.Fprintln(out, msg)
		}
	}

	results, err := strategy.Search(ctx, req)
	if err != nil {
		return nil, wrapTimeoutError(err)
	}
	return results, nil
}

// stopoverProgressMessage returns a progress line listing stopover cities for
// the given route. Returns empty string when no stopovers are known.
func stopoverProgressMessage(origin, dest string) string {
	stopovers := multicity.StopoversForRoute(origin, dest)
	if len(stopovers) == 0 {
		return ""
	}
	var cities []string
	for _, s := range stopovers {
		cities = append(cities, s.City)
	}
	n := len(cities)
	preview := cities
	if n > 4 {
		preview = cities[:4]
	}
	suffix := ""
	if n > 4 {
		suffix = ", ..."
	}
	return fmt.Sprintf("Searching via %d stopover cities (%s%s)...", n, strings.Join(preview, ", "), suffix)
}

// wrapTimeoutError returns a user-friendly message for context deadline/cancellation errors.
func wrapTimeoutError(err error) error {
	if isContextError(err) {
		return fmt.Errorf("search timed out — try a more specific route or add filters to narrow results")
	}
	return err
}

// isContextError checks if an error is a context timeout or cancellation.
func isContextError(err error) bool {
	return err == context.DeadlineExceeded || err == context.Canceled
}

// resultSummaryForChat builds a summary of search results for the conversation
// history, including top 3 results with price, airline, duration, and stops so
// the LLM can explain recommendations and answer comparison questions.
// When pi has a non-empty PriceLevel, typical price range info is appended
// so the LLM can reference whether current prices are good or bad.
func resultSummaryForChat(results []search.Itinerary, params tripParams, pi search.PriceInsights) string {
	if len(results) == 0 {
		return "No results found."
	}

	minPrice, maxPrice := results[0].TotalPrice.Amount, results[0].TotalPrice.Amount
	for _, r := range results[1:] {
		if r.TotalPrice.Amount < minPrice {
			minPrice = r.TotalPrice.Amount
		}
		if r.TotalPrice.Amount > maxPrice {
			maxPrice = r.TotalPrice.Amount
		}
	}

	var b strings.Builder
	word := "results"
	if len(results) == 1 {
		word = "result"
	}
	fmt.Fprintf(&b, "I found %d %s. Prices range from $%.0f to $%.0f USD.", len(results), word, minPrice, maxPrice)

	// Top N result details (up to 3).
	n := len(results)
	if n > 3 {
		n = 3
	}
	for i := 0; i < n; i++ {
		r := results[i]
		if len(r.Legs) == 0 || len(r.Legs[0].Flight.Outbound) == 0 {
			continue
		}
		seg := r.Legs[0].Flight.Outbound[0]
		airline := seg.AirlineName
		if airline == "" {
			airline = seg.Airline
		}
		layoverInfo := formatLayoverSummary(r.Legs[0].Flight.Outbound)
		dateStr := ""
		if params.FlexDays > 0 && !seg.DepartureTime.IsZero() {
			dateStr = seg.DepartureTime.Format("Jan 2") + ", "
		}
		fmt.Fprintf(&b, " %d) %s, %s%s, %s, $%.0f.",
			i+1, airline, dateStr, formatFlightDuration(r.Legs[0].Flight.TotalDuration), layoverInfo, r.TotalPrice.Amount)
		if r.Reasoning != "" {
			fmt.Fprintf(&b, " Reason: %s.", r.Reasoning)
		}
	}

	// Search params context.
	fmt.Fprintf(&b, " Search: %s->%s on %s", params.Origin, params.Destination, params.DepartureDate)
	if params.Cabin != "" {
		fmt.Fprintf(&b, ", %s", params.Cabin)
	}
	if params.MaxPrice > 0 {
		fmt.Fprintf(&b, ", max $%.0f", params.MaxPrice)
	}
	b.WriteString(".")
	if pi.PriceLevel != "" {
		fmt.Fprintf(&b, " Typical prices for this route: $%.0f-$%.0f (price level: %s).",
			pi.TypicalPriceRange[0], pi.TypicalPriceRange[1], pi.PriceLevel)
	}
	// Include fare trend when flex-date search produced multi-date results.
	if params.FlexDays > 0 {
		ft := search.ComputeFareTrend(results)
		if ft.CheapestDate != "" && ft.CheapestDate != ft.PriciestDate {
			fmt.Fprintf(&b, " Fare trend: %s is cheapest ($%.0f), %s most expensive ($%.0f).",
				ft.CheapestDate, ft.MinPrice, ft.PriciestDate, ft.MaxPrice)
		}
	}
	return b.String()
}

// formatFlightDuration formats a duration as e.g. "14h30m".
func formatFlightDuration(d time.Duration) string {
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	if m == 0 {
		return fmt.Sprintf("%dh", h)
	}
	return fmt.Sprintf("%dh%dm", h, m)
}

// formatLayoverSummary returns a human-readable stop description from segments.
// "nonstop" for 1 segment, "1 stop (3h IST)" when layover data is available,
// or "N stops" as fallback when LayoverDuration is zero.
func formatLayoverSummary(segs []types.Segment) string {
	stops := len(segs) - 1
	if stops <= 0 {
		return "nonstop"
	}

	// Check if all intermediate segments have layover data.
	var layovers []string
	for i := 0; i < stops; i++ {
		if segs[i].LayoverDuration == 0 {
			// Missing data — fall back to count only.
			if stops == 1 {
				return "1 stop"
			}
			return fmt.Sprintf("%d stops", stops)
		}
		layovers = append(layovers, fmt.Sprintf("%s %s",
			formatFlightDuration(segs[i].LayoverDuration), segs[i].Destination))
	}

	word := "stop"
	if stops > 1 {
		word = "stops"
	}
	return fmt.Sprintf("%d %s (%s)", stops, word, strings.Join(layovers, ", "))
}

// jsonFieldName extracts the JSON field name from a struct field's tag.
// Returns empty string if no json tag is present.
func jsonFieldName(f reflect.StructField) string {
	tag := f.Tag.Get("json")
	if tag == "" || tag == "-" {
		return ""
	}
	name, _, _ := strings.Cut(tag, ",")
	return name
}

// mergeParams fills zero-value fields in partial from prev, producing
// a complete set of params for a follow-up search. Fields listed in
// partial.ClearFields are zeroed before merge so sticky filters can be reset.
//
// Uses reflection to iterate struct fields, so new tripParams fields
// are automatically supported without modifying this function.
func mergeParams(prev, partial tripParams) tripParams {
	// Phase 1: Apply clear_fields — zero specified fields on prev.
	if len(partial.ClearFields) > 0 {
		cleared := make(map[string]bool, len(partial.ClearFields))
		for _, f := range partial.ClearFields {
			cleared[f] = true
		}
		prevV := reflect.ValueOf(&prev).Elem()
		for i := 0; i < prevV.NumField(); i++ {
			if name := jsonFieldName(prevV.Type().Field(i)); cleared[name] {
				prevV.Field(i).Set(reflect.Zero(prevV.Field(i).Type()))
			}
		}
	}

	// Phase 2: Merge — fill zero-value fields in partial from prev.
	// ClearFields is ephemeral and never carried over.
	merged := partial
	mergedV := reflect.ValueOf(&merged).Elem()
	prevV := reflect.ValueOf(&prev).Elem()
	for i := 0; i < mergedV.NumField(); i++ {
		name := jsonFieldName(mergedV.Type().Field(i))
		if name == "" || name == "clear_fields" {
			continue
		}
		if mergedV.Field(i).IsZero() {
			mergedV.Field(i).Set(prevV.Field(i))
		}
	}
	return merged
}

// anyFieldSet returns true if any field in a tripParams struct is non-zero.
// Uses reflection so new fields are automatically supported.
func anyFieldSet(p tripParams) bool {
	v := reflect.ValueOf(p)
	for i := 0; i < v.NumField(); i++ {
		if !v.Field(i).IsZero() {
			return true
		}
	}
	return false
}

// parsePartialParams extracts trip parameters from an LLM response,
// accepting partial JSON (at least one recognized field set). Used for
// follow-up refinements where the LLM only emits changed fields.
func parsePartialParams(response string) (tripParams, bool) {
	for _, line := range strings.Split(response, "\n") {
		line = llm.StripCodeFences(line)

		if !strings.HasPrefix(line, "{") {
			continue
		}

		var p tripParams
		if err := json.Unmarshal([]byte(line), &p); err != nil {
			continue
		}
		if anyFieldSet(p) {
			return p, true
		}
	}
	return tripParams{}, false
}

// inferProfile scans recent user messages for preference keywords and returns
// a ranking profile name. Returns empty string when no clear signal is detected.
func inferProfile(history []llm.Message) string {
	var budget, comfort, eco int
	for _, msg := range history {
		if msg.Role != llm.RoleUser {
			continue
		}
		lower := strings.ToLower(msg.Content)
		for _, kw := range []string{"cheapest", "budget", "save money", "lowest price", "low cost"} {
			if strings.Contains(lower, kw) {
				budget++
			}
		}
		for _, kw := range []string{"comfortable", "comfort", "hate layover", "short layover", "business class", "first class"} {
			if strings.Contains(lower, kw) {
				comfort++
			}
		}
		for _, kw := range []string{"eco", "green", "carbon", "environment", "emission"} {
			if strings.Contains(lower, kw) {
				eco++
			}
		}
	}
	switch {
	case budget > 0 && budget >= comfort && budget >= eco:
		return "budget"
	case comfort > 0 && comfort >= eco:
		return "comfort"
	case eco > 0:
		return "eco"
	default:
		return ""
	}
}

// contextWeights scans user messages for specific preference signals and returns
// additive weight deltas. These are applied on top of the base profile weights.
// Returns zero weights when no signals are detected.
func contextWeights(history []llm.Message) multicity.RankingWeights {
	type signal struct {
		keywords []string
		field    string
	}
	signals := []signal{
		{keywords: []string{"hate layover", "short layover", "long layover", "connection time"}, field: "layover"},
		{keywords: []string{"carbon", "environment", "emission", "eco", "green"}, field: "carbon"},
		{keywords: []string{"schedule matters", "departure time", "arrival time", "timing"}, field: "schedule"},
		{keywords: []string{"short flight", "long flight", "flight duration", "travel time"}, field: "duration"},
	}

	hits := make(map[string]bool)
	for _, msg := range history {
		if msg.Role != llm.RoleUser {
			continue
		}
		lower := strings.ToLower(msg.Content)
		for _, s := range signals {
			for _, kw := range s.keywords {
				if strings.Contains(lower, kw) {
					hits[s.field] = true
				}
			}
		}
	}

	const boost = 10
	var delta multicity.RankingWeights
	if hits["layover"] {
		delta.LayoverQuality = boost
	}
	if hits["carbon"] {
		delta.Carbon = boost
	}
	if hits["schedule"] {
		delta.Schedule = boost
	}
	if hits["duration"] {
		delta.FlightDuration = boost
	}
	return delta
}

// addWeights returns the sum of base and delta weights.
func addWeights(base, delta multicity.RankingWeights) multicity.RankingWeights {
	return multicity.RankingWeights{
		Cost:               base.Cost + delta.Cost,
		AirlineConsistency: base.AirlineConsistency + delta.AirlineConsistency,
		LayoverQuality:     base.LayoverQuality + delta.LayoverQuality,
		FlightDuration:     base.FlightDuration + delta.FlightDuration,
		StopoverCity:       base.StopoverCity + delta.StopoverCity,
		Schedule:           base.Schedule + delta.Schedule,
		Carbon:             base.Carbon + delta.Carbon,
	}
}

// profileWeights maps a profile name to the corresponding ranking weights.
// Unknown or empty profiles default to budget.
func profileWeights(name string) multicity.RankingWeights {
	if w, ok := profiles[name]; ok {
		return w
	}
	return multicity.WeightsBudget
}

// refinementHint returns a system message listing available refinement levers.
// Appended to conversation history after results so the LLM knows what to suggest.
func refinementHint() string {
	return "The user can refine their search. Available options: " +
		"try different dates, search nearby airports for cheaper fares, " +
		"change cabin class (economy/business/first), filter to direct flights only, " +
		"adjust number of passengers, add a return date for round-trip pricing, " +
		"set leg2_date for multi-city trips (YYYY-MM-DD, date to leave the stopover city), " +
		"adjust date flexibility (flex_days), " +
		"change ranking profile (budget/comfort/balanced/eco), " +
		"filter by preferred_alliance (Star Alliance/OneWorld/SkyTeam), " +
		"filter by departure time (departure_after/departure_before in HH:MM), " +
		"filter by arrival time (arrival_after/arrival_before in HH:MM), " +
		"limit flight duration with max_duration_hours, " +
		"sort results by sort_by (price/duration/departure/score), " +
		"exclude airlines with avoid_airlines (comma-separated IATA codes, e.g. \"BA,LH\"), " +
		"keep only specific airlines with preferred_airlines (comma-separated IATA codes, e.g. \"AC,UA\"), " +
		"adjust stopover duration with min_stopover_hours/max_stopover_hours (default 48-144h), " +
		"To remove a previously set filter, use clear_fields with the field names to reset " +
		"(e.g. {\"clear_fields\":[\"direct_only\",\"max_price\"]}). " +
		"When the user requests a change, re-emit a JSON object with ONLY the changed fields. " +
		"For example, to switch to business class: {\"cabin\":\"business\"}"
}

// filterLabels maps json tag names to human-readable labels for filter
// suggestion output. Fields not in this map are not considered filters.
// Grouped fields (departure_after/departure_before) share a label so
// the suggestion reads "departure time window" rather than listing both.
var filterLabels = map[string]string{
	"direct_only":        "direct_only",
	"max_price":          "max_price",
	"departure_after":    "departure time window",
	"departure_before":   "departure time window",
	"arrival_after":      "arrival time window",
	"arrival_before":     "arrival time window",
	"max_duration_hours": "max_duration_hours",
	"preferred_alliance": "preferred_alliance",
	"avoid_airlines":     "avoid_airlines",
	"preferred_airlines": "preferred_airlines",
}

// filterSuggestion returns a hint about which active filters might be causing
// zero results. Uses reflection to auto-support new filter fields when added
// to filterLabels. Returns empty string when no optional filters are active.
func filterSuggestion(p tripParams) string {
	seen := make(map[string]bool)
	var filters []string
	v := reflect.ValueOf(p)
	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		name := jsonFieldName(t.Field(i))
		label, isFilter := filterLabels[name]
		if !isFilter || v.Field(i).IsZero() || seen[label] {
			continue
		}
		seen[label] = true
		filters = append(filters, label)
	}
	if len(filters) == 0 {
		return ""
	}
	return "Active filters that may be limiting results: " + strings.Join(filters, ", ") +
		". Try relaxing some of these constraints."
}

// priceInsightHint formats a message about typical prices for the route.
// Returns empty string when no insights are available.
func priceInsightHint(pi search.PriceInsights) string {
	if pi.PriceLevel == "" {
		return ""
	}
	return fmt.Sprintf("Typical prices for this route: $%.0f-$%.0f (price level: %s)",
		pi.TypicalPriceRange[0], pi.TypicalPriceRange[1], pi.PriceLevel)
}

// relaxFilters removes the strictest active filter from params and returns
// the relaxed params along with a human-readable description of what changed.
// Returns empty description when no optional filters are active.
// Relaxation priority: direct_only -> preferred_alliance -> preferred_airlines
// -> max_price (50% increase) -> departure/arrival time -> max_duration.
func relaxFilters(p tripParams) (tripParams, string) {
	switch {
	case p.DirectOnly:
		p.DirectOnly = false
		return p, "Relaxed direct_only — now including connecting flights."
	case p.PreferredAlliance != "":
		p.PreferredAlliance = ""
		return p, "Relaxed preferred_alliance — now searching all alliances."
	case p.PreferredAirlines != "":
		p.PreferredAirlines = ""
		return p, "Relaxed preferred_airlines — now searching all airlines."
	case p.MaxPrice > 0:
		p.MaxPrice *= 1.5
		return p, fmt.Sprintf("Relaxed max_price — increased budget to $%.0f.", p.MaxPrice)
	case p.DepartureAfter != "" || p.DepartureBefore != "":
		p.DepartureAfter = ""
		p.DepartureBefore = ""
		return p, "Relaxed departure time window — now searching all departure times."
	case p.ArrivalAfter != "" || p.ArrivalBefore != "":
		p.ArrivalAfter = ""
		p.ArrivalBefore = ""
		return p, "Relaxed arrival time window — now searching all arrival times."
	case p.MaxDurationHours > 0:
		p.MaxDurationHours = 0
		return p, "Relaxed max_duration — now searching all flight durations."
	default:
		return p, ""
	}
}

// formatSearchParams returns a human-readable summary of search parameters.
// Example: "Searching DEL -> YYZ on 2025-06-15 (business, flex +/-3 days, max $1200, direct only)"
func formatSearchParams(p tripParams) string {
	base := fmt.Sprintf("Searching %s -> %s on %s", p.Origin, p.Destination, p.DepartureDate)

	var opts []string
	if p.ReturnDate != "" {
		opts = append(opts, "return "+p.ReturnDate)
	}
	if p.Leg2Date != "" {
		opts = append(opts, "leg2 "+p.Leg2Date)
	}
	if p.Cabin != "" && p.Cabin != defaultCabin {
		opts = append(opts, p.Cabin)
	}
	if p.Passengers > 1 {
		opts = append(opts, fmt.Sprintf("%d pax", p.Passengers))
	}
	if p.FlexDays > 0 {
		opts = append(opts, fmt.Sprintf("flex +/-%d days", p.FlexDays))
	}
	if p.DirectOnly {
		opts = append(opts, "direct only")
	}
	if p.MaxPrice > 0 {
		opts = append(opts, fmt.Sprintf("max $%.0f", p.MaxPrice))
	}
	if p.PreferredAlliance != "" {
		opts = append(opts, p.PreferredAlliance)
	}
	if p.DepartureAfter != "" || p.DepartureBefore != "" {
		opts = append(opts, "depart "+formatTimeRange(p.DepartureAfter, p.DepartureBefore))
	}
	if p.ArrivalAfter != "" || p.ArrivalBefore != "" {
		opts = append(opts, "arrive "+formatTimeRange(p.ArrivalAfter, p.ArrivalBefore))
	}
	if p.MaxDurationHours > 0 {
		opts = append(opts, fmt.Sprintf("max %dh", p.MaxDurationHours))
	}
	if p.AvoidAirlines != "" {
		opts = append(opts, "avoid "+p.AvoidAirlines)
	}
	if p.PreferredAirlines != "" {
		opts = append(opts, "only "+p.PreferredAirlines)
	}
	if p.Profile != "" && p.Profile != "budget" {
		opts = append(opts, p.Profile+" profile")
	}

	if len(opts) == 0 {
		return base
	}
	return base + " (" + strings.Join(opts, ", ") + ")"
}

// formatTimeRange returns "HH:MM-HH:MM", handling empty after/before.
func formatTimeRange(after, before string) string {
	switch {
	case after != "" && before != "":
		return after + "-" + before
	case after != "":
		return "after " + after
	default:
		return "before " + before
	}
}

// zeroResultsSuggestion returns proactive suggestions when a search returns no
// results, including nearby airports and flex-date advice.
func zeroResultsSuggestion(params tripParams) string {
	var parts []string

	// Nearby airport suggestions.
	originNearby := search.NearbyAirports(params.Origin)
	destNearby := search.NearbyAirports(params.Destination)
	if len(originNearby) > 0 {
		parts = append(parts, fmt.Sprintf("origin %s also has %s", params.Origin, strings.Join(originNearby, ", ")))
	}
	if len(destNearby) > 0 {
		parts = append(parts, fmt.Sprintf("destination %s also has %s", params.Destination, strings.Join(destNearby, ", ")))
	}

	// Flex-date advice.
	switch {
	case params.FlexDays > 0:
		parts = append(parts, fmt.Sprintf("flex_days is already set to %d", params.FlexDays))
	default:
		parts = append(parts, "consider setting flex_days to 2-3 to search nearby dates")
	}

	if len(parts) == 0 {
		return ""
	}
	return "Try nearby airports: " + strings.Join(parts, ". ") + "."
}

// truncateHistory keeps the first system message and the most recent maxRecent
// non-system messages, dropping older messages to prevent token overflow.
func truncateHistory(history []llm.Message, maxRecent int) []llm.Message {
	if len(history) <= maxRecent+1 {
		return history
	}
	return append([]llm.Message{history[0]}, history[len(history)-maxRecent:]...)
}

// stopoverSuggestion returns a tip suggesting multi-city routing via stopover
// cities if the route has known stopovers. Returns empty for multi-city trips
// (leg2Date non-empty) since the user already has a stopover planned.
func stopoverSuggestion(origin, dest, leg2Date string) string {
	if leg2Date != "" {
		return ""
	}
	stopovers := multicity.StopoversForRoute(origin, dest)
	if len(stopovers) == 0 {
		return ""
	}
	// Show up to 3 city names.
	n := len(stopovers)
	if n > 3 {
		n = 3
	}
	var cities []string
	for _, s := range stopovers[:n] {
		cities = append(cities, s.City)
	}
	return fmt.Sprintf("Flying via %s often saves money on this route. Would you like me to search a two-leg journey with a stopover?",
		strings.Join(cities, " or "))
}
