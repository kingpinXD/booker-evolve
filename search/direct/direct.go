// Package direct implements a simple origin-to-destination search strategy
// with no stopovers. This is the default strategy for most flight searches.
package direct

import (
	"context"
	"fmt"
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
// and optionally ranks via LLM. When req.ReturnDate is set, it searches both
// directions and combines outbound x return into 2-leg itineraries.
func (s *Searcher) Search(ctx context.Context, req search.Request) ([]search.Itinerary, error) {
	outbound, err := s.searchFlights(ctx, req.Origin, req.Destination, req.DepartureDate, req)
	if err != nil {
		return nil, err
	}

	var itineraries []search.Itinerary

	switch {
	case req.ReturnDate != "":
		// Round-trip: search return flights and combine with outbound.
		returnFlights, err := s.searchFlights(ctx, req.Destination, req.Origin, req.ReturnDate, req)
		if err != nil {
			return nil, err
		}
		itineraries = combineRoundTrip(outbound, returnFlights)
	default:
		// One-way: convert each outbound flight to a single-leg itinerary.
		itineraries = make([]search.Itinerary, 0, len(outbound))
		for _, f := range outbound {
			itineraries = append(itineraries, flightToItinerary(f))
		}
	}

	// Sort results (default: price).
	search.SortResults(itineraries, req.SortBy)

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

// searchFlights handles date expansion, provider fetching, and filtering for
// a single direction (origin -> dest on the given date string).
func (s *Searcher) searchFlights(ctx context.Context, origin, dest, dateStr string, req search.Request) ([]types.Flight, error) {
	baseDate, err := time.Parse(DateLayout, dateStr)
	if err != nil {
		return nil, fmt.Errorf("parsing date %q: %w", dateStr, err)
	}

	// Build the list of dates to search. When FlexDays > 0, search each
	// date in the range [base-flex, base+flex] to get results across all days.
	dates := []time.Time{baseDate}
	if req.FlexDays > 0 {
		dates = make([]time.Time, 0, 2*req.FlexDays+1)
		for d := -req.FlexDays; d <= req.FlexDays; d++ {
			dates = append(dates, baseDate.AddDate(0, 0, d))
		}
	}

	// Fetch from all providers for each date.
	var flights []types.Flight
	for _, date := range dates {
		searchReq := types.SearchRequest{
			Origin:        origin,
			Destination:   dest,
			DepartureDate: date,
			Passengers:    req.Passengers,
			CabinClass:    req.CabinClass,
			MaxStops:      req.MaxStops,
		}
		for _, p := range s.registry.All() {
			results, err := p.Search(ctx, searchReq)
			if err != nil {
				continue
			}
			flights = append(flights, results...)
		}
	}

	// Filter pipeline.
	flights = search.FilterFlights(flights)
	flights = search.FilterZeroPrices(flights)
	flights = search.FilterByMaxStops(flights, req.MaxStops)
	flights = search.FilterByMaxPrice(flights, req.MaxPrice)
	flights = search.FilterByAlliance(flights, req.PreferredAlliance)
	flights = search.FilterByDepartureTime(flights, req.DepartureAfter, req.DepartureBefore)
	if req.FlexDays > 0 {
		earliest := baseDate.AddDate(0, 0, -req.FlexDays)
		latest := baseDate.AddDate(0, 0, req.FlexDays).Add(24*time.Hour - time.Nanosecond)
		flights = search.FilterByDateRange(flights, earliest, latest)
	}

	return flights, nil
}

// combineRoundTrip pairs every outbound flight with every return flight into
// 2-leg itineraries with summed prices and computed trip durations.
func combineRoundTrip(outbound, returnFlights []types.Flight) []search.Itinerary {
	itineraries := make([]search.Itinerary, 0, len(outbound)*len(returnFlights))
	for _, out := range outbound {
		for _, ret := range returnFlights {
			itin := search.Itinerary{
				Legs:        []search.Leg{{Flight: out}, {Flight: ret}},
				TotalPrice:  types.Money{Amount: out.Price.Amount + ret.Price.Amount, Currency: out.Price.Currency},
				TotalTravel: out.TotalDuration + ret.TotalDuration,
			}
			// TotalTrip = return arrival - outbound departure (wall-clock).
			if len(out.Outbound) > 0 && len(ret.Outbound) > 0 {
				firstDep := out.Outbound[0].DepartureTime
				lastArr := ret.Outbound[len(ret.Outbound)-1].ArrivalTime
				itin.TotalTrip = lastArr.Sub(firstDep)
			}
			itineraries = append(itineraries, itin)
		}
	}
	return itineraries
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
