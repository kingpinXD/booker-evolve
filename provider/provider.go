package provider

import (
	"context"
	"fmt"
	"sync"

	"booker/config"
	"booker/types"
)

// Provider is the interface every flight data source must implement.
type Provider interface {
	// Name returns the typed provider identifier.
	Name() config.ProviderName

	// Search performs a flight search and returns normalized results.
	// Implementations must respect context cancellation.
	Search(ctx context.Context, req types.SearchRequest) ([]types.Flight, error)
}

// MultiCitySearcher is an optional interface for providers that support
// native multi-city search with combined pricing.
type MultiCitySearcher interface {
	SearchMultiCity(ctx context.Context, req MultiCityRequest) ([]MultiCityResult, error)
}

// MultiCityRequest defines a multi-city search with a stopover.
type MultiCityRequest struct {
	Origin      string
	Stopover    string
	Destination string
	Leg1Date    string // YYYY-MM-DD
	Leg2Date    string // YYYY-MM-DD
	Passengers  int
	CabinClass  types.CabinClass
	TopN        int
}

// MultiCityResult pairs two flight groups with a combined price.
type MultiCityResult struct {
	Leg1  types.Flight
	Leg2  types.Flight
	Price types.Money
}

// Registry holds registered providers and allows lookup by name.
type Registry struct {
	mu        sync.RWMutex
	providers map[config.ProviderName]Provider
}

func NewRegistry() *Registry {
	return &Registry{providers: make(map[config.ProviderName]Provider)}
}

// Register adds a provider. Returns an error if a provider with the
// same name is already registered.
func (r *Registry) Register(p Provider) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	name := p.Name()
	if _, exists := r.providers[name]; exists {
		return fmt.Errorf("provider %q already registered", name)
	}
	r.providers[name] = p
	return nil
}

// Get returns a provider by name.
func (r *Registry) Get(name config.ProviderName) (Provider, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, ok := r.providers[name]
	return p, ok
}

// All returns every registered provider.
func (r *Registry) All() []Provider {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]Provider, 0, len(r.providers))
	for _, p := range r.providers {
		out = append(out, p)
	}
	return out
}
