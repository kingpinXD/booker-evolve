package cache

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"booker/config"
	"booker/provider"
	"booker/types"
)

// fakeProvider returns a fixed set of flights or an error if set.
type fakeProvider struct {
	calls   int
	flights []types.Flight
	err     error
}

func (f *fakeProvider) Name() config.ProviderName { return "fake" }

func (f *fakeProvider) Search(_ context.Context, _ types.SearchRequest) ([]types.Flight, error) {
	f.calls++
	return f.flights, f.err
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
	err     error
}

func (f *fakeMultiCityProvider) Name() config.ProviderName { return "fakemc" }
func (f *fakeMultiCityProvider) Search(_ context.Context, _ types.SearchRequest) ([]types.Flight, error) {
	return nil, nil
}
func (f *fakeMultiCityProvider) SearchMultiCity(_ context.Context, _ provider.MultiCityRequest) ([]provider.MultiCityResult, error) {
	f.calls++
	return f.results, f.err
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

func TestName(t *testing.T) {
	fake := &fakeProvider{}
	cached := Wrap(fake, t.TempDir(), 0)
	if cached.Name() != "fake" {
		t.Fatalf("expected 'fake', got %q", cached.Name())
	}
}

func TestSearchInnerError(t *testing.T) {
	fake := &fakeProvider{err: errors.New("api down")}
	cached := Wrap(fake, t.TempDir(), 0)

	req := types.SearchRequest{
		Origin:        "DEL",
		Destination:   "HKG",
		DepartureDate: time.Date(2026, 3, 24, 0, 0, 0, 0, time.UTC),
		Passengers:    1,
		CabinClass:    types.CabinEconomy,
	}

	_, err := cached.Search(context.Background(), req)
	if err == nil || err.Error() != "api down" {
		t.Fatalf("expected 'api down' error, got %v", err)
	}
}

func TestCorruptedCacheFile(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "cache")

	fake := &fakeProvider{
		flights: []types.Flight{{Provider: "fake", Price: types.Money{Amount: 99, Currency: "USD"}}},
	}
	cached := Wrap(fake, dir, 0)

	req := types.SearchRequest{
		Origin:        "SFO",
		Destination:   "LAX",
		DepartureDate: time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC),
		Passengers:    1,
		CabinClass:    types.CabinEconomy,
	}

	// Populate the cache.
	_, _ = cached.Search(context.Background(), req)

	// Corrupt the cache file with invalid JSON.
	path := cached.cachePath(req)
	os.WriteFile(path, []byte("{invalid json!!!"), 0o644)

	// Should treat corrupted file as a miss, re-call the inner provider.
	flights, err := cached.Search(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fake.calls != 2 {
		t.Fatalf("expected 2 calls (miss after corruption), got %d", fake.calls)
	}
	if len(flights) != 1 || flights[0].Price.Amount != 99 {
		t.Fatalf("unexpected flights: %+v", flights)
	}
}

func TestCorruptedMultiCityCacheFile(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "cache")

	fake := &fakeMultiCityProvider{
		results: []provider.MultiCityResult{
			{Price: types.Money{Amount: 500, Currency: "USD"}},
		},
	}
	cached := Wrap(fake, dir, 0)

	req := provider.MultiCityRequest{
		Origin: "SFO", Stopover: "ORD", Destination: "JFK",
		Leg1Date: "2026-05-01", Leg2Date: "2026-05-05",
		Passengers: 1, CabinClass: types.CabinEconomy,
	}

	// Populate the cache.
	_, _ = cached.SearchMultiCity(context.Background(), req)

	// Corrupt the cache file.
	path := cached.multiCityCachePath(req)
	os.WriteFile(path, []byte("not json"), 0o644)

	// Should treat corrupted file as a miss.
	results, err := cached.SearchMultiCity(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fake.calls != 2 {
		t.Fatalf("expected 2 calls (miss after corruption), got %d", fake.calls)
	}
	if len(results) != 1 || results[0].Price.Amount != 500 {
		t.Fatalf("unexpected results: %+v", results)
	}
}

func TestMultiCityTTLExpiry(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "cache")

	fake := &fakeMultiCityProvider{
		results: []provider.MultiCityResult{
			{Price: types.Money{Amount: 300, Currency: "USD"}},
		},
	}

	cached := Wrap(fake, dir, 1*time.Millisecond)

	req := provider.MultiCityRequest{
		Origin: "LAX", Stopover: "NRT", Destination: "BKK",
		Leg1Date: "2026-06-01", Leg2Date: "2026-06-10",
		Passengers: 1, CabinClass: types.CabinEconomy,
	}

	_, _ = cached.SearchMultiCity(context.Background(), req)
	time.Sleep(5 * time.Millisecond)
	_, _ = cached.SearchMultiCity(context.Background(), req)

	if fake.calls != 2 {
		t.Fatalf("expected 2 calls after TTL expiry, got %d", fake.calls)
	}
}

func TestMultiCitySearchInnerError(t *testing.T) {
	fake := &fakeMultiCityProvider{err: errors.New("timeout")}
	cached := Wrap(fake, t.TempDir(), 0)

	req := provider.MultiCityRequest{
		Origin: "JFK", Stopover: "IST", Destination: "DEL",
		Leg1Date: "2026-04-01", Leg2Date: "2026-04-05",
		Passengers: 1, CabinClass: types.CabinEconomy,
	}

	_, err := cached.SearchMultiCity(context.Background(), req)
	if err == nil || err.Error() != "timeout" {
		t.Fatalf("expected 'timeout' error, got %v", err)
	}
}

func TestStoreToUnwritableDir(t *testing.T) {
	// Use /dev/null/impossible as cache dir to trigger MkdirAll failure.
	fake := &fakeProvider{
		flights: []types.Flight{{Provider: "fake", Price: types.Money{Amount: 50, Currency: "USD"}}},
	}
	cached := Wrap(fake, "/dev/null/impossible", 0)

	req := types.SearchRequest{
		Origin:        "DEL",
		Destination:   "BOM",
		DepartureDate: time.Date(2026, 3, 24, 0, 0, 0, 0, time.UTC),
		Passengers:    1,
		CabinClass:    types.CabinEconomy,
	}

	// Should still return results even though store fails.
	flights, err := cached.Search(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(flights) != 1 || flights[0].Price.Amount != 50 {
		t.Fatalf("unexpected flights: %+v", flights)
	}
}

func TestStoreMultiCityToUnwritableDir(t *testing.T) {
	fake := &fakeMultiCityProvider{
		results: []provider.MultiCityResult{
			{Price: types.Money{Amount: 600, Currency: "USD"}},
		},
	}
	cached := Wrap(fake, "/dev/null/impossible", 0)

	req := provider.MultiCityRequest{
		Origin: "JFK", Stopover: "LHR", Destination: "CDG",
		Leg1Date: "2026-07-01", Leg2Date: "2026-07-05",
		Passengers: 1, CabinClass: types.CabinEconomy,
	}

	// Should still return results even though store fails.
	results, err := cached.SearchMultiCity(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 || results[0].Price.Amount != 600 {
		t.Fatalf("unexpected results: %+v", results)
	}
}
