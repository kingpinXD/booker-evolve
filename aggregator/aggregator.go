package aggregator

import (
	"context"
	"sort"
	"sync"

	"booker/provider"
	"booker/types"
)

// Aggregator fans out a search to all registered providers concurrently
// and merges results into a single SearchResult.
type Aggregator struct {
	registry *provider.Registry
}

func New(registry *provider.Registry) *Aggregator {
	return &Aggregator{registry: registry}
}

// Search queries all providers in parallel, collects results, and returns
// them sorted by price ascending. Individual provider failures are captured
// as ProviderErrors rather than aborting the whole search.
func (a *Aggregator) Search(ctx context.Context, req types.SearchRequest) types.SearchResult {
	providers := a.registry.All()

	var (
		mu      sync.Mutex
		flights []types.Flight
		errs    []types.ProviderError
		wg      sync.WaitGroup
	)

	for _, p := range providers {
		wg.Add(1)
		go func(p provider.Provider) {
			defer wg.Done()
			results, err := p.Search(ctx, req)

			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				errs = append(errs, types.ProviderError{
					Provider: p.Name(),
					Err:      err,
				})
				return
			}
			flights = append(flights, results...)
		}(p)
	}

	wg.Wait()

	sort.Slice(flights, func(i, j int) bool {
		return flights[i].Price.Amount < flights[j].Price.Amount
	})

	return types.SearchResult{
		Request: req,
		Flights: flights,
		Errors:  errs,
	}
}
