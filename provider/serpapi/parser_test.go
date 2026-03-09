package serpapi

import (
	"encoding/json"
	"os"
	"testing"
)

// TestParseResponse_CachedData validates the parser against real SerpAPI output
// saved during the initial API test.
func TestParseResponse_CachedData(t *testing.T) {
	data, err := os.ReadFile("/tmp/serpapi_del_hkg.json")
	if err != nil {
		t.Skipf("no cached response at /tmp/serpapi_del_hkg.json: %v", err)
	}

	var resp Response
	if err := json.Unmarshal(data, &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	flights, err := ParseResponse(&resp)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	t.Logf("Parsed %d flights", len(flights))

	for i, f := range flights {
		t.Logf("#%d $%.0f %s  %d segments",
			i+1, f.Price.Amount, f.Provider, len(f.Outbound))
		for _, seg := range f.Outbound {
			t.Logf("  %s %s(%s)→%s(%s) %s→%s [%s] %s",
				seg.FlightNumber,
				seg.OriginName, seg.Origin,
				seg.DestinationName, seg.Destination,
				seg.DepartureTime.Format("Jan 02 15:04"),
				seg.ArrivalTime.Format("Jan 02 15:04"),
				seg.Duration,
				seg.AirlineName)
			if seg.LayoverDuration > 0 {
				t.Logf("    layover: %s", seg.LayoverDuration)
			}
		}
	}

	if len(flights) == 0 {
		t.Fatal("expected flights, got 0")
	}

	// Verify first flight has expected structure.
	f := flights[0]
	if f.Price.Amount <= 0 {
		t.Errorf("expected positive price, got %f", f.Price.Amount)
	}
	if f.Price.Currency != "USD" {
		t.Errorf("expected USD currency, got %s", f.Price.Currency)
	}
	if len(f.Outbound) == 0 {
		t.Error("expected outbound segments")
	}
	if f.Outbound[0].Origin != "DEL" {
		t.Errorf("expected DEL origin, got %s", f.Outbound[0].Origin)
	}
}
