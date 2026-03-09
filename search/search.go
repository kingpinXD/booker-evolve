// Package search contains the core search logic for finding flights.
//
// The search package is organized into sub-packages by search strategy:
//
//   - search/multicity: Multi-city halt search. Breaks a long-haul journey
//     into two legs with a stopover city in between. For example,
//     Delhi → Hong Kong (stay 3 days) → Toronto. This is useful when
//     direct routes are unavailable or expensive, and when the traveler
//     wants to visit an intermediate city.
//
// # Architecture
//
// Each search strategy follows the same pipeline:
//
//  1. EXPAND  — Generate candidate leg searches from the user's intent.
//  2. FETCH   — Hit the flight provider APIs for each leg.
//  3. FILTER  — Remove results that violate hard constraints (blocked
//     airlines, closed airspace, bad layover durations).
//  4. COMBINE — Pair/merge legs into complete itineraries.
//  5. RANK    — Send top candidates to the LLM for intelligent ranking
//     based on soft preferences (cost, airline consistency,
//     stopover city attractiveness, time of day, etc.).
//
// # Iteration Plan
//
// This system is designed to be improved daily. Current limitations and
// next steps are documented inline with "TODO(iterate)" comments.
// Key areas for future optimization:
//
//   - Dynamic stopover city selection based on route geometry
//   - Date flexibility scoring (shift ±3 days to find cheaper combos)
//   - Airline alliance grouping (prefer same alliance across legs)
//   - Historical price trends to judge if a price is "good"
//   - Parallel search across multiple providers (currently Kiwi only)
package search

import (
	"time"

	"booker/types"
)

// Itinerary represents a complete multi-leg journey that the user can book.
// Unlike types.Flight which is a single provider result, an Itinerary may
// combine results from different searches and includes stopover information.
type Itinerary struct {
	Legs         []Leg
	TotalPrice   types.Money
	TotalTravel  time.Duration // sum of in-air time across all legs
	TotalTrip    time.Duration // wall-clock time from first departure to last arrival
	Score        float64       // LLM-assigned score (0-100), 0 means unscored
	Reasoning    string        // LLM explanation of the score
}

// Leg is one bookable segment of the itinerary (e.g. Delhi → Hong Kong).
// Each leg corresponds to a single types.Flight from a provider search.
type Leg struct {
	Flight   types.Flight
	Stopover *Stopover // nil if this is the final leg
}

// Stopover describes a halt between two legs.
type Stopover struct {
	City     string        // e.g. "Hong Kong"
	Airport  string        // IATA code, e.g. "HKG"
	Duration time.Duration // how long the traveler stays in the city
}
