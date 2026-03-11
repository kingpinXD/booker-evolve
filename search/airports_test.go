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

func TestNearbyAirports_IndianClusters(t *testing.T) {
	tests := []struct {
		code string
		want []string
	}{
		{"DEL", []string{"JAI"}},
		{"JAI", []string{"DEL"}},
		{"BOM", []string{"PNQ"}},
		{"PNQ", []string{"BOM"}},
	}
	for _, tt := range tests {
		got := NearbyAirports(tt.code)
		sort.Strings(got)
		want := make([]string, len(tt.want))
		copy(want, tt.want)
		sort.Strings(want)
		if len(got) != len(want) {
			t.Errorf("NearbyAirports(%q) = %v, want %v", tt.code, got, want)
			continue
		}
		for i := range got {
			if got[i] != want[i] {
				t.Errorf("NearbyAirports(%q) = %v, want %v", tt.code, got, want)
				break
			}
		}
	}
}

func TestNearbyAirports_NewClusters(t *testing.T) {
	tests := []struct {
		code string
		want []string
	}{
		{"BKK", []string{"DMK"}},
		{"DMK", []string{"BKK"}},
		{"IST", []string{"SAW"}},
		{"SAW", []string{"IST"}},
		{"PEK", []string{"PKX"}},
		{"PKX", []string{"PEK"}},
		{"KIX", []string{"ITM"}},
		{"ITM", []string{"KIX"}},
		{"FCO", []string{"CIA"}},
		{"CIA", []string{"FCO"}},
		{"TPE", []string{"TSA"}},
		{"TSA", []string{"TPE"}},
		{"MIA", []string{"FLL"}},
		{"FLL", []string{"MIA"}},
		{"GRU", []string{"CGH", "VCP"}},
		{"CGH", []string{"GRU", "VCP"}},
		{"VCP", []string{"CGH", "GRU"}},
	}
	for _, tt := range tests {
		got := NearbyAirports(tt.code)
		sort.Strings(got)
		want := make([]string, len(tt.want))
		copy(want, tt.want)
		sort.Strings(want)
		if len(got) != len(want) {
			t.Errorf("NearbyAirports(%q) = %v, want %v", tt.code, got, want)
			continue
		}
		for i := range got {
			if got[i] != want[i] {
				t.Errorf("NearbyAirports(%q) = %v, want %v", tt.code, got, want)
				break
			}
		}
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
