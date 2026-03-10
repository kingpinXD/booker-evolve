package types

import (
	"errors"
	"testing"
	"time"

	"booker/config"
)

func TestSearchRequest_IsRoundTrip(t *testing.T) {
	oneWay := SearchRequest{
		Origin:        "YYZ",
		Destination:   "DEL",
		DepartureDate: time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC),
	}
	if oneWay.IsRoundTrip() {
		t.Error("one-way request should not be round trip")
	}

	roundTrip := SearchRequest{
		Origin:        "YYZ",
		Destination:   "DEL",
		DepartureDate: time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC),
		ReturnDate:    time.Date(2026, 3, 25, 0, 0, 0, 0, time.UTC),
	}
	if !roundTrip.IsRoundTrip() {
		t.Error("round-trip request should be round trip")
	}
}

func TestFlight_Stops(t *testing.T) {
	tests := []struct {
		name     string
		outbound []Segment
		want     int
	}{
		{"no segments", nil, 0},
		{"direct", []Segment{{Origin: "A", Destination: "B"}}, 0},
		{"one stop", []Segment{{}, {}}, 1},
		{"two stops", []Segment{{}, {}, {}}, 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := Flight{Outbound: tt.outbound}
			if got := f.Stops(); got != tt.want {
				t.Errorf("Stops() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestProviderError_Error(t *testing.T) {
	pe := ProviderError{
		Provider: config.ProviderName("serpapi"),
		Err:      errors.New("rate limited"),
	}
	want := "serpapi: rate limited"
	if got := pe.Error(); got != want {
		t.Errorf("Error() = %q, want %q", got, want)
	}
}
