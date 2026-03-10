package multicity

import (
	"testing"

	"booker/search"
	"booker/types"
)

func TestStrategy_Name(t *testing.T) {
	s := NewStrategy(nil, "2026-03-30")
	if got := s.Name(); got != "multicity" {
		t.Errorf("Name() = %q, want %q", got, "multicity")
	}
}

func TestStrategy_Description(t *testing.T) {
	s := NewStrategy(nil, "2026-03-30")
	if s.Description() == "" {
		t.Error("Description() should not be empty")
	}
}

// TestStrategy_ImplementsInterface verifies Strategy satisfies search.Strategy.
func TestStrategy_ImplementsInterface(t *testing.T) {
	var _ search.Strategy = (*Strategy)(nil)
}

func TestNewStrategy(t *testing.T) {
	s := NewStrategy(nil, "2026-04-01")
	if s == nil {
		t.Fatal("NewStrategy returned nil")
	}
	if s.leg2Date != "2026-04-01" {
		t.Errorf("leg2Date = %q, want %q", s.leg2Date, "2026-04-01")
	}
}

// TestStrategy_RequestMapping verifies the Request→SearchParams mapping.
func TestStrategy_RequestMapping(t *testing.T) {
	req := search.Request{
		Origin:        "DEL",
		Destination:   "YYZ",
		DepartureDate: "2026-03-24",
		Passengers:    2,
		CabinClass:    types.CabinBusiness,
		FlexDays:      3,
		MaxStops:      1,
		MaxResults:    5,
	}

	s := &Strategy{leg2Date: "2026-03-30"}
	params := s.toSearchParams(req)

	if params.Origin != "DEL" {
		t.Errorf("Origin = %q, want %q", params.Origin, "DEL")
	}
	if params.Destination != "YYZ" {
		t.Errorf("Destination = %q, want %q", params.Destination, "YYZ")
	}
	if params.DepartureDate != "2026-03-24" {
		t.Errorf("DepartureDate = %q, want %q", params.DepartureDate, "2026-03-24")
	}
	if params.Leg2Date != "2026-03-30" {
		t.Errorf("Leg2Date = %q, want %q", params.Leg2Date, "2026-03-30")
	}
	if params.Passengers != 2 {
		t.Errorf("Passengers = %d, want 2", params.Passengers)
	}
	if params.CabinClass != types.CabinBusiness {
		t.Errorf("CabinClass = %q, want %q", params.CabinClass, types.CabinBusiness)
	}
	if params.FlexDays != 3 {
		t.Errorf("FlexDays = %d, want 3", params.FlexDays)
	}
	if params.MaxLayoversLeg1 != 1 {
		t.Errorf("MaxLayoversLeg1 = %d, want 1", params.MaxLayoversLeg1)
	}
	if params.MaxLayoversLeg2 != 1 {
		t.Errorf("MaxLayoversLeg2 = %d, want 1", params.MaxLayoversLeg2)
	}
	if params.MaxResults != 5 {
		t.Errorf("MaxResults = %d, want 5", params.MaxResults)
	}
}
