package serpapi

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"booker/config"
	"booker/httpclient"
	"booker/provider/cache"
	"booker/types"
)

// TestSeedCache makes the minimum API calls needed for the multi-city test
// and caches them on disk. Run this ONCE, then all other tests use the cache.
//
// Calls: 4 stopover cities × 2 legs = 8 searches, 1 already cached = 7 new.
func TestSeedCache(t *testing.T) {
	if os.Getenv("SERPAPI_KEY") == "" {
		t.Skip("skipping: SERPAPI_KEY not set (integration test)")
	}
	cacheDir := "../../.cache/flights"

	searches := []struct {
		origin string
		dest   string
		date   string
	}{
		{"DEL", "HKG", "2026-03-24"},
		{"HKG", "YYZ", "2026-03-30"},
		{"DEL", "BKK", "2026-03-24"},
		{"BKK", "YYZ", "2026-03-30"},
		{"DEL", "IST", "2026-03-24"},
		{"IST", "YYZ", "2026-03-30"},
		{"DEL", "NRT", "2026-03-24"},
		{"NRT", "YYZ", "2026-03-30"},
	}

	cfg := config.Default()
	httpClient := httpclient.New(cfg.HTTP)
	serpCfg := cfg.Providers[config.ProviderSerpAPI]
	raw := New(serpCfg, httpClient)
	cached := cache.Wrap(raw, cacheDir, 0)

	ctx := context.Background()

	for i, s := range searches {
		date, _ := time.Parse("2006-01-02", s.date)
		req := types.SearchRequest{
			Origin:        s.origin,
			Destination:   s.dest,
			DepartureDate: date,
			Passengers:    1,
			CabinClass:    types.CabinEconomy,
		}

		flights, err := cached.Search(ctx, req)
		if err != nil {
			t.Fatalf("search %s→%s %s: %v", s.origin, s.dest, s.date, err)
		}

		t.Logf("[%d/%d] %s→%s %s: %d flights", i+1, len(searches),
			s.origin, s.dest, s.date, len(flights))

		for j, f := range flights {
			segs := ""
			for _, seg := range f.Outbound {
				if segs != "" {
					segs += " → "
				}
				segs += fmt.Sprintf("%s(%s→%s)", seg.Airline, seg.Origin, seg.Destination)
			}
			t.Logf("  #%d $%.0f %s", j+1, f.Price.Amount, segs)
		}

		// Rate limit: wait between API calls (cache hits are instant).
		if i < len(searches)-1 {
			time.Sleep(3 * time.Second)
		}
	}

	// Verify cache files.
	entries, _ := os.ReadDir(cacheDir)
	t.Logf("\nCache files in %s:", cacheDir)
	for _, e := range entries {
		data, _ := os.ReadFile(cacheDir + "/" + e.Name())
		var entry struct {
			Flights []json.RawMessage `json:"flights"`
		}
		_ = json.Unmarshal(data, &entry)
		t.Logf("  %s (%d flights)", e.Name(), len(entry.Flights))
	}
}
