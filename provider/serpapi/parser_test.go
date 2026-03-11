package serpapi

import (
	"testing"
	"time"

	"booker/config"
	"booker/types"
)

func TestParseFlightGroup_OneWay(t *testing.T) {
	g := FlightGroup{
		Flights: []FlightSegment{{
			DepartureAirport: Airport{Name: "Indira Gandhi International", ID: "DEL", Time: "2026-03-24 03:30"},
			ArrivalAirport:   Airport{Name: "Hong Kong International", ID: "HKG", Time: "2026-03-24 11:30"},
			Duration:         480,
			Airline:          "IndiGo",
			FlightNumber:     "6E 1234",
			TravelClass:      "Economy",
		}},
		TotalDuration: 480,
		Price:         350,
		BookingToken:  "token_abc",
	}

	f, err := parseFlightGroup(g)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if f.Provider != config.ProviderSerpAPI {
		t.Errorf("provider = %q, want %q", f.Provider, config.ProviderSerpAPI)
	}
	if f.Price.Amount != 350 {
		t.Errorf("price = %v, want 350", f.Price.Amount)
	}
	if f.Price.Currency != "USD" {
		t.Errorf("currency = %q, want USD", f.Price.Currency)
	}
	if f.TotalDuration != 480*time.Minute {
		t.Errorf("total duration = %v, want %v", f.TotalDuration, 480*time.Minute)
	}
	if f.BookingURL != "token_abc" {
		t.Errorf("booking URL = %q, want %q", f.BookingURL, "token_abc")
	}
	if len(f.Outbound) != 1 {
		t.Fatalf("outbound segments = %d, want 1", len(f.Outbound))
	}
	if len(f.Return) != 0 {
		t.Errorf("return segments = %d, want 0", len(f.Return))
	}
}

func TestParseFlightGroup_MultiSegment(t *testing.T) {
	g := FlightGroup{
		Flights: []FlightSegment{
			{
				DepartureAirport: Airport{Name: "DEL Airport", ID: "DEL", Time: "2026-03-24 06:00"},
				ArrivalAirport:   Airport{Name: "BKK Airport", ID: "BKK", Time: "2026-03-24 12:00"},
				Duration:         360,
				Airline:          "Thai Airways",
				FlightNumber:     "TG 332",
				TravelClass:      "Business",
			},
			{
				DepartureAirport: Airport{Name: "BKK Airport", ID: "BKK", Time: "2026-03-24 14:30"},
				ArrivalAirport:   Airport{Name: "HKG Airport", ID: "HKG", Time: "2026-03-24 18:00"},
				Duration:         210,
				Airline:          "Thai Airways",
				FlightNumber:     "TG 600",
				TravelClass:      "Business",
			},
		},
		Layovers: []Layover{{
			Duration: 150,
			Name:     "Suvarnabhumi Airport",
			ID:       "BKK",
		}},
		TotalDuration: 720,
		Price:         850,
		BookingToken:  "token_multi",
	}

	f, err := parseFlightGroup(g)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(f.Outbound) != 2 {
		t.Fatalf("outbound segments = %d, want 2", len(f.Outbound))
	}
	if f.TotalDuration != 720*time.Minute {
		t.Errorf("total duration = %v, want %v", f.TotalDuration, 720*time.Minute)
	}
	if f.Outbound[0].LayoverDuration != 150*time.Minute {
		t.Errorf("seg[0] layover = %v, want %v", f.Outbound[0].LayoverDuration, 150*time.Minute)
	}
	if f.Outbound[1].LayoverDuration != 0 {
		t.Errorf("seg[1] layover = %v, want 0", f.Outbound[1].LayoverDuration)
	}
}

func TestParseFlightGroup_InvalidTime(t *testing.T) {
	g := FlightGroup{
		Flights: []FlightSegment{{
			DepartureAirport: Airport{ID: "DEL", Time: "not-a-time"},
			ArrivalAirport:   Airport{ID: "HKG", Time: "2026-03-24 11:00"},
			FlightNumber:     "XX 100",
		}},
		Price: 100,
	}

	if _, err := parseFlightGroup(g); err == nil {
		t.Fatal("expected error for invalid departure time, got nil")
	}
}

func TestParseSegments(t *testing.T) {
	segs := []FlightSegment{
		{
			DepartureAirport: Airport{Name: "JFK Airport", ID: "JFK", Time: "2026-04-01 08:00"},
			ArrivalAirport:   Airport{Name: "LHR Airport", ID: "LHR", Time: "2026-04-01 20:00"},
			Duration:         420,
			Airline:          "British Airways",
			FlightNumber:     "BA 117",
			TravelClass:      "Premium Economy",
		},
		{
			DepartureAirport: Airport{Name: "LHR Airport", ID: "LHR", Time: "2026-04-01 22:00"},
			ArrivalAirport:   Airport{Name: "CDG Airport", ID: "CDG", Time: "2026-04-01 23:30"},
			Duration:         90,
			Airline:          "British Airways",
			FlightNumber:     "BA 334",
			TravelClass:      "Economy",
		},
	}
	layovers := []Layover{
		{Duration: 120, Name: "Heathrow", ID: "LHR"},
	}

	result, err := parseSegments(segs, layovers)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 2 {
		t.Fatalf("segments = %d, want 2", len(result))
	}

	s0 := result[0]
	if s0.Airline != "BA" {
		t.Errorf("seg[0].Airline = %q, want %q", s0.Airline, "BA")
	}
	if s0.AirlineName != "British Airways" {
		t.Errorf("seg[0].AirlineName = %q, want %q", s0.AirlineName, "British Airways")
	}
	if s0.FlightNumber != "BA117" {
		t.Errorf("seg[0].FlightNumber = %q, want %q", s0.FlightNumber, "BA117")
	}
	if s0.Origin != "JFK" {
		t.Errorf("seg[0].Origin = %q, want %q", s0.Origin, "JFK")
	}
	if s0.OriginName != "JFK Airport" {
		t.Errorf("seg[0].OriginName = %q, want %q", s0.OriginName, "JFK Airport")
	}
	if s0.Destination != "LHR" {
		t.Errorf("seg[0].Destination = %q, want %q", s0.Destination, "LHR")
	}
	if s0.DestinationName != "LHR Airport" {
		t.Errorf("seg[0].DestinationName = %q, want %q", s0.DestinationName, "LHR Airport")
	}
	if s0.Duration != 420*time.Minute {
		t.Errorf("seg[0].Duration = %v, want %v", s0.Duration, 420*time.Minute)
	}
	if s0.CabinClass != types.CabinPremiumEconomy {
		t.Errorf("seg[0].CabinClass = %q, want %q", s0.CabinClass, types.CabinPremiumEconomy)
	}
	if s0.LayoverDuration != 120*time.Minute {
		t.Errorf("seg[0].LayoverDuration = %v, want %v", s0.LayoverDuration, 120*time.Minute)
	}

	// Departure and arrival time parsing.
	wantDep := time.Date(2026, 4, 1, 8, 0, 0, 0, time.UTC)
	if !s0.DepartureTime.Equal(wantDep) {
		t.Errorf("seg[0].DepartureTime = %v, want %v", s0.DepartureTime, wantDep)
	}
	wantArr := time.Date(2026, 4, 1, 20, 0, 0, 0, time.UTC)
	if !s0.ArrivalTime.Equal(wantArr) {
		t.Errorf("seg[0].ArrivalTime = %v, want %v", s0.ArrivalTime, wantArr)
	}

	// Second segment has no layover (only one layover entry for the gap before it).
	s1 := result[1]
	if s1.CabinClass != types.CabinEconomy {
		t.Errorf("seg[1].CabinClass = %q, want %q", s1.CabinClass, types.CabinEconomy)
	}
	if s1.LayoverDuration != 0 {
		t.Errorf("seg[1].LayoverDuration = %v, want 0", s1.LayoverDuration)
	}
}

func TestParseSegments_BadArrivalTime(t *testing.T) {
	segs := []FlightSegment{{
		DepartureAirport: Airport{ID: "JFK", Time: "2026-04-01 08:00"},
		ArrivalAirport:   Airport{ID: "LHR", Time: "bad-time"},
		FlightNumber:     "BA 117",
	}}

	if _, err := parseSegments(segs, nil); err == nil {
		t.Fatal("expected error for invalid arrival time, got nil")
	}
}

func TestParseFlightNumber(t *testing.T) {
	tests := []struct {
		input       string
		wantAirline string
		wantNumber  string
	}{
		{input: "TG 332", wantAirline: "TG", wantNumber: "332"},
		{input: "BA 117", wantAirline: "BA", wantNumber: "117"},
		{input: "AA100", wantAirline: "AA100", wantNumber: ""},
		{input: "", wantAirline: "", wantNumber: ""},
		{input: "6E 1234", wantAirline: "6E", wantNumber: "1234"},
		{input: "UA 123 456", wantAirline: "UA", wantNumber: "123 456"}, // SplitN(_, 2) keeps remainder
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			airline, number := parseFlightNumber(tt.input)
			if airline != tt.wantAirline {
				t.Errorf("airline = %q, want %q", airline, tt.wantAirline)
			}
			if number != tt.wantNumber {
				t.Errorf("number = %q, want %q", number, tt.wantNumber)
			}
		})
	}
}

func TestMapCabinClass(t *testing.T) {
	tests := []struct {
		input string
		want  types.CabinClass
	}{
		{"Economy", types.CabinEconomy},
		{"economy", types.CabinEconomy},
		{"Premium Economy", types.CabinPremiumEconomy},
		{"premium economy", types.CabinPremiumEconomy},
		{"Business", types.CabinBusiness},
		{"business", types.CabinBusiness},
		{"First", types.CabinFirst},
		{"first", types.CabinFirst},
		{"", types.CabinEconomy},
		{"Unknown Class", types.CabinEconomy},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := mapCabinClass(tt.input)
			if got != tt.want {
				t.Errorf("mapCabinClass(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseMultiCityResult(t *testing.T) {
	leg1 := FlightGroup{
		Flights: []FlightSegment{{
			DepartureAirport: Airport{Name: "DEL Airport", ID: "DEL", Time: "2026-03-24 06:00"},
			ArrivalAirport:   Airport{Name: "HKG Airport", ID: "HKG", Time: "2026-03-24 14:00"},
			Duration:         480,
			Airline:          "Cathay Pacific",
			FlightNumber:     "CX 694",
			TravelClass:      "Economy",
		}},
		TotalDuration: 480,
		Price:         300,
	}

	leg2 := FlightGroup{
		Flights: []FlightSegment{{
			DepartureAirport: Airport{Name: "HKG Airport", ID: "HKG", Time: "2026-03-27 10:00"},
			ArrivalAirport:   Airport{Name: "YYZ Airport", ID: "YYZ", Time: "2026-03-27 22:00"},
			Duration:         720,
			Airline:          "Air Canada",
			FlightNumber:     "AC 16",
			TravelClass:      "Economy",
		}},
		TotalDuration: 720,
		Price:         500,
	}

	mcr := MultiCityResult{
		Leg1:  leg1,
		Leg2:  leg2,
		Price: 800,
	}

	it, err := ParseMultiCityResult(mcr, "Hong Kong", "HKG")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Total price is the combined price.
	if it.TotalPrice.Amount != 800 {
		t.Errorf("TotalPrice = %v, want 800", it.TotalPrice.Amount)
	}
	if it.TotalPrice.Currency != "USD" {
		t.Errorf("TotalPrice.Currency = %q, want USD", it.TotalPrice.Currency)
	}

	// Two legs.
	if len(it.Legs) != 2 {
		t.Fatalf("legs = %d, want 2", len(it.Legs))
	}

	// Leg 1 carries the total price; leg 2 has zero price.
	if it.Legs[0].Flight.Price.Amount != 800 {
		t.Errorf("leg1 price = %v, want 800", it.Legs[0].Flight.Price.Amount)
	}
	if it.Legs[1].Flight.Price.Amount != 0 {
		t.Errorf("leg2 price = %v, want 0", it.Legs[1].Flight.Price.Amount)
	}

	// Stopover is on leg 1.
	if it.Legs[0].Stopover == nil {
		t.Fatal("leg1 stopover is nil, want non-nil")
	}
	if it.Legs[0].Stopover.City != "Hong Kong" {
		t.Errorf("stopover city = %q, want %q", it.Legs[0].Stopover.City, "Hong Kong")
	}
	if it.Legs[0].Stopover.Airport != "HKG" {
		t.Errorf("stopover airport = %q, want %q", it.Legs[0].Stopover.Airport, "HKG")
	}

	// Stopover duration: leg1 arrives 2026-03-24 14:00, leg2 departs 2026-03-27 10:00 = 68h.
	wantStopover := 68 * time.Hour
	if it.Legs[0].Stopover.Duration != wantStopover {
		t.Errorf("stopover duration = %v, want %v", it.Legs[0].Stopover.Duration, wantStopover)
	}

	// Leg 2 has no stopover.
	if it.Legs[1].Stopover != nil {
		t.Errorf("leg2 stopover = %v, want nil", it.Legs[1].Stopover)
	}

	// TotalTravel = leg1 duration + leg2 duration.
	wantTravel := (480 + 720) * time.Minute
	if it.TotalTravel != wantTravel {
		t.Errorf("TotalTravel = %v, want %v", it.TotalTravel, wantTravel)
	}

	// TotalTrip = last arrival - first departure.
	// 2026-03-27 22:00 - 2026-03-24 06:00 = 3 days 16 hours = 88 hours.
	wantTrip := 88 * time.Hour
	if it.TotalTrip != wantTrip {
		t.Errorf("TotalTrip = %v, want %v", it.TotalTrip, wantTrip)
	}
}

func TestParseMultiCityResult_InvalidLeg1(t *testing.T) {
	mcr := MultiCityResult{
		Leg1: FlightGroup{
			Flights: []FlightSegment{{
				DepartureAirport: Airport{Time: "bad"},
				ArrivalAirport:   Airport{Time: "2026-03-24 14:00"},
				FlightNumber:     "XX 1",
			}},
		},
		Leg2: FlightGroup{
			Flights: []FlightSegment{{
				DepartureAirport: Airport{Time: "2026-03-27 10:00"},
				ArrivalAirport:   Airport{Time: "2026-03-27 22:00"},
				FlightNumber:     "YY 2",
			}},
		},
		Price: 500,
	}

	if _, err := ParseMultiCityResult(mcr, "City", "APT"); err == nil {
		t.Fatal("expected error for invalid leg1, got nil")
	}
}

func TestParseMultiCityResult_InvalidLeg2(t *testing.T) {
	mcr := MultiCityResult{
		Leg1: FlightGroup{
			Flights: []FlightSegment{{
				DepartureAirport: Airport{Time: "2026-03-24 06:00"},
				ArrivalAirport:   Airport{Time: "2026-03-24 14:00"},
				FlightNumber:     "XX 1",
			}},
		},
		Leg2: FlightGroup{
			Flights: []FlightSegment{{
				DepartureAirport: Airport{Time: "bad"},
				ArrivalAirport:   Airport{Time: "2026-03-27 22:00"},
				FlightNumber:     "YY 2",
			}},
		},
		Price: 500,
	}

	if _, err := ParseMultiCityResult(mcr, "City", "APT"); err == nil {
		t.Fatal("expected error for invalid leg2, got nil")
	}
}

func TestParseFlightGroup_CarbonEmissions(t *testing.T) {
	g := FlightGroup{
		Flights: []FlightSegment{{
			DepartureAirport: Airport{ID: "DEL", Time: "2026-03-24 03:30"},
			ArrivalAirport:   Airport{ID: "YYZ", Time: "2026-03-24 19:30"},
			Duration:         960,
			Airline:          "Air Canada",
			FlightNumber:     "AC 42",
			TravelClass:      "Economy",
		}},
		TotalDuration: 960,
		Price:         850,
		BookingToken:  "token_co2",
		CarbonEmissions: CarbonEmissions{
			ThisFlight:          1106000,
			TypicalForThisRoute: 949000,
			DifferencePercent:   17,
		},
	}

	f, err := parseFlightGroup(g)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f.CarbonKg != 1106 {
		t.Errorf("CarbonKg = %d, want 1106", f.CarbonKg)
	}
	if f.TypicalCarbonKg != 949 {
		t.Errorf("TypicalCarbonKg = %d, want 949", f.TypicalCarbonKg)
	}
	if f.CarbonDiffPct != 17 {
		t.Errorf("CarbonDiffPct = %d, want 17", f.CarbonDiffPct)
	}
}

func TestParseFlightGroup_CarbonBenchmark_Zero(t *testing.T) {
	g := FlightGroup{
		Flights: []FlightSegment{{
			DepartureAirport: Airport{ID: "DEL", Time: "2026-03-24 03:30"},
			ArrivalAirport:   Airport{ID: "HKG", Time: "2026-03-24 11:30"},
			Duration:         480,
			Airline:          "IndiGo",
			FlightNumber:     "6E 1234",
			TravelClass:      "Economy",
		}},
		TotalDuration: 480,
		Price:         350,
		CarbonEmissions: CarbonEmissions{
			ThisFlight: 500000,
		},
	}

	f, err := parseFlightGroup(g)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f.TypicalCarbonKg != 0 {
		t.Errorf("TypicalCarbonKg = %d, want 0 when TypicalForThisRoute is 0", f.TypicalCarbonKg)
	}
	if f.CarbonDiffPct != 0 {
		t.Errorf("CarbonDiffPct = %d, want 0 when DifferencePercent is 0", f.CarbonDiffPct)
	}
}

func TestParseFlightGroup_CarbonEmissions_Zero(t *testing.T) {
	g := FlightGroup{
		Flights: []FlightSegment{{
			DepartureAirport: Airport{ID: "DEL", Time: "2026-03-24 03:30"},
			ArrivalAirport:   Airport{ID: "HKG", Time: "2026-03-24 11:30"},
			Duration:         480,
			Airline:          "IndiGo",
			FlightNumber:     "6E 1234",
			TravelClass:      "Economy",
		}},
		TotalDuration: 480,
		Price:         350,
	}

	f, err := parseFlightGroup(g)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f.CarbonKg != 0 {
		t.Errorf("CarbonKg = %d, want 0 for missing emissions", f.CarbonKg)
	}
}

func TestParseSegments_Overnight(t *testing.T) {
	segs := []FlightSegment{{
		DepartureAirport: Airport{ID: "DEL", Time: "2026-03-24 23:00"},
		ArrivalAirport:   Airport{ID: "BKK", Time: "2026-03-25 05:00"},
		Duration:         360,
		Airline:          "Thai Airways",
		FlightNumber:     "TG 316",
		TravelClass:      "Economy",
		Overnight:        true,
	}}

	result, err := parseSegments(segs, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result[0].Overnight {
		t.Error("segment should be marked Overnight")
	}
}

func TestParseSegments_NotOvernight(t *testing.T) {
	segs := []FlightSegment{{
		DepartureAirport: Airport{ID: "DEL", Time: "2026-03-24 10:00"},
		ArrivalAirport:   Airport{ID: "BKK", Time: "2026-03-24 16:00"},
		Duration:         360,
		Airline:          "Thai Airways",
		FlightNumber:     "TG 316",
		TravelClass:      "Economy",
		Overnight:        false,
	}}

	result, err := parseSegments(segs, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result[0].Overnight {
		t.Error("segment should not be marked Overnight")
	}
}

func TestParseSegments_Aircraft(t *testing.T) {
	segs := []FlightSegment{{
		DepartureAirport: Airport{ID: "DEL", Time: "2026-03-24 10:00"},
		ArrivalAirport:   Airport{ID: "BKK", Time: "2026-03-24 16:00"},
		Duration:         360,
		Airline:          "Thai Airways",
		FlightNumber:     "TG 316",
		TravelClass:      "Economy",
		Airplane:         "Boeing 787-9",
	}}

	result, err := parseSegments(segs, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result[0].Aircraft != "Boeing 787-9" {
		t.Errorf("Aircraft = %q, want %q", result[0].Aircraft, "Boeing 787-9")
	}
}

func TestParseSegments_Legroom(t *testing.T) {
	segs := []FlightSegment{{
		DepartureAirport: Airport{ID: "DEL", Time: "2026-03-24 10:00"},
		ArrivalAirport:   Airport{ID: "BKK", Time: "2026-03-24 16:00"},
		Duration:         360,
		Airline:          "Thai Airways",
		FlightNumber:     "TG 316",
		TravelClass:      "Economy",
		Legroom:          "30 in",
	}}

	result, err := parseSegments(segs, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result[0].Legroom != "30 in" {
		t.Errorf("Legroom = %q, want %q", result[0].Legroom, "30 in")
	}
}

func TestParseFlightGroup_CarbonEmissions_Rounding(t *testing.T) {
	g := FlightGroup{
		Flights: []FlightSegment{{
			DepartureAirport: Airport{ID: "DEL", Time: "2026-03-24 03:30"},
			ArrivalAirport:   Airport{ID: "HKG", Time: "2026-03-24 11:30"},
			Duration:         480,
			Airline:          "IndiGo",
			FlightNumber:     "6E 1234",
			TravelClass:      "Economy",
		}},
		TotalDuration: 480,
		Price:         350,
		CarbonEmissions: CarbonEmissions{
			ThisFlight: 800, // 800 grams should round to 1 kg, not truncate to 0
		},
	}

	f, err := parseFlightGroup(g)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f.CarbonKg != 1 {
		t.Errorf("CarbonKg = %d, want 1 (800g should round to 1kg)", f.CarbonKg)
	}
}

func TestParsePriceInsights(t *testing.T) {
	resp := &Response{
		PriceInsights: PriceInsights{
			LowestPrice:       320,
			PriceLevel:        "low",
			TypicalPriceRange: []int{400, 800},
		},
	}

	pi := ParsePriceInsights(resp)
	if pi.LowestPrice != 320 {
		t.Errorf("LowestPrice = %v, want 320", pi.LowestPrice)
	}
	if pi.PriceLevel != "low" {
		t.Errorf("PriceLevel = %q, want %q", pi.PriceLevel, "low")
	}
	if pi.TypicalPriceRange != [2]float64{400, 800} {
		t.Errorf("TypicalPriceRange = %v, want [400 800]", pi.TypicalPriceRange)
	}
}

func TestParsePriceInsights_Empty(t *testing.T) {
	resp := &Response{}
	pi := ParsePriceInsights(resp)
	if pi.LowestPrice != 0 {
		t.Errorf("LowestPrice = %v, want 0", pi.LowestPrice)
	}
	if pi.PriceLevel != "" {
		t.Errorf("PriceLevel = %q, want empty", pi.PriceLevel)
	}
	if pi.TypicalPriceRange != [2]float64{0, 0} {
		t.Errorf("TypicalPriceRange = %v, want [0 0]", pi.TypicalPriceRange)
	}
}

func TestParseResponse(t *testing.T) {
	resp := &Response{
		BestFlights: []FlightGroup{{
			Flights: []FlightSegment{{
				DepartureAirport: Airport{ID: "DEL", Time: "2026-03-24 06:00"},
				ArrivalAirport:   Airport{ID: "HKG", Time: "2026-03-24 14:00"},
				Duration:         480,
				Airline:          "CX",
				FlightNumber:     "CX 694",
			}},
			TotalDuration: 480,
			Price:         350,
		}},
		OtherFlights: []FlightGroup{
			{
				Flights: []FlightSegment{{
					DepartureAirport: Airport{ID: "DEL", Time: "2026-03-24 09:00"},
					ArrivalAirport:   Airport{ID: "HKG", Time: "2026-03-24 17:00"},
					Duration:         480,
					Airline:          "AI",
					FlightNumber:     "AI 310",
				}},
				TotalDuration: 480,
				Price:         280,
			},
			// Invalid flight group to test the continue/skip behavior.
			{
				Flights: []FlightSegment{{
					DepartureAirport: Airport{Time: "bad-time"},
					ArrivalAirport:   Airport{Time: "also-bad"},
					FlightNumber:     "XX 1",
				}},
				Price: 100,
			},
		},
	}

	flights, err := ParseResponse(resp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 1 best + 1 valid other = 2 (invalid one is skipped).
	if len(flights) != 2 {
		t.Fatalf("flights = %d, want 2", len(flights))
	}
	if flights[0].Price.Amount != 350 {
		t.Errorf("flights[0].Price = %v, want 350", flights[0].Price.Amount)
	}
	if flights[1].Price.Amount != 280 {
		t.Errorf("flights[1].Price = %v, want 280", flights[1].Price.Amount)
	}
}
