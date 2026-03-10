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

	"booker/config"
	"booker/httpclient"
	"booker/llm"
	"booker/provider"
	"booker/provider/cache"
	"booker/provider/serpapi"
	"booker/search"
	"booker/search/direct"
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
	f.BoolP(keyVerbose, "v", false, "show debug output")
	_ = viper.BindPFlags(f)
}

// tripParams holds extracted flight search parameters from the LLM dialogue.
type tripParams struct {
	Origin        string `json:"origin"`
	Destination   string `json:"destination"`
	DepartureDate string `json:"departure_date"`
	ReturnDate    string `json:"return_date,omitempty"`
	Passengers    int    `json:"passengers,omitempty"`
	Cabin         string `json:"cabin,omitempty"`
	Context       string `json:"context,omitempty"`
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
- context: any preferences like "cheapest option" or "prefer direct flights"

Ask clarifying questions to gather missing information. Be conversational but concise.

When you have at least the origin, destination, and departure_date, output the parameters as a JSON object on its own line. You may include optional fields if the user mentioned them. Example:
{"origin":"DEL","destination":"YYZ","departure_date":"2025-06-15","passengers":2,"cabin":"economy","context":"budget trip"}

After outputting the JSON, briefly explain what you're searching for.`
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
	return search.Request{
		Origin:        p.Origin,
		Destination:   p.Destination,
		DepartureDate: p.DepartureDate,
		ReturnDate:    p.ReturnDate,
		Passengers:    passengers,
		CabinClass:    types.CabinClass(cabin),
		FlexDays:      defaultFlexDays,
		MaxStops:      defaultMaxStops,
		MaxResults:    defaultMaxResults,
		Context:       p.Context,
	}
}

func runChat(cmd *cobra.Command, _ []string) error {
	if !viper.GetBool(keyVerbose) {
		log.SetOutput(io.Discard)
	}

	cfg := config.Default()
	httpClient := httpclient.New(cfg.HTTP)
	llmClient := llm.New(cfg.LLM, httpClient)

	// Build search infrastructure.
	registry := provider.NewRegistry()
	raw := serpapi.New(cfg.Providers[config.ProviderSerpAPI], httpClient)
	cached := cache.Wrap(raw, ".cache/flights", 0)
	if err := registry.Register(cached); err != nil {
		return fmt.Errorf("registering serpapi: %w", err)
	}
	ranker := multicity.NewRanker(llmClient, multicity.WeightsBudget)
	directStrategy := direct.NewSearcher(registry, ranker)
	mcSearcher := multicity.NewSearcher(registry, llmClient, multicity.WeightsBudget)
	mcStrategy := multicity.NewStrategy(mcSearcher, "")
	picker := search.NewPicker(llmClient, directStrategy, mcStrategy)

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
	fmt.Fprintln(out, "Where would you like to fly? (type 'quit' to exit)")

	for {
		fmt.Fprint(out, "\n> ")
		if !scanner.Scan() {
			break
		}
		userInput := strings.TrimSpace(scanner.Text())
		if userInput == "" {
			continue
		}
		if userInput == "quit" || userInput == "exit" {
			fmt.Fprintln(out, "Goodbye!")
			return nil
		}

		history = append(history, llm.Message{Role: llm.RoleUser, Content: userInput})

		response, err := llmClient.ChatCompletion(ctx, history)
		if err != nil {
			fmt.Fprintf(out, "Error: %v\n", err)
			continue
		}

		history = append(history, llm.Message{Role: llm.RoleAssistant, Content: response})

		// Check if the LLM extracted search parameters.
		params, found := parseTripParams(response)
		if !found {
			fmt.Fprintln(out, response)
			continue
		}

		// Build and execute the search.
		req := buildRequestFromParams(params)
		fmt.Fprintf(out, "Searching %s -> %s on %s...\n", req.Origin, req.Destination, req.DepartureDate)

		strategy, err := picker.Pick(ctx, req)
		if err != nil {
			fmt.Fprintf(out, "Error picking strategy: %v\n", err)
			continue
		}

		results, err := strategy.Search(ctx, req)
		if err != nil {
			fmt.Fprintf(out, "Search error: %v\n", err)
			continue
		}

		if len(results) == 0 {
			fmt.Fprintln(out, "No flights found. Try different dates or airports.")
			continue
		}

		cur := viper.GetString(keyCurrency)
		switch viper.GetString(keyFormat) {
		case "json":
			if err := printJSON(results, cur); err != nil {
				fmt.Fprintf(out, "Error: %v\n", err)
			}
		default:
			printTable(results, cur)
		}

		fmt.Fprintln(out, "Want to refine? (e.g., 'show cheaper', 'try business class', or 'quit')")
	}

	return scanner.Err()
}
