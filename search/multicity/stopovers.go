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
		Region:  "east_asia",
		MinStay: types.DefaultMinStopover,
		MaxStay: types.DefaultMaxStopover,
		Notes:   "Major Cathay Pacific hub. Excellent DEL-HKG and HKG-YYZ frequency. Great food, easy transit city.",
	},
	{
		City:    "Singapore",
		Airport: "SIN",
		Region:  "southeast_asia",
		MinStay: types.DefaultMinStopover,
		MaxStay: types.DefaultMaxStopover,
		Notes:   "Singapore Airlines hub. Strong DEL-SIN connectivity. SIN-YYZ may require connection. Clean, safe, great food.",
	},
	{
		City:    "Bangkok",
		Airport: "BKK",
		Region:  "southeast_asia",
		MinStay: types.DefaultMinStopover,
		MaxStay: types.DefaultMaxStopover,
		Notes:   "Thai Airways hub. Very cheap. DEL-BKK frequent, BKK-YYZ usually via Tokyo or Hong Kong.",
	},
	{
		City:    "Tokyo",
		Airport: "NRT",
		Region:  "east_asia",
		MinStay: types.DefaultMinStopover,
		MaxStay: types.DefaultMaxStopover,
		Notes:   "ANA/JAL hub. NRT-YYZ direct on Air Canada. Slightly north of great-circle but excellent connectivity.",
	},
	{
		City:    "Seoul",
		Airport: "ICN",
		Region:  "east_asia",
		MinStay: types.DefaultMinStopover,
		MaxStay: types.DefaultMaxStopover,
		Notes:   "Korean Air hub. ICN-YYZ direct on Korean Air and Air Canada. DEL-ICN on Korean Air/Air India.",
	},
	{
		City:    "Kuala Lumpur",
		Airport: "KUL",
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
		Region:  "europe",
		MinStay: types.DefaultMinStopover,
		MaxStay: types.DefaultMaxStopover,
		Notes:   "Turkish Airlines mega-hub. IST is OPEN (north of conflict zone). IST-YYZ direct. DEL-IST on Turkish. Strong option.",
	},
	{
		City:    "London",
		Airport: "LHR",
		Region:  "europe",
		MinStay: types.DefaultMinStopover,
		MaxStay: types.DefaultMaxStopover,
		Notes:   "BA hub. DEL-LHR on Air India/BA. LHR-YYZ very frequent. Visa may be needed.",
	},
	{
		City:    "Frankfurt",
		Airport: "FRA",
		Region:  "europe",
		MinStay: types.DefaultMinStopover,
		MaxStay: types.DefaultMaxStopover,
		Notes:   "Lufthansa hub. DEL-FRA on Lufthansa/Air India. FRA-YYZ direct. Schengen visa needed.",
	},
	{
		City:    "Paris",
		Airport: "CDG",
		Region:  "europe",
		MinStay: types.DefaultMinStopover,
		MaxStay: types.DefaultMaxStopover,
		Notes:   "Air France hub. DEL-CDG on Air France/Air India. CDG-YYZ direct. Schengen visa needed.",
	},
}

// BOMToYYZStopovers are the candidate stopover cities for Mumbai → Toronto.
//
// Route geometry: BOM is at ~19°N, 73°E. YYZ is at ~43°N, 79°W.
// Similar corridor to DEL→YYZ but BOM has different hub connectivity.
var BOMToYYZStopovers = []StopoverCity{
	{
		City:    "Bangkok",
		Airport: "BKK",
		Region:  "southeast_asia",
		MinStay: types.DefaultMinStopover,
		MaxStay: types.DefaultMaxStopover,
		Notes:   "Thai Airways hub. BOM-BKK very frequent and cheap. BKK-YYZ via connection.",
	},
	{
		City:    "Singapore",
		Airport: "SIN",
		Region:  "southeast_asia",
		MinStay: types.DefaultMinStopover,
		MaxStay: types.DefaultMaxStopover,
		Notes:   "Singapore Airlines hub. Strong BOM-SIN frequency. Clean, safe city.",
	},
	{
		City:    "Hong Kong",
		Airport: "HKG",
		Region:  "east_asia",
		MinStay: types.DefaultMinStopover,
		MaxStay: types.DefaultMaxStopover,
		Notes:   "Cathay Pacific hub. BOM-HKG on Cathay/Air India. HKG-YYZ on Cathay.",
	},
	{
		City:    "Kuala Lumpur",
		Airport: "KUL",
		Region:  "southeast_asia",
		MinStay: types.DefaultMinStopover,
		MaxStay: types.DefaultMaxStopover,
		Notes:   "AirAsia/MAS hub. Very cheap BOM-KUL. KUL-YYZ needs connection.",
	},
	{
		City:    "Tokyo",
		Airport: "NRT",
		Region:  "east_asia",
		MinStay: types.DefaultMinStopover,
		MaxStay: types.DefaultMaxStopover,
		Notes:   "ANA/JAL hub. NRT-YYZ direct on Air Canada.",
	},
	{
		City:    "Istanbul",
		Airport: "IST",
		Region:  "europe",
		MinStay: types.DefaultMinStopover,
		MaxStay: types.DefaultMaxStopover,
		Notes:   "Turkish Airlines mega-hub. BOM-IST on Turkish. IST-YYZ direct.",
	},
	{
		City:    "London",
		Airport: "LHR",
		Region:  "europe",
		MinStay: types.DefaultMinStopover,
		MaxStay: types.DefaultMaxStopover,
		Notes:   "BA hub. BOM-LHR on BA/Air India/Virgin. LHR-YYZ very frequent.",
	},
}

// DELToYVRStopovers are the candidate stopover cities for Delhi → Vancouver.
//
// Route geometry: DEL is at ~28°N, 77°E. YVR is at ~49°N, 123°W.
// Pacific routing via East Asia is the primary corridor.
var DELToYVRStopovers = []StopoverCity{
	{
		City:    "Tokyo",
		Airport: "NRT",
		Region:  "east_asia",
		MinStay: types.DefaultMinStopover,
		MaxStay: types.DefaultMaxStopover,
		Notes:   "ANA/JAL hub. NRT-YVR direct on Air Canada/ANA. Natural Pacific waypoint.",
	},
	{
		City:    "Seoul",
		Airport: "ICN",
		Region:  "east_asia",
		MinStay: types.DefaultMinStopover,
		MaxStay: types.DefaultMaxStopover,
		Notes:   "Korean Air hub. ICN-YVR direct on Korean Air. DEL-ICN on Korean Air.",
	},
	{
		City:    "Hong Kong",
		Airport: "HKG",
		Region:  "east_asia",
		MinStay: types.DefaultMinStopover,
		MaxStay: types.DefaultMaxStopover,
		Notes:   "Cathay Pacific hub. HKG-YVR on Cathay. DEL-HKG frequent.",
	},
	{
		City:    "Bangkok",
		Airport: "BKK",
		Region:  "southeast_asia",
		MinStay: types.DefaultMinStopover,
		MaxStay: types.DefaultMaxStopover,
		Notes:   "Thai Airways hub. BKK-YVR via connection (NRT/ICN). DEL-BKK frequent and cheap.",
	},
	{
		City:    "Singapore",
		Airport: "SIN",
		Region:  "southeast_asia",
		MinStay: types.DefaultMinStopover,
		MaxStay: types.DefaultMaxStopover,
		Notes:   "SQ hub. SIN-YVR via connection. Strong DEL-SIN frequency.",
	},
	{
		City:    "Istanbul",
		Airport: "IST",
		Region:  "europe",
		MinStay: types.DefaultMinStopover,
		MaxStay: types.DefaultMaxStopover,
		Notes:   "Turkish Airlines mega-hub. IST-YVR via connection. DEL-IST on Turkish.",
	},
}

// DELToJFKStopovers are the candidate stopover cities for Delhi → New York JFK.
//
// Route geometry: DEL is at ~28°N, 77°E. JFK is at ~40°N, 73°W.
// High-demand corridor for the Indian diaspora to the US East Coast.
// Eastbound via Asia-Pacific and westbound via Europe are both viable.
var DELToJFKStopovers = []StopoverCity{
	// === EAST/SOUTHEAST ASIA — Primary corridor ===
	{
		City:    "Hong Kong",
		Airport: "HKG",
		Region:  "east_asia",
		MinStay: types.DefaultMinStopover,
		MaxStay: types.DefaultMaxStopover,
		Notes:   "Cathay Pacific hub. Excellent DEL-HKG frequency. HKG-JFK direct on Cathay.",
	},
	{
		City:    "Singapore",
		Airport: "SIN",
		Region:  "southeast_asia",
		MinStay: types.DefaultMinStopover,
		MaxStay: types.DefaultMaxStopover,
		Notes:   "SQ hub. Strong DEL-SIN connectivity. SIN-JFK direct on Singapore Airlines.",
	},
	{
		City:    "Bangkok",
		Airport: "BKK",
		Region:  "southeast_asia",
		MinStay: types.DefaultMinStopover,
		MaxStay: types.DefaultMaxStopover,
		Notes:   "Thai Airways hub. Cheap DEL-BKK. BKK-JFK via connection through NRT or ICN.",
	},
	{
		City:    "Tokyo",
		Airport: "NRT",
		Region:  "east_asia",
		MinStay: types.DefaultMinStopover,
		MaxStay: types.DefaultMaxStopover,
		Notes:   "ANA/JAL hub. NRT-JFK direct on multiple carriers. Strong Pacific gateway.",
	},
	{
		City:    "Seoul",
		Airport: "ICN",
		Region:  "east_asia",
		MinStay: types.DefaultMinStopover,
		MaxStay: types.DefaultMaxStopover,
		Notes:   "Korean Air hub. ICN-JFK direct on Korean Air. DEL-ICN on Korean Air/Air India.",
	},

	// === EUROPE — Secondary corridor ===
	{
		City:    "Istanbul",
		Airport: "IST",
		Region:  "europe",
		MinStay: types.DefaultMinStopover,
		MaxStay: types.DefaultMaxStopover,
		Notes:   "Turkish Airlines mega-hub. IST-JFK direct. DEL-IST on Turkish. Strong option.",
	},
	{
		City:    "London",
		Airport: "LHR",
		Region:  "europe",
		MinStay: types.DefaultMinStopover,
		MaxStay: types.DefaultMaxStopover,
		Notes:   "BA hub. DEL-LHR on Air India/BA. LHR-JFK very frequent.",
	},
	{
		City:    "Frankfurt",
		Airport: "FRA",
		Region:  "europe",
		MinStay: types.DefaultMinStopover,
		MaxStay: types.DefaultMaxStopover,
		Notes:   "Lufthansa hub. DEL-FRA on Lufthansa/Air India. FRA-JFK direct.",
	},
}

// BOMToJFKStopovers are the candidate stopover cities for Mumbai → New York JFK.
//
// Route geometry: BOM is at ~19°N, 73°E. JFK is at ~40°N, 73°W.
// High-demand corridor for the Indian diaspora to the US East Coast.
var BOMToJFKStopovers = []StopoverCity{
	{
		City:    "Bangkok",
		Airport: "BKK",
		Region:  "southeast_asia",
		MinStay: types.DefaultMinStopover,
		MaxStay: types.DefaultMaxStopover,
		Notes:   "Thai Airways hub. BOM-BKK very frequent and cheap. BKK-JFK via connection.",
	},
	{
		City:    "Singapore",
		Airport: "SIN",
		Region:  "southeast_asia",
		MinStay: types.DefaultMinStopover,
		MaxStay: types.DefaultMaxStopover,
		Notes:   "Singapore Airlines hub. Strong BOM-SIN frequency. SIN-JFK direct on SQ.",
	},
	{
		City:    "Hong Kong",
		Airport: "HKG",
		Region:  "east_asia",
		MinStay: types.DefaultMinStopover,
		MaxStay: types.DefaultMaxStopover,
		Notes:   "Cathay Pacific hub. BOM-HKG on Cathay/Air India. HKG-JFK direct on Cathay.",
	},
	{
		City:    "Tokyo",
		Airport: "NRT",
		Region:  "east_asia",
		MinStay: types.DefaultMinStopover,
		MaxStay: types.DefaultMaxStopover,
		Notes:   "ANA/JAL hub. NRT-JFK direct on multiple carriers.",
	},
	{
		City:    "Istanbul",
		Airport: "IST",
		Region:  "europe",
		MinStay: types.DefaultMinStopover,
		MaxStay: types.DefaultMaxStopover,
		Notes:   "Turkish Airlines mega-hub. BOM-IST on Turkish. IST-JFK direct.",
	},
	{
		City:    "London",
		Airport: "LHR",
		Region:  "europe",
		MinStay: types.DefaultMinStopover,
		MaxStay: types.DefaultMaxStopover,
		Notes:   "BA hub. BOM-LHR on BA/Air India/Virgin. LHR-JFK very frequent.",
	},
	{
		City:    "Frankfurt",
		Airport: "FRA",
		Region:  "europe",
		MinStay: types.DefaultMinStopover,
		MaxStay: types.DefaultMaxStopover,
		Notes:   "Lufthansa hub. BOM-FRA on Lufthansa. FRA-JFK direct.",
	},
}

// DELToLHRStopovers are the candidate stopover cities for Delhi → London Heathrow.
//
// Route geometry: DEL is at ~28°N, 77°E. LHR is at ~51°N, 0°W.
// Southeast Asia eastbound corridor and Istanbul westbound are the
// primary safe corridors avoiding Middle East airspace.
var DELToLHRStopovers = []StopoverCity{
	// === SOUTHEAST ASIA — Primary corridor ===
	{
		City:    "Bangkok",
		Airport: "BKK",
		Region:  "southeast_asia",
		MinStay: types.DefaultMinStopover,
		MaxStay: types.DefaultMaxStopover,
		Notes:   "Thai Airways hub. DEL-BKK frequent and cheap. BKK-LHR direct on Thai/BA. Temples, street food, nightlife.",
	},
	{
		City:    "Singapore",
		Airport: "SIN",
		Region:  "southeast_asia",
		MinStay: types.DefaultMinStopover,
		MaxStay: types.DefaultMaxStopover,
		Notes:   "SQ hub. SIN-LHR direct on Singapore Airlines. Strong DEL-SIN frequency. Clean, safe, great food.",
	},
	{
		City:    "Kuala Lumpur",
		Airport: "KUL",
		Region:  "southeast_asia",
		MinStay: types.DefaultMinStopover,
		MaxStay: types.DefaultMaxStopover,
		Notes:   "Malaysia Airlines hub. KUL-LHR direct on MAS. Cheap DEL-KUL. Petronas Towers, Batu Caves.",
	},
	{
		City:    "Hong Kong",
		Airport: "HKG",
		Region:  "east_asia",
		MinStay: types.DefaultMinStopover,
		MaxStay: types.DefaultMaxStopover,
		Notes:   "Cathay Pacific hub. HKG-LHR direct on Cathay/BA. DEL-HKG frequent. Victoria Peak, harbour, dim sum.",
	},
	{
		City:    "Colombo",
		Airport: "CMB",
		Region:  "south_asia",
		MinStay: types.DefaultMinStopover,
		MaxStay: types.DefaultMaxStopover,
		Notes:   "SriLankan Airlines hub. CMB-LHR direct on SriLankan. DEL-CMB frequent. Beaches, tea country, temples.",
	},

	// === EUROPE — Secondary corridor ===
	{
		City:    "Istanbul",
		Airport: "IST",
		Region:  "europe",
		MinStay: types.DefaultMinStopover,
		MaxStay: types.DefaultMaxStopover,
		Notes:   "Turkish Airlines mega-hub. IST-LHR very frequent. DEL-IST on Turkish. Bosphorus, bazaars, history.",
	},
}

// BOMToLHRStopovers are the candidate stopover cities for Mumbai → London Heathrow.
//
// Route geometry: BOM is at ~19°N, 73°E. LHR is at ~51°N, 0°W.
// Similar corridor to DEL→LHR but BOM has different hub connectivity.
var BOMToLHRStopovers = []StopoverCity{
	{
		City:    "Bangkok",
		Airport: "BKK",
		Region:  "southeast_asia",
		MinStay: types.DefaultMinStopover,
		MaxStay: types.DefaultMaxStopover,
		Notes:   "Thai Airways hub. BOM-BKK very frequent and cheap. BKK-LHR direct on Thai/BA.",
	},
	{
		City:    "Singapore",
		Airport: "SIN",
		Region:  "southeast_asia",
		MinStay: types.DefaultMinStopover,
		MaxStay: types.DefaultMaxStopover,
		Notes:   "SQ hub. SIN-LHR direct on Singapore Airlines. Strong BOM-SIN frequency.",
	},
	{
		City:    "Kuala Lumpur",
		Airport: "KUL",
		Region:  "southeast_asia",
		MinStay: types.DefaultMinStopover,
		MaxStay: types.DefaultMaxStopover,
		Notes:   "Malaysia Airlines hub. KUL-LHR direct on MAS. Cheap BOM-KUL.",
	},
	{
		City:    "Hong Kong",
		Airport: "HKG",
		Region:  "east_asia",
		MinStay: types.DefaultMinStopover,
		MaxStay: types.DefaultMaxStopover,
		Notes:   "Cathay Pacific hub. HKG-LHR direct on Cathay/BA. BOM-HKG on Cathay/Air India.",
	},
	{
		City:    "Colombo",
		Airport: "CMB",
		Region:  "south_asia",
		MinStay: types.DefaultMinStopover,
		MaxStay: types.DefaultMaxStopover,
		Notes:   "SriLankan Airlines hub. CMB-LHR direct on SriLankan. BOM-CMB short hop. Beaches, wildlife.",
	},
	{
		City:    "Istanbul",
		Airport: "IST",
		Region:  "europe",
		MinStay: types.DefaultMinStopover,
		MaxStay: types.DefaultMaxStopover,
		Notes:   "Turkish Airlines mega-hub. IST-LHR very frequent. BOM-IST on Turkish.",
	},
}

// routeKey creates a lookup key for origin-destination pairs.
func routeKey(origin, destination string) string {
	return origin + "→" + destination
}

// stopoversMap maps origin→destination to their stopover lists.
var stopoversMap = map[string][]StopoverCity{
	routeKey("DEL", "YYZ"): DELToYYZStopovers,
	routeKey("BOM", "YYZ"): BOMToYYZStopovers,
	routeKey("DEL", "YVR"): DELToYVRStopovers,
	routeKey("DEL", "JFK"): DELToJFKStopovers,
	routeKey("BOM", "JFK"): BOMToJFKStopovers,
	routeKey("DEL", "LHR"): DELToLHRStopovers,
	routeKey("BOM", "LHR"): BOMToLHRStopovers,
}

// GlobalFallbackHubs are well-connected hub airports used as stopover
// candidates when no route-specific stopovers are configured.
var GlobalFallbackHubs = []StopoverCity{
	{City: "Istanbul", Airport: "IST", Region: "europe",
		MinStay: types.DefaultMinStopover, MaxStay: types.DefaultMaxStopover,
		Notes: "Turkish Airlines mega-hub with global connectivity."},
	{City: "Singapore", Airport: "SIN", Region: "southeast_asia",
		MinStay: types.DefaultMinStopover, MaxStay: types.DefaultMaxStopover,
		Notes: "Singapore Airlines hub, major Asia-Pacific crossroads."},
	{City: "Hong Kong", Airport: "HKG", Region: "east_asia",
		MinStay: types.DefaultMinStopover, MaxStay: types.DefaultMaxStopover,
		Notes: "Cathay Pacific hub with extensive long-haul network."},
	{City: "Tokyo", Airport: "NRT", Region: "east_asia",
		MinStay: types.DefaultMinStopover, MaxStay: types.DefaultMaxStopover,
		Notes: "ANA/JAL hub, key Pacific gateway."},
	{City: "London", Airport: "LHR", Region: "europe",
		MinStay: types.DefaultMinStopover, MaxStay: types.DefaultMaxStopover,
		Notes: "BA hub, largest European long-haul airport."},
	{City: "Paris", Airport: "CDG", Region: "europe",
		MinStay: types.DefaultMinStopover, MaxStay: types.DefaultMaxStopover,
		Notes: "Air France hub with global reach."},
	{City: "Seoul", Airport: "ICN", Region: "east_asia",
		MinStay: types.DefaultMinStopover, MaxStay: types.DefaultMaxStopover,
		Notes: "Korean Air hub, strong Americas and Asia connectivity."},
	{City: "Bangkok", Airport: "BKK", Region: "southeast_asia",
		MinStay: types.DefaultMinStopover, MaxStay: types.DefaultMaxStopover,
		Notes: "Thai Airways hub, affordable Southeast Asia gateway."},
}

// StopoversForRoute returns the candidate stopover cities for a given
// origin-destination pair. When no route-specific stopovers exist, it
// returns the global fallback hubs with origin/destination airports excluded.
func StopoversForRoute(origin, destination string) []StopoverCity {
	if route := stopoversMap[routeKey(origin, destination)]; route != nil {
		return route
	}
	var hubs []StopoverCity
	for _, h := range GlobalFallbackHubs {
		if h.Airport != origin && h.Airport != destination {
			hubs = append(hubs, h)
		}
	}
	return hubs
}
