// Package nearby implements a search strategy that expands origin and
// destination to include nearby airports from the same metro cluster.
// For example, searching JFK->YYZ also searches EWR->YYZ, LGA->YYZ,
// JFK->YTZ, etc., finding cheaper fares at alternate airports.
package nearby

import (
	"context"
	"sync"

	"booker/search"
)

// Searcher wraps a delegate strategy and fans out searches to nearby airports.
type Searcher struct {
	delegate search.Strategy
}

// NewSearcher creates a nearby-airport search strategy that expands
// origin/destination to cluster siblings before delegating.
func NewSearcher(delegate search.Strategy) *Searcher {
	return &Searcher{delegate: delegate}
}

// Name returns the strategy identifier.
func (s *Searcher) Name() string { return "nearby" }

// Description returns a human-readable explanation for LLM strategy selection.
func (s *Searcher) Description() string {
	return "Expands origin and destination to nearby airports in the same metro area, " +
		"searching all combinations to find cheaper fares at alternate airports."
}

// Search expands origin/destination to include cluster airports, fans out
// delegate searches concurrently, then merges, deduplicates, sorts by
// req.SortBy (defaulting to price), and caps at MaxResults.
func (s *Searcher) Search(ctx context.Context, req search.Request) ([]search.Itinerary, error) {
	origins := expandCode(req.Origin)
	dests := expandCode(req.Destination)

	type result struct {
		itins []search.Itinerary
		err   error
	}

	pairs := make([]result, 0, len(origins)*len(dests))
	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, o := range origins {
		for _, d := range dests {
			wg.Add(1)
			go func(origin, dest string) {
				defer wg.Done()
				r := req
				r.Origin = origin
				r.Destination = dest
				itins, err := s.delegate.Search(ctx, r)
				mu.Lock()
				pairs = append(pairs, result{itins: itins, err: err})
				mu.Unlock()
			}(o, d)
		}
	}
	wg.Wait()

	// Merge results, tolerating partial failures.
	var merged []search.Itinerary
	var lastErr error
	for _, r := range pairs {
		if r.err != nil {
			lastErr = r.err
			continue
		}
		merged = append(merged, r.itins...)
	}

	if len(merged) == 0 && lastErr != nil {
		return nil, lastErr
	}

	merged = search.DeduplicateItineraries(merged)

	search.SortResults(merged, req.SortBy)

	if req.MaxResults > 0 && len(merged) > req.MaxResults {
		merged = merged[:req.MaxResults]
	}

	return merged, nil
}

// expandCode returns the full cluster for an IATA code (including the code
// itself). If the code is not in any cluster, returns just the code.
func expandCode(code string) []string {
	nearby := search.NearbyAirports(code)
	if len(nearby) == 0 {
		return []string{code}
	}
	// Cluster = code + its siblings.
	cluster := make([]string, 0, len(nearby)+1)
	cluster = append(cluster, code)
	cluster = append(cluster, nearby...)
	return cluster
}

