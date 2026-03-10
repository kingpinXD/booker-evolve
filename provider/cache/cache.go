// Package cache provides a caching decorator for flight providers.
//
// It wraps any provider.Provider and stores search results on disk as JSON.
// On cache hit, it returns stored results without making an API call.
// On cache miss, it calls the underlying provider and stores the result.
//
// Cache key is derived from the search request fields (origin, destination,
// date, passengers, cabin class). Files are stored in a configurable directory
// with the naming pattern: {provider}_{origin}_{dest}_{date}_{cabin}.json
//
// Usage:
//
//	realProvider := serpapi.New(cfg, httpClient)
//	cached := cache.Wrap(realProvider, ".cache/flights")
//	registry.Register(cached) // same interface as any provider
package cache

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"booker/config"
	"booker/provider"
	"booker/types"
)

// Provider wraps another provider with disk-based caching.
type Provider struct {
	inner provider.Provider
	dir   string
	ttl   time.Duration
}

// Wrap creates a caching decorator around the given provider.
// Results are cached in dir with the given TTL. A zero TTL means
// cached entries never expire (useful during development).
func Wrap(inner provider.Provider, dir string, ttl time.Duration) *Provider {
	return &Provider{inner: inner, dir: dir, ttl: ttl}
}

// Name returns the underlying provider's name.
func (p *Provider) Name() config.ProviderName {
	return p.inner.Name()
}

// Search checks the disk cache first. On miss, calls the underlying
// provider and stores the result.
func (p *Provider) Search(ctx context.Context, req types.SearchRequest) ([]types.Flight, error) {
	path := p.cachePath(req)

	if flights, ok := p.load(path); ok {
		log.Printf("[cache] HIT %s %s→%s %s (%d flights)",
			p.inner.Name(), req.Origin, req.Destination,
			req.DepartureDate.Format("2006-01-02"), len(flights))
		return flights, nil
	}

	log.Printf("[cache] MISS %s %s→%s %s",
		p.inner.Name(), req.Origin, req.Destination,
		req.DepartureDate.Format("2006-01-02"))

	flights, err := p.inner.Search(ctx, req)
	if err != nil {
		return nil, err
	}

	p.store(path, flights)
	return flights, nil
}

func (p *Provider) cachePath(req types.SearchRequest) string {
	key := fmt.Sprintf("%s_%s_%s_%s_%d_%s",
		p.inner.Name(),
		req.Origin, req.Destination,
		req.DepartureDate.Format("2006-01-02"),
		req.Passengers,
		req.CabinClass,
	)
	hash := fmt.Sprintf("%x", sha256.Sum256([]byte(key)))[:12]
	filename := fmt.Sprintf("%s_%s_%s_%s_%s.json",
		p.inner.Name(), req.Origin, req.Destination,
		req.DepartureDate.Format("2006-01-02"), hash)
	return filepath.Join(p.dir, filename)
}

// cacheEntry wraps flights with a timestamp for TTL checks.
type cacheEntry struct {
	CachedAt time.Time      `json:"cached_at"`
	Flights  []types.Flight `json:"flights"`
}

func (p *Provider) load(path string) ([]types.Flight, bool) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, false
	}

	var entry cacheEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return nil, false
	}

	if p.ttl > 0 && time.Since(entry.CachedAt) > p.ttl {
		return nil, false
	}

	return entry.Flights, true
}

// SearchMultiCity checks if the inner provider implements MultiCitySearcher
// and if so, provides caching for multi-city search results.
func (p *Provider) SearchMultiCity(ctx context.Context, req provider.MultiCityRequest) ([]provider.MultiCityResult, error) {
	mc, ok := p.inner.(provider.MultiCitySearcher)
	if !ok {
		return nil, fmt.Errorf("provider %s does not support multi-city search", p.inner.Name())
	}

	path := p.multiCityCachePath(req)

	if results, ok := p.loadMultiCity(path); ok {
		log.Printf("[cache] HIT mc %s %s→%s→%s (%d results)",
			p.inner.Name(), req.Origin, req.Stopover, req.Destination, len(results))
		return results, nil
	}

	log.Printf("[cache] MISS mc %s %s→%s→%s",
		p.inner.Name(), req.Origin, req.Stopover, req.Destination)

	results, err := mc.SearchMultiCity(ctx, req)
	if err != nil {
		return nil, err
	}

	p.storeMultiCity(path, results)
	return results, nil
}

func (p *Provider) multiCityCachePath(req provider.MultiCityRequest) string {
	key := fmt.Sprintf("%s_mc_%s_%s_%s_%s_%s_%d_%s",
		p.inner.Name(),
		req.Origin, req.Stopover, req.Destination,
		req.Leg1Date, req.Leg2Date,
		req.Passengers, req.CabinClass,
	)
	hash := fmt.Sprintf("%x", sha256.Sum256([]byte(key)))[:12]
	filename := fmt.Sprintf("%s_mc_%s_%s_%s_%s_%s.json",
		p.inner.Name(), req.Origin, req.Stopover, req.Destination,
		req.Leg1Date, hash)
	return filepath.Join(p.dir, filename)
}

type multiCityCacheEntry struct {
	CachedAt time.Time                  `json:"cached_at"`
	Results  []provider.MultiCityResult `json:"results"`
}

func (p *Provider) loadMultiCity(path string) ([]provider.MultiCityResult, bool) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, false
	}

	var entry multiCityCacheEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return nil, false
	}

	if p.ttl > 0 && time.Since(entry.CachedAt) > p.ttl {
		return nil, false
	}

	return entry.Results, true
}

func (p *Provider) storeMultiCity(path string, results []provider.MultiCityResult) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		log.Printf("[cache] failed to create dir: %v", err)
		return
	}

	entry := multiCityCacheEntry{
		CachedAt: time.Now(),
		Results:  results,
	}
	data, err := json.MarshalIndent(entry, "", "  ")
	if err != nil {
		log.Printf("[cache] failed to marshal mc: %v", err)
		return
	}

	if err := os.WriteFile(path, data, 0o644); err != nil {
		log.Printf("[cache] failed to write mc %s: %v", path, err)
	}
}

func (p *Provider) store(path string, flights []types.Flight) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		log.Printf("[cache] failed to create dir: %v", err)
		return
	}

	entry := cacheEntry{
		CachedAt: time.Now(),
		Flights:  flights,
	}
	data, err := json.MarshalIndent(entry, "", "  ")
	if err != nil {
		log.Printf("[cache] failed to marshal: %v", err)
		return
	}

	if err := os.WriteFile(path, data, 0o644); err != nil {
		log.Printf("[cache] failed to write %s: %v", path, err)
	}
}
