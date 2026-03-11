//go:build integration

package multicity

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"booker/config"
	"booker/httpclient"
	"booker/provider/kiwi"
	"booker/search"
	"booker/types"
)

// TestDiagnostic_RawData fetches a single route and inspects what Kiwi returns:
// what dates, which airlines, which hubs. This helps diagnose why the full
// pipeline returns 0 results.
func TestDiagnostic_RawData(t *testing.T) {
	if os.Getenv("KIWI_API_KEY") == "" {
		t.Skip("skipping: KIWI_API_KEY not set (integration test)")
	}
	cfg := config.Default()
	httpClient := httpclient.New(cfg.HTTP)
	kiwiCfg := cfg.Providers[config.ProviderKiwi]
	p := kiwi.New(kiwiCfg, httpClient)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	routes := []struct {
		name   string
		origin string
		dest   string
	}{
		{"DEL→HKG", "DEL", "HKG"},
		{"DEL→FRA", "DEL", "FRA"},
	}

	for _, r := range routes {
		t.Run(r.name, func(t *testing.T) {
			// Small delay to avoid rate limiting between routes
			time.Sleep(2 * time.Second)

			req := types.SearchRequest{
				Origin:      r.origin,
				Destination: r.dest,
				Passengers:  1,
				CabinClass:  types.CabinEconomy,
			}

			flights, err := p.Search(ctx, req)
			if err != nil {
				t.Fatalf("search error: %v", err)
			}

			t.Logf("Raw results: %d flights", len(flights))

			// Analyze dates
			t.Log("\n=== DATE DISTRIBUTION ===")
			dateCounts := map[string]int{}
			for _, f := range flights {
				if len(f.Outbound) > 0 {
					date := f.Outbound[0].DepartureTime.Format("2006-01-02")
					dateCounts[date]++
				}
			}
			for date, count := range dateCounts {
				t.Logf("  %s: %d flights", date, count)
			}

			// Analyze which flights are blocked and why
			t.Log("\n=== BLOCKED ANALYSIS ===")
			blocked := 0
			blockedReasons := map[string]int{}
			for _, f := range flights {
				for _, seg := range f.Outbound {
					if search.IsAirlineBlocked(seg.Airline) {
						reason := fmt.Sprintf("airline:%s(%s)", seg.Airline, seg.AirlineName)
						blockedReasons[reason]++
						blocked++
						break
					}
					if search.IsHubBlocked(seg.Origin) || search.IsHubBlocked(seg.Destination) {
						hub := seg.Origin
						if search.IsHubBlocked(seg.Destination) {
							hub = seg.Destination
						}
						reason := fmt.Sprintf("hub:%s", hub)
						blockedReasons[reason]++
						blocked++
						break
					}
				}
			}
			t.Logf("  Blocked: %d / %d", blocked, len(flights))
			for reason, count := range blockedReasons {
				t.Logf("    %s: %d", reason, count)
			}

			// Show unblocked flights
			t.Log("\n=== UNBLOCKED FLIGHTS ===")
			unblocked := search.FilterFlights(flights)
			for i, f := range unblocked {
				if len(f.Outbound) == 0 {
					continue
				}
				dep := f.Outbound[0].DepartureTime.Format("Jan 02")
				arr := f.Outbound[len(f.Outbound)-1].ArrivalTime.Format("Jan 02")
				airlines := ""
				for _, seg := range f.Outbound {
					if airlines != "" {
						airlines += " → "
					}
					airlines += fmt.Sprintf("%s(%s→%s)", seg.Airline, seg.Origin, seg.Destination)
				}
				t.Logf("  #%d $%.0f  %s–%s  %s", i+1, f.Price.Amount, dep, arr, airlines)
			}
		})
	}
}
