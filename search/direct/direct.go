// Package direct implements a simple origin-to-destination search strategy
// with no stopovers. This is the default strategy for most flight searches.
package direct

import (
	"context"
	"fmt"
	"sort"
	"time"

	"booker/provider"
	"booker/search"
	"booker/types"
)

// DateLayout is the format for parsing date strings.
const DateLayout = "2006-01-02"

// Searcher implements search.Strategy for direct (non-stopover) flights.
type Searcher struct {
	registry *provider.Registry
	ranker   search.Ranker
}

// NewSearcher creates a direct search strategy. If ranker is nil, results
// are returned sorted by price only.
func NewSearcher(registry *provider.Registry, ranker search.Ranker) *Searcher {
	return &Searcher{registry: registry, ranker: ranker}
}

// Name returns the strategy identifier.
func (s *Searcher) Name() string { return "direct" }

// Description returns a human-readable explanation for LLM strategy selection.
func (s *Searcher) Description() string {
	return "Direct flight search from origin to destination with no stopovers. " +
		"Best for simple point-to-point trips or short-haul routes."
}

// Search fetches flights from all providers, filters, converts to itineraries,
// and optionally ranks via LLM.
func (s *Searcher) Search(ctx context.Context, req search.Request) ([]search.Itinerary, error) {
	depDate, err := time.Parse(DateLayout, req.DepartureDate)
	if err != nil {
		return nil, fmt.Errorf("parsing departure date %q: %w", req.DepartureDate, err)
	}

	searchReq := types.SearchRequest{
		Origin:        req.Origin,
		Destination:   req.Destination,
		DepartureDate: depDate,
		Passengers:    req.Passengers,
		CabinClass:    req.CabinClass,
		MaxStops:      req.MaxStops,
	}

	// Fetch from all providers.
	var allFlights []types.Flight
	for _, p := range s.registry.All() {
		results, err := p.Search(ctx, searchReq)
		if err != nil {
			continue
		}
		allFlights = append(allFlights, results...)
	}

	// Filter pipeline.
	allFlights = search.FilterFlights(allFlights)
	allFlights = search.FilterZeroPrices(allFlights)
	allFlights = search.FilterByMaxStops(allFlights, req.MaxStops)
	if req.FlexDays > 0 {
		earliest := depDate.AddDate(0, 0, -req.FlexDays)
		latest := depDate.AddDate(0, 0, req.FlexDays).Add(24*time.Hour - time.Nanosecond)
		allFlights = search.FilterByDateRange(allFlights, earliest, latest)
	}

	// Convert to itineraries.
	itineraries := make([]search.Itinerary, 0, len(allFlights))
	for _, f := range allFlights {
		itineraries = append(itineraries, flightToItinerary(f))
	}

	// Sort by price.
	sort.Slice(itineraries, func(i, j int) bool {
		return itineraries[i].TotalPrice.Amount < itineraries[j].TotalPrice.Amount
	})

	// Rank via LLM if available.
	if s.ranker != nil {
		ranked, err := s.ranker.Rank(ctx, itineraries)
		if err == nil {
			itineraries = ranked
		}
		// On ranker failure, fall back to price-sorted results.
	}

	// Cap results.
	if req.MaxResults > 0 && len(itineraries) > req.MaxResults {
		itineraries = itineraries[:req.MaxResults]
	}

	return itineraries, nil
}

func flightToItinerary(f types.Flight) search.Itinerary {
	itin := search.Itinerary{
		Legs:        []search.Leg{{Flight: f}},
		TotalPrice:  f.Price,
		TotalTravel: f.TotalDuration,
	}

	// Compute total trip duration from first departure to last arrival.
	if len(f.Outbound) > 0 {
		first := f.Outbound[0].DepartureTime
		last := f.Outbound[len(f.Outbound)-1].ArrivalTime
		itin.TotalTrip = last.Sub(first)
	}

	return itin
}
