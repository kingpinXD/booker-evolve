package search

import (
	"sort"
	"testing"
)

func TestNearbyAirports_KnownCluster(t *testing.T) {
	got := NearbyAirports("JFK")
	if len(got) != 2 {
		t.Fatalf("NearbyAirports(JFK) returned %d codes, want 2", len(got))
	}
	sort.Strings(got)
	if got[0] != "EWR" || got[1] != "LGA" {
		t.Errorf("NearbyAirports(JFK) = %v, want [EWR LGA]", got)
	}
}

func TestNearbyAirports_UnknownCode(t *testing.T) {
	if got := NearbyAirports("XYZ"); got != nil {
		t.Errorf("NearbyAirports(XYZ) = %v, want nil", got)
	}
}

func TestNearbyAirports_SingleAirportMetro(t *testing.T) {
	if got := NearbyAirports("DEL"); got != nil {
		t.Errorf("NearbyAirports(DEL) = %v, want nil", got)
	}
	if got := NearbyAirports("BOM"); got != nil {
		t.Errorf("NearbyAirports(BOM) = %v, want nil", got)
	}
}

func TestNearbyAirports_AllClusters(t *testing.T) {
	for metro, codes := range airportClusters {
		for _, code := range codes {
			got := NearbyAirports(code)
			if got == nil {
				t.Errorf("NearbyAirports(%q) [%s] = nil, want siblings", code, metro)
				continue
			}
			if len(got) != len(codes)-1 {
				t.Errorf("NearbyAirports(%q) [%s] returned %d siblings, want %d",
					code, metro, len(got), len(codes)-1)
			}
			// The input code must not appear in the result.
			for _, s := range got {
				if s == code {
					t.Errorf("NearbyAirports(%q) [%s] contains input code in result", code, metro)
				}
			}
		}
	}
}
