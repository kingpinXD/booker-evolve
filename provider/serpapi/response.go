// Package serpapi implements the SerpAPI Google Flights provider.
package serpapi

// Response is the top-level SerpAPI Google Flights response.
type Response struct {
	BestFlights   []FlightGroup `json:"best_flights"`
	OtherFlights  []FlightGroup `json:"other_flights"`
	PriceInsights PriceInsights `json:"price_insights"`
}

// FlightGroup is a single itinerary (one or more segments + layovers).
type FlightGroup struct {
	Flights         []FlightSegment `json:"flights"`
	Layovers        []Layover       `json:"layovers"`
	TotalDuration   int             `json:"total_duration"` // minutes
	Price           int             `json:"price"`
	Type            string          `json:"type"`
	BookingToken    string          `json:"booking_token"`
	DepartureToken  string          `json:"departure_token"` // multi-city step 1
	CarbonEmissions CarbonEmissions `json:"carbon_emissions"`
}

// CarbonEmissions holds CO2 data for a flight group.
type CarbonEmissions struct {
	ThisFlight          int `json:"this_flight"`            // grams CO2
	TypicalForThisRoute int `json:"typical_for_this_route"` // grams CO2
	DifferencePercent   int `json:"difference_percent"`     // vs typical
}

// MultiCityResult pairs leg1 and leg2 flight groups with the combined price.
type MultiCityResult struct {
	Leg1  FlightGroup
	Leg2  FlightGroup
	Price int // combined price from step 2
}

// FlightSegment is a single non-stop flight within an itinerary.
type FlightSegment struct {
	DepartureAirport Airport `json:"departure_airport"`
	ArrivalAirport   Airport `json:"arrival_airport"`
	Duration         int     `json:"duration"` // minutes
	Airplane         string  `json:"airplane"`
	Airline          string  `json:"airline"`
	AirlineLogo      string  `json:"airline_logo"`
	TravelClass      string  `json:"travel_class"`
	FlightNumber     string  `json:"flight_number"` // e.g. "TG 332"
	Legroom          string  `json:"legroom"`
	Overnight        bool    `json:"overnight"`
}

// Airport holds departure or arrival airport info.
type Airport struct {
	Name string `json:"name"`
	ID   string `json:"id"`   // IATA code
	Time string `json:"time"` // "2026-03-24 03:30"
}

// Layover describes a connection between two segments.
type Layover struct {
	Duration  int    `json:"duration"` // minutes
	Name      string `json:"name"`
	ID        string `json:"id"` // IATA code
	Overnight bool   `json:"overnight"`
}

// PriceInsights contains pricing context from Google Flights.
type PriceInsights struct {
	LowestPrice       int    `json:"lowest_price"`
	PriceLevel        string `json:"price_level"`
	TypicalPriceRange []int  `json:"typical_price_range"`
}
