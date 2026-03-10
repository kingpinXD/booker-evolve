package cmd

import (
	"context"
	"strings"
	"testing"

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
	responses []string
	idx       int
}

func (m *chatMockLLM) ChatCompletion(_ context.Context, _ []llm.Message) (string, error) {
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
