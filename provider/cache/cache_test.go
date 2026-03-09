package cache

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"booker/config"
	"booker/types"
)

// fakeProvider returns a fixed set of flights.
type fakeProvider struct {
	calls  int
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

	cached.Search(context.Background(), req)
	time.Sleep(5 * time.Millisecond)
	cached.Search(context.Background(), req)

	if fake.calls != 2 {
		t.Fatalf("expected 2 calls after TTL expiry, got %d", fake.calls)
	}
}
