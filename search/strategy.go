package search

import (
	"context"

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
	Origin        string           // IATA code, e.g. "DEL"
	Destination   string           // IATA code, e.g. "YYZ"
	DepartureDate string           // YYYY-MM-DD
	ReturnDate    string           // YYYY-MM-DD, empty for one-way
	Passengers    int
	CabinClass    types.CabinClass
	FlexDays      int
	MaxStops      int // -1 = no limit
	MaxResults    int
	Context       string // User's natural language context/preferences
}

// Ranker scores and reorders itineraries. Decoupled from strategies so
// any strategy can reuse ranking logic (e.g. LLM-based ranking).
type Ranker interface {
	Rank(ctx context.Context, itineraries []Itinerary) ([]Itinerary, error)
}
