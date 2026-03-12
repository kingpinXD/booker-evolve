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
//   - Parallel search across multiple providers (currently SerpAPI only)
package search

import (
	"time"

	"booker/types"
)

// PriceInsights contains pricing context from the flight search provider.
// Helps users gauge whether results are a good deal.
type PriceInsights struct {
	LowestPrice       float64    // cheapest available fare in USD
	PriceLevel        string     // "low", "typical", or "high"
	TypicalPriceRange [2]float64 // [low, high] typical fare range in USD
}

// Itinerary represents a complete multi-leg journey that the user can book.
// Unlike types.Flight which is a single provider result, an Itinerary may
// combine results from different searches and includes stopover information.
type Itinerary struct {
	Legs        []Leg
	TotalPrice  types.Money
	TotalTravel time.Duration // sum of in-air time across all legs
	TotalTrip   time.Duration // wall-clock time from first departure to last arrival
	Score       float64       // LLM-assigned score (0-100), 0 means unscored
	Reasoning   string        // LLM explanation of the score
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
	Notes    string        // context about the stopover city (connectivity, food, visa, etc.)
}

// FareTrend summarizes price variation across dates in a flex-date search.
// Empty when FlexDays is 0 or there are no results.
type FareTrend struct {
	CheapestDate string  // YYYY-MM-DD of the cheapest departure
	PriciestDate string  // YYYY-MM-DD of the most expensive departure
	MinPrice     float64 // lowest price across all dates
	MaxPrice     float64 // highest price across all dates
}

// ComputeFareTrend analyzes itineraries from a flex-date search and returns
// per-date price extremes. Returns a zero FareTrend when itineraries is empty.
func ComputeFareTrend(itineraries []Itinerary) FareTrend {
	if len(itineraries) == 0 {
		return FareTrend{}
	}

	// Find cheapest itinerary per date, track overall min/max.
	dateBest := make(map[string]float64) // date -> cheapest price
	for _, itin := range itineraries {
		date := itinDate(itin)
		if date == "" {
			continue
		}
		price := itin.TotalPrice.Amount
		if best, ok := dateBest[date]; !ok || price < best {
			dateBest[date] = price
		}
	}

	if len(dateBest) == 0 {
		return FareTrend{}
	}

	var ft FareTrend
	first := true
	for date, price := range dateBest {
		if first || price < ft.MinPrice {
			ft.MinPrice = price
			ft.CheapestDate = date
		}
		if first || price > ft.MaxPrice {
			ft.MaxPrice = price
			ft.PriciestDate = date
		}
		first = false
	}
	return ft
}

// itinDate extracts the departure date (YYYY-MM-DD) from an itinerary's first leg.
func itinDate(itin Itinerary) string {
	if len(itin.Legs) > 0 && len(itin.Legs[0].Flight.Outbound) > 0 {
		t := itin.Legs[0].Flight.Outbound[0].DepartureTime
		if !t.IsZero() {
			return t.Format("2006-01-02")
		}
	}
	return ""
}
