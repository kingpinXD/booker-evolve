package types

import (
	"time"

	"booker/config"
)

// CabinClass represents the travel class for a flight search.
type CabinClass string

const (
	CabinEconomy        CabinClass = "economy"
	CabinPremiumEconomy CabinClass = "premium_economy"
	CabinBusiness       CabinClass = "business"
	CabinFirst          CabinClass = "first"
)

const MaxStopsAny = -1

// Layover bounds — connections within a single leg (you stay in the airport).
const (
	MaxLayover = 6 * time.Hour
	MinLayover = 1 * time.Hour
)

// Stopover bounds — the gap between legs (you leave the airport, explore the city).
const (
	DefaultMinStopover = 48 * time.Hour  // 2 days
	DefaultMaxStopover = 144 * time.Hour // 6 days
)

// SearchRequest is the normalized input for any flight search.
type SearchRequest struct {
	Origin        string // IATA airport code, e.g. "JFK"
	Destination   string // IATA airport code, e.g. "LHR"
	DepartureDate time.Time
	ReturnDate    time.Time // zero value means one-way
	Passengers    int
	CabinClass    CabinClass
	MaxStops      int // MaxStopsAny means no preference

	// SortBy hints to the provider how to sort results.
	// Providers may ignore this if they don't support sorting.
	SortBy string // e.g. "QUALITY", "PRICE", "DURATION"

	// Provider-specific overrides. These are optional and allow callers
	// to pass provider-native location IDs instead of raw IATA codes.
	OriginKiwiID      string // e.g. "City:new_delhi_in", "Airport:DEL"
	DestinationKiwiID string // e.g. "City:toronto_ca", "Airport:YYZ"
}

// IsRoundTrip reports whether the search includes a return leg.
func (r SearchRequest) IsRoundTrip() bool {
	return !r.ReturnDate.IsZero()
}

// Flight is a single itinerary option returned by a provider.
// All providers normalize their API responses into this shape.
type Flight struct {
	Provider      config.ProviderName
	Price         Money
	Outbound      []Segment // outbound leg segments
	Return        []Segment // return leg segments (empty for one-way)
	TotalDuration time.Duration
	BookingURL    string
	SeatsLeft     int // 0 means unknown
	BagsIncluded  BagsIncluded
}

// Stops returns the number of connections on the outbound leg.
func (f Flight) Stops() int {
	if len(f.Outbound) == 0 {
		return 0
	}
	return len(f.Outbound) - 1
}

// BagsIncluded describes what baggage is included in the price.
type BagsIncluded struct {
	HandBags    int
	CheckedBags int
}

// Segment is a single non-stop portion of a flight.
type Segment struct {
	Airline          string // IATA carrier code, e.g. "BA"
	AirlineName      string
	FlightNumber     string // e.g. "BA117"
	Origin           string // IATA airport code
	OriginName       string // full airport name
	OriginCity       string
	Destination      string // IATA airport code
	DestinationName  string
	DestinationCity  string
	DepartureTime    time.Time
	ArrivalTime      time.Time
	Duration         time.Duration
	CabinClass       CabinClass
	OperatingCarrier string        // if different from marketing carrier
	LayoverDuration  time.Duration // time until next segment (0 if last)
}

// Money represents a price with its currency.
type Money struct {
	Amount   float64
	Currency string // ISO 4217, e.g. "USD"
}

// SearchResult is what the aggregator produces after merging all providers.
type SearchResult struct {
	Request SearchRequest
	Flights []Flight
	Errors  []ProviderError
}

// ProviderError captures a failure from a single provider so remaining
// providers can still return results.
type ProviderError struct {
	Provider config.ProviderName
	Err      error
}

func (e ProviderError) Error() string {
	return string(e.Provider) + ": " + e.Err.Error()
}
