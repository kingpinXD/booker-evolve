package multicity

import (
	"time"

	"booker/search"
	"booker/types"
)

// CombineLegs takes flight results for leg1 (origin → stopover) and leg2
// (stopover → destination) and produces candidate Itineraries.
//
// # Combination Logic
//
// For each (leg1_flight, leg2_flight) pair, we check:
//
//  1. STOPOVER DURATION — The gap between leg1 arrival and leg2 departure
//     must be within [minStay, maxStay]. We want the traveler to have
//     2-4 days in the stopover city, NOT a 16-hour airport layover.
//     A 16-hour gap means you're stuck in the airport.
//     A 2-day gap means you check into a hotel and explore the city.
//
//  2. AIRLINE CONSISTENCY — We prefer itineraries where both legs use
//     the same airline or at least the same airline alliance. Mixed
//     airlines mean separate bookings, no baggage transfer, and more
//     risk if one leg is delayed.
//
//  3. LAYOVER QUALITY (within each leg) — If a single leg has connections
//     (e.g. DEL → BKK → HKG), we reject legs where any single layover
//     exceeds MaxAirportLayover. A 2-hour connection is fine. An 8-hour
//     airport wait is miserable.
//
// TODO(iterate): Score airline alliance matching (Star Alliance, OneWorld, SkyTeam).
// TODO(iterate): Add time-of-day preferences (avoid red-eye on long legs).
// TODO(iterate): Weight stopover city attractiveness into the combination.


// CombineParams controls how legs are paired.
type CombineParams struct {
	Stopover StopoverCity

	// MinStay overrides the stopover city's default if set.
	MinStay time.Duration
	// MaxStay overrides the stopover city's default if set.
	MaxStay time.Duration
}

// CombineLegs pairs leg1 and leg2 flights into valid Itineraries.
// It returns only combinations that pass all hard constraints.
func CombineLegs(leg1, leg2 []types.Flight, params CombineParams) []search.Itinerary {
	minStay := params.Stopover.MinStay
	if params.MinStay > 0 {
		minStay = params.MinStay
	}
	maxStay := params.Stopover.MaxStay
	if params.MaxStay > 0 {
		maxStay = params.MaxStay
	}

	var results []search.Itinerary
	for _, f1 := range leg1 {
		if hasLongLayover(f1.Outbound) {
			continue
		}

		leg1Arrival := lastArrival(f1.Outbound)
		if leg1Arrival.IsZero() {
			continue
		}

		for _, f2 := range leg2 {
			if hasLongLayover(f2.Outbound) {
				continue
			}

			leg2Departure := firstDeparture(f2.Outbound)
			if leg2Departure.IsZero() {
				continue
			}

			// Check stopover duration is within bounds.
			gap := leg2Departure.Sub(leg1Arrival)
			if gap < minStay || gap > maxStay {
				continue
			}

			itin := buildItinerary(f1, f2, params.Stopover, gap)
			results = append(results, itin)
		}
	}
	return results
}

// hasLongLayover checks if any connection within a leg exceeds the maximum
// airport layover time. This catches the "16 hours stuck in the airport"
// scenario that we want to avoid.
func hasLongLayover(segments []types.Segment) bool {
	for _, seg := range segments {
		if seg.LayoverDuration > types.MaxLayover && seg.LayoverDuration > 0 {
			return true
		}
		if seg.LayoverDuration > 0 && seg.LayoverDuration < types.MinLayover {
			return true
		}
	}
	return false
}

func lastArrival(segments []types.Segment) time.Time {
	if len(segments) == 0 {
		return time.Time{}
	}
	return segments[len(segments)-1].ArrivalTime
}

func firstDeparture(segments []types.Segment) time.Time {
	if len(segments) == 0 {
		return time.Time{}
	}
	return segments[0].DepartureTime
}

// buildItinerary constructs an Itinerary from two legs and stopover info.
func buildItinerary(leg1, leg2 types.Flight, stop StopoverCity, gap time.Duration) search.Itinerary {
	totalPrice := types.Money{
		Amount:   leg1.Price.Amount + leg2.Price.Amount,
		Currency: leg1.Price.Currency,
	}

	totalTravel := leg1.TotalDuration + leg2.TotalDuration
	firstDep := firstDeparture(leg1.Outbound)
	lastArr := lastArrival(leg2.Outbound)
	totalTrip := lastArr.Sub(firstDep)

	return search.Itinerary{
		Legs: []search.Leg{
			{
				Flight: leg1,
				Stopover: &search.Stopover{
					City:     stop.City,
					Airport:  stop.Airport,
					Duration: gap,
				},
			},
			{
				Flight: leg2,
			},
		},
		TotalPrice:  totalPrice,
		TotalTravel: totalTravel,
		TotalTrip:   totalTrip,
	}
}

// PrimaryAirline returns the most-used airline IATA code across all
// outbound segments of a flight. Used to check airline consistency
// between legs.
func PrimaryAirline(f types.Flight) string {
	counts := make(map[string]int)
	for _, seg := range f.Outbound {
		counts[seg.Airline]++
	}
	var best string
	var bestCount int
	for code, count := range counts {
		if count > bestCount {
			best = code
			bestCount = count
		}
	}
	return best
}

// SameAirline returns true if both flights primarily use the same airline.
func SameAirline(a, b types.Flight) bool {
	return PrimaryAirline(a) == PrimaryAirline(b)
}
