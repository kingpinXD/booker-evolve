package multicity

import (
	"testing"
)

func TestStopoversForRoute_KnownRoute(t *testing.T) {
	got := StopoversForRoute("DEL", "YYZ")
	if len(got) == 0 {
		t.Fatal("expected stopovers for DEL->YYZ, got none")
	}
	// Should return the route-specific list, not fallback.
	if &got[0] != &DELToYYZStopovers[0] {
		t.Error("DEL->YYZ should return the route-specific slice, not a copy")
	}
}

func TestStopoversForRoute_UnknownRoute_ReturnsFallback(t *testing.T) {
	got := StopoversForRoute("JFK", "NRT")
	if len(got) == 0 {
		t.Fatal("expected fallback hubs for unknown route JFK->NRT, got none")
	}
	// Should contain several global hubs.
	if len(got) < 4 {
		t.Errorf("expected at least 4 fallback hubs, got %d", len(got))
	}
}

func TestStopoversForRoute_FallbackExcludesOrigin(t *testing.T) {
	// IST is a fallback hub. When origin is IST, it must be excluded.
	got := StopoversForRoute("IST", "CDG")
	for _, s := range got {
		if s.Airport == "IST" {
			t.Error("fallback hubs should not include origin airport IST")
		}
	}
}

func TestStopoversForRoute_FallbackExcludesDestination(t *testing.T) {
	// LHR is a fallback hub. When destination is LHR, it must be excluded.
	got := StopoversForRoute("JFK", "LHR")
	for _, s := range got {
		if s.Airport == "LHR" {
			t.Error("fallback hubs should not include destination airport LHR")
		}
	}
}

func TestStopoversForRoute_FallbackExcludesBoth(t *testing.T) {
	// Both IST (origin) and LHR (destination) are fallback hubs.
	got := StopoversForRoute("IST", "LHR")
	for _, s := range got {
		if s.Airport == "IST" || s.Airport == "LHR" {
			t.Errorf("fallback hubs should not include origin or destination, found %s", s.Airport)
		}
	}
	// Should still have remaining hubs.
	if len(got) < 4 {
		t.Errorf("expected at least 4 remaining hubs after excluding IST and LHR, got %d", len(got))
	}
}

func TestStopoversForRoute_DELToJFK(t *testing.T) {
	got := StopoversForRoute("DEL", "JFK")
	if len(got) == 0 {
		t.Fatal("expected stopovers for DEL->JFK, got none")
	}

	// Should return route-specific list, not fallback.
	if &got[0] != &DELToJFKStopovers[0] {
		t.Error("DEL->JFK should return the route-specific slice, not a copy")
	}

	// Verify expected cities are present.
	airports := make(map[string]bool)
	for _, s := range got {
		airports[s.Airport] = true
	}
	for _, want := range []string{"HKG", "IST", "NRT", "ICN", "SIN", "BKK", "LHR", "FRA"} {
		if !airports[want] {
			t.Errorf("DEL->JFK stopovers missing expected airport %s", want)
		}
	}

	// Origin and destination must not appear in the stopover list.
	if airports["DEL"] {
		t.Error("origin DEL should not be in stopover list")
	}
	if airports["JFK"] {
		t.Error("destination JFK should not be in stopover list")
	}
}

func TestStopoversForRoute_BOMToJFK(t *testing.T) {
	got := StopoversForRoute("BOM", "JFK")
	if len(got) == 0 {
		t.Fatal("expected stopovers for BOM->JFK, got none")
	}

	// Should return route-specific list, not fallback.
	if &got[0] != &BOMToJFKStopovers[0] {
		t.Error("BOM->JFK should return the route-specific slice, not a copy")
	}

	// Verify expected cities are present.
	airports := make(map[string]bool)
	for _, s := range got {
		airports[s.Airport] = true
	}
	for _, want := range []string{"BKK", "IST", "LHR", "SIN", "HKG", "NRT", "FRA"} {
		if !airports[want] {
			t.Errorf("BOM->JFK stopovers missing expected airport %s", want)
		}
	}

	// Origin and destination must not appear in the stopover list.
	if airports["BOM"] {
		t.Error("origin BOM should not be in stopover list")
	}
	if airports["JFK"] {
		t.Error("destination JFK should not be in stopover list")
	}
}

func TestStopoversForRoute_DELToLHR(t *testing.T) {
	got := StopoversForRoute("DEL", "LHR")
	if len(got) == 0 {
		t.Fatal("expected stopovers for DEL->LHR, got none")
	}

	// Should return route-specific list, not fallback.
	if &got[0] != &DELToLHRStopovers[0] {
		t.Error("DEL->LHR should return the route-specific slice, not a copy")
	}

	// Verify expected cities are present.
	airports := make(map[string]bool)
	for _, s := range got {
		airports[s.Airport] = true
	}
	for _, want := range []string{"BKK", "SIN", "KUL", "HKG", "IST", "CMB"} {
		if !airports[want] {
			t.Errorf("DEL->LHR stopovers missing expected airport %s", want)
		}
	}

	// Origin and destination must not appear in the stopover list.
	if airports["DEL"] {
		t.Error("origin DEL should not be in stopover list")
	}
	if airports["LHR"] {
		t.Error("destination LHR should not be in stopover list")
	}
}

func TestStopoversForRoute_ReverseRoute(t *testing.T) {
	// YYZ->DEL should return route-specific stopovers (from DEL->YYZ),
	// not global fallback hubs.
	got := StopoversForRoute("YYZ", "DEL")
	if len(got) == 0 {
		t.Fatal("expected stopovers for reverse route YYZ->DEL, got none")
	}

	airports := make(map[string]bool)
	for _, s := range got {
		airports[s.Airport] = true
	}
	// KUL and FRA are in DEL->YYZ route-specific list but NOT in GlobalFallbackHubs.
	// Their presence proves reverse lookup uses route-specific data.
	for _, want := range []string{"KUL", "FRA"} {
		if !airports[want] {
			t.Errorf("reverse YYZ->DEL missing route-specific airport %s (would not be in fallback)", want)
		}
	}

	// Origin and destination must be excluded.
	if airports["YYZ"] {
		t.Error("origin YYZ should not be in stopover list")
	}
	if airports["DEL"] {
		t.Error("destination DEL should not be in stopover list")
	}
}

func TestStopoversForRoute_ReverseExcludesOriginDest(t *testing.T) {
	// LHR->DEL reverse: DEL->LHR has CMB which is not in fallback hubs.
	// CMB presence proves route-specific reverse lookup.
	got := StopoversForRoute("LHR", "DEL")
	if len(got) == 0 {
		t.Fatal("expected stopovers for reverse LHR->DEL, got none")
	}
	airports := make(map[string]bool)
	for _, s := range got {
		airports[s.Airport] = true
	}
	if !airports["CMB"] {
		t.Error("reverse LHR->DEL missing route-specific airport CMB (proves non-fallback)")
	}
	// Neither origin nor destination should appear.
	if airports["LHR"] {
		t.Error("origin LHR should not be in stopover list")
	}
	if airports["DEL"] {
		t.Error("destination DEL should not be in stopover list")
	}
}

func TestStopoversForRoute_BOMToLHR(t *testing.T) {
	got := StopoversForRoute("BOM", "LHR")
	if len(got) == 0 {
		t.Fatal("expected stopovers for BOM->LHR, got none")
	}

	// Should return route-specific list, not fallback.
	if &got[0] != &BOMToLHRStopovers[0] {
		t.Error("BOM->LHR should return the route-specific slice, not a copy")
	}

	// Verify expected cities are present.
	airports := make(map[string]bool)
	for _, s := range got {
		airports[s.Airport] = true
	}
	for _, want := range []string{"BKK", "SIN", "KUL", "HKG", "IST", "CMB"} {
		if !airports[want] {
			t.Errorf("BOM->LHR stopovers missing expected airport %s", want)
		}
	}

	// Origin and destination must not appear in the stopover list.
	if airports["BOM"] {
		t.Error("origin BOM should not be in stopover list")
	}
	if airports["LHR"] {
		t.Error("destination LHR should not be in stopover list")
	}
}
