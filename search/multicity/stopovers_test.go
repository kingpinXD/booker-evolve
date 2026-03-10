package multicity

import "testing"

func TestStopoversForRoute_DELtoYYZ(t *testing.T) {
	stopovers := StopoversForRoute("DEL", "YYZ")
	if len(stopovers) != len(DELToYYZStopovers) {
		t.Fatalf("got %d stopovers, want %d", len(stopovers), len(DELToYYZStopovers))
	}
}

func TestStopoversForRoute_UnknownReturnsNil(t *testing.T) {
	stopovers := StopoversForRoute("SFO", "NRT")
	if stopovers != nil {
		t.Fatalf("unknown route should return nil, got %d stopovers", len(stopovers))
	}
}

func TestStopoversForRoute_BOMtoYYZ(t *testing.T) {
	stopovers := StopoversForRoute("BOM", "YYZ")
	if len(stopovers) == 0 {
		t.Fatal("BOM→YYZ should have stopover cities")
	}
	// BOM→YYZ should have geographically appropriate stopovers (SE Asia, East Asia, Europe)
	// and NOT include cities that only make sense for DEL corridors.
	airports := map[string]bool{}
	for _, s := range stopovers {
		airports[s.Airport] = true
	}
	// Bangkok and Singapore are geographically close to Mumbai.
	if !airports["BKK"] && !airports["SIN"] {
		t.Error("BOM→YYZ should include at least BKK or SIN")
	}
}

func TestStopoversForRoute_DELtoYVR(t *testing.T) {
	stopovers := StopoversForRoute("DEL", "YVR")
	if len(stopovers) == 0 {
		t.Fatal("DEL→YVR should have stopover cities")
	}
	airports := map[string]bool{}
	for _, s := range stopovers {
		airports[s.Airport] = true
	}
	// Tokyo/Seoul are natural Pacific stopovers for DEL→YVR.
	if !airports["NRT"] && !airports["ICN"] {
		t.Error("DEL→YVR should include at least NRT or ICN")
	}
}

func TestStopoversForRoute_CityCount(t *testing.T) {
	stopovers := StopoversForRoute("DEL", "YYZ")
	if len(stopovers) != 10 {
		t.Fatalf("expected 10 stopover cities, got %d", len(stopovers))
	}
}

func TestStopoverCityFields(t *testing.T) {
	for _, sc := range DELToYYZStopovers {
		t.Run(sc.City, func(t *testing.T) {
			if sc.City == "" {
				t.Error("City is empty")
			}
			if len(sc.Airport) != 3 {
				t.Errorf("Airport %q is not a 3-letter IATA code", sc.Airport)
			}
			if sc.KiwiID == "" {
				t.Error("KiwiID is empty")
			}
			if sc.Region == "" {
				t.Error("Region is empty")
			}
			if sc.MinStay <= 0 {
				t.Error("MinStay must be positive")
			}
			if sc.MaxStay <= 0 {
				t.Error("MaxStay must be positive")
			}
			if sc.MaxStay < sc.MinStay {
				t.Errorf("MaxStay (%v) < MinStay (%v)", sc.MaxStay, sc.MinStay)
			}
			if sc.Notes == "" {
				t.Error("Notes is empty")
			}
		})
	}
}

func TestAllRouteStopoversValid(t *testing.T) {
	routes := []struct {
		origin, dest string
	}{
		{"DEL", "YYZ"},
		{"BOM", "YYZ"},
		{"DEL", "YVR"},
	}
	for _, r := range routes {
		stopovers := StopoversForRoute(r.origin, r.dest)
		for _, sc := range stopovers {
			t.Run(r.origin+"→"+r.dest+"/"+sc.City, func(t *testing.T) {
				if sc.City == "" {
					t.Error("City is empty")
				}
				if len(sc.Airport) != 3 {
					t.Errorf("Airport %q is not a 3-letter IATA code", sc.Airport)
				}
				if sc.Region == "" {
					t.Error("Region is empty")
				}
				if sc.MinStay <= 0 || sc.MaxStay <= 0 || sc.MaxStay < sc.MinStay {
					t.Errorf("invalid stay range: %v-%v", sc.MinStay, sc.MaxStay)
				}
			})
		}
	}
}

func TestStopoverAirportCodes(t *testing.T) {
	expected := map[string]bool{
		"HKG": true, "SIN": true, "BKK": true, "NRT": true, "ICN": true,
		"KUL": true, "IST": true, "LHR": true, "FRA": true, "CDG": true,
	}
	for _, sc := range DELToYYZStopovers {
		if !expected[sc.Airport] {
			t.Errorf("unexpected airport code %q", sc.Airport)
		}
		delete(expected, sc.Airport)
	}
	for code := range expected {
		t.Errorf("missing airport code %q", code)
	}
}
