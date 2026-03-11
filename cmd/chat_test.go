package cmd

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"booker/llm"
	"booker/search"
	"booker/types"
)

func TestParseTripParams_ValidJSON(t *testing.T) {
	input := `Based on your requirements, here are the parameters:
{"origin":"DEL","destination":"YYZ","departure_date":"2025-06-15","passengers":2,"cabin":"economy"}
I'll search for flights now.`

	params, ok := parseTripParams(input)
	if !ok {
		t.Fatal("expected to find trip params")
	}
	if params.Origin != "DEL" {
		t.Errorf("Origin = %q, want %q", params.Origin, "DEL")
	}
	if params.Destination != "YYZ" {
		t.Errorf("Destination = %q, want %q", params.Destination, "YYZ")
	}
	if params.DepartureDate != "2025-06-15" {
		t.Errorf("DepartureDate = %q, want %q", params.DepartureDate, "2025-06-15")
	}
	if params.Passengers != 2 {
		t.Errorf("Passengers = %d, want %d", params.Passengers, 2)
	}
	if params.Cabin != "economy" {
		t.Errorf("Cabin = %q, want %q", params.Cabin, "economy")
	}
}

func TestParseTripParams_InCodeFence(t *testing.T) {
	input := "Here are your trip details:\n```json\n{\"origin\":\"BOM\",\"destination\":\"YVR\",\"departure_date\":\"2025-07-01\",\"passengers\":1,\"cabin\":\"business\"}\n```\nSearching now."

	params, ok := parseTripParams(input)
	if !ok {
		t.Fatal("expected to find trip params in code fence")
	}
	if params.Origin != "BOM" {
		t.Errorf("Origin = %q, want %q", params.Origin, "BOM")
	}
	if params.Cabin != "business" {
		t.Errorf("Cabin = %q, want %q", params.Cabin, "business")
	}
}

func TestParseTripParams_NoJSON(t *testing.T) {
	input := "Can you tell me where you'd like to fly? Which airport are you departing from?"

	_, ok := parseTripParams(input)
	if ok {
		t.Error("expected no trip params in conversational response")
	}
}

func TestParseTripParams_IncompleteJSON(t *testing.T) {
	// Missing destination -- should not match.
	input := `{"origin":"DEL","passengers":1}`

	_, ok := parseTripParams(input)
	if ok {
		t.Error("expected no match for incomplete params (missing destination and date)")
	}
}

func TestParseTripParams_Defaults(t *testing.T) {
	input := `{"origin":"DEL","destination":"YYZ","departure_date":"2025-06-15"}`

	params, ok := parseTripParams(input)
	if !ok {
		t.Fatal("expected to find trip params")
	}
	if params.Passengers != 0 {
		t.Errorf("Passengers = %d, want 0 (unset)", params.Passengers)
	}
	if params.Cabin != "" {
		t.Errorf("Cabin = %q, want empty (unset)", params.Cabin)
	}
}

func TestBuildRequestFromParams(t *testing.T) {
	params := tripParams{
		Origin:        "DEL",
		Destination:   "YYZ",
		DepartureDate: "2025-06-15",
		ReturnDate:    "2025-06-25",
		Passengers:    2,
		Cabin:         "business",
	}
	req := buildRequestFromParams(params)
	if req.Origin != "DEL" {
		t.Errorf("Origin = %q, want %q", req.Origin, "DEL")
	}
	if req.Destination != "YYZ" {
		t.Errorf("Destination = %q, want %q", req.Destination, "YYZ")
	}
	if req.DepartureDate != "2025-06-15" {
		t.Errorf("DepartureDate = %q, want %q", req.DepartureDate, "2025-06-15")
	}
	if req.ReturnDate != "2025-06-25" {
		t.Errorf("ReturnDate = %q, want %q", req.ReturnDate, "2025-06-25")
	}
	if req.Passengers != 2 {
		t.Errorf("Passengers = %d, want %d", req.Passengers, 2)
	}
	if req.CabinClass != "business" {
		t.Errorf("CabinClass = %q, want %q", req.CabinClass, "business")
	}
}

func TestBuildRequestFromParams_Defaults(t *testing.T) {
	params := tripParams{
		Origin:        "DEL",
		Destination:   "YYZ",
		DepartureDate: "2025-06-15",
	}
	req := buildRequestFromParams(params)
	if req.Passengers != defaultPassengers {
		t.Errorf("Passengers = %d, want %d (default)", req.Passengers, defaultPassengers)
	}
	if string(req.CabinClass) != defaultCabin {
		t.Errorf("CabinClass = %q, want %q (default)", req.CabinClass, defaultCabin)
	}
	if req.MaxResults != defaultMaxResults {
		t.Errorf("MaxResults = %d, want %d (default)", req.MaxResults, defaultMaxResults)
	}
	if req.FlexDays != defaultFlexDays {
		t.Errorf("FlexDays = %d, want %d (default)", req.FlexDays, defaultFlexDays)
	}
}

func TestChatLoop_QuitImmediately(t *testing.T) {
	mock := &chatMockLLM{}
	picker := search.NewPicker(mock)

	in := strings.NewReader("quit\n")
	var out strings.Builder

	err := chatLoop(context.Background(), mock, picker, in, &out)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out.String(), "Goodbye!") {
		t.Error("expected goodbye message")
	}
}

func TestChatLoop_ConversationThenSearch(t *testing.T) {
	responses := []string{
		"Where would you like to go?",
		`Great! Here are your trip details:
{"origin":"DEL","destination":"YYZ","departure_date":"2025-06-15","passengers":1,"cabin":"economy"}
Searching for flights now.`,
	}
	mock := &chatMockLLM{responses: responses}
	// Use a fakeSearchStrategy so Search returns results.
	fakeStrat := &fakeSearchStrategy{
		results: []search.Itinerary{
			{
				Legs:       []search.Leg{{Flight: types.Flight{Price: types.Money{Amount: 800, Currency: "USD"}, Outbound: []types.Segment{{Origin: "DEL", Destination: "YYZ"}}}}},
				TotalPrice: types.Money{Amount: 800, Currency: "USD"},
			},
		},
	}
	picker := search.NewPicker(mock, fakeStrat)

	in := strings.NewReader("I want to go to Toronto\nfrom Delhi on June 15\nquit\n")
	var out strings.Builder

	err := chatLoop(context.Background(), mock, picker, in, &out)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	output := out.String()
	if !strings.Contains(output, "Searching DEL -> YYZ") {
		t.Errorf("expected search execution message, got:\n%s", output)
	}
}

// chatMockLLM returns pre-set responses in order.
type chatMockLLM struct {
	responses      []string
	idx            int
	captureHistory bool
	historyLog     [][]llm.Message // recorded history per call
}

func (m *chatMockLLM) ChatCompletion(_ context.Context, msgs []llm.Message) (string, error) {
	if m.captureHistory {
		cp := make([]llm.Message, len(msgs))
		copy(cp, msgs)
		m.historyLog = append(m.historyLog, cp)
	}
	if m.idx >= len(m.responses) {
		return "I don't understand.", nil
	}
	resp := m.responses[m.idx]
	m.idx++
	return resp, nil
}

// fakeSearchStrategy is a test double that returns canned results.
type fakeSearchStrategy struct {
	results []search.Itinerary
}

func (f *fakeSearchStrategy) Name() string        { return "direct" }
func (f *fakeSearchStrategy) Description() string { return "fake" }
func (f *fakeSearchStrategy) Search(_ context.Context, _ search.Request) ([]search.Itinerary, error) {
	return f.results, nil
}

func TestChatLoop_ResultSummaryInHistory(t *testing.T) {
	// After search results are shown, the LLM should receive a result summary
	// in the conversation history so it can help the user refine.
	responses := []string{
		`{"origin":"DEL","destination":"YYZ","departure_date":"2025-06-15"}
Searching for flights.`,
		"The cheapest option is $500 on Air Canada.",
	}
	mock := &chatMockLLM{responses: responses, captureHistory: true}
	fakeStrat := &fakeSearchStrategy{
		results: []search.Itinerary{
			{
				Legs:       []search.Leg{{Flight: types.Flight{Price: types.Money{Amount: 500, Currency: "USD"}, Outbound: []types.Segment{{Origin: "DEL", Destination: "YYZ"}}}}},
				TotalPrice: types.Money{Amount: 500, Currency: "USD"},
			},
			{
				Legs:       []search.Leg{{Flight: types.Flight{Price: types.Money{Amount: 800, Currency: "USD"}, Outbound: []types.Segment{{Origin: "DEL", Destination: "YYZ"}}}}},
				TotalPrice: types.Money{Amount: 800, Currency: "USD"},
			},
		},
	}
	picker := search.NewPicker(mock, fakeStrat)

	// First message triggers search, second is a refinement request, third quits.
	in := strings.NewReader("find flights\nshow me cheaper\nquit\n")
	var out strings.Builder

	err := chatLoop(context.Background(), mock, picker, in, &out)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// The second LLM call (for "show me cheaper") should have a result summary in history.
	if len(mock.historyLog) < 2 {
		t.Fatalf("expected at least 2 LLM calls, got %d", len(mock.historyLog))
	}
	secondCallHistory := mock.historyLog[1]
	found := false
	for _, msg := range secondCallHistory {
		if msg.Role == "assistant" && strings.Contains(msg.Content, "2 results") && strings.Contains(msg.Content, "500") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected result summary in second LLM call history, got messages:\n")
		for _, msg := range secondCallHistory {
			t.Errorf("  [%s] %s", msg.Role, msg.Content[:min(len(msg.Content), 100)])
		}
	}
}

func TestResultSummaryForChat(t *testing.T) {
	itins := []search.Itinerary{
		{
			Legs: []search.Leg{{Flight: types.Flight{
				Price:         types.Money{Amount: 500, Currency: "USD"},
				TotalDuration: 14*time.Hour + 30*time.Minute,
				Outbound:      []types.Segment{{Origin: "DEL", Destination: "YYZ", Airline: "AC", AirlineName: "Air Canada"}},
			}}},
			TotalPrice: types.Money{Amount: 500, Currency: "USD"},
		},
		{TotalPrice: types.Money{Amount: 800, Currency: "USD"}},
		{TotalPrice: types.Money{Amount: 1200, Currency: "USD"}},
	}
	params := tripParams{Origin: "DEL", Destination: "YYZ", DepartureDate: "2025-06-15", Cabin: "economy"}
	summary := resultSummaryForChat(itins, params)

	// Price range and count.
	if !strings.Contains(summary, "3") {
		t.Errorf("summary should contain result count, got: %s", summary)
	}
	if !strings.Contains(summary, "500") {
		t.Errorf("summary should contain cheapest price, got: %s", summary)
	}
	if !strings.Contains(summary, "1200") {
		t.Errorf("summary should contain most expensive price, got: %s", summary)
	}
	// Top result details.
	if !strings.Contains(summary, "DEL") || !strings.Contains(summary, "YYZ") {
		t.Errorf("summary should contain route, got: %s", summary)
	}
	if !strings.Contains(summary, "Air Canada") {
		t.Errorf("summary should contain airline name, got: %s", summary)
	}
	if !strings.Contains(summary, "14h30m") {
		t.Errorf("summary should contain duration, got: %s", summary)
	}
	// Search params context.
	if !strings.Contains(summary, "2025-06-15") {
		t.Errorf("summary should contain departure date, got: %s", summary)
	}
	if !strings.Contains(summary, "economy") {
		t.Errorf("summary should contain cabin class, got: %s", summary)
	}
}

func TestResultSummaryForChat_Empty(t *testing.T) {
	summary := resultSummaryForChat(nil, tripParams{})
	if !strings.Contains(summary, "No results") {
		t.Errorf("expected no-results message, got: %s", summary)
	}
}

func TestChatSystemPrompt(t *testing.T) {
	prompt := chatSystemPrompt()
	if prompt == "" {
		t.Fatal("system prompt should not be empty")
	}
	// Must instruct LLM to return JSON with required fields.
	for _, field := range []string{"origin", "destination", "departure_date"} {
		if !contains(prompt, field) {
			t.Errorf("system prompt missing required field %q", field)
		}
	}
	// Must mention nearby airports so the LLM can suggest alternatives.
	if !strings.Contains(prompt, "nearby") {
		t.Error("system prompt should mention nearby airports")
	}
	// Must mention direct_only optional field.
	if !strings.Contains(prompt, "direct_only") {
		t.Error("system prompt should mention direct_only option")
	}
}

func TestNearbyAirportHint(t *testing.T) {
	// JFK has nearby airports (EWR, LGA).
	hint := nearbyAirportHint("JFK", "YYZ")
	if hint == "" {
		t.Fatal("expected hint for JFK (NYC metro)")
	}
	if !strings.Contains(hint, "EWR") || !strings.Contains(hint, "LGA") {
		t.Errorf("hint should mention EWR and LGA, got: %s", hint)
	}

	// DEL and BOM have no nearby airports -- hint should be empty.
	hint = nearbyAirportHint("DEL", "BOM")
	if hint != "" {
		t.Errorf("expected empty hint for DEL->BOM (single airports), got: %s", hint)
	}

	// Both have nearby airports.
	hint = nearbyAirportHint("JFK", "LHR")
	if !strings.Contains(hint, "EWR") || !strings.Contains(hint, "LGW") {
		t.Errorf("hint should mention nearby airports for both origin and destination, got: %s", hint)
	}
}

func TestRefinementHint(t *testing.T) {
	hint := refinementHint()
	for _, lever := range []string{"dates", "nearby airports", "cabin class", "direct flights"} {
		if !strings.Contains(hint, lever) {
			t.Errorf("refinement hint missing lever %q, got: %s", lever, hint)
		}
	}
}

func TestChatLoop_RefinementHintInHistory(t *testing.T) {
	responses := []string{
		`{"origin":"DEL","destination":"YYZ","departure_date":"2025-06-15"}
Searching for flights.`,
		"Here are some options to consider.",
	}
	mock := &chatMockLLM{responses: responses, captureHistory: true}
	fakeStrat := &fakeSearchStrategy{
		results: []search.Itinerary{
			{
				Legs:       []search.Leg{{Flight: types.Flight{Price: types.Money{Amount: 500, Currency: "USD"}, Outbound: []types.Segment{{Origin: "DEL", Destination: "YYZ"}}}}},
				TotalPrice: types.Money{Amount: 500, Currency: "USD"},
			},
		},
	}
	picker := search.NewPicker(mock, fakeStrat)

	in := strings.NewReader("find flights\nshow me options\nquit\n")
	var out strings.Builder

	err := chatLoop(context.Background(), mock, picker, in, &out)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// The second LLM call should have a refinement hint in history.
	if len(mock.historyLog) < 2 {
		t.Fatalf("expected at least 2 LLM calls, got %d", len(mock.historyLog))
	}
	secondCallHistory := mock.historyLog[1]
	found := false
	for _, msg := range secondCallHistory {
		if msg.Role == "system" && strings.Contains(msg.Content, "nearby airports") && strings.Contains(msg.Content, "cabin class") {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected refinement hint with levers in second LLM call history")
		for _, msg := range secondCallHistory {
			t.Logf("  [%s] %s", msg.Role, msg.Content[:min(len(msg.Content), 120)])
		}
	}
}

func TestMergeParams(t *testing.T) {
	tests := []struct {
		name    string
		prev    tripParams
		partial tripParams
		want    tripParams
	}{
		{
			name:    "full merge from empty prev",
			prev:    tripParams{},
			partial: tripParams{Origin: "DEL", Destination: "YYZ", DepartureDate: "2025-06-15"},
			want:    tripParams{Origin: "DEL", Destination: "YYZ", DepartureDate: "2025-06-15"},
		},
		{
			name:    "partial cabin override",
			prev:    tripParams{Origin: "DEL", Destination: "YYZ", DepartureDate: "2025-06-15", Cabin: "economy"},
			partial: tripParams{Cabin: "business"},
			want:    tripParams{Origin: "DEL", Destination: "YYZ", DepartureDate: "2025-06-15", Cabin: "business"},
		},
		{
			name:    "partial date override",
			prev:    tripParams{Origin: "DEL", Destination: "YYZ", DepartureDate: "2025-06-15", Cabin: "economy"},
			partial: tripParams{DepartureDate: "2025-07-01"},
			want:    tripParams{Origin: "DEL", Destination: "YYZ", DepartureDate: "2025-07-01", Cabin: "economy"},
		},
		{
			name:    "zero-value fields preserved from prev",
			prev:    tripParams{Origin: "DEL", Destination: "YYZ", DepartureDate: "2025-06-15", Passengers: 2, MaxPrice: 1200},
			partial: tripParams{Cabin: "business"},
			want:    tripParams{Origin: "DEL", Destination: "YYZ", DepartureDate: "2025-06-15", Passengers: 2, MaxPrice: 1200, Cabin: "business"},
		},
		{
			name:    "partial max_price override",
			prev:    tripParams{Origin: "DEL", Destination: "YYZ", DepartureDate: "2025-06-15", MaxPrice: 1200},
			partial: tripParams{MaxPrice: 800},
			want:    tripParams{Origin: "DEL", Destination: "YYZ", DepartureDate: "2025-06-15", MaxPrice: 800},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mergeParams(tt.prev, tt.partial)
			if got != tt.want {
				t.Errorf("mergeParams() =\n  %+v\nwant\n  %+v", got, tt.want)
			}
		})
	}
}

func TestParsePartialParams(t *testing.T) {
	// Full JSON should parse.
	_, ok := parsePartialParams(`{"origin":"DEL","destination":"YYZ","departure_date":"2025-06-15"}`)
	if !ok {
		t.Error("expected full JSON to parse as partial")
	}

	// Partial JSON with only cabin should parse.
	p, ok := parsePartialParams(`{"cabin":"business"}`)
	if !ok {
		t.Error("expected partial JSON with cabin to parse")
	}
	if p.Cabin != "business" {
		t.Errorf("Cabin = %q, want %q", p.Cabin, "business")
	}

	// Empty JSON should NOT parse (no recognized fields set).
	_, ok = parsePartialParams(`{}`)
	if ok {
		t.Error("expected empty JSON to not parse as partial")
	}

	// Non-JSON should NOT parse.
	_, ok = parsePartialParams("Can you try a different date?")
	if ok {
		t.Error("expected non-JSON to not parse as partial")
	}
}

func TestChatLoop_FollowUpPartialParams(t *testing.T) {
	// First response: full params -> triggers search.
	// Second response: partial params (cabin only) -> should merge and re-search.
	responses := []string{
		`{"origin":"DEL","destination":"YYZ","departure_date":"2025-06-15","cabin":"economy"}
Searching for flights.`,
		`{"cabin":"business"}
Switching to business class.`,
	}
	mock := &chatMockLLM{responses: responses, captureHistory: true}
	fakeStrat := &fakeSearchStrategy{
		results: []search.Itinerary{
			{
				Legs:       []search.Leg{{Flight: types.Flight{Price: types.Money{Amount: 800, Currency: "USD"}, Outbound: []types.Segment{{Origin: "DEL", Destination: "YYZ", Airline: "AC"}}}}},
				TotalPrice: types.Money{Amount: 800, Currency: "USD"},
			},
		},
	}
	picker := search.NewPicker(mock, fakeStrat)

	in := strings.NewReader("find flights\ntry business class\nquit\n")
	var out strings.Builder

	err := chatLoop(context.Background(), mock, picker, in, &out)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := out.String()
	// First search: economy.
	if !strings.Contains(output, "Searching DEL -> YYZ") {
		t.Errorf("expected first search, got:\n%s", output)
	}
	// Second search should still be DEL -> YYZ (merged from prev).
	// Count occurrences of "Searching DEL -> YYZ".
	count := strings.Count(output, "Searching DEL -> YYZ")
	if count < 2 {
		t.Errorf("expected 2 searches for DEL -> YYZ, got %d in:\n%s", count, output)
	}
}

func TestBuildRequestFromParams_DirectOnly(t *testing.T) {
	params := tripParams{
		Origin:        "DEL",
		Destination:   "YYZ",
		DepartureDate: "2025-06-15",
		DirectOnly:    true,
	}
	req := buildRequestFromParams(params)
	if req.MaxStops != 0 {
		t.Errorf("MaxStops = %d, want 0 when DirectOnly=true", req.MaxStops)
	}
}

func TestBuildRequestFromParams_DirectOnlyFalse(t *testing.T) {
	params := tripParams{
		Origin:        "DEL",
		Destination:   "YYZ",
		DepartureDate: "2025-06-15",
		DirectOnly:    false,
	}
	req := buildRequestFromParams(params)
	if req.MaxStops != defaultMaxStops {
		t.Errorf("MaxStops = %d, want %d when DirectOnly=false", req.MaxStops, defaultMaxStops)
	}
}

func TestParsePartialParams_DirectOnly(t *testing.T) {
	p, ok := parsePartialParams(`{"direct_only":true}`)
	if !ok {
		t.Fatal("expected partial JSON with direct_only to parse")
	}
	if !p.DirectOnly {
		t.Error("DirectOnly should be true")
	}
}

func TestMergeParams_DirectOnly(t *testing.T) {
	prev := tripParams{Origin: "DEL", Destination: "YYZ", DepartureDate: "2025-06-15"}
	partial := tripParams{DirectOnly: true}
	got := mergeParams(prev, partial)
	if !got.DirectOnly {
		t.Error("expected DirectOnly=true to be preserved after merge")
	}
	if got.Origin != "DEL" {
		t.Errorf("Origin = %q, want %q", got.Origin, "DEL")
	}

	// When prev has DirectOnly=true and partial doesn't set it, it should carry forward.
	prev2 := tripParams{Origin: "DEL", Destination: "YYZ", DepartureDate: "2025-06-15", DirectOnly: true}
	partial2 := tripParams{Cabin: "business"}
	got2 := mergeParams(prev2, partial2)
	if !got2.DirectOnly {
		t.Error("expected DirectOnly=true to carry forward from prev")
	}
}

func TestChatLoop_OutputContainsResultData(t *testing.T) {
	// Verify that chatLoop writes search result data (price, route) to the
	// provided io.Writer, not to os.Stdout.
	responses := []string{
		`{"origin":"DEL","destination":"YYZ","departure_date":"2025-06-15"}
Searching for flights.`,
	}
	mock := &chatMockLLM{responses: responses}
	fakeStrat := &fakeSearchStrategy{
		results: []search.Itinerary{
			{
				Legs: []search.Leg{{Flight: types.Flight{
					Price:    types.Money{Amount: 750, Currency: "USD"},
					Outbound: []types.Segment{{Origin: "DEL", Destination: "YYZ", Airline: "AC", DepartureTime: time.Date(2025, 6, 15, 8, 0, 0, 0, time.UTC), ArrivalTime: time.Date(2025, 6, 15, 20, 0, 0, 0, time.UTC)}},
				}}},
				TotalPrice: types.Money{Amount: 750, Currency: "USD"},
				Score:      90,
			},
		},
	}
	picker := search.NewPicker(mock, fakeStrat)

	in := strings.NewReader("find flights\nquit\n")
	var out strings.Builder

	err := chatLoop(context.Background(), mock, picker, in, &out)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := out.String()
	// The table output (price, route) must appear in the writer, not just os.Stdout.
	if !strings.Contains(output, "DEL") || !strings.Contains(output, "YYZ") {
		t.Errorf("chatLoop output missing route data, got:\n%s", output)
	}
	// Price is converted to display currency (CAD by default), so check for
	// the table structure rather than the raw USD amount.
	if !strings.Contains(output, "PRICE") {
		t.Errorf("chatLoop output missing table header, got:\n%s", output)
	}
}

func TestProfileWeights(t *testing.T) {
	tests := []struct {
		name    string
		profile string
		want    string // check a distinctive weight to verify correct profile
	}{
		{"budget", "budget", "budget"},
		{"comfort", "comfort", "comfort"},
		{"balanced", "balanced", "balanced"},
		{"unknown defaults to budget", "deluxe", "budget"},
		{"empty defaults to budget", "", "budget"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := profileWeights(tt.profile)
			want := profileWeights(tt.want)
			if got != want {
				t.Errorf("profileWeights(%q) != profileWeights(%q)", tt.profile, tt.want)
			}
		})
	}
}

func TestTruncateHistory_PreservesSystemPrompt(t *testing.T) {
	// Build history: system prompt + 30 user/assistant messages.
	history := []llm.Message{
		{Role: llm.RoleSystem, Content: "You are a flight assistant."},
	}
	for i := 0; i < 30; i++ {
		history = append(history, llm.Message{
			Role:    llm.RoleUser,
			Content: fmt.Sprintf("message %d", i),
		})
	}

	got := truncateHistory(history, 20)

	// Should be system prompt + last 20 messages = 21 total.
	if len(got) != 21 {
		t.Fatalf("len = %d, want 21", len(got))
	}
	if got[0].Role != llm.RoleSystem {
		t.Errorf("first message role = %q, want %q", got[0].Role, llm.RoleSystem)
	}
	if got[0].Content != "You are a flight assistant." {
		t.Errorf("system prompt content changed")
	}
	// First non-system message should be message 10 (skipped 0-9).
	if got[1].Content != "message 10" {
		t.Errorf("got[1].Content = %q, want %q", got[1].Content, "message 10")
	}
	// Last message should be message 29.
	if got[20].Content != "message 29" {
		t.Errorf("got[20].Content = %q, want %q", got[20].Content, "message 29")
	}
}

func TestTruncateHistory_ShortHistoryUnchanged(t *testing.T) {
	history := []llm.Message{
		{Role: llm.RoleSystem, Content: "system prompt"},
		{Role: llm.RoleUser, Content: "hello"},
		{Role: llm.RoleAssistant, Content: "hi"},
		{Role: llm.RoleUser, Content: "search"},
		{Role: llm.RoleAssistant, Content: "results"},
		{Role: llm.RoleUser, Content: "thanks"},
	}

	got := truncateHistory(history, 20)

	if len(got) != len(history) {
		t.Fatalf("len = %d, want %d (unchanged)", len(got), len(history))
	}
	for i := range history {
		if got[i] != history[i] {
			t.Errorf("message[%d] changed: got %+v, want %+v", i, got[i], history[i])
		}
	}
}

func TestChatLoop_HistoryTruncation(t *testing.T) {
	// Generate enough LLM responses to build >20 non-system messages.
	// Each search cycle adds: user, assistant (with JSON), assistant (summary), system (hint).
	// So 6 search cycles = 6 user + 6 assistant + 6 summary + 6 hint = 24 non-system messages.
	// Plus 1 extra user+assistant for final conversational turn = 26 non-system.
	var responses []string
	for i := 0; i < 6; i++ {
		responses = append(responses, fmt.Sprintf(
			`{"origin":"DEL","destination":"YYZ","departure_date":"2025-06-%02d"}
Searching.`, 15+i))
	}
	// Final conversational response (no JSON).
	responses = append(responses, "Anything else I can help with?")

	mock := &chatMockLLM{responses: responses, captureHistory: true}
	fakeStrat := &fakeSearchStrategy{
		results: []search.Itinerary{
			{
				Legs:       []search.Leg{{Flight: types.Flight{Price: types.Money{Amount: 500, Currency: "USD"}, Outbound: []types.Segment{{Origin: "DEL", Destination: "YYZ"}}}}},
				TotalPrice: types.Money{Amount: 500, Currency: "USD"},
			},
		},
	}
	picker := search.NewPicker(mock, fakeStrat)

	// 6 search inputs + 1 conversational + quit.
	var inputLines []string
	for i := 0; i < 6; i++ {
		inputLines = append(inputLines, fmt.Sprintf("search %d", i))
	}
	inputLines = append(inputLines, "what else?", "quit")
	in := strings.NewReader(strings.Join(inputLines, "\n") + "\n")
	var out strings.Builder

	err := chatLoop(context.Background(), mock, picker, in, &out)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// The last LLM call should have at most maxHistoryMessages+1 messages
	// (1 system + maxHistoryMessages recent).
	lastCall := mock.historyLog[len(mock.historyLog)-1]
	maxAllowed := maxHistoryMessages + 1
	if len(lastCall) > maxAllowed {
		t.Errorf("last LLM call history len = %d, want <= %d", len(lastCall), maxAllowed)
	}
	// System prompt must still be first.
	if lastCall[0].Role != llm.RoleSystem {
		t.Errorf("first message role = %q, want %q", lastCall[0].Role, llm.RoleSystem)
	}
}

func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && indexOf(s, substr) >= 0
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
