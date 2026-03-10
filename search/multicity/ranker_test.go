package multicity

import (
	"testing"
	"time"

	"booker/search"
	"booker/types"
)

func TestParseRankingResponse_ValidJSON(t *testing.T) {
	input := `[{"index":0,"score":85,"reasoning":"cheap"},{"index":1,"score":70,"reasoning":"ok"}]`
	results, err := parseRankingResponse(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("got %d results, want 2", len(results))
	}
	if results[0].Score != 85 || results[1].Score != 70 {
		t.Fatalf("wrong scores: %v", results)
	}
}

func TestParseRankingResponse_MarkdownFenced(t *testing.T) {
	input := "```json\n[{\"index\":0,\"score\":90,\"reasoning\":\"best\"}]\n```"
	results, err := parseRankingResponse(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 || results[0].Score != 90 {
		t.Fatalf("wrong result: %v", results)
	}
}

func TestParseRankingResponse_InvalidJSON(t *testing.T) {
	_, err := parseRankingResponse("not json at all")
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		name string
		d    time.Duration
		want string
	}{
		{"hours and minutes", 5*time.Hour + 30*time.Minute, "5h 30m"},
		{"zero", 0, "0h 0m"},
		{"days", 26*time.Hour + 15*time.Minute, "1d 2h 15m"},
		{"exact hours", 3 * time.Hour, "3h 0m"},
		{"multi-day", 72*time.Hour + 45*time.Minute, "3d 0h 45m"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatDuration(tt.d)
			if got != tt.want {
				t.Errorf("formatDuration(%v) = %q, want %q", tt.d, got, tt.want)
			}
		})
	}
}

func TestBuildSystemPrompt(t *testing.T) {
	prompt := buildSystemPrompt(WeightsBudget)
	// Verify weight values are interpolated.
	if !containsAll(prompt, "45%", "15%", "10%", "15%", "10%", "5%") {
		t.Errorf("prompt missing weight values: %s", prompt[:200])
	}
}

func TestBuildRankingPrompt(t *testing.T) {
	now := time.Date(2026, 3, 15, 10, 0, 0, 0, time.UTC)
	itineraries := []search.Itinerary{
		{
			TotalPrice:  types.Money{Amount: 500, Currency: "USD"},
			TotalTravel: 10 * time.Hour,
			TotalTrip:   72 * time.Hour,
			Legs: []search.Leg{
				{
					Flight: types.Flight{
						Price: types.Money{Amount: 250, Currency: "USD"},
						Outbound: []types.Segment{
							{
								FlightNumber:  "AA100",
								Origin:        "JFK",
								Destination:   "LHR",
								OriginCity:    "New York",
								DestinationCity: "London",
								DepartureTime: now,
								ArrivalTime:   now.Add(7 * time.Hour),
								Duration:      7 * time.Hour,
								AirlineName:   "American Airlines",
							},
						},
					},
					Stopover: &search.Stopover{
						City:     "London",
						Airport:  "LHR",
						Duration: 48 * time.Hour,
					},
				},
				{
					Flight: types.Flight{
						Price: types.Money{Amount: 250, Currency: "USD"},
						Outbound: []types.Segment{
							{
								FlightNumber:  "BA200",
								Origin:        "LHR",
								Destination:   "DEL",
								OriginCity:    "London",
								DestinationCity: "Delhi",
								DepartureTime: now.Add(55 * time.Hour),
								ArrivalTime:   now.Add(63 * time.Hour),
								Duration:      8 * time.Hour,
								AirlineName:   "British Airways",
							},
						},
					},
				},
			},
		},
	}

	prompt := buildRankingPrompt(itineraries)
	if !containsAll(prompt, "ITINERARY 0", "LEG 1", "LEG 2", "AA100", "BA200", "$500.00", "STOPOVER: London") {
		t.Errorf("prompt missing expected content: %s", prompt)
	}
}

// containsAll checks that s contains every substring.
func containsAll(s string, subs ...string) bool {
	for _, sub := range subs {
		if !contains(s, sub) {
			return false
		}
	}
	return true
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && searchString(s, sub)
}

func searchString(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
