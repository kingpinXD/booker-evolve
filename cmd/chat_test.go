package cmd

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"booker/llm"
	"booker/search"
	"booker/search/multicity"
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

	err := chatLoop(context.Background(), mock, picker, nil, nil, in, &out)
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

	err := chatLoop(context.Background(), mock, picker, nil, nil, in, &out)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	output := out.String()
	if !strings.Contains(output, "Searching DEL -> YYZ") {
		t.Errorf("expected search execution message, got:\n%s", output)
	}
	if !strings.Contains(output, "Using direct strategy") {
		t.Errorf("expected strategy reason in output, got:\n%s", output)
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

	err := chatLoop(context.Background(), mock, picker, nil, nil, in, &out)
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

func TestResultSummaryForChat_Top3(t *testing.T) {
	// 5 results: summary should show top 3 with price, airline, duration, stops.
	makeSummaryItin := func(airline, airlineName, origin, dest string, dur time.Duration, price float64, segments int) search.Itinerary {
		segs := make([]types.Segment, segments)
		segs[0] = types.Segment{Origin: origin, Destination: dest, Airline: airline, AirlineName: airlineName}
		for i := 1; i < segments; i++ {
			segs[i] = types.Segment{Airline: airline}
		}
		return search.Itinerary{
			Legs: []search.Leg{{Flight: types.Flight{
				Price:         types.Money{Amount: price, Currency: "USD"},
				TotalDuration: dur,
				Outbound:      segs,
			}}},
			TotalPrice: types.Money{Amount: price, Currency: "USD"},
		}
	}

	itins := []search.Itinerary{
		makeSummaryItin("AC", "Air Canada", "DEL", "YYZ", 14*time.Hour+30*time.Minute, 500, 1),
		makeSummaryItin("BA", "British Airways", "DEL", "YYZ", 16*time.Hour, 650, 2),
		makeSummaryItin("LH", "Lufthansa", "DEL", "YYZ", 18*time.Hour, 700, 1),
		makeSummaryItin("EK", "Emirates", "DEL", "YYZ", 20*time.Hour, 900, 3),
		makeSummaryItin("QR", "Qatar Airways", "DEL", "YYZ", 22*time.Hour, 1200, 1),
	}
	params := tripParams{Origin: "DEL", Destination: "YYZ", DepartureDate: "2025-06-15", Cabin: "economy"}
	summary := resultSummaryForChat(itins, params)

	// Price range and count.
	if !strings.Contains(summary, "5") {
		t.Errorf("summary should contain result count 5, got: %s", summary)
	}
	if !strings.Contains(summary, "500") {
		t.Errorf("summary should contain min price, got: %s", summary)
	}
	if !strings.Contains(summary, "1200") {
		t.Errorf("summary should contain max price, got: %s", summary)
	}

	// Top 3 results should each have airline, duration, stops, and price.
	// Result 1: Air Canada, 14h30m, 0 stops, $500
	if !strings.Contains(summary, "Air Canada") {
		t.Errorf("summary should contain 1st airline, got: %s", summary)
	}
	if !strings.Contains(summary, "14h30m") {
		t.Errorf("summary should contain 1st duration, got: %s", summary)
	}
	if !strings.Contains(summary, "0 stop") {
		t.Errorf("summary should contain 0 stops for nonstop flight, got: %s", summary)
	}

	// Result 2: British Airways, 16h, 1 stop, $650
	if !strings.Contains(summary, "British Airways") {
		t.Errorf("summary should contain 2nd airline, got: %s", summary)
	}
	if !strings.Contains(summary, "16h") {
		t.Errorf("summary should contain 2nd duration, got: %s", summary)
	}
	if !strings.Contains(summary, "1 stop") {
		t.Errorf("summary should contain 1 stop, got: %s", summary)
	}

	// Result 3: Lufthansa, 18h, 0 stops, $700
	if !strings.Contains(summary, "Lufthansa") {
		t.Errorf("summary should contain 3rd airline, got: %s", summary)
	}
	if !strings.Contains(summary, "18h") {
		t.Errorf("summary should contain 3rd duration, got: %s", summary)
	}
	if !strings.Contains(summary, "700") {
		t.Errorf("summary should contain 3rd price, got: %s", summary)
	}

	// 4th and 5th should NOT appear.
	if strings.Contains(summary, "Emirates") {
		t.Errorf("summary should not contain 4th result airline, got: %s", summary)
	}
	if strings.Contains(summary, "Qatar") {
		t.Errorf("summary should not contain 5th result airline, got: %s", summary)
	}

	// Search params context still present.
	if !strings.Contains(summary, "2025-06-15") {
		t.Errorf("summary should contain departure date, got: %s", summary)
	}
	if !strings.Contains(summary, "economy") {
		t.Errorf("summary should contain cabin class, got: %s", summary)
	}
}

func TestResultSummaryForChat_TwoResults(t *testing.T) {
	// 2 results: should show both, not crash looking for a 3rd.
	itin1 := search.Itinerary{
		Legs: []search.Leg{{Flight: types.Flight{
			Price:         types.Money{Amount: 500, Currency: "USD"},
			TotalDuration: 14 * time.Hour,
			Outbound:      []types.Segment{{Origin: "DEL", Destination: "YYZ", Airline: "AC", AirlineName: "Air Canada"}},
		}}},
		TotalPrice: types.Money{Amount: 500, Currency: "USD"},
	}
	itin2 := search.Itinerary{
		Legs: []search.Leg{{Flight: types.Flight{
			Price:         types.Money{Amount: 800, Currency: "USD"},
			TotalDuration: 18 * time.Hour,
			Outbound: []types.Segment{
				{Origin: "DEL", Destination: "LHR", Airline: "BA", AirlineName: "British Airways"},
				{Origin: "LHR", Destination: "YYZ", Airline: "BA"},
			},
		}}},
		TotalPrice: types.Money{Amount: 800, Currency: "USD"},
	}
	params := tripParams{Origin: "DEL", Destination: "YYZ", DepartureDate: "2025-06-15"}
	summary := resultSummaryForChat([]search.Itinerary{itin1, itin2}, params)

	if !strings.Contains(summary, "2 results") {
		t.Errorf("summary should say 2 results, got: %s", summary)
	}
	if !strings.Contains(summary, "Air Canada") {
		t.Errorf("summary should contain 1st airline, got: %s", summary)
	}
	if !strings.Contains(summary, "British Airways") {
		t.Errorf("summary should contain 2nd airline, got: %s", summary)
	}
}

func TestResultSummaryForChat_OneResult(t *testing.T) {
	itin := search.Itinerary{
		Legs: []search.Leg{{Flight: types.Flight{
			Price:         types.Money{Amount: 500, Currency: "USD"},
			TotalDuration: 14 * time.Hour,
			Outbound:      []types.Segment{{Origin: "DEL", Destination: "YYZ", Airline: "AC", AirlineName: "Air Canada"}},
		}}},
		TotalPrice: types.Money{Amount: 500, Currency: "USD"},
	}
	params := tripParams{Origin: "DEL", Destination: "YYZ", DepartureDate: "2025-06-15"}
	summary := resultSummaryForChat([]search.Itinerary{itin}, params)

	if !strings.Contains(summary, "1 result") {
		t.Errorf("summary should say 1 result, got: %s", summary)
	}
	if !strings.Contains(summary, "Air Canada") {
		t.Errorf("summary should contain airline, got: %s", summary)
	}
	if !strings.Contains(summary, "14h") {
		t.Errorf("summary should contain duration, got: %s", summary)
	}
	if !strings.Contains(summary, "0 stop") {
		t.Errorf("summary should contain stops, got: %s", summary)
	}
}

func TestResultSummaryForChat_Empty(t *testing.T) {
	summary := resultSummaryForChat(nil, tripParams{})
	if !strings.Contains(summary, "No results") {
		t.Errorf("expected no-results message, got: %s", summary)
	}
}

func TestChatSystemPrompt(t *testing.T) {
	now := time.Date(2025, 7, 15, 0, 0, 0, 0, time.UTC)
	prompt := chatSystemPrompt(now)
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
	// Must contain the injected date.
	if !strings.Contains(prompt, "2025-07-15") {
		t.Error("system prompt should contain the provided date")
	}
}

func TestChatSystemPrompt_ContainsDate(t *testing.T) {
	now := time.Date(2026, 1, 20, 14, 30, 0, 0, time.UTC)
	prompt := chatSystemPrompt(now)
	if !strings.Contains(prompt, "Today's date is 2026-01-20") {
		t.Errorf("system prompt should contain 'Today's date is 2026-01-20', got start: %s", prompt[:80])
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

	// AMD and HYD have no nearby airports -- hint should be empty.
	hint = nearbyAirportHint("AMD", "HYD")
	if hint != "" {
		t.Errorf("expected empty hint for AMD->HYD (single airports), got: %s", hint)
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

	err := chatLoop(context.Background(), mock, picker, nil, nil, in, &out)
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

	err := chatLoop(context.Background(), mock, picker, nil, nil, in, &out)
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

	err := chatLoop(context.Background(), mock, picker, nil, nil, in, &out)
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

func TestProfileWeights_EcoDistinctFromBudget(t *testing.T) {
	eco := profileWeights("eco")
	budget := profileWeights("budget")
	if eco == budget {
		t.Error("eco profile should differ from budget (carbon-weighted)")
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

	err := chatLoop(context.Background(), mock, picker, nil, nil, in, &out)
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

func TestParsePartialParams_FlexDays(t *testing.T) {
	p, ok := parsePartialParams(`{"flex_days":5}`)
	if !ok {
		t.Fatal("expected partial JSON with flex_days to parse")
	}
	if p.FlexDays != 5 {
		t.Errorf("FlexDays = %d, want 5", p.FlexDays)
	}
}

func TestMergeParams_FlexDays(t *testing.T) {
	// FlexDays=5 from prev is preserved when partial has FlexDays=0.
	prev := tripParams{Origin: "DEL", Destination: "YYZ", DepartureDate: "2025-06-15", FlexDays: 5}
	partial := tripParams{Cabin: "business"}
	got := mergeParams(prev, partial)
	if got.FlexDays != 5 {
		t.Errorf("FlexDays = %d, want 5 (preserved from prev)", got.FlexDays)
	}

	// Partial FlexDays=7 overrides prev FlexDays=5.
	partial2 := tripParams{FlexDays: 7}
	got2 := mergeParams(prev, partial2)
	if got2.FlexDays != 7 {
		t.Errorf("FlexDays = %d, want 7 (overridden by partial)", got2.FlexDays)
	}
}

func TestBuildRequestFromParams_FlexDays(t *testing.T) {
	// FlexDays=5 in params produces req.FlexDays=5.
	params := tripParams{
		Origin:        "DEL",
		Destination:   "YYZ",
		DepartureDate: "2025-06-15",
		FlexDays:      5,
	}
	req := buildRequestFromParams(params)
	if req.FlexDays != 5 {
		t.Errorf("FlexDays = %d, want 5", req.FlexDays)
	}

	// FlexDays=0 produces req.FlexDays=defaultFlexDays (3).
	params2 := tripParams{
		Origin:        "DEL",
		Destination:   "YYZ",
		DepartureDate: "2025-06-15",
	}
	req2 := buildRequestFromParams(params2)
	if req2.FlexDays != defaultFlexDays {
		t.Errorf("FlexDays = %d, want %d (default)", req2.FlexDays, defaultFlexDays)
	}
}

func TestParseTripParams_PreferredAlliance(t *testing.T) {
	input := `{"origin":"DEL","destination":"YYZ","departure_date":"2025-06-15","preferred_alliance":"Star Alliance"}`
	params, ok := parseTripParams(input)
	if !ok {
		t.Fatal("expected to find trip params")
	}
	if params.PreferredAlliance != "Star Alliance" {
		t.Errorf("PreferredAlliance = %q, want %q", params.PreferredAlliance, "Star Alliance")
	}
}

func TestParsePartialParams_PreferredAlliance(t *testing.T) {
	p, ok := parsePartialParams(`{"preferred_alliance":"OneWorld"}`)
	if !ok {
		t.Fatal("expected partial JSON with preferred_alliance to parse")
	}
	if p.PreferredAlliance != "OneWorld" {
		t.Errorf("PreferredAlliance = %q, want %q", p.PreferredAlliance, "OneWorld")
	}
}

func TestMergeParams_PreferredAlliance(t *testing.T) {
	// PreferredAlliance from prev is preserved when partial has none.
	prev := tripParams{Origin: "DEL", Destination: "YYZ", DepartureDate: "2025-06-15", PreferredAlliance: "Star Alliance"}
	partial := tripParams{Cabin: "business"}
	got := mergeParams(prev, partial)
	if got.PreferredAlliance != "Star Alliance" {
		t.Errorf("PreferredAlliance = %q, want %q (preserved from prev)", got.PreferredAlliance, "Star Alliance")
	}

	// Partial PreferredAlliance overrides prev.
	partial2 := tripParams{PreferredAlliance: "SkyTeam"}
	got2 := mergeParams(prev, partial2)
	if got2.PreferredAlliance != "SkyTeam" {
		t.Errorf("PreferredAlliance = %q, want %q (overridden by partial)", got2.PreferredAlliance, "SkyTeam")
	}
}

func TestBuildRequestFromParams_PreferredAlliance(t *testing.T) {
	params := tripParams{
		Origin:            "DEL",
		Destination:       "YYZ",
		DepartureDate:     "2025-06-15",
		PreferredAlliance: "OneWorld",
	}
	req := buildRequestFromParams(params)
	if req.PreferredAlliance != "OneWorld" {
		t.Errorf("PreferredAlliance = %q, want %q", req.PreferredAlliance, "OneWorld")
	}
}

func TestChatSystemPrompt_PreferredAlliance(t *testing.T) {
	prompt := chatSystemPrompt(time.Date(2025, 7, 15, 0, 0, 0, 0, time.UTC))
	if !strings.Contains(prompt, "preferred_alliance") {
		t.Error("system prompt should mention preferred_alliance")
	}
}

func TestRefinementHint_PreferredAlliance(t *testing.T) {
	hint := refinementHint()
	if !strings.Contains(hint, "preferred_alliance") {
		t.Error("refinement hint should mention preferred_alliance")
	}
}

func TestChatSystemPrompt_FlexDays(t *testing.T) {
	prompt := chatSystemPrompt(time.Date(2025, 7, 15, 0, 0, 0, 0, time.UTC))
	if !strings.Contains(prompt, "flex_days") {
		t.Error("system prompt should mention flex_days")
	}
}

func TestRefinementHint_FlexDays(t *testing.T) {
	hint := refinementHint()
	if !strings.Contains(hint, "flex_days") {
		t.Error("refinement hint should mention flex_days")
	}
}

func TestParsePartialParams_DepartureTime(t *testing.T) {
	p, ok := parsePartialParams(`{"departure_after":"06:00"}`)
	if !ok {
		t.Fatal("expected partial JSON with departure_after to parse")
	}
	if p.DepartureAfter != "06:00" {
		t.Errorf("DepartureAfter = %q, want %q", p.DepartureAfter, "06:00")
	}

	p2, ok := parsePartialParams(`{"departure_before":"22:00"}`)
	if !ok {
		t.Fatal("expected partial JSON with departure_before to parse")
	}
	if p2.DepartureBefore != "22:00" {
		t.Errorf("DepartureBefore = %q, want %q", p2.DepartureBefore, "22:00")
	}
}

func TestMergeParams_DepartureTime(t *testing.T) {
	prev := tripParams{Origin: "DEL", Destination: "YYZ", DepartureDate: "2025-06-15", DepartureAfter: "06:00", DepartureBefore: "22:00"}
	partial := tripParams{Cabin: "business"}
	got := mergeParams(prev, partial)
	if got.DepartureAfter != "06:00" {
		t.Errorf("DepartureAfter = %q, want %q (preserved from prev)", got.DepartureAfter, "06:00")
	}
	if got.DepartureBefore != "22:00" {
		t.Errorf("DepartureBefore = %q, want %q (preserved from prev)", got.DepartureBefore, "22:00")
	}

	// Partial overrides prev.
	partial2 := tripParams{DepartureAfter: "08:00"}
	got2 := mergeParams(prev, partial2)
	if got2.DepartureAfter != "08:00" {
		t.Errorf("DepartureAfter = %q, want %q (overridden by partial)", got2.DepartureAfter, "08:00")
	}
}

func TestBuildRequestFromParams_DepartureTime(t *testing.T) {
	params := tripParams{
		Origin:          "DEL",
		Destination:     "YYZ",
		DepartureDate:   "2025-06-15",
		DepartureAfter:  "06:00",
		DepartureBefore: "22:00",
	}
	req := buildRequestFromParams(params)
	if req.DepartureAfter != "06:00" {
		t.Errorf("DepartureAfter = %q, want %q", req.DepartureAfter, "06:00")
	}
	if req.DepartureBefore != "22:00" {
		t.Errorf("DepartureBefore = %q, want %q", req.DepartureBefore, "22:00")
	}
}

func TestChatSystemPrompt_DepartureTime(t *testing.T) {
	prompt := chatSystemPrompt(time.Date(2025, 7, 15, 0, 0, 0, 0, time.UTC))
	if !strings.Contains(prompt, "departure_after") {
		t.Error("system prompt should mention departure_after")
	}
	if !strings.Contains(prompt, "departure_before") {
		t.Error("system prompt should mention departure_before")
	}
}

func TestParsePartialParams_SortBy(t *testing.T) {
	p, ok := parsePartialParams(`{"sort_by":"duration"}`)
	if !ok {
		t.Fatal("expected partial JSON with sort_by to parse")
	}
	if p.SortBy != "duration" {
		t.Errorf("SortBy = %q, want %q", p.SortBy, "duration")
	}
}

func TestMergeParams_SortBy(t *testing.T) {
	prev := tripParams{Origin: "DEL", Destination: "YYZ", DepartureDate: "2025-06-15", SortBy: "duration"}
	partial := tripParams{Cabin: "business"}
	got := mergeParams(prev, partial)
	if got.SortBy != "duration" {
		t.Errorf("SortBy = %q, want %q (preserved from prev)", got.SortBy, "duration")
	}

	// Partial overrides prev.
	partial2 := tripParams{SortBy: "departure"}
	got2 := mergeParams(prev, partial2)
	if got2.SortBy != "departure" {
		t.Errorf("SortBy = %q, want %q (overridden by partial)", got2.SortBy, "departure")
	}
}

func TestBuildRequestFromParams_SortBy(t *testing.T) {
	params := tripParams{
		Origin:        "DEL",
		Destination:   "YYZ",
		DepartureDate: "2025-06-15",
		SortBy:        "duration",
	}
	req := buildRequestFromParams(params)
	if req.SortBy != "duration" {
		t.Errorf("SortBy = %q, want %q", req.SortBy, "duration")
	}
}

func TestChatSystemPrompt_SortBy(t *testing.T) {
	prompt := chatSystemPrompt(time.Date(2025, 7, 15, 0, 0, 0, 0, time.UTC))
	if !strings.Contains(prompt, "sort_by") {
		t.Error("system prompt should mention sort_by")
	}
}

func TestRefinementHint_SortBy(t *testing.T) {
	hint := refinementHint()
	if !strings.Contains(hint, "sort_by") {
		t.Error("refinement hint should mention sort_by")
	}
}

func TestRefinementHint_DepartureTime(t *testing.T) {
	hint := refinementHint()
	if !strings.Contains(hint, "departure_after") {
		t.Error("refinement hint should mention departure_after")
	}
	if !strings.Contains(hint, "departure_before") {
		t.Error("refinement hint should mention departure_before")
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

// mockPriceInsighter implements priceInsighter for testing.
type mockPriceInsighter struct {
	insights search.PriceInsights
}

func (m *mockPriceInsighter) LastPriceInsights() search.PriceInsights {
	return m.insights
}

func TestParsePartialParams_AvoidAirlines(t *testing.T) {
	p, ok := parsePartialParams(`{"avoid_airlines":"BA,LH"}`)
	if !ok {
		t.Fatal("expected partial JSON with avoid_airlines to parse")
	}
	if p.AvoidAirlines != "BA,LH" {
		t.Errorf("AvoidAirlines = %q, want %q", p.AvoidAirlines, "BA,LH")
	}
}

func TestMergeParams_AvoidAirlines(t *testing.T) {
	prev := tripParams{Origin: "DEL", Destination: "YYZ", DepartureDate: "2025-06-15", AvoidAirlines: "BA,LH"}
	partial := tripParams{Cabin: "business"}
	got := mergeParams(prev, partial)
	if got.AvoidAirlines != "BA,LH" {
		t.Errorf("AvoidAirlines = %q, want %q (preserved from prev)", got.AvoidAirlines, "BA,LH")
	}

	// Partial overrides prev.
	partial2 := tripParams{AvoidAirlines: "UA"}
	got2 := mergeParams(prev, partial2)
	if got2.AvoidAirlines != "UA" {
		t.Errorf("AvoidAirlines = %q, want %q (overridden by partial)", got2.AvoidAirlines, "UA")
	}
}

func TestBuildRequestFromParams_AvoidAirlines(t *testing.T) {
	params := tripParams{
		Origin:        "DEL",
		Destination:   "YYZ",
		DepartureDate: "2025-06-15",
		AvoidAirlines: "BA,LH",
	}
	req := buildRequestFromParams(params)
	if req.AvoidAirlines != "BA,LH" {
		t.Errorf("AvoidAirlines = %q, want %q", req.AvoidAirlines, "BA,LH")
	}
}

func TestChatSystemPrompt_AvoidAirlines(t *testing.T) {
	prompt := chatSystemPrompt(time.Date(2025, 7, 15, 0, 0, 0, 0, time.UTC))
	if !strings.Contains(prompt, "avoid_airlines") {
		t.Error("system prompt should mention avoid_airlines")
	}
}

func TestRefinementHint_AvoidAirlines(t *testing.T) {
	hint := refinementHint()
	if !strings.Contains(hint, "avoid_airlines") {
		t.Error("refinement hint should mention avoid_airlines")
	}
}

func TestParsePartialParams_PreferredAirlines(t *testing.T) {
	p, ok := parsePartialParams(`{"preferred_airlines":"AC,UA"}`)
	if !ok {
		t.Fatal("expected partial JSON with preferred_airlines to parse")
	}
	if p.PreferredAirlines != "AC,UA" {
		t.Errorf("PreferredAirlines = %q, want %q", p.PreferredAirlines, "AC,UA")
	}
}

func TestMergeParams_PreferredAirlines(t *testing.T) {
	prev := tripParams{Origin: "DEL", Destination: "YYZ", DepartureDate: "2025-06-15", PreferredAirlines: "AC"}
	partial := tripParams{Cabin: "business"}
	got := mergeParams(prev, partial)
	if got.PreferredAirlines != "AC" {
		t.Errorf("PreferredAirlines = %q, want %q (preserved from prev)", got.PreferredAirlines, "AC")
	}
}

func TestBuildRequestFromParams_PreferredAirlines(t *testing.T) {
	p := tripParams{
		Origin:            "DEL",
		Destination:       "YYZ",
		DepartureDate:     "2025-06-15",
		PreferredAirlines: "AC,UA",
	}
	req := buildRequestFromParams(p)
	if req.PreferredAirlines != "AC,UA" {
		t.Errorf("PreferredAirlines = %q, want %q", req.PreferredAirlines, "AC,UA")
	}
}

func TestChatSystemPrompt_PreferredAirlines(t *testing.T) {
	prompt := chatSystemPrompt(time.Now())
	if !strings.Contains(prompt, "preferred_airlines") {
		t.Error("system prompt should mention preferred_airlines")
	}
}

func TestRefinementHint_PreferredAirlines(t *testing.T) {
	hint := refinementHint()
	if !strings.Contains(hint, "preferred_airlines") {
		t.Error("refinement hint should mention preferred_airlines")
	}
}

func TestParsePartialParams_Leg2Date(t *testing.T) {
	p, ok := parsePartialParams(`{"leg2_date":"2025-06-20"}`)
	if !ok {
		t.Fatal("expected partial JSON with leg2_date to parse")
	}
	if p.Leg2Date != "2025-06-20" {
		t.Errorf("Leg2Date = %q, want %q", p.Leg2Date, "2025-06-20")
	}
}

func TestMergeParams_Leg2Date(t *testing.T) {
	prev := tripParams{Origin: "DEL", Destination: "YYZ", DepartureDate: "2025-06-15", Leg2Date: "2025-06-20"}
	partial := tripParams{Cabin: "business"}
	got := mergeParams(prev, partial)
	if got.Leg2Date != "2025-06-20" {
		t.Errorf("Leg2Date = %q, want %q (preserved from prev)", got.Leg2Date, "2025-06-20")
	}

	// Partial overrides prev.
	partial2 := tripParams{Leg2Date: "2025-06-25"}
	got2 := mergeParams(prev, partial2)
	if got2.Leg2Date != "2025-06-25" {
		t.Errorf("Leg2Date = %q, want %q (overridden by partial)", got2.Leg2Date, "2025-06-25")
	}
}

func TestBuildRequestFromParams_Leg2Date(t *testing.T) {
	params := tripParams{
		Origin:        "DEL",
		Destination:   "YYZ",
		DepartureDate: "2025-06-15",
		Leg2Date:      "2025-06-20",
	}
	req := buildRequestFromParams(params)
	if req.Leg2Date != "2025-06-20" {
		t.Errorf("Leg2Date = %q, want %q", req.Leg2Date, "2025-06-20")
	}
}

func TestChatSystemPrompt_Leg2Date(t *testing.T) {
	prompt := chatSystemPrompt(time.Date(2025, 7, 15, 0, 0, 0, 0, time.UTC))
	if !strings.Contains(prompt, "leg2_date") {
		t.Error("system prompt should mention leg2_date")
	}
}

func TestRefinementHint_Leg2Date(t *testing.T) {
	hint := refinementHint()
	if !strings.Contains(hint, "leg2_date") {
		t.Error("refinement hint should mention leg2_date")
	}
}

func TestParsePartialParams_ArrivalTime(t *testing.T) {
	p, ok := parsePartialParams(`{"arrival_before":"18:00"}`)
	if !ok {
		t.Fatal("expected partial JSON with arrival_before to parse")
	}
	if p.ArrivalBefore != "18:00" {
		t.Errorf("ArrivalBefore = %q, want %q", p.ArrivalBefore, "18:00")
	}

	p2, ok := parsePartialParams(`{"arrival_after":"08:00"}`)
	if !ok {
		t.Fatal("expected partial JSON with arrival_after to parse")
	}
	if p2.ArrivalAfter != "08:00" {
		t.Errorf("ArrivalAfter = %q, want %q", p2.ArrivalAfter, "08:00")
	}
}

func TestMergeParams_ArrivalTime(t *testing.T) {
	prev := tripParams{Origin: "DEL", Destination: "YYZ", DepartureDate: "2025-06-15", ArrivalAfter: "08:00", ArrivalBefore: "18:00"}
	partial := tripParams{Cabin: "business"}
	got := mergeParams(prev, partial)
	if got.ArrivalAfter != "08:00" || got.ArrivalBefore != "18:00" {
		t.Errorf("ArrivalAfter/Before not preserved: got %q/%q", got.ArrivalAfter, got.ArrivalBefore)
	}
}

func TestBuildRequestFromParams_ArrivalTime(t *testing.T) {
	params := tripParams{
		Origin:        "DEL",
		Destination:   "YYZ",
		DepartureDate: "2025-06-15",
		ArrivalAfter:  "08:00",
		ArrivalBefore: "18:00",
	}
	req := buildRequestFromParams(params)
	if req.ArrivalAfter != "08:00" || req.ArrivalBefore != "18:00" {
		t.Errorf("ArrivalAfter/Before = %q/%q, want 08:00/18:00", req.ArrivalAfter, req.ArrivalBefore)
	}
}

func TestParsePartialParams_MaxDurationHours(t *testing.T) {
	p, ok := parsePartialParams(`{"max_duration_hours":12}`)
	if !ok {
		t.Fatal("expected partial JSON with max_duration_hours to parse")
	}
	if p.MaxDurationHours != 12 {
		t.Errorf("MaxDurationHours = %d, want 12", p.MaxDurationHours)
	}
}

func TestMergeParams_MaxDurationHours(t *testing.T) {
	prev := tripParams{Origin: "DEL", Destination: "YYZ", DepartureDate: "2025-06-15", MaxDurationHours: 12}
	partial := tripParams{Cabin: "business"}
	got := mergeParams(prev, partial)
	if got.MaxDurationHours != 12 {
		t.Errorf("MaxDurationHours = %d, want 12 (preserved from prev)", got.MaxDurationHours)
	}
}

func TestBuildRequestFromParams_MaxDuration(t *testing.T) {
	params := tripParams{
		Origin:           "DEL",
		Destination:      "YYZ",
		DepartureDate:    "2025-06-15",
		MaxDurationHours: 12,
	}
	req := buildRequestFromParams(params)
	if req.MaxDuration != 12*time.Hour {
		t.Errorf("MaxDuration = %v, want %v", req.MaxDuration, 12*time.Hour)
	}
}

func TestChatSystemPrompt_ArrivalTime(t *testing.T) {
	prompt := chatSystemPrompt(time.Date(2025, 7, 15, 0, 0, 0, 0, time.UTC))
	if !strings.Contains(prompt, "arrival_after") || !strings.Contains(prompt, "arrival_before") {
		t.Error("system prompt should mention arrival_after and arrival_before")
	}
}

func TestChatSystemPrompt_MaxDuration(t *testing.T) {
	prompt := chatSystemPrompt(time.Date(2025, 7, 15, 0, 0, 0, 0, time.UTC))
	if !strings.Contains(prompt, "max_duration_hours") {
		t.Error("system prompt should mention max_duration_hours")
	}
}

func TestRefinementHint_ArrivalTime(t *testing.T) {
	hint := refinementHint()
	if !strings.Contains(hint, "arrival_after") || !strings.Contains(hint, "arrival_before") {
		t.Error("refinement hint should mention arrival_after/arrival_before")
	}
}

func TestRefinementHint_MaxDuration(t *testing.T) {
	hint := refinementHint()
	if !strings.Contains(hint, "max_duration_hours") {
		t.Error("refinement hint should mention max_duration_hours")
	}
}

func TestChatLoop_PriceInsightsInOutput(t *testing.T) {
	responses := []string{
		`{"origin":"DEL","destination":"YYZ","departure_date":"2025-06-15"}
Searching for flights.`,
	}
	mock := &chatMockLLM{responses: responses}
	fakeStrat := &fakeSearchStrategy{
		results: []search.Itinerary{
			{
				Legs:       []search.Leg{{Flight: types.Flight{Price: types.Money{Amount: 600, Currency: "USD"}, Outbound: []types.Segment{{Origin: "DEL", Destination: "YYZ"}}}}},
				TotalPrice: types.Money{Amount: 600, Currency: "USD"},
			},
		},
	}
	picker := search.NewPicker(mock, fakeStrat)
	pi := &mockPriceInsighter{
		insights: search.PriceInsights{
			PriceLevel:        "low",
			LowestPrice:       450,
			TypicalPriceRange: [2]float64{500, 900},
		},
	}

	in := strings.NewReader("find flights\nquit\n")
	var out strings.Builder

	err := chatLoop(context.Background(), mock, picker, pi, nil, in, &out)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "Price level: low") {
		t.Errorf("expected price insights in output, got:\n%s", output)
	}
	if !strings.Contains(output, "500") || !strings.Contains(output, "900") {
		t.Errorf("expected typical price range in output, got:\n%s", output)
	}
}

// --- stopover duration params ---

func TestParsePartialParams_StopoverHours(t *testing.T) {
	input := `{"min_stopover_hours":24,"max_stopover_hours":72}`
	p, ok := parsePartialParams(input)
	if !ok {
		t.Fatal("expected partial params")
	}
	if p.MinStopoverHours != 24 {
		t.Errorf("MinStopoverHours = %d, want 24", p.MinStopoverHours)
	}
	if p.MaxStopoverHours != 72 {
		t.Errorf("MaxStopoverHours = %d, want 72", p.MaxStopoverHours)
	}
}

func TestMergeParams_StopoverHours(t *testing.T) {
	prev := tripParams{MinStopoverHours: 24, MaxStopoverHours: 96}
	partial := tripParams{MaxStopoverHours: 72} // only change max
	merged := mergeParams(prev, partial)
	if merged.MinStopoverHours != 24 {
		t.Errorf("MinStopoverHours = %d, want 24 (from prev)", merged.MinStopoverHours)
	}
	if merged.MaxStopoverHours != 72 {
		t.Errorf("MaxStopoverHours = %d, want 72 (from partial)", merged.MaxStopoverHours)
	}
}

func TestBuildRequestFromParams_StopoverDuration(t *testing.T) {
	p := tripParams{
		Origin:           "DEL",
		Destination:      "YYZ",
		DepartureDate:    "2026-06-15",
		MinStopoverHours: 24,
		MaxStopoverHours: 96,
	}
	req := buildRequestFromParams(p)
	if req.MinStopover != 24*time.Hour {
		t.Errorf("MinStopover = %v, want 24h", req.MinStopover)
	}
	if req.MaxStopover != 96*time.Hour {
		t.Errorf("MaxStopover = %v, want 96h", req.MaxStopover)
	}
}

// --- filterSuggestion ---

func TestFilterSuggestion_WithFilters(t *testing.T) {
	tests := []struct {
		name   string
		params tripParams
		want   []string // substrings that should appear in the suggestion
	}{
		{
			name:   "direct only",
			params: tripParams{DirectOnly: true},
			want:   []string{"direct_only"},
		},
		{
			name:   "max price",
			params: tripParams{MaxPrice: 500},
			want:   []string{"max_price"},
		},
		{
			name:   "departure time",
			params: tripParams{DepartureAfter: "08:00"},
			want:   []string{"departure"},
		},
		{
			name:   "arrival time",
			params: tripParams{ArrivalBefore: "18:00"},
			want:   []string{"arrival"},
		},
		{
			name:   "max duration",
			params: tripParams{MaxDurationHours: 8},
			want:   []string{"max_duration"},
		},
		{
			name:   "preferred alliance",
			params: tripParams{PreferredAlliance: "Star Alliance"},
			want:   []string{"preferred_alliance"},
		},
		{
			name:   "avoid airlines",
			params: tripParams{AvoidAirlines: "BA,LH"},
			want:   []string{"avoid_airlines"},
		},
		{
			name:   "preferred airlines",
			params: tripParams{PreferredAirlines: "AC"},
			want:   []string{"preferred_airlines"},
		},
		{
			name:   "multiple filters",
			params: tripParams{DirectOnly: true, MaxPrice: 500, PreferredAlliance: "OneWorld"},
			want:   []string{"direct_only", "max_price", "preferred_alliance"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := filterSuggestion(tt.params)
			for _, sub := range tt.want {
				if !strings.Contains(got, sub) {
					t.Errorf("filterSuggestion() = %q, missing %q", got, sub)
				}
			}
		})
	}
}

func TestFilterSuggestion_NoFilters(t *testing.T) {
	got := filterSuggestion(tripParams{})
	if got != "" {
		t.Errorf("filterSuggestion() = %q, want empty for no active filters", got)
	}
}

// --- zeroResultsSuggestion ---

func TestZeroResultsSuggestion_NearbyAirports(t *testing.T) {
	// YYZ has nearby YTZ.
	got := zeroResultsSuggestion(tripParams{Origin: "YYZ", Destination: "DEL"})
	if !strings.Contains(got, "YTZ") {
		t.Errorf("expected YTZ in suggestion, got: %s", got)
	}
	if !strings.Contains(got, "origin YYZ") {
		t.Errorf("expected 'origin YYZ' label, got: %s", got)
	}
}

func TestZeroResultsSuggestion_FlexDaysZero(t *testing.T) {
	// FlexDays==0: should suggest setting flex_days.
	got := zeroResultsSuggestion(tripParams{Origin: "DEL", Destination: "BOM"})
	if !strings.Contains(got, "flex_days") {
		t.Errorf("expected flex_days suggestion when FlexDays==0, got: %s", got)
	}
}

func TestZeroResultsSuggestion_FlexDaysActive(t *testing.T) {
	// FlexDays>0 and no nearby airports: should note flex is already active.
	got := zeroResultsSuggestion(tripParams{Origin: "DEL", Destination: "BOM", FlexDays: 3})
	if !strings.Contains(got, "already set to 3") {
		t.Errorf("expected 'already set to 3' when FlexDays=3, got: %s", got)
	}
}

func TestZeroResultsSuggestion_BothNearbyAndFlexDays(t *testing.T) {
	// JFK has nearby airports + FlexDays==0.
	got := zeroResultsSuggestion(tripParams{Origin: "JFK", Destination: "LHR"})
	if !strings.Contains(got, "EWR") {
		t.Errorf("expected EWR in suggestion, got: %s", got)
	}
	if !strings.Contains(got, "LGW") {
		t.Errorf("expected LGW in suggestion, got: %s", got)
	}
	if !strings.Contains(got, "flex_days") {
		t.Errorf("expected flex_days suggestion, got: %s", got)
	}
}

func TestChatLoop_ZeroResultsShowsSuggestion(t *testing.T) {
	responses := []string{
		`{"origin":"YYZ","destination":"DEL","departure_date":"2025-06-15"}
Searching for flights.`,
	}
	mock := &chatMockLLM{responses: responses}
	// Empty results triggers the zero-results block.
	fakeStrat := &fakeSearchStrategy{results: nil}
	picker := search.NewPicker(mock, fakeStrat)

	in := strings.NewReader("find flights\nquit\n")
	var out strings.Builder

	err := chatLoop(context.Background(), mock, picker, nil, nil, in, &out)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := out.String()
	// YYZ has nearby YTZ -- should appear in output.
	if !strings.Contains(output, "YTZ") {
		t.Errorf("expected nearby airport YTZ in zero-results output, got:\n%s", output)
	}
	// Should suggest flex_days since default is 0.
	if !strings.Contains(output, "flex_days") {
		t.Errorf("expected flex_days suggestion in zero-results output, got:\n%s", output)
	}
}

// mockWeightsUpdater records calls to SetWeights for test assertions.
type mockWeightsUpdater struct {
	calls []multicity.RankingWeights
}

func (m *mockWeightsUpdater) SetWeights(w multicity.RankingWeights) {
	m.calls = append(m.calls, w)
}

func TestChatLoop_ProfileSwitchUpdatesWeights(t *testing.T) {
	// First response: initial search with budget profile.
	// Second response: switch to eco profile.
	responses := []string{
		`{"origin":"DEL","destination":"YYZ","departure_date":"2025-06-15","profile":"budget"}
Searching for flights.`,
		`{"profile":"eco"}
Let me search with eco ranking.`,
		"Done.",
	}
	mock := &chatMockLLM{responses: responses}
	fakeStrat := &fakeSearchStrategy{
		results: []search.Itinerary{
			{
				Legs:       []search.Leg{{Flight: types.Flight{Price: types.Money{Amount: 500, Currency: "USD"}, Outbound: []types.Segment{{Origin: "DEL", Destination: "YYZ"}}}}},
				TotalPrice: types.Money{Amount: 500, Currency: "USD"},
			},
		},
	}
	picker := search.NewPicker(mock, fakeStrat)

	wu := &mockWeightsUpdater{}
	in := strings.NewReader("find flights\nswitch to eco\nquit\n")
	var out strings.Builder

	err := chatLoop(context.Background(), mock, picker, nil, wu, in, &out)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should have called SetWeights twice: once for "budget", once for "eco".
	if len(wu.calls) != 2 {
		t.Fatalf("expected 2 SetWeights calls, got %d", len(wu.calls))
	}
	if wu.calls[0] != multicity.WeightsBudget {
		t.Errorf("first SetWeights should be WeightsBudget, got %+v", wu.calls[0])
	}
	if wu.calls[1] != multicity.WeightsEco {
		t.Errorf("second SetWeights should be WeightsEco, got %+v", wu.calls[1])
	}
}
