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
	f.String(keyProfile, "budget", "ranking profile: budget, comfort, or balanced")
	f.BoolP(keyVerbose, "v", false, "show debug output")
	_ = viper.BindPFlags(f)
}

// tripParams holds extracted flight search parameters from the LLM dialogue.
type tripParams struct {
	Origin        string  `json:"origin"`
	Destination   string  `json:"destination"`
	DepartureDate string  `json:"departure_date"`
	ReturnDate    string  `json:"return_date,omitempty"`
	Passengers    int     `json:"passengers,omitempty"`
	Cabin         string  `json:"cabin,omitempty"`
	MaxPrice      float64 `json:"max_price,omitempty"`
	DirectOnly    bool    `json:"direct_only,omitempty"`
	Profile       string  `json:"profile,omitempty"`
	Context       string  `json:"context,omitempty"`
}

// chatSystemPrompt returns the system prompt for the chat conversation.
func chatSystemPrompt() string {
	return `You are a flight booking assistant. Help the user plan their trip by gathering the following information through natural conversation:

Required:
- origin: departure airport IATA code (e.g. "DEL", "JFK")
- destination: arrival airport IATA code (e.g. "YYZ", "LHR")
- departure_date: in YYYY-MM-DD format

Optional:
- return_date: in YYYY-MM-DD format (for round trips)
- passengers: number of travelers (default: 1)
- cabin: economy, premium_economy, business, or first (default: economy)
- max_price: maximum budget per flight in USD (e.g. 1200)
- direct_only: true to show only non-stop flights
- profile: ranking profile — "budget" (cheapest), "comfort" (best schedule/airline), or "balanced" (default: budget)
- context: any preferences like "cheapest option" or "prefer direct flights"

Ask clarifying questions to gather missing information. Be conversational but concise.

When you have at least the origin, destination, and departure_date, output the parameters as a JSON object on its own line. You may include optional fields if the user mentioned them. Example:
{"origin":"DEL","destination":"YYZ","departure_date":"2025-06-15","passengers":2,"cabin":"economy","context":"budget trip"}

If the user mentions a city served by multiple airports, suggest nearby alternatives (e.g., for New York: JFK, EWR, LGA). Searching nearby airports can reveal cheaper fares.

After outputting the JSON, briefly explain what you're searching for.`
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
		line = strings.TrimSpace(line)
		// Strip markdown code fences.
		line = strings.TrimPrefix(line, "```json")
		line = strings.TrimPrefix(line, "```")
		line = strings.TrimSuffix(line, "```")
		line = strings.TrimSpace(line)

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
	return search.Request{
		Origin:        p.Origin,
		Destination:   p.Destination,
		DepartureDate: p.DepartureDate,
		ReturnDate:    p.ReturnDate,
		Passengers:    passengers,
		CabinClass:    types.CabinClass(cabin),
		FlexDays:      defaultFlexDays,
		MaxStops:      maxStops,
		MaxPrice:      p.MaxPrice,
		MaxResults:    defaultMaxResults,
		Context:       p.Context,
	}
}

// resultSummaryForChat builds a summary of search results for the conversation
// history, including top result details and search parameters so the LLM can
// explain recommendations and answer "why that one?" questions.
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
	fmt.Fprintf(&b, "I found %d results. Prices range from $%.0f to $%.0f USD.", len(results), minPrice, maxPrice)

	// Top result details.
	top := results[0]
	if len(top.Legs) > 0 && len(top.Legs[0].Flight.Outbound) > 0 {
		seg := top.Legs[0].Flight.Outbound[0]
		airline := seg.AirlineName
		if airline == "" {
			airline = seg.Airline
		}
		fmt.Fprintf(&b, " Top result: %s->%s on %s", seg.Origin, seg.Destination, airline)
		if top.Legs[0].Flight.TotalDuration > 0 {
			fmt.Fprintf(&b, ", %s", formatFlightDuration(top.Legs[0].Flight.TotalDuration))
		}
		fmt.Fprintf(&b, ", $%.0f.", top.TotalPrice.Amount)
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
	if merged.Context == "" {
		merged.Context = prev.Context
	}
	if !merged.DirectOnly {
		merged.DirectOnly = prev.DirectOnly
	}
	return merged
}

// parsePartialParams extracts trip parameters from an LLM response,
// accepting partial JSON (at least one recognized field set). Used for
// follow-up refinements where the LLM only emits changed fields.
func parsePartialParams(response string) (tripParams, bool) {
	for _, line := range strings.Split(response, "\n") {
		line = strings.TrimSpace(line)
		line = strings.TrimPrefix(line, "```json")
		line = strings.TrimPrefix(line, "```")
		line = strings.TrimSuffix(line, "```")
		line = strings.TrimSpace(line)

		if !strings.HasPrefix(line, "{") {
			continue
		}

		var p tripParams
		if err := json.Unmarshal([]byte(line), &p); err != nil {
			continue
		}
		// At least one field must be set.
		if p.Origin != "" || p.Destination != "" || p.DepartureDate != "" ||
			p.ReturnDate != "" || p.Passengers != 0 || p.Cabin != "" ||
			p.MaxPrice != 0 || p.DirectOnly || p.Profile != "" || p.Context != "" {
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
		"or change ranking profile (budget/comfort/balanced). " +
		"When the user requests a change, re-emit a JSON object with ONLY the changed fields. " +
		"For example, to switch to business class: {\"cabin\":\"business\"}"
}

func runChat(cmd *cobra.Command, _ []string) error {
	if !viper.GetBool(keyVerbose) {
		log.SetOutput(io.Discard)
	}

	weights := profileWeights(viper.GetString(keyProfile))
	picker, llmClient, _, err := buildPicker(weights, "")
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(cmd.Context(), defaultTimeout)
	defer cancel()

	return chatLoop(ctx, llmClient, picker, os.Stdin, os.Stdout)
}

// chatLoop runs the multi-turn conversation. Separated from runChat for testability.
func chatLoop(ctx context.Context, llmClient search.ChatCompleter, picker *search.Picker, in io.Reader, out io.Writer) error {
	history := []llm.Message{
		{Role: llm.RoleSystem, Content: chatSystemPrompt()},
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

		// Build and execute the search.
		req := buildRequestFromParams(params)
		_, _ = fmt.Fprintf(out, "Searching %s -> %s on %s...\n", req.Origin, req.Destination, req.DepartureDate)
		if hint := nearbyAirportHint(req.Origin, req.Destination); hint != "" {
			_, _ = fmt.Fprintf(out, "Tip: %s\n", hint)
		}

		strategy, err := picker.Pick(ctx, req)
		if err != nil {
			_, _ = fmt.Fprintf(out, "Error picking strategy: %v\n", err)
			continue
		}

		results, err := strategy.Search(ctx, req)
		if err != nil {
			_, _ = fmt.Fprintf(out, "Search error: %v\n", err)
			continue
		}

		if len(results) == 0 {
			_, _ = fmt.Fprintln(out, "No flights found. Try different dates or airports.")
			continue
		}

		cur := viper.GetString(keyCurrency)
		switch viper.GetString(keyFormat) {
		case "json":
			if err := printJSON(results, cur); err != nil {
				_, _ = fmt.Fprintf(out, "Error: %v\n", err)
			}
		default:
			printTable(results, cur)
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
