package multicity

import (
	"context"
	"testing"
	"time"

	"booker/llm"
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
								FlightNumber:    "AA100",
								Origin:          "JFK",
								Destination:     "LHR",
								OriginCity:      "New York",
								DestinationCity: "London",
								DepartureTime:   now,
								ArrivalTime:     now.Add(7 * time.Hour),
								Duration:        7 * time.Hour,
								AirlineName:     "American Airlines",
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
								FlightNumber:    "BA200",
								Origin:          "LHR",
								Destination:     "DEL",
								OriginCity:      "London",
								DestinationCity: "Delhi",
								DepartureTime:   now.Add(55 * time.Hour),
								ArrivalTime:     now.Add(63 * time.Hour),
								Duration:        8 * time.Hour,
								AirlineName:     "British Airways",
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

func TestBuildRankingPrompt_AllianceTags(t *testing.T) {
	now := time.Date(2026, 3, 15, 10, 0, 0, 0, time.UTC)
	itineraries := []search.Itinerary{
		{
			TotalPrice:  types.Money{Amount: 600, Currency: "USD"},
			TotalTravel: 15 * time.Hour,
			TotalTrip:   96 * time.Hour,
			Legs: []search.Leg{
				{
					Flight: types.Flight{
						Price: types.Money{Amount: 300, Currency: "USD"},
						Outbound: []types.Segment{
							{
								FlightNumber: "AC100",
								Origin:       "YYZ", Destination: "LHR",
								OriginCity: "Toronto", DestinationCity: "London",
								DepartureTime: now, ArrivalTime: now.Add(7 * time.Hour),
								Duration: 7 * time.Hour, Airline: "AC", AirlineName: "Air Canada",
							},
						},
					},
					Stopover: &search.Stopover{City: "London", Airport: "LHR", Duration: 72 * time.Hour},
				},
				{
					Flight: types.Flight{
						Price: types.Money{Amount: 300, Currency: "USD"},
						Outbound: []types.Segment{
							{
								FlightNumber: "BA200",
								Origin:       "LHR", Destination: "DEL",
								OriginCity: "London", DestinationCity: "Delhi",
								DepartureTime: now.Add(79 * time.Hour), ArrivalTime: now.Add(87 * time.Hour),
								Duration: 8 * time.Hour, Airline: "BA", AirlineName: "British Airways",
							},
						},
					},
				},
			},
		},
	}

	prompt := buildRankingPrompt(itineraries)
	// AC is Star Alliance, BA is OneWorld.
	if !containsAll(prompt, "[Star Alliance]", "[OneWorld]") {
		t.Errorf("prompt missing alliance tags, got:\n%s", prompt)
	}
}

func TestBuildRankingPrompt_UnknownAirlineNoTag(t *testing.T) {
	now := time.Date(2026, 3, 15, 10, 0, 0, 0, time.UTC)
	itineraries := []search.Itinerary{
		{
			TotalPrice:  types.Money{Amount: 200, Currency: "USD"},
			TotalTravel: 5 * time.Hour,
			TotalTrip:   5 * time.Hour,
			Legs: []search.Leg{
				{
					Flight: types.Flight{
						Price: types.Money{Amount: 200, Currency: "USD"},
						Outbound: []types.Segment{
							{
								FlightNumber: "XX999",
								Origin:       "AAA", Destination: "BBB",
								OriginCity: "CityA", DestinationCity: "CityB",
								DepartureTime: now, ArrivalTime: now.Add(5 * time.Hour),
								Duration: 5 * time.Hour, Airline: "XX", AirlineName: "Unknown Air",
							},
						},
					},
				},
			},
		},
	}

	prompt := buildRankingPrompt(itineraries)
	// Unknown airline should have no alliance tag.
	for _, tag := range []string{"[Star Alliance]", "[OneWorld]", "[SkyTeam]"} {
		if searchString(prompt, tag) {
			t.Errorf("prompt should not contain %s for unknown airline, got:\n%s", tag, prompt)
		}
	}
}

func TestBuildRankingPrompt_StopoverNotes(t *testing.T) {
	now := time.Date(2026, 3, 15, 10, 0, 0, 0, time.UTC)
	itineraries := []search.Itinerary{
		{
			TotalPrice:  types.Money{Amount: 500, Currency: "USD"},
			TotalTravel: 15 * time.Hour,
			TotalTrip:   96 * time.Hour,
			Legs: []search.Leg{
				{
					Flight: types.Flight{
						Price: types.Money{Amount: 250, Currency: "USD"},
						Outbound: []types.Segment{{
							FlightNumber: "CX100", Origin: "DEL", Destination: "HKG",
							OriginCity: "Delhi", DestinationCity: "Hong Kong",
							DepartureTime: now, ArrivalTime: now.Add(7 * time.Hour),
							Duration: 7 * time.Hour, AirlineName: "Cathay Pacific",
						}},
					},
					Stopover: &search.Stopover{
						City: "Hong Kong", Airport: "HKG", Duration: 72 * time.Hour,
						Notes: "Major Cathay Pacific hub. Great food, easy transit city.",
					},
				},
				{
					Flight: types.Flight{
						Price: types.Money{Amount: 250, Currency: "USD"},
						Outbound: []types.Segment{{
							FlightNumber: "AC200", Origin: "HKG", Destination: "YYZ",
							OriginCity: "Hong Kong", DestinationCity: "Toronto",
							DepartureTime: now.Add(79 * time.Hour), ArrivalTime: now.Add(95 * time.Hour),
							Duration: 16 * time.Hour, AirlineName: "Air Canada",
						}},
					},
				},
			},
		},
	}

	prompt := buildRankingPrompt(itineraries)
	if !containsAll(prompt, "Cathay Pacific hub", "Great food") {
		t.Errorf("prompt missing stopover notes, got:\n%s", prompt)
	}
}

func TestBuildRankingPrompt_StopoverNoNotes(t *testing.T) {
	now := time.Date(2026, 3, 15, 10, 0, 0, 0, time.UTC)
	itineraries := []search.Itinerary{
		{
			TotalPrice:  types.Money{Amount: 500, Currency: "USD"},
			TotalTravel: 15 * time.Hour,
			TotalTrip:   96 * time.Hour,
			Legs: []search.Leg{
				{
					Flight: types.Flight{
						Price: types.Money{Amount: 250, Currency: "USD"},
						Outbound: []types.Segment{{
							FlightNumber: "CX100", Origin: "DEL", Destination: "HKG",
							OriginCity: "Delhi", DestinationCity: "Hong Kong",
							DepartureTime: now, ArrivalTime: now.Add(7 * time.Hour),
							Duration: 7 * time.Hour, AirlineName: "Cathay Pacific",
						}},
					},
					Stopover: &search.Stopover{
						City: "Hong Kong", Airport: "HKG", Duration: 72 * time.Hour,
					},
				},
				{
					Flight: types.Flight{
						Price: types.Money{Amount: 250, Currency: "USD"},
						Outbound: []types.Segment{{
							FlightNumber: "AC200", Origin: "HKG", Destination: "YYZ",
							OriginCity: "Hong Kong", DestinationCity: "Toronto",
							DepartureTime: now.Add(79 * time.Hour), ArrivalTime: now.Add(95 * time.Hour),
							Duration: 16 * time.Hour, AirlineName: "Air Canada",
						}},
					},
				},
			},
		},
	}

	prompt := buildRankingPrompt(itineraries)
	// With no notes, the stopover line should still be present but no notes line.
	if !containsAll(prompt, "STOPOVER: Hong Kong") {
		t.Errorf("prompt should still show stopover, got:\n%s", prompt)
	}
	if searchString(prompt, "Notes:") {
		t.Errorf("prompt should not contain Notes: when notes are empty, got:\n%s", prompt)
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

func TestIsRedEye_EarlyMorning(t *testing.T) {
	tm := time.Date(2026, 3, 15, 2, 30, 0, 0, time.UTC)
	if !isRedEye(tm) {
		t.Error("02:30 should be red-eye")
	}
}

func TestIsRedEye_Morning(t *testing.T) {
	tm := time.Date(2026, 3, 15, 10, 0, 0, 0, time.UTC)
	if isRedEye(tm) {
		t.Error("10:00 should not be red-eye")
	}
}

func TestIsRedEye_Midnight(t *testing.T) {
	tm := time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC)
	if !isRedEye(tm) {
		t.Error("00:00 should be red-eye")
	}
}

func TestIsRedEye_FiveAM(t *testing.T) {
	tm := time.Date(2026, 3, 15, 5, 0, 0, 0, time.UTC)
	if isRedEye(tm) {
		t.Error("05:00 should not be red-eye")
	}
}

func TestBuildRankingPrompt_RedEyeTag(t *testing.T) {
	dep := time.Date(2026, 3, 24, 3, 30, 0, 0, time.UTC)
	itineraries := []search.Itinerary{
		{
			TotalPrice:  types.Money{Amount: 400, Currency: "USD"},
			TotalTravel: 6*time.Hour + 30*time.Minute,
			TotalTrip:   6*time.Hour + 30*time.Minute,
			Legs: []search.Leg{
				{
					Flight: types.Flight{
						Price: types.Money{Amount: 400, Currency: "USD"},
						Outbound: []types.Segment{{
							FlightNumber: "TG332", Origin: "DEL", Destination: "HKG",
							OriginCity: "Delhi", DestinationCity: "Hong Kong",
							DepartureTime: dep, ArrivalTime: dep.Add(6*time.Hour + 30*time.Minute),
							Duration: 6*time.Hour + 30*time.Minute, AirlineName: "Thai Airways",
						}},
					},
				},
			},
		},
	}

	prompt := buildRankingPrompt(itineraries)
	if !searchString(prompt, "[Red-eye]") {
		t.Errorf("prompt should contain [Red-eye] for 03:30 departure, got:\n%s", prompt)
	}
}

func TestBuildRankingPrompt_OvernightTag(t *testing.T) {
	dep := time.Date(2026, 3, 24, 23, 0, 0, 0, time.UTC)
	itineraries := []search.Itinerary{
		{
			TotalPrice:  types.Money{Amount: 400, Currency: "USD"},
			TotalTravel: 6 * time.Hour,
			TotalTrip:   6 * time.Hour,
			Legs: []search.Leg{
				{
					Flight: types.Flight{
						Price: types.Money{Amount: 400, Currency: "USD"},
						Outbound: []types.Segment{{
							FlightNumber: "TG316", Origin: "DEL", Destination: "BKK",
							OriginCity: "Delhi", DestinationCity: "Bangkok",
							DepartureTime: dep, ArrivalTime: dep.Add(6 * time.Hour),
							Duration: 6 * time.Hour, AirlineName: "Thai Airways",
							Overnight: true,
						}},
					},
				},
			},
		},
	}

	prompt := buildRankingPrompt(itineraries)
	if !searchString(prompt, "[Overnight]") {
		t.Errorf("prompt should contain [Overnight] for overnight segment, got:\n%s", prompt)
	}
}

func TestBuildRankingPrompt_NoOvernightTag(t *testing.T) {
	dep := time.Date(2026, 3, 24, 10, 0, 0, 0, time.UTC)
	itineraries := []search.Itinerary{
		{
			TotalPrice:  types.Money{Amount: 400, Currency: "USD"},
			TotalTravel: 6 * time.Hour,
			TotalTrip:   6 * time.Hour,
			Legs: []search.Leg{
				{
					Flight: types.Flight{
						Price: types.Money{Amount: 400, Currency: "USD"},
						Outbound: []types.Segment{{
							FlightNumber: "TG316", Origin: "DEL", Destination: "BKK",
							OriginCity: "Delhi", DestinationCity: "Bangkok",
							DepartureTime: dep, ArrivalTime: dep.Add(6 * time.Hour),
							Duration: 6 * time.Hour, AirlineName: "Thai Airways",
							Overnight: false,
						}},
					},
				},
			},
		},
	}

	prompt := buildRankingPrompt(itineraries)
	if searchString(prompt, "[Overnight]") {
		t.Errorf("prompt should not contain [Overnight] for daytime segment, got:\n%s", prompt)
	}
}

func TestBuildRankingPrompt_LegroomTag(t *testing.T) {
	dep := time.Date(2026, 3, 24, 10, 0, 0, 0, time.UTC)
	itineraries := []search.Itinerary{
		{
			TotalPrice:  types.Money{Amount: 400, Currency: "USD"},
			TotalTravel: 6 * time.Hour,
			TotalTrip:   6 * time.Hour,
			Legs: []search.Leg{
				{
					Flight: types.Flight{
						Price: types.Money{Amount: 400, Currency: "USD"},
						Outbound: []types.Segment{{
							FlightNumber: "TG316", Origin: "DEL", Destination: "BKK",
							OriginCity: "Delhi", DestinationCity: "Bangkok",
							DepartureTime: dep, ArrivalTime: dep.Add(6 * time.Hour),
							Duration: 6 * time.Hour, AirlineName: "Thai Airways",
							Legroom: "32 in",
						}},
					},
				},
			},
		},
	}

	prompt := buildRankingPrompt(itineraries)
	if !searchString(prompt, "[Legroom: 32 in]") {
		t.Errorf("prompt should contain [Legroom: 32 in], got:\n%s", prompt)
	}
}

func TestBuildRankingPrompt_NoLegroomTag(t *testing.T) {
	dep := time.Date(2026, 3, 24, 10, 0, 0, 0, time.UTC)
	itineraries := []search.Itinerary{
		{
			TotalPrice:  types.Money{Amount: 400, Currency: "USD"},
			TotalTravel: 6 * time.Hour,
			TotalTrip:   6 * time.Hour,
			Legs: []search.Leg{
				{
					Flight: types.Flight{
						Price: types.Money{Amount: 400, Currency: "USD"},
						Outbound: []types.Segment{{
							FlightNumber: "TG316", Origin: "DEL", Destination: "BKK",
							OriginCity: "Delhi", DestinationCity: "Bangkok",
							DepartureTime: dep, ArrivalTime: dep.Add(6 * time.Hour),
							Duration: 6 * time.Hour, AirlineName: "Thai Airways",
						}},
					},
				},
			},
		},
	}

	prompt := buildRankingPrompt(itineraries)
	if searchString(prompt, "[Legroom:") {
		t.Errorf("prompt should not contain [Legroom:] when legroom is empty, got:\n%s", prompt)
	}
}

func TestBuildRankingPrompt_AircraftTag(t *testing.T) {
	dep := time.Date(2026, 3, 24, 10, 0, 0, 0, time.UTC)
	itineraries := []search.Itinerary{
		{
			TotalPrice:  types.Money{Amount: 400, Currency: "USD"},
			TotalTravel: 6 * time.Hour,
			TotalTrip:   6 * time.Hour,
			Legs: []search.Leg{
				{
					Flight: types.Flight{
						Price: types.Money{Amount: 400, Currency: "USD"},
						Outbound: []types.Segment{{
							FlightNumber: "TG316", Origin: "DEL", Destination: "BKK",
							OriginCity: "Delhi", DestinationCity: "Bangkok",
							DepartureTime: dep, ArrivalTime: dep.Add(6 * time.Hour),
							Duration: 6 * time.Hour, AirlineName: "Thai Airways",
							Aircraft: "Boeing 787",
						}},
					},
				},
			},
		},
	}

	prompt := buildRankingPrompt(itineraries)
	if !searchString(prompt, "[Aircraft: Boeing 787]") {
		t.Errorf("prompt should contain [Aircraft: Boeing 787], got:\n%s", prompt)
	}
}

func TestBuildRankingPrompt_NoAircraftTag(t *testing.T) {
	dep := time.Date(2026, 3, 24, 10, 0, 0, 0, time.UTC)
	itineraries := []search.Itinerary{
		{
			TotalPrice:  types.Money{Amount: 400, Currency: "USD"},
			TotalTravel: 6 * time.Hour,
			TotalTrip:   6 * time.Hour,
			Legs: []search.Leg{
				{
					Flight: types.Flight{
						Price: types.Money{Amount: 400, Currency: "USD"},
						Outbound: []types.Segment{{
							FlightNumber: "TG316", Origin: "DEL", Destination: "BKK",
							OriginCity: "Delhi", DestinationCity: "Bangkok",
							DepartureTime: dep, ArrivalTime: dep.Add(6 * time.Hour),
							Duration: 6 * time.Hour, AirlineName: "Thai Airways",
						}},
					},
				},
			},
		},
	}

	prompt := buildRankingPrompt(itineraries)
	if searchString(prompt, "[Aircraft:") {
		t.Errorf("prompt should not contain [Aircraft:] when aircraft is empty, got:\n%s", prompt)
	}
}

func TestBuildRankingPrompt_CarbonLine(t *testing.T) {
	dep := time.Date(2026, 3, 24, 10, 0, 0, 0, time.UTC)
	itineraries := []search.Itinerary{
		{
			TotalPrice:  types.Money{Amount: 400, Currency: "USD"},
			TotalTravel: 6 * time.Hour,
			TotalTrip:   6 * time.Hour,
			Legs: []search.Leg{
				{
					Flight: types.Flight{
						Price:    types.Money{Amount: 400, Currency: "USD"},
						CarbonKg: 150,
						Outbound: []types.Segment{{
							FlightNumber: "TG316", Origin: "DEL", Destination: "BKK",
							OriginCity: "Delhi", DestinationCity: "Bangkok",
							DepartureTime: dep, ArrivalTime: dep.Add(6 * time.Hour),
							Duration: 6 * time.Hour, AirlineName: "Thai Airways",
						}},
					},
				},
			},
		},
	}

	prompt := buildRankingPrompt(itineraries)
	if !searchString(prompt, "CO2: 150kg") {
		t.Errorf("prompt should contain CO2: 150kg, got:\n%s", prompt)
	}
}

func TestBuildRankingPrompt_NoCarbonLine(t *testing.T) {
	dep := time.Date(2026, 3, 24, 10, 0, 0, 0, time.UTC)
	itineraries := []search.Itinerary{
		{
			TotalPrice:  types.Money{Amount: 400, Currency: "USD"},
			TotalTravel: 6 * time.Hour,
			TotalTrip:   6 * time.Hour,
			Legs: []search.Leg{
				{
					Flight: types.Flight{
						Price: types.Money{Amount: 400, Currency: "USD"},
						Outbound: []types.Segment{{
							FlightNumber: "TG316", Origin: "DEL", Destination: "BKK",
							OriginCity: "Delhi", DestinationCity: "Bangkok",
							DepartureTime: dep, ArrivalTime: dep.Add(6 * time.Hour),
							Duration: 6 * time.Hour, AirlineName: "Thai Airways",
						}},
					},
				},
			},
		},
	}

	prompt := buildRankingPrompt(itineraries)
	if searchString(prompt, "CO2:") {
		t.Errorf("prompt should not contain CO2: when CarbonKg is 0, got:\n%s", prompt)
	}
}

func TestBuildRankingPrompt_CarbonBenchmark(t *testing.T) {
	dep := time.Date(2026, 3, 24, 10, 0, 0, 0, time.UTC)
	itineraries := []search.Itinerary{
		{
			TotalPrice:  types.Money{Amount: 400, Currency: "USD"},
			TotalTravel: 6 * time.Hour,
			TotalTrip:   6 * time.Hour,
			Legs: []search.Leg{
				{
					Flight: types.Flight{
						Price:           types.Money{Amount: 400, Currency: "USD"},
						CarbonKg:        1106,
						TypicalCarbonKg: 949,
						CarbonDiffPct:   17,
						Outbound: []types.Segment{{
							FlightNumber: "AC42", Origin: "DEL", Destination: "YYZ",
							OriginCity: "Delhi", DestinationCity: "Toronto",
							DepartureTime: dep, ArrivalTime: dep.Add(14 * time.Hour),
							Duration: 14 * time.Hour, AirlineName: "Air Canada",
						}},
					},
				},
			},
		},
	}

	prompt := buildRankingPrompt(itineraries)
	if !searchString(prompt, "CO2: 1106kg (+17% vs typical)") {
		t.Errorf("prompt should contain benchmark comparison, got:\n%s", prompt)
	}
}

func TestBuildRankingPrompt_CarbonTypicalOnly(t *testing.T) {
	dep := time.Date(2026, 3, 24, 10, 0, 0, 0, time.UTC)
	itineraries := []search.Itinerary{
		{
			TotalPrice:  types.Money{Amount: 400, Currency: "USD"},
			TotalTravel: 6 * time.Hour,
			TotalTrip:   6 * time.Hour,
			Legs: []search.Leg{
				{
					Flight: types.Flight{
						Price:           types.Money{Amount: 400, Currency: "USD"},
						CarbonKg:        949,
						TypicalCarbonKg: 949,
						Outbound: []types.Segment{{
							FlightNumber: "AC42", Origin: "DEL", Destination: "YYZ",
							OriginCity: "Delhi", DestinationCity: "Toronto",
							DepartureTime: dep, ArrivalTime: dep.Add(14 * time.Hour),
							Duration: 14 * time.Hour, AirlineName: "Air Canada",
						}},
					},
				},
			},
		},
	}

	prompt := buildRankingPrompt(itineraries)
	if !searchString(prompt, "CO2: 949kg (typical: 949kg)") {
		t.Errorf("prompt should show typical when DiffPct is 0 but typical is known, got:\n%s", prompt)
	}
}

func TestBuildRankingPrompt_SeatsTag(t *testing.T) {
	dep := time.Date(2026, 3, 24, 10, 0, 0, 0, time.UTC)
	itineraries := []search.Itinerary{
		{
			TotalPrice:  types.Money{Amount: 400, Currency: "USD"},
			TotalTravel: 6 * time.Hour,
			TotalTrip:   6 * time.Hour,
			Legs: []search.Leg{
				{
					Flight: types.Flight{
						Price: types.Money{Amount: 400, Currency: "USD"},
						Outbound: []types.Segment{{
							FlightNumber: "TG316", Origin: "DEL", Destination: "BKK",
							OriginCity: "Delhi", DestinationCity: "Bangkok",
							DepartureTime: dep, ArrivalTime: dep.Add(6 * time.Hour),
							Duration: 6 * time.Hour, AirlineName: "Thai Airways",
							SeatsLeft: 4,
						}},
					},
				},
			},
		},
	}

	prompt := buildRankingPrompt(itineraries)
	if !searchString(prompt, "[Seats: 4 left]") {
		t.Errorf("prompt should contain [Seats: 4 left], got:\n%s", prompt)
	}
}

func TestBuildRankingPrompt_NoSeatsTag(t *testing.T) {
	dep := time.Date(2026, 3, 24, 10, 0, 0, 0, time.UTC)
	itineraries := []search.Itinerary{
		{
			TotalPrice:  types.Money{Amount: 400, Currency: "USD"},
			TotalTravel: 6 * time.Hour,
			TotalTrip:   6 * time.Hour,
			Legs: []search.Leg{
				{
					Flight: types.Flight{
						Price: types.Money{Amount: 400, Currency: "USD"},
						Outbound: []types.Segment{{
							FlightNumber: "TG316", Origin: "DEL", Destination: "BKK",
							OriginCity: "Delhi", DestinationCity: "Bangkok",
							DepartureTime: dep, ArrivalTime: dep.Add(6 * time.Hour),
							Duration: 6 * time.Hour, AirlineName: "Thai Airways",
						}},
					},
				},
			},
		},
	}

	prompt := buildRankingPrompt(itineraries)
	if searchString(prompt, "[Seats:") {
		t.Errorf("prompt should not contain [Seats:] when SeatsLeft is 0, got:\n%s", prompt)
	}
}

func TestBuildRankingPrompt_EmptyCityNames(t *testing.T) {
	dep := time.Date(2026, 3, 24, 10, 0, 0, 0, time.UTC)
	itineraries := []search.Itinerary{
		{
			TotalPrice:  types.Money{Amount: 400, Currency: "USD"},
			TotalTravel: 6 * time.Hour,
			TotalTrip:   6 * time.Hour,
			Legs: []search.Leg{
				{
					Flight: types.Flight{
						Price: types.Money{Amount: 400, Currency: "USD"},
						Outbound: []types.Segment{{
							FlightNumber: "CX100", Origin: "DEL", Destination: "HKG",
							DepartureTime: dep, ArrivalTime: dep.Add(6 * time.Hour),
							Duration: 6 * time.Hour, AirlineName: "Cathay Pacific",
						}},
					},
				},
			},
		},
	}

	prompt := buildRankingPrompt(itineraries)
	// Empty city names should NOT produce "(→)" noise.
	if searchString(prompt, "(→)") {
		t.Errorf("prompt should not contain empty city parenthetical (→), got:\n%s", prompt)
	}
	// Airport codes should still be present.
	if !searchString(prompt, "DEL→HKG") {
		t.Errorf("prompt should contain airport codes DEL→HKG, got:\n%s", prompt)
	}
}

func TestBuildRankingPrompt_WithCityNames(t *testing.T) {
	dep := time.Date(2026, 3, 24, 10, 0, 0, 0, time.UTC)
	itineraries := []search.Itinerary{
		{
			TotalPrice:  types.Money{Amount: 400, Currency: "USD"},
			TotalTravel: 6 * time.Hour,
			TotalTrip:   6 * time.Hour,
			Legs: []search.Leg{
				{
					Flight: types.Flight{
						Price: types.Money{Amount: 400, Currency: "USD"},
						Outbound: []types.Segment{{
							FlightNumber: "CX100", Origin: "DEL", Destination: "HKG",
							OriginCity: "Delhi", DestinationCity: "Hong Kong",
							DepartureTime: dep, ArrivalTime: dep.Add(6 * time.Hour),
							Duration: 6 * time.Hour, AirlineName: "Cathay Pacific",
						}},
					},
				},
			},
		},
	}

	prompt := buildRankingPrompt(itineraries)
	if !searchString(prompt, "(Delhi→Hong Kong)") {
		t.Errorf("prompt should contain (Delhi→Hong Kong), got:\n%s", prompt)
	}
}

func TestBuildRankingPrompt_RiskyConnection(t *testing.T) {
	dep := time.Date(2026, 3, 24, 10, 0, 0, 0, time.UTC)
	itineraries := []search.Itinerary{
		{
			TotalPrice:  types.Money{Amount: 400, Currency: "USD"},
			TotalTravel: 12 * time.Hour,
			TotalTrip:   12 * time.Hour,
			Legs: []search.Leg{
				{
					Flight: types.Flight{
						Price: types.Money{Amount: 400, Currency: "USD"},
						Outbound: []types.Segment{
							{
								FlightNumber: "AA100", Origin: "JFK", Destination: "ORD",
								OriginCity: "New York", DestinationCity: "Chicago",
								DepartureTime: dep, ArrivalTime: dep.Add(3 * time.Hour),
								Duration: 3 * time.Hour, AirlineName: "American Airlines",
							},
							{
								FlightNumber: "AA200", Origin: "ORD", Destination: "LAX",
								OriginCity: "Chicago", DestinationCity: "Los Angeles",
								DepartureTime:   dep.Add(3*time.Hour + 45*time.Minute),
								ArrivalTime:     dep.Add(7*time.Hour + 45*time.Minute),
								Duration:        4 * time.Hour,
								AirlineName:     "American Airlines",
								LayoverDuration: 45 * time.Minute,
							},
						},
					},
				},
			},
		},
	}

	prompt := buildRankingPrompt(itineraries)
	if !searchString(prompt, "[Risky connection: 45m]") {
		t.Errorf("prompt should contain [Risky connection: 45m] for 45min layover, got:\n%s", prompt)
	}
}

func TestBuildRankingPrompt_TightConnection(t *testing.T) {
	dep := time.Date(2026, 3, 24, 10, 0, 0, 0, time.UTC)
	itineraries := []search.Itinerary{
		{
			TotalPrice:  types.Money{Amount: 400, Currency: "USD"},
			TotalTravel: 12 * time.Hour,
			TotalTrip:   12 * time.Hour,
			Legs: []search.Leg{
				{
					Flight: types.Flight{
						Price: types.Money{Amount: 400, Currency: "USD"},
						Outbound: []types.Segment{
							{
								FlightNumber: "UA100", Origin: "SFO", Destination: "DEN",
								OriginCity: "San Francisco", DestinationCity: "Denver",
								DepartureTime: dep, ArrivalTime: dep.Add(3 * time.Hour),
								Duration: 3 * time.Hour, AirlineName: "United Airlines",
							},
							{
								FlightNumber: "UA200", Origin: "DEN", Destination: "JFK",
								OriginCity: "Denver", DestinationCity: "New York",
								DepartureTime:   dep.Add(4*time.Hour + 15*time.Minute),
								ArrivalTime:     dep.Add(8*time.Hour + 15*time.Minute),
								Duration:        4 * time.Hour,
								AirlineName:     "United Airlines",
								LayoverDuration: 75 * time.Minute,
							},
						},
					},
				},
			},
		},
	}

	prompt := buildRankingPrompt(itineraries)
	if !searchString(prompt, "[Tight connection: 75m]") {
		t.Errorf("prompt should contain [Tight connection: 75m] for 75min layover, got:\n%s", prompt)
	}
}

func TestBuildRankingPrompt_NoConnectionTag(t *testing.T) {
	dep := time.Date(2026, 3, 24, 10, 0, 0, 0, time.UTC)
	itineraries := []search.Itinerary{
		{
			TotalPrice:  types.Money{Amount: 400, Currency: "USD"},
			TotalTravel: 12 * time.Hour,
			TotalTrip:   12 * time.Hour,
			Legs: []search.Leg{
				{
					Flight: types.Flight{
						Price: types.Money{Amount: 400, Currency: "USD"},
						Outbound: []types.Segment{
							{
								FlightNumber: "DL100", Origin: "ATL", Destination: "MSP",
								OriginCity: "Atlanta", DestinationCity: "Minneapolis",
								DepartureTime: dep, ArrivalTime: dep.Add(3 * time.Hour),
								Duration: 3 * time.Hour, AirlineName: "Delta Air Lines",
							},
							{
								FlightNumber: "DL200", Origin: "MSP", Destination: "SEA",
								OriginCity: "Minneapolis", DestinationCity: "Seattle",
								DepartureTime:   dep.Add(5 * time.Hour),
								ArrivalTime:     dep.Add(9 * time.Hour),
								Duration:        4 * time.Hour,
								AirlineName:     "Delta Air Lines",
								LayoverDuration: 120 * time.Minute,
							},
						},
					},
				},
			},
		},
	}

	prompt := buildRankingPrompt(itineraries)
	if searchString(prompt, "[Risky connection:") {
		t.Errorf("prompt should not contain [Risky connection:] for 120min layover, got:\n%s", prompt)
	}
	if searchString(prompt, "[Tight connection:") {
		t.Errorf("prompt should not contain [Tight connection:] for 120min layover, got:\n%s", prompt)
	}
}

// countingLLM tracks how many times ChatCompletion is called.
type countingLLM struct {
	calls    int
	response string
}

func (m *countingLLM) ChatCompletion(_ context.Context, _ []llm.Message) (string, error) {
	m.calls++
	return m.response, nil
}

func makeTestItineraries(price float64, origin, dest string) []search.Itinerary {
	dep := time.Date(2026, 3, 15, 10, 0, 0, 0, time.UTC)
	return []search.Itinerary{{
		TotalPrice:  types.Money{Amount: price, Currency: "USD"},
		TotalTravel: 6 * time.Hour,
		TotalTrip:   6 * time.Hour,
		Legs: []search.Leg{{
			Flight: types.Flight{
				Price: types.Money{Amount: price, Currency: "USD"},
				Outbound: []types.Segment{{
					FlightNumber: "AA100", Origin: origin, Destination: dest,
					OriginCity: "CityA", DestinationCity: "CityB",
					DepartureTime: dep, ArrivalTime: dep.Add(6 * time.Hour),
					Duration: 6 * time.Hour, AirlineName: "Test Air",
				}},
			},
		}},
	}}
}

func TestRankCache_HitSkipsLLM(t *testing.T) {
	mock := &countingLLM{response: `[{"index":0,"score":80,"reasoning":"good"}]`}
	ranker := NewRanker(mock, WeightsBalanced)

	itins := makeTestItineraries(500, "JFK", "LHR")
	ctx := context.Background()

	if _, err := ranker.Rank(ctx, itins); err != nil {
		t.Fatalf("first Rank: %v", err)
	}
	if _, err := ranker.Rank(ctx, itins); err != nil {
		t.Fatalf("second Rank: %v", err)
	}
	if mock.calls != 1 {
		t.Errorf("expected 1 LLM call (cache hit), got %d", mock.calls)
	}
}

func TestRankCache_DifferentItinerariesMiss(t *testing.T) {
	mock := &countingLLM{response: `[{"index":0,"score":80,"reasoning":"good"}]`}
	ranker := NewRanker(mock, WeightsBalanced)

	ctx := context.Background()
	if _, err := ranker.Rank(ctx, makeTestItineraries(500, "JFK", "LHR")); err != nil {
		t.Fatalf("first Rank: %v", err)
	}
	if _, err := ranker.Rank(ctx, makeTestItineraries(700, "SFO", "NRT")); err != nil {
		t.Fatalf("second Rank: %v", err)
	}
	if mock.calls != 2 {
		t.Errorf("expected 2 LLM calls (cache miss), got %d", mock.calls)
	}
}

func TestRankCache_DifferentWeightsMiss(t *testing.T) {
	mock := &countingLLM{response: `[{"index":0,"score":80,"reasoning":"good"}]`}
	itins := makeTestItineraries(500, "JFK", "LHR")
	ctx := context.Background()

	budget := NewRanker(mock, WeightsBudget)
	if _, err := budget.Rank(ctx, itins); err != nil {
		t.Fatalf("budget Rank: %v", err)
	}

	comfort := NewRanker(mock, WeightsComfort)
	if _, err := comfort.Rank(ctx, itins); err != nil {
		t.Fatalf("comfort Rank: %v", err)
	}

	if mock.calls != 2 {
		t.Errorf("expected 2 LLM calls (different weights), got %d", mock.calls)
	}
}

func TestRankerCacheStats_HitAndMiss(t *testing.T) {
	mock := &countingLLM{response: `[{"index":0,"score":80,"reasoning":"good"}]`}
	ranker := NewRanker(mock, WeightsBalanced)
	ctx := context.Background()

	itins := makeTestItineraries(500, "JFK", "LHR")
	if _, err := ranker.Rank(ctx, itins); err != nil {
		t.Fatalf("first Rank: %v", err)
	}
	if _, err := ranker.Rank(ctx, itins); err != nil {
		t.Fatalf("second Rank: %v", err)
	}

	hits, misses := ranker.CacheStats()
	if hits != 1 || misses != 1 {
		t.Errorf("CacheStats() = (%d, %d), want (1, 1)", hits, misses)
	}
}

func TestRankerCacheStats_AllMisses(t *testing.T) {
	mock := &countingLLM{response: `[{"index":0,"score":80,"reasoning":"good"}]`}
	ranker := NewRanker(mock, WeightsBalanced)
	ctx := context.Background()

	if _, err := ranker.Rank(ctx, makeTestItineraries(500, "JFK", "LHR")); err != nil {
		t.Fatalf("first Rank: %v", err)
	}
	if _, err := ranker.Rank(ctx, makeTestItineraries(700, "SFO", "NRT")); err != nil {
		t.Fatalf("second Rank: %v", err)
	}

	hits, misses := ranker.CacheStats()
	if hits != 0 || misses != 2 {
		t.Errorf("CacheStats() = (%d, %d), want (0, 2)", hits, misses)
	}
}

func TestRankerCacheStats_Empty(t *testing.T) {
	mock := &countingLLM{response: `[{"index":0,"score":80,"reasoning":"good"}]`}
	ranker := NewRanker(mock, WeightsBalanced)

	hits, misses := ranker.CacheStats()
	if hits != 0 || misses != 0 {
		t.Errorf("CacheStats() = (%d, %d), want (0, 0)", hits, misses)
	}
}

func TestBuildRankingPrompt_NoRedEyeTag(t *testing.T) {
	dep := time.Date(2026, 3, 24, 10, 0, 0, 0, time.UTC)
	itineraries := []search.Itinerary{
		{
			TotalPrice:  types.Money{Amount: 400, Currency: "USD"},
			TotalTravel: 6*time.Hour + 30*time.Minute,
			TotalTrip:   6*time.Hour + 30*time.Minute,
			Legs: []search.Leg{
				{
					Flight: types.Flight{
						Price: types.Money{Amount: 400, Currency: "USD"},
						Outbound: []types.Segment{{
							FlightNumber: "TG332", Origin: "DEL", Destination: "HKG",
							OriginCity: "Delhi", DestinationCity: "Hong Kong",
							DepartureTime: dep, ArrivalTime: dep.Add(6*time.Hour + 30*time.Minute),
							Duration: 6*time.Hour + 30*time.Minute, AirlineName: "Thai Airways",
						}},
					},
				},
			},
		},
	}

	prompt := buildRankingPrompt(itineraries)
	if searchString(prompt, "[Red-eye]") {
		t.Errorf("prompt should not contain [Red-eye] for 10:00 departure, got:\n%s", prompt)
	}
}
