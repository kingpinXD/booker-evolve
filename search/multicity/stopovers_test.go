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

func TestStopoversForRoute_DELToBKK(t *testing.T) {
	got := StopoversForRoute("DEL", "BKK")
	if len(got) == 0 {
		t.Fatal("expected stopovers for DEL->BKK, got none")
	}

	// Should return route-specific list, not fallback.
	if &got[0] != &DELToBKKStopovers[0] {
		t.Error("DEL->BKK should return the route-specific slice, not a copy")
	}

	// Verify expected cities: Gulf + South Asian hubs.
	airports := make(map[string]bool)
	for _, s := range got {
		airports[s.Airport] = true
	}
	for _, want := range []string{"DOH", "AUH", "DXB", "SIN", "KUL", "CCU"} {
		if !airports[want] {
			t.Errorf("DEL->BKK stopovers missing expected airport %s", want)
		}
	}

	if airports["DEL"] {
		t.Error("origin DEL should not be in stopover list")
	}
	if airports["BKK"] {
		t.Error("destination BKK should not be in stopover list")
	}
}

func TestStopoversForRoute_ReverseBKKToDEL(t *testing.T) {
	got := StopoversForRoute("BKK", "DEL")
	if len(got) == 0 {
		t.Fatal("expected stopovers for reverse route BKK->DEL, got none")
	}

	airports := make(map[string]bool)
	for _, s := range got {
		airports[s.Airport] = true
	}
	// CCU is in DEL->BKK route-specific list but NOT in GlobalFallbackHubs.
	// Its presence proves reverse lookup uses route-specific data.
	if !airports["CCU"] {
		t.Error("reverse BKK->DEL missing route-specific airport CCU (would not be in fallback)")
	}

	if airports["BKK"] {
		t.Error("origin BKK should not be in stopover list")
	}
	if airports["DEL"] {
		t.Error("destination DEL should not be in stopover list")
	}
}

func TestStopoversForRoute_BOMToBKK(t *testing.T) {
	got := StopoversForRoute("BOM", "BKK")
	if len(got) == 0 {
		t.Fatal("expected stopovers for BOM->BKK, got none")
	}

	// Should return route-specific list, not fallback.
	if &got[0] != &BOMToBKKStopovers[0] {
		t.Error("BOM->BKK should return the route-specific slice, not a copy")
	}

	// Verify expected cities: Gulf + Southeast Asian hubs.
	airports := make(map[string]bool)
	for _, s := range got {
		airports[s.Airport] = true
	}
	for _, want := range []string{"DOH", "AUH", "DXB", "SIN", "KUL"} {
		if !airports[want] {
			t.Errorf("BOM->BKK stopovers missing expected airport %s", want)
		}
	}

	if airports["BOM"] {
		t.Error("origin BOM should not be in stopover list")
	}
	if airports["BKK"] {
		t.Error("destination BKK should not be in stopover list")
	}
}

func TestStopoversForRoute_ReverseBKKToBOM(t *testing.T) {
	got := StopoversForRoute("BKK", "BOM")
	if len(got) == 0 {
		t.Fatal("expected stopovers for reverse route BKK->BOM, got none")
	}

	airports := make(map[string]bool)
	for _, s := range got {
		airports[s.Airport] = true
	}
	// KUL is in BOM->BKK route-specific list but NOT in GlobalFallbackHubs.
	// Its presence proves reverse lookup uses route-specific data.
	if !airports["KUL"] {
		t.Error("reverse BKK->BOM missing route-specific airport KUL (would not be in fallback)")
	}

	if airports["BKK"] {
		t.Error("origin BKK should not be in stopover list")
	}
	if airports["BOM"] {
		t.Error("destination BOM should not be in stopover list")
	}
}

func TestStopoversForRoute_DELToNRT(t *testing.T) {
	got := StopoversForRoute("DEL", "NRT")
	if len(got) == 0 {
		t.Fatal("expected stopovers for DEL->NRT, got none")
	}

	// Should return route-specific list, not fallback.
	if &got[0] != &DELToNRTStopovers[0] {
		t.Error("DEL->NRT should return the route-specific slice, not a copy")
	}

	// Verify expected cities: Southeast/East Asia corridor.
	airports := make(map[string]bool)
	for _, s := range got {
		airports[s.Airport] = true
	}
	for _, want := range []string{"BKK", "SIN", "HKG", "TPE", "ICN", "KUL"} {
		if !airports[want] {
			t.Errorf("DEL->NRT stopovers missing expected airport %s", want)
		}
	}

	if airports["DEL"] {
		t.Error("origin DEL should not be in stopover list")
	}
	if airports["NRT"] {
		t.Error("destination NRT should not be in stopover list")
	}
}

func TestStopoversForRoute_ReverseNRTToDEL(t *testing.T) {
	got := StopoversForRoute("NRT", "DEL")
	if len(got) == 0 {
		t.Fatal("expected stopovers for reverse route NRT->DEL, got none")
	}

	airports := make(map[string]bool)
	for _, s := range got {
		airports[s.Airport] = true
	}
	// TPE is in DEL->NRT route-specific list but NOT in GlobalFallbackHubs.
	// Its presence proves reverse lookup uses route-specific data.
	if !airports["TPE"] {
		t.Error("reverse NRT->DEL missing route-specific airport TPE (would not be in fallback)")
	}

	if airports["NRT"] {
		t.Error("origin NRT should not be in stopover list")
	}
	if airports["DEL"] {
		t.Error("destination DEL should not be in stopover list")
	}
}

func TestStopoversForRoute_BOMToNRT(t *testing.T) {
	got := StopoversForRoute("BOM", "NRT")
	if len(got) == 0 {
		t.Fatal("expected stopovers for BOM->NRT, got none")
	}

	// Should return route-specific list, not fallback.
	if &got[0] != &BOMToNRTStopovers[0] {
		t.Error("BOM->NRT should return the route-specific slice, not a copy")
	}

	// Verify expected cities: Southeast/East Asia corridor.
	airports := make(map[string]bool)
	for _, s := range got {
		airports[s.Airport] = true
	}
	for _, want := range []string{"BKK", "SIN", "HKG", "TPE", "ICN"} {
		if !airports[want] {
			t.Errorf("BOM->NRT stopovers missing expected airport %s", want)
		}
	}

	if airports["BOM"] {
		t.Error("origin BOM should not be in stopover list")
	}
	if airports["NRT"] {
		t.Error("destination NRT should not be in stopover list")
	}
}

func TestStopoversForRoute_ReverseNRTToBOM(t *testing.T) {
	got := StopoversForRoute("NRT", "BOM")
	if len(got) == 0 {
		t.Fatal("expected stopovers for reverse route NRT->BOM, got none")
	}

	airports := make(map[string]bool)
	for _, s := range got {
		airports[s.Airport] = true
	}
	// TPE is in BOM->NRT route-specific list but NOT in GlobalFallbackHubs.
	// Its presence proves reverse lookup uses route-specific data.
	if !airports["TPE"] {
		t.Error("reverse NRT->BOM missing route-specific airport TPE (would not be in fallback)")
	}

	if airports["NRT"] {
		t.Error("origin NRT should not be in stopover list")
	}
	if airports["BOM"] {
		t.Error("destination BOM should not be in stopover list")
	}
}

func TestStopoversForRoute_DELToMEL(t *testing.T) {
	got := StopoversForRoute("DEL", "MEL")
	if len(got) == 0 {
		t.Fatal("expected stopovers for DEL->MEL, got none")
	}

	// Should return route-specific list, not fallback.
	if &got[0] != &DELToMELStopovers[0] {
		t.Error("DEL->MEL should return the route-specific slice, not a copy")
	}

	// Verify expected cities: Southeast Asia corridor.
	airports := make(map[string]bool)
	for _, s := range got {
		airports[s.Airport] = true
	}
	for _, want := range []string{"SIN", "BKK", "KUL", "HKG", "NRT"} {
		if !airports[want] {
			t.Errorf("DEL->MEL stopovers missing expected airport %s", want)
		}
	}

	if airports["DEL"] {
		t.Error("origin DEL should not be in stopover list")
	}
	if airports["MEL"] {
		t.Error("destination MEL should not be in stopover list")
	}
}

func TestStopoversForRoute_BOMToMEL(t *testing.T) {
	got := StopoversForRoute("BOM", "MEL")
	if len(got) == 0 {
		t.Fatal("expected stopovers for BOM->MEL, got none")
	}

	// Should return route-specific list, not fallback.
	if &got[0] != &BOMToMELStopovers[0] {
		t.Error("BOM->MEL should return the route-specific slice, not a copy")
	}

	airports := make(map[string]bool)
	for _, s := range got {
		airports[s.Airport] = true
	}
	for _, want := range []string{"SIN", "BKK", "KUL", "HKG", "NRT"} {
		if !airports[want] {
			t.Errorf("BOM->MEL stopovers missing expected airport %s", want)
		}
	}

	if airports["BOM"] {
		t.Error("origin BOM should not be in stopover list")
	}
	if airports["MEL"] {
		t.Error("destination MEL should not be in stopover list")
	}
}

func TestStopoversForRoute_ReverseMELToDEL(t *testing.T) {
	got := StopoversForRoute("MEL", "DEL")
	if len(got) == 0 {
		t.Fatal("expected stopovers for reverse route MEL->DEL, got none")
	}

	airports := make(map[string]bool)
	for _, s := range got {
		airports[s.Airport] = true
	}
	// KUL is in DEL->MEL route-specific list but NOT in GlobalFallbackHubs.
	// Its presence proves reverse lookup uses route-specific data.
	if !airports["KUL"] {
		t.Error("reverse MEL->DEL missing route-specific airport KUL (would not be in fallback)")
	}

	if airports["MEL"] {
		t.Error("origin MEL should not be in stopover list")
	}
	if airports["DEL"] {
		t.Error("destination DEL should not be in stopover list")
	}
}

func TestStopoversForRoute_DELToCDG(t *testing.T) {
	got := StopoversForRoute("DEL", "CDG")
	if len(got) == 0 {
		t.Fatal("expected stopovers for DEL->CDG, got none")
	}

	// Should return route-specific list, not fallback.
	if &got[0] != &DELToCDGStopovers[0] {
		t.Error("DEL->CDG should return the route-specific slice, not a copy")
	}

	// Verify expected Gulf hub cities.
	airports := make(map[string]bool)
	for _, s := range got {
		airports[s.Airport] = true
	}
	for _, want := range []string{"DOH", "AUH", "DXB", "IST", "BAH"} {
		if !airports[want] {
			t.Errorf("DEL->CDG stopovers missing expected airport %s", want)
		}
	}

	if airports["DEL"] {
		t.Error("origin DEL should not be in stopover list")
	}
	if airports["CDG"] {
		t.Error("destination CDG should not be in stopover list")
	}
}

func TestStopoversForRoute_BOMToCDG(t *testing.T) {
	got := StopoversForRoute("BOM", "CDG")
	if len(got) == 0 {
		t.Fatal("expected stopovers for BOM->CDG, got none")
	}

	// Should return route-specific list, not fallback.
	if &got[0] != &BOMToCDGStopovers[0] {
		t.Error("BOM->CDG should return the route-specific slice, not a copy")
	}

	// Verify expected Gulf hub cities.
	airports := make(map[string]bool)
	for _, s := range got {
		airports[s.Airport] = true
	}
	for _, want := range []string{"DOH", "AUH", "DXB", "IST", "BAH"} {
		if !airports[want] {
			t.Errorf("BOM->CDG stopovers missing expected airport %s", want)
		}
	}

	if airports["BOM"] {
		t.Error("origin BOM should not be in stopover list")
	}
	if airports["CDG"] {
		t.Error("destination CDG should not be in stopover list")
	}
}

func TestStopoversForRoute_ReverseCDGToDEL(t *testing.T) {
	got := StopoversForRoute("CDG", "DEL")
	if len(got) == 0 {
		t.Fatal("expected stopovers for reverse route CDG->DEL, got none")
	}

	airports := make(map[string]bool)
	for _, s := range got {
		airports[s.Airport] = true
	}
	// BAH is in DEL->CDG route-specific list but NOT in GlobalFallbackHubs.
	// Its presence proves reverse lookup uses route-specific data.
	if !airports["BAH"] {
		t.Error("reverse CDG->DEL missing route-specific airport BAH (would not be in fallback)")
	}

	if airports["CDG"] {
		t.Error("origin CDG should not be in stopover list")
	}
	if airports["DEL"] {
		t.Error("destination DEL should not be in stopover list")
	}
}

func TestStopoversForRoute_DELToICN(t *testing.T) {
	got := StopoversForRoute("DEL", "ICN")
	if len(got) == 0 {
		t.Fatal("expected stopovers for DEL->ICN, got none")
	}

	// Should return route-specific list, not fallback.
	if &got[0] != &DELToICNStopovers[0] {
		t.Error("DEL->ICN should return the route-specific slice, not a copy")
	}

	// Verify expected cities: Southeast/East Asia corridor.
	airports := make(map[string]bool)
	for _, s := range got {
		airports[s.Airport] = true
	}
	for _, want := range []string{"BKK", "SIN", "HKG", "TPE", "KUL"} {
		if !airports[want] {
			t.Errorf("DEL->ICN stopovers missing expected airport %s", want)
		}
	}

	if airports["DEL"] {
		t.Error("origin DEL should not be in stopover list")
	}
	if airports["ICN"] {
		t.Error("destination ICN should not be in stopover list")
	}
}

func TestStopoversForRoute_BOMToICN(t *testing.T) {
	got := StopoversForRoute("BOM", "ICN")
	if len(got) == 0 {
		t.Fatal("expected stopovers for BOM->ICN, got none")
	}

	// Should return route-specific list, not fallback.
	if &got[0] != &BOMToICNStopovers[0] {
		t.Error("BOM->ICN should return the route-specific slice, not a copy")
	}

	// Verify expected cities: East Asian corridor.
	airports := make(map[string]bool)
	for _, s := range got {
		airports[s.Airport] = true
	}
	for _, want := range []string{"BKK", "SIN", "HKG", "TPE"} {
		if !airports[want] {
			t.Errorf("BOM->ICN stopovers missing expected airport %s", want)
		}
	}

	if airports["BOM"] {
		t.Error("origin BOM should not be in stopover list")
	}
	if airports["ICN"] {
		t.Error("destination ICN should not be in stopover list")
	}
}

func TestStopoversForRoute_DELToHKG(t *testing.T) {
	got := StopoversForRoute("DEL", "HKG")
	if len(got) == 0 {
		t.Fatal("expected stopovers for DEL->HKG, got none")
	}

	// Should return route-specific list, not fallback.
	if &got[0] != &DELToHKGStopovers[0] {
		t.Error("DEL->HKG should return the route-specific slice, not a copy")
	}

	// Verify expected cities: Southeast Asian corridor.
	airports := make(map[string]bool)
	for _, s := range got {
		airports[s.Airport] = true
	}
	for _, want := range []string{"BKK", "SIN", "KUL", "CCU", "TPE"} {
		if !airports[want] {
			t.Errorf("DEL->HKG stopovers missing expected airport %s", want)
		}
	}

	if airports["DEL"] {
		t.Error("origin DEL should not be in stopover list")
	}
	if airports["HKG"] {
		t.Error("destination HKG should not be in stopover list")
	}
}

func TestStopoversForRoute_BOMToHKG(t *testing.T) {
	got := StopoversForRoute("BOM", "HKG")
	if len(got) == 0 {
		t.Fatal("expected stopovers for BOM->HKG, got none")
	}

	// Should return route-specific list, not fallback.
	if &got[0] != &BOMToHKGStopovers[0] {
		t.Error("BOM->HKG should return the route-specific slice, not a copy")
	}

	// Verify expected cities: Southeast Asian corridor.
	airports := make(map[string]bool)
	for _, s := range got {
		airports[s.Airport] = true
	}
	for _, want := range []string{"BKK", "SIN", "KUL", "TPE"} {
		if !airports[want] {
			t.Errorf("BOM->HKG stopovers missing expected airport %s", want)
		}
	}

	if airports["BOM"] {
		t.Error("origin BOM should not be in stopover list")
	}
	if airports["HKG"] {
		t.Error("destination HKG should not be in stopover list")
	}
}

func TestStopoversForRoute_DELToLAX(t *testing.T) {
	got := StopoversForRoute("DEL", "LAX")
	if len(got) == 0 {
		t.Fatal("expected stopovers for DEL->LAX, got none")
	}

	// Should return route-specific list, not fallback.
	if &got[0] != &DELToLAXStopovers[0] {
		t.Error("DEL->LAX should return the route-specific slice, not a copy")
	}

	// Verify expected cities: East Asia Pacific corridor.
	airports := make(map[string]bool)
	for _, s := range got {
		airports[s.Airport] = true
	}
	for _, want := range []string{"BKK", "SIN", "KUL", "HKG", "NRT", "TPE"} {
		if !airports[want] {
			t.Errorf("DEL->LAX stopovers missing expected airport %s", want)
		}
	}

	if airports["DEL"] {
		t.Error("origin DEL should not be in stopover list")
	}
	if airports["LAX"] {
		t.Error("destination LAX should not be in stopover list")
	}
}

func TestStopoversForRoute_BOMToLAX(t *testing.T) {
	got := StopoversForRoute("BOM", "LAX")
	if len(got) == 0 {
		t.Fatal("expected stopovers for BOM->LAX, got none")
	}

	// Should return route-specific list, not fallback.
	if &got[0] != &BOMToLAXStopovers[0] {
		t.Error("BOM->LAX should return the route-specific slice, not a copy")
	}

	// Verify expected cities.
	airports := make(map[string]bool)
	for _, s := range got {
		airports[s.Airport] = true
	}
	for _, want := range []string{"BKK", "SIN", "HKG", "NRT", "TPE"} {
		if !airports[want] {
			t.Errorf("BOM->LAX stopovers missing expected airport %s", want)
		}
	}

	if airports["BOM"] {
		t.Error("origin BOM should not be in stopover list")
	}
	if airports["LAX"] {
		t.Error("destination LAX should not be in stopover list")
	}
}

func TestStopoversForRoute_ReverseLAXToDEL(t *testing.T) {
	got := StopoversForRoute("LAX", "DEL")
	if len(got) == 0 {
		t.Fatal("expected stopovers for reverse route LAX->DEL, got none")
	}

	airports := make(map[string]bool)
	for _, s := range got {
		airports[s.Airport] = true
	}
	// KUL is in DEL->LAX route-specific list but NOT in GlobalFallbackHubs.
	// Its presence proves reverse lookup uses route-specific data.
	if !airports["KUL"] {
		t.Error("reverse LAX->DEL missing route-specific airport KUL (would not be in fallback)")
	}

	if airports["LAX"] {
		t.Error("origin LAX should not be in stopover list")
	}
	if airports["DEL"] {
		t.Error("destination DEL should not be in stopover list")
	}
}

func TestStopoversForRoute_DELToORD(t *testing.T) {
	got := StopoversForRoute("DEL", "ORD")
	if len(got) == 0 {
		t.Fatal("expected stopovers for DEL->ORD, got none")
	}

	// Should return route-specific list, not fallback.
	if &got[0] != &DELToORDStopovers[0] {
		t.Error("DEL->ORD should return the route-specific slice, not a copy")
	}

	// Verify expected cities: East Asia Pacific + European corridor.
	airports := make(map[string]bool)
	for _, s := range got {
		airports[s.Airport] = true
	}
	for _, want := range []string{"BKK", "SIN", "HKG", "NRT", "ICN", "IST"} {
		if !airports[want] {
			t.Errorf("DEL->ORD stopovers missing expected airport %s", want)
		}
	}

	if airports["DEL"] {
		t.Error("origin DEL should not be in stopover list")
	}
	if airports["ORD"] {
		t.Error("destination ORD should not be in stopover list")
	}
}

func TestStopoversForRoute_BOMToORD(t *testing.T) {
	got := StopoversForRoute("BOM", "ORD")
	if len(got) == 0 {
		t.Fatal("expected stopovers for BOM->ORD, got none")
	}

	// Should return route-specific list, not fallback.
	if &got[0] != &BOMToORDStopovers[0] {
		t.Error("BOM->ORD should return the route-specific slice, not a copy")
	}

	// Verify expected cities: East Asia Pacific + European corridor.
	airports := make(map[string]bool)
	for _, s := range got {
		airports[s.Airport] = true
	}
	for _, want := range []string{"BKK", "SIN", "HKG", "NRT", "ICN", "IST"} {
		if !airports[want] {
			t.Errorf("BOM->ORD stopovers missing expected airport %s", want)
		}
	}

	if airports["BOM"] {
		t.Error("origin BOM should not be in stopover list")
	}
	if airports["ORD"] {
		t.Error("destination ORD should not be in stopover list")
	}
}

func TestStopoversForRoute_ReverseORDToDEL(t *testing.T) {
	got := StopoversForRoute("ORD", "DEL")
	if len(got) == 0 {
		t.Fatal("expected stopovers for reverse route ORD->DEL, got none")
	}

	airports := make(map[string]bool)
	for _, s := range got {
		airports[s.Airport] = true
	}
	// ICN is in both DEL->ORD route-specific list AND GlobalFallbackHubs,
	// but NRT is also in fallback. Use HKG which is also in fallback.
	// Instead check that IST is present — IST is in fallback too.
	// Better: check count matches DEL->ORD minus origin/dest filtering.
	// The reverse lookup should return all DEL->ORD stopovers since
	// neither ORD nor DEL are stopover airports in that list.
	if len(got) != len(DELToORDStopovers) {
		t.Errorf("reverse ORD->DEL should return %d stopovers, got %d", len(DELToORDStopovers), len(got))
	}

	if airports["ORD"] {
		t.Error("origin ORD should not be in stopover list")
	}
	if airports["DEL"] {
		t.Error("destination DEL should not be in stopover list")
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
