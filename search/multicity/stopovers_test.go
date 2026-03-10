package multicity

import "testing"

func TestStopoversForRoute_DELtoYYZ(t *testing.T) {
	stopovers := StopoversForRoute("DEL", "YYZ")
	if len(stopovers) != len(DELToYYZStopovers) {
		t.Fatalf("got %d stopovers, want %d", len(stopovers), len(DELToYYZStopovers))
	}
}

func TestStopoversForRoute_UnknownFallback(t *testing.T) {
	stopovers := StopoversForRoute("SFO", "NRT")
	if len(stopovers) != len(DELToYYZStopovers) {
		t.Fatalf("unknown route should fallback to DELToYYZ, got %d stopovers", len(stopovers))
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
