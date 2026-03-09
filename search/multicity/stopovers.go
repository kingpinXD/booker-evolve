// Package multicity implements the multi-city halt search strategy.
//
// # What is a multi-city halt?
//
// Instead of booking one long-haul flight (e.g. Delhi → Toronto, 16+ hours),
// the traveler breaks the journey into two legs with a stopover of 2-4 days
// in an intermediate city. This provides:
//
//   - Often cheaper total cost than a single long-haul ticket
//   - A chance to visit an extra city (tourism value)
//   - Shorter individual flights (less fatigue)
//   - More routing options when direct paths are blocked (e.g. Middle East)
//
// # How stopover cities are chosen
//
// Stopover cities must satisfy ALL of the following:
//
//  1. GEOGRAPHIC SENSE — The city should be roughly "on the way" between
//     origin and destination, or at least not a major detour. We don't
//     want Delhi → São Paulo → Toronto.
//
//  2. GOOD CONNECTIVITY — The city must have frequent flights to both the
//     origin and destination. Small regional airports won't work.
//
//  3. SAFE AIRSPACE — The city and its approach routes must not cross
//     blocked airspace (currently: Middle East).
//
//  4. TOURIST VALUE — Since the traveler will stay 2-4 days, the city
//     should be interesting to visit. This is subjective and can be
//     adjusted based on traveler preferences.
//
// TODO(iterate): Make stopover selection dynamic based on origin/destination.
// Currently these are hand-picked for the DEL → YYZ corridor avoiding
// Middle East routing. Future versions should:
//   - Compute great-circle waypoints between origin and destination
//   - Query an airport database for major hubs near those waypoints
//   - Filter by connectivity (min flights/day to both endpoints)
//   - Score by tourist value using LLM or a static rating
package multicity

import (
	"time"

	"booker/types"
)

// StopoverCity defines a candidate intermediate city for a multi-city halt.
type StopoverCity struct {
	// City is the human-readable city name.
	City string

	// Airport is the primary IATA airport code.
	Airport string

	// KiwiID is the Kiwi API location identifier (e.g. "City:hong_kong_hk").
	// Kiwi uses its own location format, not raw IATA codes.
	KiwiID string

	// Region helps group stopovers for diverse itinerary suggestions.
	Region string

	// MinStay is the minimum recommended stopover duration.
	MinStay time.Duration

	// MaxStay is the maximum recommended stopover duration.
	MaxStay time.Duration

	// Notes documents why this city is a good stopover for the current route.
	Notes string
}

// DELToYYZStopovers are the candidate stopover cities for Delhi → Toronto
// that avoid Middle East airspace entirely.
//
// Route geometry: DEL is at ~28°N, 77°E. YYZ is at ~43°N, 79°W.
// Eastbound via Asia-Pacific is the primary safe corridor.
// Northbound via Europe/Istanbul is the secondary corridor.
//
// TODO(iterate): Add more cities as airspace situation evolves.
// TODO(iterate): Score each city by current flight frequency + price trends.
var DELToYYZStopovers = []StopoverCity{
	// === EAST/SOUTHEAST ASIA — Primary corridor ===
	// These route eastbound from Delhi, then across the Pacific to Toronto.

	{
		City:    "Hong Kong",
		Airport: "HKG",
		KiwiID:  "Airport:HKG",
		Region:  "east_asia",
		MinStay: types.DefaultMinStopover,
		MaxStay: types.DefaultMaxStopover,
		Notes:   "Major Cathay Pacific hub. Excellent DEL-HKG and HKG-YYZ frequency. Great food, easy transit city.",
	},
	{
		City:    "Singapore",
		Airport: "SIN",
		KiwiID:  "Airport:SIN",
		Region:  "southeast_asia",
		MinStay: types.DefaultMinStopover,
		MaxStay: types.DefaultMaxStopover,
		Notes:   "Singapore Airlines hub. Strong DEL-SIN connectivity. SIN-YYZ may require connection. Clean, safe, great food.",
	},
	{
		City:    "Bangkok",
		Airport: "BKK",
		KiwiID:  "Airport:BKK",
		Region:  "southeast_asia",
		MinStay: types.DefaultMinStopover,
		MaxStay: types.DefaultMaxStopover,
		Notes:   "Thai Airways hub. Very cheap. DEL-BKK frequent, BKK-YYZ usually via Tokyo or Hong Kong.",
	},
	{
		City:    "Tokyo",
		Airport: "NRT",
		KiwiID:  "Airport:NRT",
		Region:  "east_asia",
		MinStay: types.DefaultMinStopover,
		MaxStay: types.DefaultMaxStopover,
		Notes:   "ANA/JAL hub. NRT-YYZ direct on Air Canada. Slightly north of great-circle but excellent connectivity.",
	},
	{
		City:    "Seoul",
		Airport: "ICN",
		KiwiID:  "Airport:ICN",
		Region:  "east_asia",
		MinStay: types.DefaultMinStopover,
		MaxStay: types.DefaultMaxStopover,
		Notes:   "Korean Air hub. ICN-YYZ direct on Korean Air and Air Canada. DEL-ICN on Korean Air/Air India.",
	},
	{
		City:    "Kuala Lumpur",
		Airport: "KUL",
		KiwiID:  "Airport:KUL",
		Region:  "southeast_asia",
		MinStay: types.DefaultMinStopover,
		MaxStay: types.DefaultMaxStopover,
		Notes:   "Malaysia Airlines / AirAsia hub. Cheap DEL-KUL. KUL-YYZ needs connection but affordable.",
	},

	// === EUROPE — Secondary corridor ===
	// These route westbound via Turkey or northern Europe, avoiding
	// Middle East airspace by going north of Iran.

	{
		City:    "Istanbul",
		Airport: "IST",
		KiwiID:  "Airport:IST",
		Region:  "europe",
		MinStay: types.DefaultMinStopover,
		MaxStay: types.DefaultMaxStopover,
		Notes:   "Turkish Airlines mega-hub. IST is OPEN (north of conflict zone). IST-YYZ direct. DEL-IST on Turkish. Strong option.",
	},
	{
		City:    "London",
		Airport: "LHR",
		KiwiID:  "Airport:LHR",
		Region:  "europe",
		MinStay: types.DefaultMinStopover,
		MaxStay: types.DefaultMaxStopover,
		Notes:   "BA hub. DEL-LHR on Air India/BA. LHR-YYZ very frequent. Visa may be needed.",
	},
	{
		City:    "Frankfurt",
		Airport: "FRA",
		KiwiID:  "Airport:FRA",
		Region:  "europe",
		MinStay: types.DefaultMinStopover,
		MaxStay: types.DefaultMaxStopover,
		Notes:   "Lufthansa hub. DEL-FRA on Lufthansa/Air India. FRA-YYZ direct. Schengen visa needed.",
	},
	{
		City:    "Paris",
		Airport: "CDG",
		KiwiID:  "Airport:CDG",
		Region:  "europe",
		MinStay: types.DefaultMinStopover,
		MaxStay: types.DefaultMaxStopover,
		Notes:   "Air France hub. DEL-CDG on Air France/Air India. CDG-YYZ direct. Schengen visa needed.",
	},
}

// StopoversForRoute returns the candidate stopover cities for a given
// origin-destination pair.
//
// TODO(iterate): Currently returns a hardcoded list for DEL→YYZ. Future
// versions should dynamically compute stopovers based on the route.
func StopoversForRoute(origin, destination string) []StopoverCity {
	// For now, we only have one route defined.
	// As we add more origin-destination pairs, this becomes a lookup.
	if origin == "DEL" && destination == "YYZ" {
		return DELToYYZStopovers
	}
	// Fallback: return all known stopovers.
	return DELToYYZStopovers
}
