package cache

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"booker/config"
	"booker/provider"
	"booker/types"
)

// fakeProvider returns a fixed set of flights.
type fakeProvider struct {
	calls   int
	flights []types.Flight
}

func (f *fakeProvider) Name() config.ProviderName { return "fake" }

func (f *fakeProvider) Search(_ context.Context, _ types.SearchRequest) ([]types.Flight, error) {
	f.calls++
	return f.flights, nil
}

func TestCacheHitAndMiss(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "cache")

	fake := &fakeProvider{
		flights: []types.Flight{
			{
				Provider: "fake",
				Price:    types.Money{Amount: 264, Currency: "USD"},
				Outbound: []types.Segment{{
					Origin:      "DEL",
					Destination: "HKG",
					Airline:     "6E",
				}},
			},
		},
	}

	cached := Wrap(fake, dir, 0)

	req := types.SearchRequest{
		Origin:        "DEL",
		Destination:   "HKG",
		DepartureDate: time.Date(2026, 3, 24, 0, 0, 0, 0, time.UTC),
		Passengers:    1,
		CabinClass:    types.CabinEconomy,
	}

	// First call: cache miss, should call underlying provider.
	flights, err := cached.Search(context.Background(), req)
	if err != nil {
		t.Fatalf("first search: %v", err)
	}
	if fake.calls != 1 {
		t.Fatalf("expected 1 call, got %d", fake.calls)
	}
	if len(flights) != 1 || flights[0].Price.Amount != 264 {
		t.Fatalf("unexpected flights: %+v", flights)
	}

	// Second call: cache hit, should NOT call underlying provider.
	flights, err = cached.Search(context.Background(), req)
	if err != nil {
		t.Fatalf("second search: %v", err)
	}
	if fake.calls != 1 {
		t.Fatalf("expected still 1 call after cache hit, got %d", fake.calls)
	}
	if len(flights) != 1 || flights[0].Price.Amount != 264 {
		t.Fatalf("unexpected cached flights: %+v", flights)
	}

	// Verify cache file exists.
	entries, _ := os.ReadDir(dir)
	if len(entries) != 1 {
		t.Fatalf("expected 1 cache file, got %d", len(entries))
	}
	t.Logf("Cache file: %s", entries[0].Name())
}

func TestCacheTTLExpiry(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "cache")

	fake := &fakeProvider{
		flights: []types.Flight{{Provider: "fake", Price: types.Money{Amount: 100, Currency: "USD"}}},
	}

	// 1ms TTL — will expire immediately.
	cached := Wrap(fake, dir, 1*time.Millisecond)

	req := types.SearchRequest{
		Origin:        "DEL",
		Destination:   "HKG",
		DepartureDate: time.Date(2026, 3, 24, 0, 0, 0, 0, time.UTC),
		Passengers:    1,
		CabinClass:    types.CabinEconomy,
	}

	_, _ = cached.Search(context.Background(), req)
	time.Sleep(5 * time.Millisecond)
	_, _ = cached.Search(context.Background(), req)

	if fake.calls != 2 {
		t.Fatalf("expected 2 calls after TTL expiry, got %d", fake.calls)
	}
}

// fakeMultiCityProvider implements both provider.Provider and provider.MultiCitySearcher.
type fakeMultiCityProvider struct {
	calls   int
	results []provider.MultiCityResult
}

func (f *fakeMultiCityProvider) Name() config.ProviderName { return "fakemc" }
func (f *fakeMultiCityProvider) Search(_ context.Context, _ types.SearchRequest) ([]types.Flight, error) {
	return nil, nil
}
func (f *fakeMultiCityProvider) SearchMultiCity(_ context.Context, _ provider.MultiCityRequest) ([]provider.MultiCityResult, error) {
	f.calls++
	return f.results, nil
}

func TestMultiCityCacheMiss(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "cache")
	fake := &fakeMultiCityProvider{
		results: []provider.MultiCityResult{
			{
				Leg1:  types.Flight{Price: types.Money{Amount: 200, Currency: "USD"}},
				Leg2:  types.Flight{Price: types.Money{Amount: 150, Currency: "USD"}},
				Price: types.Money{Amount: 350, Currency: "USD"},
			},
		},
	}

	cached := Wrap(fake, dir, 0)
	req := provider.MultiCityRequest{
		Origin: "JFK", Stopover: "IST", Destination: "DEL",
		Leg1Date: "2026-04-01", Leg2Date: "2026-04-05",
		Passengers: 1, CabinClass: types.CabinEconomy,
	}

	results, err := cached.SearchMultiCity(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fake.calls != 1 {
		t.Fatalf("expected 1 call on miss, got %d", fake.calls)
	}
	if len(results) != 1 || results[0].Price.Amount != 350 {
		t.Fatalf("unexpected results: %+v", results)
	}
}

func TestMultiCityCacheHit(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "cache")
	fake := &fakeMultiCityProvider{
		results: []provider.MultiCityResult{
			{Price: types.Money{Amount: 400, Currency: "USD"}},
		},
	}

	cached := Wrap(fake, dir, 0)
	req := provider.MultiCityRequest{
		Origin: "JFK", Stopover: "IST", Destination: "DEL",
		Leg1Date: "2026-04-01", Leg2Date: "2026-04-05",
		Passengers: 1, CabinClass: types.CabinEconomy,
	}

	// First call: miss.
	_, _ = cached.SearchMultiCity(context.Background(), req)
	// Second call: hit.
	results, err := cached.SearchMultiCity(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fake.calls != 1 {
		t.Fatalf("expected 1 call after cache hit, got %d", fake.calls)
	}
	if len(results) != 1 || results[0].Price.Amount != 400 {
		t.Fatalf("unexpected cached results: %+v", results)
	}
}

func TestMultiCityCachePath_Deterministic(t *testing.T) {
	fake := &fakeMultiCityProvider{}
	cached := Wrap(fake, "/tmp/test", 0)
	req := provider.MultiCityRequest{
		Origin: "JFK", Stopover: "IST", Destination: "DEL",
		Leg1Date: "2026-04-01", Leg2Date: "2026-04-05",
		Passengers: 1, CabinClass: types.CabinEconomy,
	}

	path1 := cached.multiCityCachePath(req)
	path2 := cached.multiCityCachePath(req)
	if path1 != path2 {
		t.Errorf("cache paths differ:\n  %s\n  %s", path1, path2)
	}
}

func TestMultiCityNotSupported(t *testing.T) {
	// fakeProvider does NOT implement MultiCitySearcher.
	fake := &fakeProvider{}
	cached := Wrap(fake, t.TempDir(), 0)
	req := provider.MultiCityRequest{
		Origin: "JFK", Stopover: "IST", Destination: "DEL",
	}

	_, err := cached.SearchMultiCity(context.Background(), req)
	if err == nil {
		t.Fatal("expected error for non-multi-city provider, got nil")
	}
}
