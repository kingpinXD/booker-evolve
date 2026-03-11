package search

import "testing"

func TestAlliance_StarAlliance(t *testing.T) {
	if got := Alliance("AC"); got != "Star Alliance" {
		t.Errorf("Alliance(AC) = %q, want %q", got, "Star Alliance")
	}
}

func TestAlliance_OneWorld(t *testing.T) {
	if got := Alliance("BA"); got != "OneWorld" {
		t.Errorf("Alliance(BA) = %q, want %q", got, "OneWorld")
	}
}

func TestAlliance_SkyTeam(t *testing.T) {
	if got := Alliance("DL"); got != "SkyTeam" {
		t.Errorf("Alliance(DL) = %q, want %q", got, "SkyTeam")
	}
}

func TestAlliance_Unknown(t *testing.T) {
	if got := Alliance("XX"); got != "" {
		t.Errorf("Alliance(XX) = %q, want %q", got, "")
	}
}

func TestSameAlliance_SameAlliance(t *testing.T) {
	if !SameAlliance("AC", "LH") {
		t.Error("SameAlliance(AC, LH) = false, want true")
	}
}

func TestSameAlliance_DifferentAlliance(t *testing.T) {
	if SameAlliance("AC", "AA") {
		t.Error("SameAlliance(AC, AA) = true, want false")
	}
}

func TestSameAlliance_UnknownCode(t *testing.T) {
	if SameAlliance("AC", "XX") {
		t.Error("SameAlliance(AC, XX) = true, want false")
	}
	if SameAlliance("XX", "YY") {
		t.Error("SameAlliance(XX, YY) = true, want false")
	}
}

func TestSameAlliance_SameCode(t *testing.T) {
	if !SameAlliance("AC", "AC") {
		t.Error("SameAlliance(AC, AC) = false, want true")
	}
}
