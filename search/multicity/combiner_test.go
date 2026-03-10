package multicity

import (
	"testing"
	"time"

	"booker/types"
)

// basetime is a fixed reference point for test data construction.
var basetime = time.Date(2026, 3, 24, 10, 0, 0, 0, time.UTC)

// makeFlight builds a minimal Flight with one outbound segment.
func makeFlight(airline, origin, dest string, dep, arr time.Time, price float64) types.Flight {
	return types.Flight{
		Price: types.Money{Amount: price, Currency: "USD"},
		Outbound: []types.Segment{{
			Airline:       airline,
			Origin:        origin,
			Destination:   dest,
			DepartureTime: dep,
			ArrivalTime:   arr,
		}},
		TotalDuration: arr.Sub(dep),
	}
}

// makeConnectingFlight builds a Flight with two outbound segments and a layover.
func makeConnectingFlight(airline, orig, mid, dest string, dep1, arr1, dep2, arr2 time.Time, layover time.Duration, price float64) types.Flight {
	return types.Flight{
		Price: types.Money{Amount: price, Currency: "USD"},
		Outbound: []types.Segment{
			{
				Airline:         airline,
				Origin:          orig,
				Destination:     mid,
				DepartureTime:   dep1,
				ArrivalTime:     arr1,
				LayoverDuration: layover,
			},
			{
				Airline:       airline,
				Origin:        mid,
				Destination:   dest,
				DepartureTime: dep2,
				ArrivalTime:   arr2,
			},
		},
		TotalDuration: arr2.Sub(dep1),
	}
}

func defaultParams() CombineParams {
	return CombineParams{
		Stopover: StopoverCity{
			City:    "Hong Kong",
			Airport: "HKG",
			MinStay: types.DefaultMinStopover,
			MaxStay: types.DefaultMaxStopover,
		},
	}
}

func TestCombineLegs(t *testing.T) {
	// leg1: DEL -> HKG, arrives Mar 24 18:00
	// leg2: HKG -> YYZ, departs Mar 27 10:00 (3 days gap = valid)
	leg1 := makeFlight("CX", "DEL", "HKG", basetime, basetime.Add(8*time.Hour), 300)
	leg2 := makeFlight("AC", "HKG", "YYZ", basetime.Add(72*time.Hour), basetime.Add(88*time.Hour), 500)

	tests := []struct {
		name      string
		leg1      []types.Flight
		leg2      []types.Flight
		params    CombineParams
		wantCount int
	}{
		{
			name:      "valid pair within stopover bounds",
			leg1:      []types.Flight{leg1},
			leg2:      []types.Flight{leg2},
			params:    defaultParams(),
			wantCount: 1,
		},
		{
			name: "gap too short (under MinStay)",
			leg1: []types.Flight{leg1},
			leg2: []types.Flight{
				makeFlight("AC", "HKG", "YYZ", basetime.Add(10*time.Hour), basetime.Add(26*time.Hour), 500),
			},
			params:    defaultParams(),
			wantCount: 0,
		},
		{
			name: "gap too long (over MaxStay)",
			leg1: []types.Flight{leg1},
			leg2: []types.Flight{
				makeFlight("AC", "HKG", "YYZ", basetime.Add(200*time.Hour), basetime.Add(216*time.Hour), 500),
			},
			params:    defaultParams(),
			wantCount: 0,
		},
		{
			name:      "empty leg1",
			leg1:      nil,
			leg2:      []types.Flight{leg2},
			params:    defaultParams(),
			wantCount: 0,
		},
		{
			name:      "empty leg2",
			leg1:      []types.Flight{leg1},
			leg2:      nil,
			params:    defaultParams(),
			wantCount: 0,
		},
		{
			name: "leg1 with long layover rejected",
			leg1: []types.Flight{
				makeConnectingFlight("CX", "DEL", "BKK", "HKG",
					basetime, basetime.Add(4*time.Hour),
					basetime.Add(12*time.Hour), basetime.Add(14*time.Hour),
					8*time.Hour, 300),
			},
			leg2:      []types.Flight{leg2},
			params:    defaultParams(),
			wantCount: 0,
		},
		{
			name: "leg2 with long layover rejected",
			leg1: []types.Flight{leg1},
			leg2: []types.Flight{
				makeConnectingFlight("AC", "HKG", "YVR", "YYZ",
					basetime.Add(72*time.Hour), basetime.Add(82*time.Hour),
					basetime.Add(92*time.Hour), basetime.Add(95*time.Hour),
					10*time.Hour, 500),
			},
			params:    defaultParams(),
			wantCount: 0,
		},
		{
			name: "multiple valid combinations",
			leg1: []types.Flight{
				makeFlight("CX", "DEL", "HKG", basetime, basetime.Add(8*time.Hour), 300),
				makeFlight("AI", "DEL", "HKG", basetime.Add(2*time.Hour), basetime.Add(10*time.Hour), 280),
			},
			leg2: []types.Flight{
				makeFlight("AC", "HKG", "YYZ", basetime.Add(72*time.Hour), basetime.Add(88*time.Hour), 500),
				makeFlight("CX", "HKG", "YYZ", basetime.Add(73*time.Hour), basetime.Add(89*time.Hour), 520),
			},
			params:    defaultParams(),
			wantCount: 4,
		},
		{
			name: "params override MinStay/MaxStay",
			leg1: []types.Flight{leg1},
			leg2: []types.Flight{
				// Gap is 24h from leg1 arrival. Default MinStay is 48h so this would fail.
				// But we override MinStay to 20h.
				makeFlight("AC", "HKG", "YYZ", basetime.Add(32*time.Hour), basetime.Add(48*time.Hour), 500),
			},
			params: CombineParams{
				Stopover: StopoverCity{
					City:    "Hong Kong",
					Airport: "HKG",
					MinStay: types.DefaultMinStopover,
					MaxStay: types.DefaultMaxStopover,
				},
				MinStay: 20 * time.Hour,
				MaxStay: 48 * time.Hour,
			},
			wantCount: 1,
		},
		{
			name: "flight with empty outbound segments skipped",
			leg1: []types.Flight{{
				Price:    types.Money{Amount: 300, Currency: "USD"},
				Outbound: nil,
			}},
			leg2:      []types.Flight{leg2},
			params:    defaultParams(),
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CombineLegs(tt.leg1, tt.leg2, tt.params)
			if len(got) != tt.wantCount {
				t.Errorf("CombineLegs() returned %d itineraries, want %d", len(got), tt.wantCount)
			}
		})
	}
}

func TestCombineLegs_ItineraryContents(t *testing.T) {
	leg1 := makeFlight("CX", "DEL", "HKG", basetime, basetime.Add(8*time.Hour), 300)
	leg2 := makeFlight("AC", "HKG", "YYZ", basetime.Add(72*time.Hour), basetime.Add(88*time.Hour), 500)

	results := CombineLegs([]types.Flight{leg1}, []types.Flight{leg2}, defaultParams())
	if len(results) != 1 {
		t.Fatalf("expected 1 itinerary, got %d", len(results))
	}

	itin := results[0]

	// Check total price.
	if itin.TotalPrice.Amount != 800 {
		t.Errorf("TotalPrice = %v, want 800", itin.TotalPrice.Amount)
	}
	if itin.TotalPrice.Currency != "USD" {
		t.Errorf("Currency = %v, want USD", itin.TotalPrice.Currency)
	}

	// Check legs.
	if len(itin.Legs) != 2 {
		t.Fatalf("expected 2 legs, got %d", len(itin.Legs))
	}

	// First leg should have stopover.
	if itin.Legs[0].Stopover == nil {
		t.Fatal("first leg stopover is nil")
	}
	if itin.Legs[0].Stopover.City != "Hong Kong" {
		t.Errorf("stopover city = %q, want %q", itin.Legs[0].Stopover.City, "Hong Kong")
	}
	if itin.Legs[0].Stopover.Airport != "HKG" {
		t.Errorf("stopover airport = %q, want %q", itin.Legs[0].Stopover.Airport, "HKG")
	}
	if itin.Legs[0].Stopover.Duration != 64*time.Hour {
		t.Errorf("stopover duration = %v, want %v", itin.Legs[0].Stopover.Duration, 64*time.Hour)
	}

	// Second leg should not have stopover.
	if itin.Legs[1].Stopover != nil {
		t.Error("second leg should not have a stopover")
	}

	// TotalTravel = leg1 duration + leg2 duration.
	wantTravel := 8*time.Hour + 16*time.Hour
	if itin.TotalTravel != wantTravel {
		t.Errorf("TotalTravel = %v, want %v", itin.TotalTravel, wantTravel)
	}

	// TotalTrip = last arrival - first departure.
	wantTrip := basetime.Add(88 * time.Hour).Sub(basetime)
	if itin.TotalTrip != wantTrip {
		t.Errorf("TotalTrip = %v, want %v", itin.TotalTrip, wantTrip)
	}
}

func TestHasLongLayover(t *testing.T) {
	tests := []struct {
		name string
		segs []types.Segment
		want bool
	}{
		{
			name: "no layover (single segment)",
			segs: []types.Segment{{LayoverDuration: 0}},
			want: false,
		},
		{
			name: "acceptable layover",
			segs: []types.Segment{{LayoverDuration: 2 * time.Hour}},
			want: false,
		},
		{
			name: "layover exceeds MaxLayover",
			segs: []types.Segment{{LayoverDuration: 7 * time.Hour}},
			want: true,
		},
		{
			name: "layover below MinLayover",
			segs: []types.Segment{{LayoverDuration: 30 * time.Minute}},
			want: true,
		},
		{
			name: "empty segments",
			segs: nil,
			want: false,
		},
		{
			name: "multiple segments, one bad",
			segs: []types.Segment{
				{LayoverDuration: 2 * time.Hour},
				{LayoverDuration: 8 * time.Hour},
			},
			want: true,
		},
		{
			name: "exactly at MaxLayover boundary",
			segs: []types.Segment{{LayoverDuration: types.MaxLayover}},
			want: false,
		},
		{
			name: "exactly at MinLayover boundary",
			segs: []types.Segment{{LayoverDuration: types.MinLayover}},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := hasLongLayover(tt.segs); got != tt.want {
				t.Errorf("hasLongLayover() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLastArrival(t *testing.T) {
	t1 := basetime
	t2 := basetime.Add(4 * time.Hour)

	tests := []struct {
		name string
		segs []types.Segment
		want time.Time
	}{
		{"empty", nil, time.Time{}},
		{"single", []types.Segment{{ArrivalTime: t1}}, t1},
		{"multiple returns last", []types.Segment{{ArrivalTime: t1}, {ArrivalTime: t2}}, t2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := lastArrival(tt.segs); !got.Equal(tt.want) {
				t.Errorf("lastArrival() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFirstDeparture(t *testing.T) {
	t1 := basetime
	t2 := basetime.Add(4 * time.Hour)

	tests := []struct {
		name string
		segs []types.Segment
		want time.Time
	}{
		{"empty", nil, time.Time{}},
		{"single", []types.Segment{{DepartureTime: t1}}, t1},
		{"multiple returns first", []types.Segment{{DepartureTime: t1}, {DepartureTime: t2}}, t1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := firstDeparture(tt.segs); !got.Equal(tt.want) {
				t.Errorf("firstDeparture() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPrimaryAirline(t *testing.T) {
	tests := []struct {
		name string
		f    types.Flight
		want string
	}{
		{
			name: "single segment",
			f:    makeFlight("CX", "DEL", "HKG", basetime, basetime.Add(8*time.Hour), 300),
			want: "CX",
		},
		{
			name: "majority airline wins",
			f: types.Flight{
				Outbound: []types.Segment{
					{Airline: "CX"}, {Airline: "CX"}, {Airline: "AI"},
				},
			},
			want: "CX",
		},
		{
			name: "empty flight",
			f:    types.Flight{},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := PrimaryAirline(tt.f); got != tt.want {
				t.Errorf("PrimaryAirline() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestSameAirline(t *testing.T) {
	cx := makeFlight("CX", "DEL", "HKG", basetime, basetime.Add(8*time.Hour), 300)
	ai := makeFlight("AI", "DEL", "HKG", basetime, basetime.Add(8*time.Hour), 280)

	if !SameAirline(cx, cx) {
		t.Error("SameAirline(CX, CX) should be true")
	}
	if SameAirline(cx, ai) {
		t.Error("SameAirline(CX, AI) should be false")
	}
}
