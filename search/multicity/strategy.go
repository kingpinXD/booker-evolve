package multicity

import (
	"context"

	"booker/search"
)

// Strategy adapts the multicity Searcher to implement search.Strategy.
type Strategy struct {
	searcher *Searcher
	leg2Date string // provided at construction time from CLI flag
}

// NewStrategy wraps a Searcher as a search.Strategy.
func NewStrategy(searcher *Searcher, leg2Date string) *Strategy {
	return &Strategy{searcher: searcher, leg2Date: leg2Date}
}

// Name returns the strategy identifier.
func (s *Strategy) Name() string { return "multicity" }

// Description returns a human-readable explanation for LLM strategy selection.
func (s *Strategy) Description() string {
	return "Multi-city search with a stopover. Breaks a long-haul journey into " +
		"two legs with a 2-6 day halt in an intermediate city. Best for long " +
		"routes where direct options are expensive or unavailable, or when the " +
		"traveler wants to visit an extra city."
}

// Search maps search.Request to SearchParams and delegates to the Searcher.
func (s *Strategy) Search(ctx context.Context, req search.Request) ([]search.Itinerary, error) {
	return s.searcher.Search(ctx, s.toSearchParams(req))
}

func (s *Strategy) toSearchParams(req search.Request) SearchParams {
	return SearchParams{
		Origin:          req.Origin,
		Destination:     req.Destination,
		DepartureDate:   req.DepartureDate,
		Leg2Date:        s.leg2Date,
		Passengers:      req.Passengers,
		CabinClass:      req.CabinClass,
		FlexDays:        req.FlexDays,
		MaxLayoversLeg1:   req.MaxStops,
		MaxLayoversLeg2:   req.MaxStops,
		MaxResults:        req.MaxResults,
		PreferredAlliance: req.PreferredAlliance,
		MaxPrice:          req.MaxPrice,
	}
}
