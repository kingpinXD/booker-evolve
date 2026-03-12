package search

import (
	"context"
	"fmt"
	"sync"
)

// CompositeStrategy runs multiple strategies concurrently and merges their results.
// Useful when the Picker cannot determine a single best strategy and wants to
// compare results from different approaches (e.g., direct vs multicity).
type CompositeStrategy struct {
	strategies []Strategy
	ranker     Ranker
}

// NewCompositeStrategy creates a composite that fans out to all child strategies.
// An optional ranker re-orders the merged results. Pass nil to skip ranking.
func NewCompositeStrategy(ranker Ranker, strategies ...Strategy) *CompositeStrategy {
	return &CompositeStrategy{strategies: strategies, ranker: ranker}
}

// Name returns "composite".
func (c *CompositeStrategy) Name() string { return "composite" }

// Description returns a human-readable explanation.
func (c *CompositeStrategy) Description() string {
	return "Runs multiple strategies in parallel and merges their results."
}

// Search executes all child strategies concurrently, merges results, deduplicates,
// and optionally re-ranks. Partial failures are tolerated -- if at least one
// strategy succeeds, its results are returned. Returns an error only if all fail.
func (c *CompositeStrategy) Search(ctx context.Context, req Request) ([]Itinerary, error) {
	if len(c.strategies) == 0 {
		return nil, nil
	}

	type result struct {
		itins []Itinerary
		err   error
	}

	results := make([]result, len(c.strategies))
	var wg sync.WaitGroup
	for i, s := range c.strategies {
		wg.Add(1)
		go func(idx int, strat Strategy) {
			defer wg.Done()
			itins, err := strat.Search(ctx, req)
			results[idx] = result{itins: itins, err: err}
		}(i, s)
	}
	wg.Wait()

	// Merge results, tolerating partial failures.
	var merged []Itinerary
	var errs []error
	for _, r := range results {
		if r.err != nil {
			errs = append(errs, r.err)
			continue
		}
		merged = append(merged, r.itins...)
	}

	if len(merged) == 0 && len(errs) > 0 {
		return nil, fmt.Errorf("all strategies failed: %v", errs[0])
	}

	merged = DeduplicateItineraries(merged)

	if c.ranker != nil {
		ranked, err := c.ranker.Rank(ctx, merged)
		if err == nil {
			merged = ranked
		}
	}

	merged = DiversifyResults(merged, req.MaxResults)

	return merged, nil
}

// DeduplicateItineraries removes itineraries with identical route and price.
func DeduplicateItineraries(itins []Itinerary) []Itinerary {
	type key struct {
		route string
		price float64
	}
	seen := make(map[key]bool, len(itins))
	out := make([]Itinerary, 0, len(itins))
	for _, itin := range itins {
		k := key{route: ItinRoute(itin), price: itin.TotalPrice.Amount}
		if seen[k] {
			continue
		}
		seen[k] = true
		out = append(out, itin)
	}
	return out
}

// ItinRoute builds a string key from the itinerary's segments.
func ItinRoute(itin Itinerary) string {
	route := ""
	for _, leg := range itin.Legs {
		for _, seg := range leg.Flight.Outbound {
			route += seg.Origin + "-" + seg.Destination + "|"
		}
	}
	return route
}
