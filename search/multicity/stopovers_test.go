package multicity

import (
	"regexp"
	"strings"
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

func TestStopoversForRoute_DELToSFO(t *testing.T) {
	got := StopoversForRoute("DEL", "SFO")
	if len(got) == 0 {
		t.Fatal("expected stopovers for DEL->SFO, got none")
	}

	// Should return route-specific list, not fallback.
	if &got[0] != &DELToSFOStopovers[0] {
		t.Error("DEL->SFO should return the route-specific slice, not a copy")
	}

	// Verify expected cities: primary Pacific corridor via East Asia.
	airports := make(map[string]bool)
	for _, s := range got {
		airports[s.Airport] = true
	}
	for _, want := range []string{"NRT", "ICN", "HKG", "BKK", "SIN"} {
		if !airports[want] {
			t.Errorf("DEL->SFO stopovers missing expected airport %s", want)
		}
	}

	if airports["DEL"] {
		t.Error("origin DEL should not be in stopover list")
	}
	if airports["SFO"] {
		t.Error("destination SFO should not be in stopover list")
	}
}

func TestStopoversForRoute_BOMToSFO(t *testing.T) {
	got := StopoversForRoute("BOM", "SFO")
	if len(got) == 0 {
		t.Fatal("expected stopovers for BOM->SFO, got none")
	}

	// Should return route-specific list, not fallback.
	if &got[0] != &BOMToSFOStopovers[0] {
		t.Error("BOM->SFO should return the route-specific slice, not a copy")
	}

	airports := make(map[string]bool)
	for _, s := range got {
		airports[s.Airport] = true
	}
	for _, want := range []string{"NRT", "HKG", "BKK", "SIN"} {
		if !airports[want] {
			t.Errorf("BOM->SFO stopovers missing expected airport %s", want)
		}
	}

	if airports["BOM"] {
		t.Error("origin BOM should not be in stopover list")
	}
	if airports["SFO"] {
		t.Error("destination SFO should not be in stopover list")
	}
}

func TestStopoversForRoute_DELToSYD(t *testing.T) {
	got := StopoversForRoute("DEL", "SYD")
	if len(got) == 0 {
		t.Fatal("expected stopovers for DEL->SYD, got none")
	}

	// Should return route-specific list, not fallback.
	if &got[0] != &DELToSYDStopovers[0] {
		t.Error("DEL->SYD should return the route-specific slice, not a copy")
	}

	// Verify expected cities: Southeast Asia corridor.
	airports := make(map[string]bool)
	for _, s := range got {
		airports[s.Airport] = true
	}
	for _, want := range []string{"SIN", "BKK", "KUL", "HKG", "NRT", "KIX"} {
		if !airports[want] {
			t.Errorf("DEL->SYD stopovers missing expected airport %s", want)
		}
	}

	if airports["DEL"] {
		t.Error("origin DEL should not be in stopover list")
	}
	if airports["SYD"] {
		t.Error("destination SYD should not be in stopover list")
	}
}

func TestStopoversForRoute_BOMToSYD(t *testing.T) {
	got := StopoversForRoute("BOM", "SYD")
	if len(got) == 0 {
		t.Fatal("expected stopovers for BOM->SYD, got none")
	}

	// Should return route-specific list, not fallback.
	if &got[0] != &BOMToSYDStopovers[0] {
		t.Error("BOM->SYD should return the route-specific slice, not a copy")
	}

	airports := make(map[string]bool)
	for _, s := range got {
		airports[s.Airport] = true
	}
	for _, want := range []string{"SIN", "BKK", "KUL", "HKG", "NRT"} {
		if !airports[want] {
			t.Errorf("BOM->SYD stopovers missing expected airport %s", want)
		}
	}

	if airports["BOM"] {
		t.Error("origin BOM should not be in stopover list")
	}
	if airports["SYD"] {
		t.Error("destination SYD should not be in stopover list")
	}
}

func TestStopoversForRoute_ReverseSYDToDEL(t *testing.T) {
	got := StopoversForRoute("SYD", "DEL")
	if len(got) == 0 {
		t.Fatal("expected stopovers for reverse route SYD->DEL, got none")
	}

	airports := make(map[string]bool)
	for _, s := range got {
		airports[s.Airport] = true
	}
	// KIX is in DEL->SYD route-specific list but NOT in GlobalFallbackHubs.
	// Its presence proves reverse lookup uses route-specific data.
	if !airports["KIX"] {
		t.Error("reverse SYD->DEL missing route-specific airport KIX (would not be in fallback)")
	}

	if airports["SYD"] {
		t.Error("origin SYD should not be in stopover list")
	}
	if airports["DEL"] {
		t.Error("destination DEL should not be in stopover list")
	}
}

func TestStopoversForRoute_ReverseSYDToBOM(t *testing.T) {
	got := StopoversForRoute("SYD", "BOM")
	if len(got) == 0 {
		t.Fatal("expected stopovers for reverse route SYD->BOM, got none")
	}

	airports := make(map[string]bool)
	for _, s := range got {
		airports[s.Airport] = true
	}
	// KUL is in BOM->SYD route-specific list but NOT in GlobalFallbackHubs.
	// Its presence proves reverse lookup uses route-specific data.
	if !airports["KUL"] {
		t.Error("reverse SYD->BOM missing route-specific airport KUL (would not be in fallback)")
	}

	if airports["SYD"] {
		t.Error("origin SYD should not be in stopover list")
	}
	if airports["BOM"] {
		t.Error("destination BOM should not be in stopover list")
	}
}

func TestStopoversForRoute_DELToFRA(t *testing.T) {
	got := StopoversForRoute("DEL", "FRA")
	if len(got) == 0 {
		t.Fatal("expected stopovers for DEL->FRA, got none")
	}

	// Should return route-specific list, not fallback.
	if &got[0] != &DELToFRAStopovers[0] {
		t.Error("DEL->FRA should return the route-specific slice, not a copy")
	}

	// Verify expected Gulf hub cities.
	airports := make(map[string]bool)
	for _, s := range got {
		airports[s.Airport] = true
	}
	for _, want := range []string{"DOH", "AUH", "DXB", "IST", "BAH", "KWI"} {
		if !airports[want] {
			t.Errorf("DEL->FRA stopovers missing expected airport %s", want)
		}
	}

	if airports["DEL"] {
		t.Error("origin DEL should not be in stopover list")
	}
	if airports["FRA"] {
		t.Error("destination FRA should not be in stopover list")
	}
}

func TestStopoversForRoute_ReverseFRAToDEL(t *testing.T) {
	got := StopoversForRoute("FRA", "DEL")
	if len(got) == 0 {
		t.Fatal("expected stopovers for reverse route FRA->DEL, got none")
	}

	airports := make(map[string]bool)
	for _, s := range got {
		airports[s.Airport] = true
	}
	// KWI is in DEL->FRA route-specific list but NOT in GlobalFallbackHubs.
	// Its presence proves reverse lookup uses route-specific data.
	if !airports["KWI"] {
		t.Error("reverse FRA->DEL missing route-specific airport KWI (would not be in fallback)")
	}

	if airports["FRA"] {
		t.Error("origin FRA should not be in stopover list")
	}
	if airports["DEL"] {
		t.Error("destination DEL should not be in stopover list")
	}
}

func TestStopoversForRoute_BOMToFRA(t *testing.T) {
	got := StopoversForRoute("BOM", "FRA")
	if len(got) == 0 {
		t.Fatal("expected stopovers for BOM->FRA, got none")
	}

	// Should return route-specific list, not fallback.
	if &got[0] != &BOMToFRAStopovers[0] {
		t.Error("BOM->FRA should return the route-specific slice, not a copy")
	}

	// Verify expected Gulf hub cities.
	airports := make(map[string]bool)
	for _, s := range got {
		airports[s.Airport] = true
	}
	for _, want := range []string{"DOH", "AUH", "DXB", "IST", "BAH"} {
		if !airports[want] {
			t.Errorf("BOM->FRA stopovers missing expected airport %s", want)
		}
	}

	if airports["BOM"] {
		t.Error("origin BOM should not be in stopover list")
	}
	if airports["FRA"] {
		t.Error("destination FRA should not be in stopover list")
	}
}

func TestStopoversForRoute_ReverseFRAToBOM(t *testing.T) {
	got := StopoversForRoute("FRA", "BOM")
	if len(got) == 0 {
		t.Fatal("expected stopovers for reverse route FRA->BOM, got none")
	}

	airports := make(map[string]bool)
	for _, s := range got {
		airports[s.Airport] = true
	}
	// BAH is in BOM->FRA route-specific list but NOT in GlobalFallbackHubs.
	// Its presence proves reverse lookup uses route-specific data.
	if !airports["BAH"] {
		t.Error("reverse FRA->BOM missing route-specific airport BAH (would not be in fallback)")
	}

	if airports["FRA"] {
		t.Error("origin FRA should not be in stopover list")
	}
	if airports["BOM"] {
		t.Error("destination BOM should not be in stopover list")
	}
}

// TestStopoverDataConsistency validates all stopover data for correctness.
func TestStopoverDataConsistency(t *testing.T) {
	iataRe := regexp.MustCompile(`^[A-Z]{3}$`)

	// Validate route-specific stopovers.
	for key, stopovers := range stopoversMap {
		parts := strings.SplitN(key, "\u2192", 2) // "→" unicode
		if len(parts) != 2 {
			t.Errorf("invalid route key format: %q", key)
			continue
		}
		origin, dest := parts[0], parts[1]

		for i, s := range stopovers {
			label := key + "[" + s.Airport + "]"

			if !iataRe.MatchString(s.Airport) {
				t.Errorf("%s: invalid IATA code %q", label, s.Airport)
			}
			if s.Airport == origin {
				t.Errorf("%s: stopover airport matches origin %s", label, origin)
			}
			if s.Airport == dest {
				t.Errorf("%s: stopover airport matches destination %s", label, dest)
			}
			if s.MinStay >= s.MaxStay {
				t.Errorf("%s: MinStay (%v) >= MaxStay (%v)", label, s.MinStay, s.MaxStay)
			}
			if s.City == "" {
				t.Errorf("%s: empty City at index %d", label, i)
			}
			if s.Notes == "" {
				t.Errorf("%s: empty Notes at index %d", label, i)
			}
			if s.Region == "" {
				t.Errorf("%s: empty Region at index %d", label, i)
			}
		}
	}

	// Validate global fallback hubs.
	for i, s := range GlobalFallbackHubs {
		label := "GlobalFallbackHubs[" + s.Airport + "]"

		if !iataRe.MatchString(s.Airport) {
			t.Errorf("%s: invalid IATA code %q", label, s.Airport)
		}
		if s.MinStay >= s.MaxStay {
			t.Errorf("%s: MinStay (%v) >= MaxStay (%v)", label, s.MinStay, s.MaxStay)
		}
		if s.City == "" {
			t.Errorf("%s: empty City at index %d", label, i)
		}
		if s.Notes == "" {
			t.Errorf("%s: empty Notes at index %d", label, i)
		}
	}
}
