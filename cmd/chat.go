package cmd

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"booker/llm"
	"booker/search"
	"booker/search/multicity"
	"booker/types"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var chatCmd = &cobra.Command{
	Use:   "chat",
	Short: "Plan a trip through conversation",
	Long:  `Start a conversational session where the agent gathers your travel preferences and finds flights.`,
	RunE:  runChat,
}

func init() {
	rootCmd.AddCommand(chatCmd)

	f := chatCmd.Flags()
	f.String(keyCurrency, defaultCurrency, "display currency (e.g. CAD, USD, EUR)")
	f.String(keyFormat, "table", "output format: table or json")
	f.String(keyProfile, "budget", "ranking profile: budget, comfort, balanced, or eco")
	f.BoolP(keyVerbose, "v", false, "show debug output")
	_ = viper.BindPFlags(f)
}

// tripParams holds extracted flight search parameters from the LLM dialogue.
type tripParams struct {
	Origin            string  `json:"origin"`
	Destination       string  `json:"destination"`
	DepartureDate     string  `json:"departure_date"`
	ReturnDate        string  `json:"return_date,omitempty"`
	Leg2Date          string  `json:"leg2_date,omitempty"`
	Passengers        int     `json:"passengers,omitempty"`
	Cabin             string  `json:"cabin,omitempty"`
	MaxPrice          float64 `json:"max_price,omitempty"`
	DirectOnly        bool    `json:"direct_only,omitempty"`
	FlexDays          int     `json:"flex_days,omitempty"`
	Profile           string  `json:"profile,omitempty"`
	PreferredAlliance string  `json:"preferred_alliance,omitempty"`
	DepartureAfter    string  `json:"departure_after,omitempty"`
	DepartureBefore   string  `json:"departure_before,omitempty"`
	ArrivalAfter      string  `json:"arrival_after,omitempty"`
	ArrivalBefore     string  `json:"arrival_before,omitempty"`
	MaxDurationHours  int     `json:"max_duration_hours,omitempty"`
	SortBy            string  `json:"sort_by,omitempty"`
	AvoidAirlines     string  `json:"avoid_airlines,omitempty"`
	PreferredAirlines string  `json:"preferred_airlines,omitempty"`
	MinStopoverHours  int     `json:"min_stopover_hours,omitempty"`
	MaxStopoverHours  int     `json:"max_stopover_hours,omitempty"`
	Context           string  `json:"context,omitempty"`
}

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

Be a helpful travel advisor:
- When the user is flexible on routing, suggest stopover cities that could save money (e.g. "Flying Delhi to Toronto via Bangkok often saves $300-400 and you get a 2-day city break").
- If the user mentions a city served by multiple airports, suggest nearby alternatives (e.g., for New York: JFK, EWR, LGA). Searching nearby airports can reveal cheaper fares.
- Explain tradeoffs briefly when relevant (e.g. "A 3-hour longer layover saves $200" or "Business class on this route is only $400 more than premium economy").
- Ask about flexibility on dates and routing — small changes often unlock much better prices.

When you have at least the origin, destination, and departure_date, output the parameters as a JSON object on its own line. You may include optional fields if the user mentioned them. Example:
{"origin":"DEL","destination":"YYZ","departure_date":"2025-06-15","passengers":2,"cabin":"economy","context":"budget trip"}

After outputting the JSON, briefly explain what you're searching for and why you chose that approach.`
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

// resultSummaryForChat builds a summary of search results for the conversation
// history, including top 3 results with price, airline, duration, and stops so
// the LLM can explain recommendations and answer comparison questions.
func resultSummaryForChat(results []search.Itinerary, params tripParams) string {
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
		stops := itineraryStops(r)
		stopWord := "stops"
		if stops == 1 {
			stopWord = "stop"
		}
		fmt.Fprintf(&b, " %d) %s, %s, %d %s, $%.0f.",
			i+1, airline, formatFlightDuration(r.Legs[0].Flight.TotalDuration), stops, stopWord, r.TotalPrice.Amount)
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

// mergeParams fills zero-value fields in partial from prev, producing
// a complete set of params for a follow-up search.
func mergeParams(prev, partial tripParams) tripParams {
	merged := partial
	if merged.Origin == "" {
		merged.Origin = prev.Origin
	}
	if merged.Destination == "" {
		merged.Destination = prev.Destination
	}
	if merged.DepartureDate == "" {
		merged.DepartureDate = prev.DepartureDate
	}
	if merged.ReturnDate == "" {
		merged.ReturnDate = prev.ReturnDate
	}
	if merged.Leg2Date == "" {
		merged.Leg2Date = prev.Leg2Date
	}
	if merged.Passengers == 0 {
		merged.Passengers = prev.Passengers
	}
	if merged.Cabin == "" {
		merged.Cabin = prev.Cabin
	}
	if merged.MaxPrice == 0 {
		merged.MaxPrice = prev.MaxPrice
	}
	if merged.Profile == "" {
		merged.Profile = prev.Profile
	}
	if merged.PreferredAlliance == "" {
		merged.PreferredAlliance = prev.PreferredAlliance
	}
	if merged.DepartureAfter == "" {
		merged.DepartureAfter = prev.DepartureAfter
	}
	if merged.DepartureBefore == "" {
		merged.DepartureBefore = prev.DepartureBefore
	}
	if merged.ArrivalAfter == "" {
		merged.ArrivalAfter = prev.ArrivalAfter
	}
	if merged.ArrivalBefore == "" {
		merged.ArrivalBefore = prev.ArrivalBefore
	}
	if merged.MaxDurationHours == 0 {
		merged.MaxDurationHours = prev.MaxDurationHours
	}
	if merged.SortBy == "" {
		merged.SortBy = prev.SortBy
	}
	if merged.AvoidAirlines == "" {
		merged.AvoidAirlines = prev.AvoidAirlines
	}
	if merged.PreferredAirlines == "" {
		merged.PreferredAirlines = prev.PreferredAirlines
	}
	if merged.MinStopoverHours == 0 {
		merged.MinStopoverHours = prev.MinStopoverHours
	}
	if merged.MaxStopoverHours == 0 {
		merged.MaxStopoverHours = prev.MaxStopoverHours
	}
	if merged.Context == "" {
		merged.Context = prev.Context
	}
	if !merged.DirectOnly {
		merged.DirectOnly = prev.DirectOnly
	}
	if merged.FlexDays == 0 {
		merged.FlexDays = prev.FlexDays
	}
	return merged
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
		// At least one field must be set.
		if p.Origin != "" || p.Destination != "" || p.DepartureDate != "" ||
			p.ReturnDate != "" || p.Leg2Date != "" || p.Passengers != 0 || p.Cabin != "" ||
			p.MaxPrice != 0 || p.DirectOnly || p.FlexDays != 0 ||
			p.Profile != "" || p.PreferredAlliance != "" ||
			p.DepartureAfter != "" || p.DepartureBefore != "" ||
			p.ArrivalAfter != "" || p.ArrivalBefore != "" ||
			p.MaxDurationHours != 0 ||
			p.SortBy != "" || p.AvoidAirlines != "" || p.PreferredAirlines != "" ||
			p.MinStopoverHours != 0 || p.MaxStopoverHours != 0 || p.Context != "" {
			return p, true
		}
	}
	return tripParams{}, false
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
		"When the user requests a change, re-emit a JSON object with ONLY the changed fields. " +
		"For example, to switch to business class: {\"cabin\":\"business\"}"
}

// filterSuggestion returns a hint about which active filters might be causing
// zero results. Returns empty string when no optional filters are active.
func filterSuggestion(p tripParams) string {
	var filters []string
	if p.DirectOnly {
		filters = append(filters, "direct_only")
	}
	if p.MaxPrice > 0 {
		filters = append(filters, "max_price")
	}
	if p.DepartureAfter != "" || p.DepartureBefore != "" {
		filters = append(filters, "departure time window")
	}
	if p.ArrivalAfter != "" || p.ArrivalBefore != "" {
		filters = append(filters, "arrival time window")
	}
	if p.MaxDurationHours > 0 {
		filters = append(filters, "max_duration_hours")
	}
	if p.PreferredAlliance != "" {
		filters = append(filters, "preferred_alliance")
	}
	if p.AvoidAirlines != "" {
		filters = append(filters, "avoid_airlines")
	}
	if p.PreferredAirlines != "" {
		filters = append(filters, "preferred_airlines")
	}
	if len(filters) == 0 {
		return ""
	}
	return "Active filters that may be limiting results: " + strings.Join(filters, ", ") +
		". Try relaxing some of these constraints."
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

// priceInsighter provides access to the last price insights from a search.
type priceInsighter interface {
	LastPriceInsights() search.PriceInsights
}

// weightsUpdater allows dynamic ranking weight changes mid-session.
type weightsUpdater interface {
	SetWeights(multicity.RankingWeights)
}

func runChat(cmd *cobra.Command, _ []string) error {
	if !viper.GetBool(keyVerbose) {
		log.SetOutput(io.Discard)
	}

	weights := profileWeights(viper.GetString(keyProfile))
	picker, llmClient, rawProvider, ranker, err := buildPicker(weights, "")
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(cmd.Context(), defaultTimeout)
	defer cancel()

	return chatLoop(ctx, llmClient, picker, rawProvider, ranker, os.Stdin, os.Stdout)
}

// chatLoop runs the multi-turn conversation. Separated from runChat for testability.
// insights may be nil when no price insight provider is available.
// wu may be nil when dynamic weight updates are not supported.
func chatLoop(ctx context.Context, llmClient search.ChatCompleter, picker *search.Picker, insights priceInsighter, wu weightsUpdater, in io.Reader, out io.Writer) error {
	history := []llm.Message{
		{Role: llm.RoleSystem, Content: chatSystemPrompt(time.Now())},
	}

	scanner := bufio.NewScanner(in)
	_, _ = fmt.Fprintln(out, "Where would you like to fly? (type 'quit' to exit)")

	var lastParams tripParams

	for {
		_, _ = fmt.Fprint(out, "\n> ")
		if !scanner.Scan() {
			break
		}
		userInput := strings.TrimSpace(scanner.Text())
		if userInput == "" {
			continue
		}
		if userInput == "quit" || userInput == "exit" {
			_, _ = fmt.Fprintln(out, "Goodbye!")
			return nil
		}

		history = append(history, llm.Message{Role: llm.RoleUser, Content: userInput})
		history = truncateHistory(history, maxHistoryMessages)

		response, err := llmClient.ChatCompletion(ctx, history)
		if err != nil {
			_, _ = fmt.Fprintf(out, "Error: %v\n", err)
			continue
		}

		history = append(history, llm.Message{Role: llm.RoleAssistant, Content: response})

		// Try full parse first, then partial merge for follow-ups.
		params, found := parseTripParams(response)
		if !found && lastParams.Origin != "" {
			partial, partialFound := parsePartialParams(response)
			if partialFound {
				params = mergeParams(lastParams, partial)
				found = true
			}
		}
		if !found {
			_, _ = fmt.Fprintln(out, response)
			continue
		}

		lastParams = params

		// Update ranking weights when the profile changes mid-session.
		if params.Profile != "" && wu != nil {
			wu.SetWeights(profileWeights(params.Profile))
		}

		// Build and execute the search.
		req := buildRequestFromParams(params)
		_, _ = fmt.Fprintf(out, "Searching %s -> %s on %s...\n", req.Origin, req.Destination, req.DepartureDate)
		if hint := nearbyAirportHint(req.Origin, req.Destination); hint != "" {
			_, _ = fmt.Fprintf(out, "Tip: %s\n", hint)
		}

		strategy, reason, err := picker.Pick(ctx, req)
		if err != nil {
			_, _ = fmt.Fprintf(out, "Error picking strategy: %v\n", err)
			continue
		}
		_, _ = fmt.Fprintf(out, "Using %s strategy (%s)\n", strategy.Name(), reason)

		results, err := strategy.Search(ctx, req)
		if err != nil {
			_, _ = fmt.Fprintf(out, "Search error: %v\n", err)
			continue
		}

		if len(results) == 0 {
			_, _ = fmt.Fprintln(out, "No flights found. Try different dates or airports.")
			if hint := filterSuggestion(params); hint != "" {
				_, _ = fmt.Fprintln(out, hint)
			}
			if hint := zeroResultsSuggestion(params); hint != "" {
				_, _ = fmt.Fprintln(out, hint)
			}
			continue
		}

		cur := viper.GetString(keyCurrency)
		var pi search.PriceInsights
		if insights != nil {
			pi = insights.LastPriceInsights()
		}
		switch viper.GetString(keyFormat) {
		case "json":
			if err := printJSONWithInsights(out, results, cur, pi); err != nil {
				_, _ = fmt.Fprintf(out, "Error: %v\n", err)
			}
		default:
			printTable(out, results, cur)
			if s := formatPriceInsights(pi); s != "" {
				_, _ = fmt.Fprintln(out, s)
			}
		}

		// Add result summary and refinement guidance to conversation history
		// so the LLM knows what was shown and what levers are available.
		summary := resultSummaryForChat(results, params)
		history = append(history, llm.Message{Role: llm.RoleAssistant, Content: summary})
		history = append(history, llm.Message{Role: llm.RoleSystem, Content: refinementHint()})

		_, _ = fmt.Fprintln(out, "Want to refine? (e.g., 'show cheaper', 'try business class', or 'quit')")
	}

	return scanner.Err()
}
