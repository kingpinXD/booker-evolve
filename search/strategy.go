package search

import (
	"context"
	"time"

	"booker/types"
)

// Strategy is a search approach that finds flight itineraries.
// Each implementation (direct, multicity, etc.) encodes a different
// way to break down and search for flights.
type Strategy interface {
	// Name returns a short identifier (e.g. "direct", "multicity").
	Name() string
	// Description returns a human-readable explanation of what this strategy
	// does. Used by the LLM picker to understand available options.
	Description() string
	// Search executes the strategy and returns ranked itineraries.
	Search(ctx context.Context, req Request) ([]Itinerary, error)
}

// Request is the common input for all search strategies.
type Request struct {
	Origin            string // IATA code, e.g. "DEL"
	Destination       string // IATA code, e.g. "YYZ"
	DepartureDate     string // YYYY-MM-DD
	ReturnDate        string // YYYY-MM-DD, empty for one-way
	Leg2Date          string // YYYY-MM-DD, for multicity second leg departure
	Passengers        int
	CabinClass        types.CabinClass
	FlexDays          int
	MaxStops          int           // -1 = no limit
	MaxPrice          float64       // 0 = no limit (USD)
	PreferredAlliance string        // "Star Alliance", "OneWorld", "SkyTeam", or "" for no filter
	DepartureAfter    string        // time-of-day "HH:MM" e.g. "06:00" — only keep flights departing at/after this
	DepartureBefore   string        // time-of-day "HH:MM" e.g. "22:00" — only keep flights departing at/before this
	ArrivalAfter      string        // time-of-day "HH:MM" — only keep flights arriving at/after this
	ArrivalBefore     string        // time-of-day "HH:MM" — only keep flights arriving at/before this
	MaxDuration       time.Duration // max total flight duration; 0 = no limit
	SortBy            string        // "price" (default), "duration", or "departure"
	AvoidAirlines     string        // comma-separated IATA codes to exclude (e.g. "BA,LH")
	PreferredAirlines string        // comma-separated IATA codes to keep (e.g. "AC,UA")
	MaxResults        int
	Context           string // User's natural language context/preferences
}

// Ranker scores and reorders itineraries. Decoupled from strategies so
// any strategy can reuse ranking logic (e.g. LLM-based ranking).
type Ranker interface {
	Rank(ctx context.Context, itineraries []Itinerary) ([]Itinerary, error)
}
