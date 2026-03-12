package cmd

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"booker/llm"
	"booker/search"
	"booker/search/multicity"

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
	f.String(keyFormat, "", "output format: bullet (default), table, or json")
	f.String(keyProfile, "budget", "ranking profile: budget, comfort, balanced, or eco")
	f.BoolP(keyVerbose, "v", false, "show debug output")
	_ = viper.BindPFlags(f)
}

// tripParams holds extracted flight search parameters from the LLM dialogue.
type tripParams struct {
	Origin            string   `json:"origin"`
	Destination       string   `json:"destination"`
	DepartureDate     string   `json:"departure_date"`
	ReturnDate        string   `json:"return_date,omitempty"`
	Leg2Date          string   `json:"leg2_date,omitempty"`
	Passengers        int      `json:"passengers,omitempty"`
	Cabin             string   `json:"cabin,omitempty"`
	MaxPrice          float64  `json:"max_price,omitempty"`
	DirectOnly        bool     `json:"direct_only,omitempty"`
	FlexDays          int      `json:"flex_days,omitempty"`
	Profile           string   `json:"profile,omitempty"`
	PreferredAlliance string   `json:"preferred_alliance,omitempty"`
	DepartureAfter    string   `json:"departure_after,omitempty"`
	DepartureBefore   string   `json:"departure_before,omitempty"`
	ArrivalAfter      string   `json:"arrival_after,omitempty"`
	ArrivalBefore     string   `json:"arrival_before,omitempty"`
	MaxDurationHours  int      `json:"max_duration_hours,omitempty"`
	SortBy            string   `json:"sort_by,omitempty"`
	AvoidAirlines     string   `json:"avoid_airlines,omitempty"`
	PreferredAirlines string   `json:"preferred_airlines,omitempty"`
	MinStopoverHours  int      `json:"min_stopover_hours,omitempty"`
	MaxStopoverHours  int      `json:"max_stopover_hours,omitempty"`
	Context           string   `json:"context,omitempty"`
	ClearFields       []string `json:"clear_fields,omitempty"`
}

// priceInsighter provides access to the last price insights from a search.
type priceInsighter interface {
	LastPriceInsights() search.PriceInsights
}

// weightsUpdater allows dynamic ranking weight changes mid-session.
type weightsUpdater interface {
	SetWeights(multicity.RankingWeights)
}

// looksLikeHelp returns true if the user input is a help request.
func looksLikeHelp(input string) bool {
	lower := strings.ToLower(input)
	return lower == "help" || lower == "?" ||
		strings.HasPrefix(lower, "what can")
}

// chatHelpText returns a summary of available chat capabilities.
func chatHelpText() string {
	return `I can help you find flights. Here's what I can do:

  Search: Tell me where and when you want to fly
    Example: "I want to fly from Delhi to Toronto on June 15"

  Refine: After results, ask me to adjust
    Example: "show business class" or "try a later date"

  Compare: Compare results side by side
    Example: "compare 1 and 3"

  Details: Get full details on a result
    Example: "details on option 2"

  Available filters:
    - cabin (economy/business/first)
    - max_price, direct_only, flex_days
    - departure_after/before, arrival_after/before (HH:MM)
    - preferred_alliance (Star Alliance/OneWorld/SkyTeam)
    - avoid_airlines, preferred_airlines (IATA codes)
    - sort_by (price/duration/departure/score)
    - profile (budget/comfort/balanced/eco)
    - leg2_date for multi-city trips

  Reset filters: "clear my price limit" or clear_fields

  Type 'quit' to exit.`
}

// looksLikeComparison returns true if the user input looks like a request
// to compare search results (e.g. "compare 1 and 3").
func looksLikeComparison(input string) bool {
	lower := strings.ToLower(input)
	return strings.HasPrefix(lower, "compare ")
}

// looksLikeDetail returns true if the user input looks like a request for
// details on a specific result (e.g. "details on option 2", "more about 1").
func looksLikeDetail(input string) bool {
	lower := strings.ToLower(input)
	return strings.HasPrefix(lower, "detail") ||
		strings.HasPrefix(lower, "more about") ||
		strings.HasPrefix(lower, "more info") ||
		strings.Contains(lower, "more about")
}

var numberRe = regexp.MustCompile(`\d+`)

// parseOptionIndices extracts 1-based option numbers from user input.
func parseOptionIndices(input string) []int {
	matches := numberRe.FindAllString(input, -1)
	var indices []int
	for _, m := range matches {
		n, err := strconv.Atoi(m)
		if err == nil && n > 0 {
			indices = append(indices, n)
		}
	}
	return indices
}

// legSummary returns "airline, route, duration, stops" for a single leg.
// Returns empty string when the leg has no outbound segments.
func legSummary(leg search.Leg) string {
	segs := leg.Flight.Outbound
	if len(segs) == 0 {
		return ""
	}
	airline := segs[0].AirlineName
	if airline == "" {
		airline = segs[0].Airline
	}
	route := segs[0].Origin + " -> " + segs[len(segs)-1].Destination
	stops := formatLayoverSummary(segs)
	return fmt.Sprintf("%s, %s, %s, %s", airline, route, formatFlightDuration(leg.Flight.TotalDuration), stops)
}

// formatOptionDetail returns a detailed summary of a single search result.
// idx is 1-based.
func formatOptionDetail(results []search.Itinerary, idx int) string {
	if idx < 1 || idx > len(results) {
		return fmt.Sprintf("Option %d is out of range (1-%d).", idx, len(results))
	}
	itin := results[idx-1]
	var b strings.Builder
	fmt.Fprintf(&b, "Option %d:\n", idx)
	fmt.Fprintf(&b, "  Price: $%.0f USD\n", itin.TotalPrice.Amount)
	fmt.Fprintf(&b, "  Duration: %s\n", formatFlightDuration(itin.TotalTravel))
	for i, leg := range itin.Legs {
		summary := legSummary(leg)
		if summary == "" {
			continue
		}
		fmt.Fprintf(&b, "  Leg %d: %s\n", i+1, summary)
		if leg.Flight.CarbonKg > 0 {
			fmt.Fprintf(&b, "    CO2: %dkg\n", leg.Flight.CarbonKg)
		}
		if leg.Flight.BookingURL != "" {
			fmt.Fprintf(&b, "    Book: %s\n", leg.Flight.BookingURL)
		}
	}
	if itin.Reasoning != "" {
		fmt.Fprintf(&b, "  Ranking reason: %s\n", itin.Reasoning)
	}
	return b.String()
}

// formatComparison returns a side-by-side comparison of the specified results.
// indices are 1-based. For multi-leg itineraries, shows per-leg details.
func formatComparison(results []search.Itinerary, indices []int) string {
	var b strings.Builder
	b.WriteString("Comparison:\n")
	for _, idx := range indices {
		if idx < 1 || idx > len(results) {
			fmt.Fprintf(&b, "  Option %d: out of range (1-%d)\n", idx, len(results))
			continue
		}
		itin := results[idx-1]
		if len(itin.Legs) <= 1 {
			// Single-leg: compact one-line format.
			summary := legSummary(itin.Legs[0])
			fmt.Fprintf(&b, "  Option %d: %s, $%.0f\n", idx, summary, itin.TotalPrice.Amount)
		} else {
			// Multi-leg: show each leg.
			fmt.Fprintf(&b, "  Option %d: $%.0f, %s\n", idx, itin.TotalPrice.Amount, formatFlightDuration(itin.TotalTravel))
			for i, leg := range itin.Legs {
				if summary := legSummary(leg); summary != "" {
					fmt.Fprintf(&b, "    Leg %d: %s\n", i+1, summary)
				}
			}
		}
	}
	return b.String()
}

// displayChatResults renders search results to out in the configured format
// and displays price insights if available. Defaults to bullet format for
// readability; "table" and "json" are available as explicit overrides.
func displayChatResults(out io.Writer, results []search.Itinerary, pi search.PriceInsights) {
	cur := viper.GetString(keyCurrency)
	switch viper.GetString(keyFormat) {
	case "json":
		if err := printJSONWithInsights(out, results, cur, pi); err != nil {
			_, _ = fmt.Fprintf(out, "Error: %v\n", err)
		}
	case "table":
		printTable(out, results, cur)
	default:
		printBulletResults(out, results, cur)
	}
	if s := formatPriceInsights(pi); s != "" {
		_, _ = fmt.Fprintln(out, s)
	}
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
	var lastResults []search.Itinerary

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

		// Intercept help requests before the LLM call.
		if looksLikeHelp(userInput) {
			_, _ = fmt.Fprintln(out, chatHelpText())
			continue
		}

		// Intercept comparison/detail requests using cached results.
		if len(lastResults) > 0 {
			indices := parseOptionIndices(userInput)
			if looksLikeComparison(userInput) && len(indices) >= 2 {
				text := formatComparison(lastResults, indices)
				_, _ = fmt.Fprintln(out, text)
				history = append(history, llm.Message{Role: llm.RoleUser, Content: userInput})
				history = append(history, llm.Message{Role: llm.RoleAssistant, Content: text})
				continue
			}
			if looksLikeDetail(userInput) && len(indices) >= 1 {
				text := formatOptionDetail(lastResults, indices[0])
				_, _ = fmt.Fprintln(out, text)
				history = append(history, llm.Message{Role: llm.RoleUser, Content: userInput})
				history = append(history, llm.Message{Role: llm.RoleAssistant, Content: text})
				continue
			}
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

		// Infer profile from conversation when LLM didn't set one explicitly.
		if params.Profile == "" {
			params.Profile = inferProfile(history)
		}

		lastParams = params

		// Update ranking weights: base profile + context-aware deltas from conversation.
		if wu != nil {
			base := profileWeights(params.Profile)
			delta := contextWeights(history)
			wu.SetWeights(addWeights(base, delta))
		}

		// Build and execute the search.
		results, err := chatSearch(ctx, params, picker, out)
		if err != nil {
			_, _ = fmt.Fprintf(out, "Error: %v\n", err)
			continue
		}

		lastResults = results

		if len(results) == 0 {
			// Auto-retry with relaxed filters before showing zero-results suggestions.
			if relaxed, desc := relaxFilters(params); desc != "" {
				_, _ = fmt.Fprintln(out, desc)
				retryResults, retryErr := chatSearch(ctx, relaxed, picker, out)
				if retryErr == nil && len(retryResults) > 0 {
					results = retryResults
					lastResults = results
					params = relaxed
					lastParams = params
					// Fall through to display results below.
					goto showResults
				}
			}
			_, _ = fmt.Fprintln(out, "No flights found. Try different dates or airports.")
			if hint := filterSuggestion(params); hint != "" {
				_, _ = fmt.Fprintln(out, hint)
			}
			if hint := zeroResultsSuggestion(params); hint != "" {
				_, _ = fmt.Fprintln(out, hint)
			}
			if insights != nil {
				if hint := priceInsightHint(insights.LastPriceInsights()); hint != "" {
					_, _ = fmt.Fprintln(out, hint)
				}
			}
			continue
		}
	showResults:

		var pi search.PriceInsights
		if insights != nil {
			pi = insights.LastPriceInsights()
		}
		displayChatResults(out, results, pi)

		// Show fare trend for flex-date searches.
		if params.FlexDays > 0 {
			if hint := formatFareTrend(search.ComputeFareTrend(results)); hint != "" {
				_, _ = fmt.Fprintln(out, hint)
			}
		}

		// Suggest multi-city routing for single-leg trips with known stopovers.
		// Add to LLM history so it can guide the user through stopover setup.
		if tip := stopoverSuggestion(params.Origin, params.Destination, params.Leg2Date); tip != "" {
			_, _ = fmt.Fprintln(out, tip)
			history = append(history, llm.Message{Role: llm.RoleAssistant, Content: tip})
		}

		// Add result summary and refinement guidance to conversation history
		// so the LLM knows what was shown and what levers are available.
		summary := resultSummaryForChat(results, params, pi)
		history = append(history, llm.Message{Role: llm.RoleAssistant, Content: summary})
		history = append(history, llm.Message{Role: llm.RoleSystem, Content: refinementHint()})

		_, _ = fmt.Fprintln(out, "Want to refine? (e.g., 'show cheaper', 'try business class', or 'quit')")
	}

	return scanner.Err()
}
