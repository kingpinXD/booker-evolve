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

// DELToSFOStopovers are the candidate stopover cities for Delhi → San Francisco.
//
// Route geometry: DEL is at ~28°N, 77°E. SFO is at ~37°N, 122°W.
// High-demand corridor for the Indian tech diaspora to the US West Coast.
// Pacific routing via East Asia is the primary safe corridor.
var DELToSFOStopovers = []StopoverCity{
	// === EAST ASIA — Primary corridor ===
	{
		City:    "Tokyo",
		Airport: "NRT",
		Region:  "east_asia",
		MinStay: types.DefaultMinStopover,
		MaxStay: types.DefaultMaxStopover,
		Notes:   "ANA/JAL hub. NRT-SFO direct on ANA/JAL/United. DEL-NRT on JAL/ANA. Natural Pacific gateway.",
	},
	{
		City:    "Seoul",
		Airport: "ICN",
		Region:  "east_asia",
		MinStay: types.DefaultMinStopover,
		MaxStay: types.DefaultMaxStopover,
		Notes:   "Korean Air hub. ICN-SFO direct on Korean Air/Asiana/United. DEL-ICN on Korean Air.",
	},
	{
		City:    "Hong Kong",
		Airport: "HKG",
		Region:  "east_asia",
		MinStay: types.DefaultMinStopover,
		MaxStay: types.DefaultMaxStopover,
		Notes:   "Cathay Pacific hub. HKG-SFO direct on Cathay/United. DEL-HKG frequent.",
	},

	// === SOUTHEAST ASIA ===
	{
		City:    "Bangkok",
		Airport: "BKK",
		Region:  "southeast_asia",
		MinStay: types.DefaultMinStopover,
		MaxStay: types.DefaultMaxStopover,
		Notes:   "Thai Airways hub. BKK-SFO via NRT/ICN connection. DEL-BKK frequent and cheap.",
	},
	{
		City:    "Singapore",
		Airport: "SIN",
		Region:  "southeast_asia",
		MinStay: types.DefaultMinStopover,
		MaxStay: types.DefaultMaxStopover,
		Notes:   "SQ hub. SIN-SFO direct on Singapore Airlines/United. Strong DEL-SIN frequency.",
	},

	// === EUROPE — Secondary corridor ===
	{
		City:    "Istanbul",
		Airport: "IST",
		Region:  "europe",
		MinStay: types.DefaultMinStopover,
		MaxStay: types.DefaultMaxStopover,
		Notes:   "Turkish Airlines mega-hub. IST-SFO direct on Turkish. DEL-IST on Turkish. Westbound option.",
	},
}

// BOMToSFOStopovers are the candidate stopover cities for Mumbai → San Francisco.
//
// Route geometry: BOM is at ~19°N, 73°E. SFO is at ~37°N, 122°W.
// High-demand corridor for the Indian tech diaspora to the US West Coast.
var BOMToSFOStopovers = []StopoverCity{
	{
		City:    "Tokyo",
		Airport: "NRT",
		Region:  "east_asia",
		MinStay: types.DefaultMinStopover,
		MaxStay: types.DefaultMaxStopover,
		Notes:   "ANA/JAL hub. NRT-SFO direct on ANA/JAL/United. BOM-NRT on ANA.",
	},
	{
		City:    "Hong Kong",
		Airport: "HKG",
		Region:  "east_asia",
		MinStay: types.DefaultMinStopover,
		MaxStay: types.DefaultMaxStopover,
		Notes:   "Cathay Pacific hub. HKG-SFO direct on Cathay/United. BOM-HKG on Cathay/Air India.",
	},
	{
		City:    "Bangkok",
		Airport: "BKK",
		Region:  "southeast_asia",
		MinStay: types.DefaultMinStopover,
		MaxStay: types.DefaultMaxStopover,
		Notes:   "Thai Airways hub. BOM-BKK very frequent and cheap. BKK-SFO via connection.",
	},
	{
		City:    "Singapore",
		Airport: "SIN",
		Region:  "southeast_asia",
		MinStay: types.DefaultMinStopover,
		MaxStay: types.DefaultMaxStopover,
		Notes:   "SQ hub. SIN-SFO direct on Singapore Airlines/United. Strong BOM-SIN frequency.",
	},
	{
		City:    "Istanbul",
		Airport: "IST",
		Region:  "europe",
		MinStay: types.DefaultMinStopover,
		MaxStay: types.DefaultMaxStopover,
		Notes:   "Turkish Airlines mega-hub. IST-SFO direct on Turkish. BOM-IST on Turkish.",
	},
}

// DELToSYDStopovers are the candidate stopover cities for Delhi → Sydney.
//
// Route geometry: DEL is at ~28°N, 77°E. SYD is at ~33°S, 151°E.
// Southeast Asia is the primary corridor — these cities are directly
// on the route and have strong connectivity to both India and Australia.
var DELToSYDStopovers = []StopoverCity{
	// === SOUTHEAST ASIA — Primary corridor ===
	{
		City:    "Singapore",
		Airport: "SIN",
		Region:  "southeast_asia",
		MinStay: types.DefaultMinStopover,
		MaxStay: types.DefaultMaxStopover,
		Notes:   "SQ hub. SIN-SYD very frequent on SQ/Qantas. Strong DEL-SIN frequency. Clean, safe, great food.",
	},
	{
		City:    "Bangkok",
		Airport: "BKK",
		Region:  "southeast_asia",
		MinStay: types.DefaultMinStopover,
		MaxStay: types.DefaultMaxStopover,
		Notes:   "Thai Airways hub. BKK-SYD direct on Thai. DEL-BKK frequent and cheap. Temples, street food.",
	},
	{
		City:    "Kuala Lumpur",
		Airport: "KUL",
		Region:  "southeast_asia",
		MinStay: types.DefaultMinStopover,
		MaxStay: types.DefaultMaxStopover,
		Notes:   "Malaysia Airlines/AirAsia hub. KUL-SYD direct on MAS. Cheap DEL-KUL. Petronas Towers, Batu Caves.",
	},

	// === EAST ASIA ===
	{
		City:    "Hong Kong",
		Airport: "HKG",
		Region:  "east_asia",
		MinStay: types.DefaultMinStopover,
		MaxStay: types.DefaultMaxStopover,
		Notes:   "Cathay Pacific hub. HKG-SYD direct on Cathay/Qantas. DEL-HKG frequent. Victoria Peak, dim sum.",
	},
	{
		City:    "Tokyo",
		Airport: "NRT",
		Region:  "east_asia",
		MinStay: types.DefaultMinStopover,
		MaxStay: types.DefaultMaxStopover,
		Notes:   "ANA/JAL hub. NRT-SYD direct on ANA/JAL/Qantas. DEL-NRT on JAL. Slight detour but great city.",
	},
	{
		City:    "Osaka",
		Airport: "KIX",
		Region:  "east_asia",
		MinStay: types.DefaultMinStopover,
		MaxStay: types.DefaultMaxStopover,
		Notes:   "Peach/JAL hub. KIX-SYD on Jetstar/JAL. Less crowded than Tokyo. Dotonbori, Osaka Castle.",
	},
}

// BOMToSYDStopovers are the candidate stopover cities for Mumbai → Sydney.
//
// Route geometry: BOM is at ~19°N, 73°E. SYD is at ~33°S, 151°E.
// Southeast Asia is the natural corridor with strong BOM connectivity.
var BOMToSYDStopovers = []StopoverCity{
	{
		City:    "Singapore",
		Airport: "SIN",
		Region:  "southeast_asia",
		MinStay: types.DefaultMinStopover,
		MaxStay: types.DefaultMaxStopover,
		Notes:   "SQ hub. SIN-SYD very frequent on SQ/Qantas. Strong BOM-SIN frequency.",
	},
	{
		City:    "Bangkok",
		Airport: "BKK",
		Region:  "southeast_asia",
		MinStay: types.DefaultMinStopover,
		MaxStay: types.DefaultMaxStopover,
		Notes:   "Thai Airways hub. BOM-BKK very frequent and cheap. BKK-SYD direct on Thai.",
	},
	{
		City:    "Kuala Lumpur",
		Airport: "KUL",
		Region:  "southeast_asia",
		MinStay: types.DefaultMinStopover,
		MaxStay: types.DefaultMaxStopover,
		Notes:   "Malaysia Airlines/AirAsia hub. KUL-SYD direct on MAS. Cheap BOM-KUL.",
	},
	{
		City:    "Hong Kong",
		Airport: "HKG",
		Region:  "east_asia",
		MinStay: types.DefaultMinStopover,
		MaxStay: types.DefaultMaxStopover,
		Notes:   "Cathay Pacific hub. HKG-SYD direct on Cathay/Qantas. BOM-HKG on Cathay/Air India.",
	},
	{
		City:    "Tokyo",
		Airport: "NRT",
		Region:  "east_asia",
		MinStay: types.DefaultMinStopover,
		MaxStay: types.DefaultMaxStopover,
		Notes:   "ANA/JAL hub. NRT-SYD direct on ANA/JAL/Qantas. BOM-NRT on ANA.",
	},
}

// DELToFRAStopovers are the candidate stopover cities for Delhi → Frankfurt.
//
// Route geometry: DEL is at ~28°N, 77°E. FRA is at ~50°N, 8°E.
// Gulf carrier hubs are the primary corridor — strong connectivity
// to both India and Frankfurt via ME3 airlines and Lufthansa partners.
var DELToFRAStopovers = []StopoverCity{
	// === GULF — Primary corridor ===
	{
		City:    "Doha",
		Airport: "DOH",
		Region:  "gulf",
		MinStay: types.DefaultMinStopover,
		MaxStay: types.DefaultMaxStopover,
		Notes:   "Qatar Airways hub. DEL-DOH very frequent on QR. DOH-FRA direct on QR/Lufthansa. Museum of Islamic Art, desert safaris.",
	},
	{
		City:    "Abu Dhabi",
		Airport: "AUH",
		Region:  "gulf",
		MinStay: types.DefaultMinStopover,
		MaxStay: types.DefaultMaxStopover,
		Notes:   "Etihad hub. DEL-AUH on Etihad/Air India. AUH-FRA direct on Etihad. Louvre Abu Dhabi, Sheikh Zayed Mosque.",
	},
	{
		City:    "Dubai",
		Airport: "DXB",
		Region:  "gulf",
		MinStay: types.DefaultMinStopover,
		MaxStay: types.DefaultMaxStopover,
		Notes:   "Emirates mega-hub. DEL-DXB very frequent on Emirates/AI. DXB-FRA direct on Emirates/Lufthansa.",
	},
	{
		City:    "Istanbul",
		Airport: "IST",
		Region:  "europe",
		MinStay: types.DefaultMinStopover,
		MaxStay: types.DefaultMaxStopover,
		Notes:   "Turkish Airlines mega-hub. DEL-IST on Turkish. IST-FRA direct on Turkish/Lufthansa. Bosphorus, bazaars.",
	},
	{
		City:    "Bahrain",
		Airport: "BAH",
		Region:  "gulf",
		MinStay: types.DefaultMinStopover,
		MaxStay: types.DefaultMaxStopover,
		Notes:   "Gulf Air hub. DEL-BAH on Gulf Air. BAH-FRA direct on Gulf Air. Compact, walkable, historic sites.",
	},
	{
		City:    "Kuwait City",
		Airport: "KWI",
		Region:  "gulf",
		MinStay: types.DefaultMinStopover,
		MaxStay: types.DefaultMaxStopover,
		Notes:   "Kuwait Airways hub. DEL-KWI on Kuwait Airways/AI. KWI-FRA direct on Kuwait Airways. Kuwait Towers, souks.",
	},
}

// BOMToFRAStopovers are the candidate stopover cities for Mumbai → Frankfurt.
//
// Route geometry: BOM is at ~19°N, 73°E. FRA is at ~50°N, 8°E.
// Gulf carrier hubs are the primary corridor with strong BOM connectivity.
var BOMToFRAStopovers = []StopoverCity{
	{
		City:    "Doha",
		Airport: "DOH",
		Region:  "gulf",
		MinStay: types.DefaultMinStopover,
		MaxStay: types.DefaultMaxStopover,
		Notes:   "Qatar Airways hub. BOM-DOH very frequent on QR. DOH-FRA direct on QR/Lufthansa.",
	},
	{
		City:    "Abu Dhabi",
		Airport: "AUH",
		Region:  "gulf",
		MinStay: types.DefaultMinStopover,
		MaxStay: types.DefaultMaxStopover,
		Notes:   "Etihad hub. BOM-AUH on Etihad/Air India. AUH-FRA direct on Etihad.",
	},
	{
		City:    "Dubai",
		Airport: "DXB",
		Region:  "gulf",
		MinStay: types.DefaultMinStopover,
		MaxStay: types.DefaultMaxStopover,
		Notes:   "Emirates mega-hub. BOM-DXB very frequent on Emirates/AI. DXB-FRA direct on Emirates/Lufthansa.",
	},
	{
		City:    "Istanbul",
		Airport: "IST",
		Region:  "europe",
		MinStay: types.DefaultMinStopover,
		MaxStay: types.DefaultMaxStopover,
		Notes:   "Turkish Airlines mega-hub. BOM-IST on Turkish. IST-FRA direct on Turkish/Lufthansa.",
	},
	{
		City:    "Bahrain",
		Airport: "BAH",
		Region:  "gulf",
		MinStay: types.DefaultMinStopover,
		MaxStay: types.DefaultMaxStopover,
		Notes:   "Gulf Air hub. BOM-BAH on Gulf Air. BAH-FRA direct on Gulf Air.",
	},
}

// DELToBKKStopovers are the candidate stopover cities for Delhi → Bangkok.
//
// Route geometry: DEL is at ~28°N, 77°E. BKK is at ~13°N, 100°E.
// Gulf carrier hubs and South Asian crossroads are the primary corridor.
var DELToBKKStopovers = []StopoverCity{
	// === GULF — Primary corridor ===
	{
		City:    "Doha",
		Airport: "DOH",
		Region:  "gulf",
		MinStay: types.DefaultMinStopover,
		MaxStay: types.DefaultMaxStopover,
		Notes:   "Qatar Airways hub. DEL-DOH very frequent on QR. DOH-BKK direct on QR. Desert safaris, Museum of Islamic Art.",
	},
	{
		City:    "Abu Dhabi",
		Airport: "AUH",
		Region:  "gulf",
		MinStay: types.DefaultMinStopover,
		MaxStay: types.DefaultMaxStopover,
		Notes:   "Etihad hub. DEL-AUH on Etihad/Air India. AUH-BKK direct on Etihad. Louvre Abu Dhabi, Sheikh Zayed Mosque.",
	},
	{
		City:    "Dubai",
		Airport: "DXB",
		Region:  "gulf",
		MinStay: types.DefaultMinStopover,
		MaxStay: types.DefaultMaxStopover,
		Notes:   "Emirates mega-hub. DEL-DXB very frequent on Emirates/AI. DXB-BKK direct on Emirates.",
	},

	// === SOUTHEAST ASIA ===
	{
		City:    "Singapore",
		Airport: "SIN",
		Region:  "southeast_asia",
		MinStay: types.DefaultMinStopover,
		MaxStay: types.DefaultMaxStopover,
		Notes:   "SQ hub. DEL-SIN frequent. SIN-BKK very frequent on SQ/Thai. Clean, safe, great food.",
	},
	{
		City:    "Kuala Lumpur",
		Airport: "KUL",
		Region:  "southeast_asia",
		MinStay: types.DefaultMinStopover,
		MaxStay: types.DefaultMaxStopover,
		Notes:   "Malaysia Airlines/AirAsia hub. DEL-KUL on AirAsia/MAS. KUL-BKK very frequent. Petronas Towers, Batu Caves.",
	},

	// === SOUTH ASIA ===
	{
		City:    "Kolkata",
		Airport: "CCU",
		Region:  "south_asia",
		MinStay: types.DefaultMinStopover,
		MaxStay: types.DefaultMaxStopover,
		Notes:   "Eastern India gateway. DEL-CCU very frequent on AI/IndiGo. CCU-BKK direct on Thai/IndiGo. Victoria Memorial, street food.",
	},
}

// BOMToBKKStopovers are the candidate stopover cities for Mumbai → Bangkok.
//
// Route geometry: BOM is at ~19°N, 73°E. BKK is at ~13°N, 100°E.
// Gulf carrier hubs and Southeast Asian crossroads are the primary corridor.
var BOMToBKKStopovers = []StopoverCity{
	{
		City:    "Doha",
		Airport: "DOH",
		Region:  "gulf",
		MinStay: types.DefaultMinStopover,
		MaxStay: types.DefaultMaxStopover,
		Notes:   "Qatar Airways hub. BOM-DOH very frequent on QR. DOH-BKK direct on QR.",
	},
	{
		City:    "Abu Dhabi",
		Airport: "AUH",
		Region:  "gulf",
		MinStay: types.DefaultMinStopover,
		MaxStay: types.DefaultMaxStopover,
		Notes:   "Etihad hub. BOM-AUH on Etihad/Air India. AUH-BKK direct on Etihad.",
	},
	{
		City:    "Dubai",
		Airport: "DXB",
		Region:  "gulf",
		MinStay: types.DefaultMinStopover,
		MaxStay: types.DefaultMaxStopover,
		Notes:   "Emirates mega-hub. BOM-DXB very frequent on Emirates/AI. DXB-BKK direct on Emirates.",
	},
	{
		City:    "Singapore",
		Airport: "SIN",
		Region:  "southeast_asia",
		MinStay: types.DefaultMinStopover,
		MaxStay: types.DefaultMaxStopover,
		Notes:   "SQ hub. BOM-SIN frequent. SIN-BKK very frequent on SQ/Thai.",
	},
	{
		City:    "Kuala Lumpur",
		Airport: "KUL",
		Region:  "southeast_asia",
		MinStay: types.DefaultMinStopover,
		MaxStay: types.DefaultMaxStopover,
		Notes:   "Malaysia Airlines/AirAsia hub. BOM-KUL on AirAsia/MAS. KUL-BKK very frequent.",
	},
}

// DELToNRTStopovers are the candidate stopover cities for Delhi → Tokyo Narita.
//
// Route geometry: DEL is at ~28°N, 77°E. NRT is at ~35°N, 140°E.
// Southeast and East Asia corridor — cities with strong connectivity
// to both India and Japan.
var DELToNRTStopovers = []StopoverCity{
	// === SOUTHEAST ASIA ===
	{
		City:    "Bangkok",
		Airport: "BKK",
		Region:  "southeast_asia",
		MinStay: types.DefaultMinStopover,
		MaxStay: types.DefaultMaxStopover,
		Notes:   "Thai Airways hub. DEL-BKK frequent and cheap. BKK-NRT direct on Thai/ANA. Temples, street food.",
	},
	{
		City:    "Singapore",
		Airport: "SIN",
		Region:  "southeast_asia",
		MinStay: types.DefaultMinStopover,
		MaxStay: types.DefaultMaxStopover,
		Notes:   "SQ hub. DEL-SIN frequent. SIN-NRT direct on SQ/ANA/JAL. Clean, safe, great food.",
	},
	{
		City:    "Kuala Lumpur",
		Airport: "KUL",
		Region:  "southeast_asia",
		MinStay: types.DefaultMinStopover,
		MaxStay: types.DefaultMaxStopover,
		Notes:   "Malaysia Airlines/AirAsia hub. DEL-KUL on AirAsia/MAS. KUL-NRT direct on MAS/AirAsia X.",
	},

	// === EAST ASIA ===
	{
		City:    "Hong Kong",
		Airport: "HKG",
		Region:  "east_asia",
		MinStay: types.DefaultMinStopover,
		MaxStay: types.DefaultMaxStopover,
		Notes:   "Cathay Pacific hub. DEL-HKG frequent. HKG-NRT direct on Cathay/ANA/JAL. Victoria Peak, dim sum.",
	},
	{
		City:    "Taipei",
		Airport: "TPE",
		Region:  "east_asia",
		MinStay: types.DefaultMinStopover,
		MaxStay: types.DefaultMaxStopover,
		Notes:   "EVA Air/China Airlines hub. TPE-NRT very frequent. DEL-TPE on EVA/China Airlines. Night markets, temples.",
	},
	{
		City:    "Seoul",
		Airport: "ICN",
		Region:  "east_asia",
		MinStay: types.DefaultMinStopover,
		MaxStay: types.DefaultMaxStopover,
		Notes:   "Korean Air hub. ICN-NRT very frequent. DEL-ICN on Korean Air/Air India. Palaces, street food.",
	},
}

// BOMToNRTStopovers are the candidate stopover cities for Mumbai → Tokyo Narita.
//
// Route geometry: BOM is at ~19°N, 73°E. NRT is at ~35°N, 140°E.
// Similar corridor to DEL→NRT but BOM has different hub connectivity.
var BOMToNRTStopovers = []StopoverCity{
	{
		City:    "Bangkok",
		Airport: "BKK",
		Region:  "southeast_asia",
		MinStay: types.DefaultMinStopover,
		MaxStay: types.DefaultMaxStopover,
		Notes:   "Thai Airways hub. BOM-BKK very frequent and cheap. BKK-NRT direct on Thai/ANA.",
	},
	{
		City:    "Singapore",
		Airport: "SIN",
		Region:  "southeast_asia",
		MinStay: types.DefaultMinStopover,
		MaxStay: types.DefaultMaxStopover,
		Notes:   "SQ hub. BOM-SIN frequent. SIN-NRT direct on SQ/ANA/JAL.",
	},
	{
		City:    "Hong Kong",
		Airport: "HKG",
		Region:  "east_asia",
		MinStay: types.DefaultMinStopover,
		MaxStay: types.DefaultMaxStopover,
		Notes:   "Cathay Pacific hub. BOM-HKG on Cathay/Air India. HKG-NRT direct on Cathay/ANA/JAL.",
	},
	{
		City:    "Taipei",
		Airport: "TPE",
		Region:  "east_asia",
		MinStay: types.DefaultMinStopover,
		MaxStay: types.DefaultMaxStopover,
		Notes:   "EVA Air/China Airlines hub. TPE-NRT very frequent. BOM-TPE on EVA/China Airlines.",
	},
	{
		City:    "Seoul",
		Airport: "ICN",
		Region:  "east_asia",
		MinStay: types.DefaultMinStopover,
		MaxStay: types.DefaultMaxStopover,
		Notes:   "Korean Air hub. ICN-NRT very frequent. BOM-ICN on Korean Air.",
	},
}

// DELToMELStopovers are the candidate stopover cities for Delhi → Melbourne.
//
// Route geometry: DEL is at ~28°N, 77°E. MEL is at ~37°S, 144°E.
// Southeast Asia is the primary corridor — same hubs as the SYD route
// minus Osaka (weak KIX-MEL connectivity).
var DELToMELStopovers = []StopoverCity{
	// === SOUTHEAST ASIA — Primary corridor ===
	{
		City:    "Singapore",
		Airport: "SIN",
		Region:  "southeast_asia",
		MinStay: types.DefaultMinStopover,
		MaxStay: types.DefaultMaxStopover,
		Notes:   "SQ hub. SIN-MEL very frequent on SQ/Qantas. Strong DEL-SIN frequency. Clean, safe, great food.",
	},
	{
		City:    "Bangkok",
		Airport: "BKK",
		Region:  "southeast_asia",
		MinStay: types.DefaultMinStopover,
		MaxStay: types.DefaultMaxStopover,
		Notes:   "Thai Airways hub. BKK-MEL direct on Thai. DEL-BKK frequent and cheap. Temples, street food.",
	},
	{
		City:    "Kuala Lumpur",
		Airport: "KUL",
		Region:  "southeast_asia",
		MinStay: types.DefaultMinStopover,
		MaxStay: types.DefaultMaxStopover,
		Notes:   "Malaysia Airlines/AirAsia hub. KUL-MEL direct on MAS/AirAsia X. Cheap DEL-KUL.",
	},

	// === EAST ASIA ===
	{
		City:    "Hong Kong",
		Airport: "HKG",
		Region:  "east_asia",
		MinStay: types.DefaultMinStopover,
		MaxStay: types.DefaultMaxStopover,
		Notes:   "Cathay Pacific hub. HKG-MEL direct on Cathay/Qantas. DEL-HKG frequent.",
	},
	{
		City:    "Tokyo",
		Airport: "NRT",
		Region:  "east_asia",
		MinStay: types.DefaultMinStopover,
		MaxStay: types.DefaultMaxStopover,
		Notes:   "ANA/JAL hub. NRT-MEL direct on JAL/Qantas. DEL-NRT on JAL. Slight detour but great city.",
	},
}

// BOMToMELStopovers are the candidate stopover cities for Mumbai → Melbourne.
//
// Route geometry: BOM is at ~19°N, 73°E. MEL is at ~37°S, 144°E.
// Southeast Asia is the natural corridor with strong BOM connectivity.
var BOMToMELStopovers = []StopoverCity{
	{
		City:    "Singapore",
		Airport: "SIN",
		Region:  "southeast_asia",
		MinStay: types.DefaultMinStopover,
		MaxStay: types.DefaultMaxStopover,
		Notes:   "SQ hub. SIN-MEL very frequent on SQ/Qantas. Strong BOM-SIN frequency.",
	},
	{
		City:    "Bangkok",
		Airport: "BKK",
		Region:  "southeast_asia",
		MinStay: types.DefaultMinStopover,
		MaxStay: types.DefaultMaxStopover,
		Notes:   "Thai Airways hub. BOM-BKK very frequent and cheap. BKK-MEL direct on Thai.",
	},
	{
		City:    "Kuala Lumpur",
		Airport: "KUL",
		Region:  "southeast_asia",
		MinStay: types.DefaultMinStopover,
		MaxStay: types.DefaultMaxStopover,
		Notes:   "Malaysia Airlines/AirAsia hub. KUL-MEL direct on MAS/AirAsia X. Cheap BOM-KUL.",
	},
	{
		City:    "Hong Kong",
		Airport: "HKG",
		Region:  "east_asia",
		MinStay: types.DefaultMinStopover,
		MaxStay: types.DefaultMaxStopover,
		Notes:   "Cathay Pacific hub. HKG-MEL direct on Cathay/Qantas. BOM-HKG on Cathay/Air India.",
	},
	{
		City:    "Tokyo",
		Airport: "NRT",
		Region:  "east_asia",
		MinStay: types.DefaultMinStopover,
		MaxStay: types.DefaultMaxStopover,
		Notes:   "ANA/JAL hub. NRT-MEL direct on JAL/Qantas. BOM-NRT on ANA.",
	},
}

// DELToCDGStopovers are the candidate stopover cities for Delhi → Paris CDG.
//
// Route geometry: DEL is at ~28°N, 77°E. CDG is at ~49°N, 2°E.
// Gulf carrier hubs are the primary corridor — strong connectivity
// to both India and Paris via ME3 airlines and Air France partners.
var DELToCDGStopovers = []StopoverCity{
	// === GULF — Primary corridor ===
	{
		City:    "Doha",
		Airport: "DOH",
		Region:  "gulf",
		MinStay: types.DefaultMinStopover,
		MaxStay: types.DefaultMaxStopover,
		Notes:   "Qatar Airways hub. DEL-DOH very frequent on QR. DOH-CDG direct on QR/Air France.",
	},
	{
		City:    "Abu Dhabi",
		Airport: "AUH",
		Region:  "gulf",
		MinStay: types.DefaultMinStopover,
		MaxStay: types.DefaultMaxStopover,
		Notes:   "Etihad hub. DEL-AUH on Etihad/Air India. AUH-CDG direct on Etihad.",
	},
	{
		City:    "Dubai",
		Airport: "DXB",
		Region:  "gulf",
		MinStay: types.DefaultMinStopover,
		MaxStay: types.DefaultMaxStopover,
		Notes:   "Emirates mega-hub. DEL-DXB very frequent on Emirates/AI. DXB-CDG direct on Emirates/Air France.",
	},
	{
		City:    "Istanbul",
		Airport: "IST",
		Region:  "europe",
		MinStay: types.DefaultMinStopover,
		MaxStay: types.DefaultMaxStopover,
		Notes:   "Turkish Airlines mega-hub. DEL-IST on Turkish. IST-CDG direct on Turkish/Air France. Bosphorus, bazaars.",
	},
	{
		City:    "Bahrain",
		Airport: "BAH",
		Region:  "gulf",
		MinStay: types.DefaultMinStopover,
		MaxStay: types.DefaultMaxStopover,
		Notes:   "Gulf Air hub. DEL-BAH on Gulf Air. BAH-CDG direct on Gulf Air. Compact, walkable.",
	},
}

// BOMToCDGStopovers are the candidate stopover cities for Mumbai → Paris CDG.
//
// Route geometry: BOM is at ~19°N, 73°E. CDG is at ~49°N, 2°E.
// Gulf carrier hubs are the primary corridor with strong BOM connectivity.
var BOMToCDGStopovers = []StopoverCity{
	{
		City:    "Doha",
		Airport: "DOH",
		Region:  "gulf",
		MinStay: types.DefaultMinStopover,
		MaxStay: types.DefaultMaxStopover,
		Notes:   "Qatar Airways hub. BOM-DOH very frequent on QR. DOH-CDG direct on QR/Air France.",
	},
	{
		City:    "Abu Dhabi",
		Airport: "AUH",
		Region:  "gulf",
		MinStay: types.DefaultMinStopover,
		MaxStay: types.DefaultMaxStopover,
		Notes:   "Etihad hub. BOM-AUH on Etihad/Air India. AUH-CDG direct on Etihad.",
	},
	{
		City:    "Dubai",
		Airport: "DXB",
		Region:  "gulf",
		MinStay: types.DefaultMinStopover,
		MaxStay: types.DefaultMaxStopover,
		Notes:   "Emirates mega-hub. BOM-DXB very frequent on Emirates/AI. DXB-CDG direct on Emirates/Air France.",
	},
	{
		City:    "Istanbul",
		Airport: "IST",
		Region:  "europe",
		MinStay: types.DefaultMinStopover,
		MaxStay: types.DefaultMaxStopover,
		Notes:   "Turkish Airlines mega-hub. BOM-IST on Turkish. IST-CDG direct on Turkish/Air France.",
	},
	{
		City:    "Bahrain",
		Airport: "BAH",
		Region:  "gulf",
		MinStay: types.DefaultMinStopover,
		MaxStay: types.DefaultMaxStopover,
		Notes:   "Gulf Air hub. BOM-BAH on Gulf Air. BAH-CDG direct on Gulf Air.",
	},
}

// DELToICNStopovers are the candidate stopover cities for Delhi → Seoul Incheon.
//
// Route geometry: DEL is at ~28°N, 77°E. ICN is at ~37°N, 126°E.
// Southeast and East Asian corridor.
var DELToICNStopovers = []StopoverCity{
	// === SOUTHEAST ASIA ===
	{
		City:    "Bangkok",
		Airport: "BKK",
		Region:  "southeast_asia",
		MinStay: types.DefaultMinStopover,
		MaxStay: types.DefaultMaxStopover,
		Notes:   "Street food capital, temples, vibrant nightlife",
	},
	{
		City:    "Singapore",
		Airport: "SIN",
		Region:  "southeast_asia",
		MinStay: types.DefaultMinStopover,
		MaxStay: types.DefaultMaxStopover,
		Notes:   "Singapore Airlines hub, major Asia-Pacific crossroads",
	},
	{
		City:    "Kuala Lumpur",
		Airport: "KUL",
		Region:  "southeast_asia",
		MinStay: types.DefaultMinStopover,
		MaxStay: types.DefaultMaxStopover,
		Notes:   "Petronas Towers, diverse cuisine, affordable luxury",
	},

	// === EAST ASIA ===
	{
		City:    "Hong Kong",
		Airport: "HKG",
		Region:  "east_asia",
		MinStay: types.DefaultMinStopover,
		MaxStay: types.DefaultMaxStopover,
		Notes:   "Victoria Peak, dim sum, world-class shopping",
	},
	{
		City:    "Taipei",
		Airport: "TPE",
		Region:  "east_asia",
		MinStay: types.DefaultMinStopover,
		MaxStay: types.DefaultMaxStopover,
		Notes:   "Night markets, Taipei 101, hot springs",
	},
}

// BOMToICNStopovers are the candidate stopover cities for Mumbai → Seoul Incheon.
//
// Route geometry: BOM is at ~19°N, 73°E. ICN is at ~37°N, 126°E.
// East Asian corridor with strong BOM connectivity.
var BOMToICNStopovers = []StopoverCity{
	{
		City:    "Bangkok",
		Airport: "BKK",
		Region:  "southeast_asia",
		MinStay: types.DefaultMinStopover,
		MaxStay: types.DefaultMaxStopover,
		Notes:   "Thai Airways hub. BOM-BKK very frequent and cheap. BKK-ICN direct on Thai/Korean Air.",
	},
	{
		City:    "Singapore",
		Airport: "SIN",
		Region:  "southeast_asia",
		MinStay: types.DefaultMinStopover,
		MaxStay: types.DefaultMaxStopover,
		Notes:   "SQ hub. BOM-SIN frequent. SIN-ICN direct on SQ/Korean Air.",
	},
	{
		City:    "Hong Kong",
		Airport: "HKG",
		Region:  "east_asia",
		MinStay: types.DefaultMinStopover,
		MaxStay: types.DefaultMaxStopover,
		Notes:   "Cathay Pacific hub. BOM-HKG on Cathay/Air India. HKG-ICN direct on Cathay/Korean Air.",
	},
	{
		City:    "Taipei",
		Airport: "TPE",
		Region:  "east_asia",
		MinStay: types.DefaultMinStopover,
		MaxStay: types.DefaultMaxStopover,
		Notes:   "EVA Air/China Airlines hub. TPE-ICN very frequent. BOM-TPE on EVA/China Airlines.",
	},
}

// DELToHKGStopovers are the candidate stopover cities for Delhi → Hong Kong.
//
// Route geometry: DEL is at ~28°N, 77°E. HKG is at ~22°N, 114°E.
// Southeast Asian corridor.
var DELToHKGStopovers = []StopoverCity{
	// === SOUTHEAST ASIA ===
	{
		City:    "Bangkok",
		Airport: "BKK",
		Region:  "southeast_asia",
		MinStay: types.DefaultMinStopover,
		MaxStay: types.DefaultMaxStopover,
		Notes:   "Street food capital, temples, vibrant nightlife",
	},
	{
		City:    "Singapore",
		Airport: "SIN",
		Region:  "southeast_asia",
		MinStay: types.DefaultMinStopover,
		MaxStay: types.DefaultMaxStopover,
		Notes:   "Singapore Airlines hub, major Asia-Pacific crossroads",
	},
	{
		City:    "Kuala Lumpur",
		Airport: "KUL",
		Region:  "southeast_asia",
		MinStay: types.DefaultMinStopover,
		MaxStay: types.DefaultMaxStopover,
		Notes:   "Petronas Towers, diverse cuisine, affordable luxury",
	},

	// === SOUTH ASIA ===
	{
		City:    "Kolkata",
		Airport: "CCU",
		Region:  "south_asia",
		MinStay: types.DefaultMinStopover,
		MaxStay: 72 * time.Hour,
		Notes:   "Colonial heritage, literary culture, gateway to East India",
	},

	// === EAST ASIA ===
	{
		City:    "Taipei",
		Airport: "TPE",
		Region:  "east_asia",
		MinStay: types.DefaultMinStopover,
		MaxStay: types.DefaultMaxStopover,
		Notes:   "Night markets, Taipei 101, hot springs",
	},
}

// BOMToHKGStopovers are the candidate stopover cities for Mumbai → Hong Kong.
//
// Route geometry: BOM is at ~19°N, 73°E. HKG is at ~22°N, 114°E.
// Southeast Asian corridor with strong BOM connectivity.
var BOMToHKGStopovers = []StopoverCity{
	{
		City:    "Bangkok",
		Airport: "BKK",
		Region:  "southeast_asia",
		MinStay: types.DefaultMinStopover,
		MaxStay: types.DefaultMaxStopover,
		Notes:   "Thai Airways hub. BOM-BKK very frequent and cheap. BKK-HKG direct on Thai/Cathay.",
	},
	{
		City:    "Singapore",
		Airport: "SIN",
		Region:  "southeast_asia",
		MinStay: types.DefaultMinStopover,
		MaxStay: types.DefaultMaxStopover,
		Notes:   "SQ hub. BOM-SIN frequent. SIN-HKG direct on SQ/Cathay.",
	},
	{
		City:    "Kuala Lumpur",
		Airport: "KUL",
		Region:  "southeast_asia",
		MinStay: types.DefaultMinStopover,
		MaxStay: types.DefaultMaxStopover,
		Notes:   "Malaysia Airlines/AirAsia hub. BOM-KUL on AirAsia/MAS. KUL-HKG direct on MAS/Cathay.",
	},
	{
		City:    "Taipei",
		Airport: "TPE",
		Region:  "east_asia",
		MinStay: types.DefaultMinStopover,
		MaxStay: types.DefaultMaxStopover,
		Notes:   "EVA Air/China Airlines hub. TPE-HKG very frequent. BOM-TPE on EVA/China Airlines.",
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
	routeKey("DEL", "SFO"): DELToSFOStopovers,
	routeKey("BOM", "SFO"): BOMToSFOStopovers,
	routeKey("DEL", "SYD"): DELToSYDStopovers,
	routeKey("BOM", "SYD"): BOMToSYDStopovers,
	routeKey("DEL", "FRA"): DELToFRAStopovers,
	routeKey("BOM", "FRA"): BOMToFRAStopovers,
	routeKey("DEL", "BKK"): DELToBKKStopovers,
	routeKey("BOM", "BKK"): BOMToBKKStopovers,
	routeKey("DEL", "NRT"): DELToNRTStopovers,
	routeKey("BOM", "NRT"): BOMToNRTStopovers,
	routeKey("DEL", "MEL"): DELToMELStopovers,
	routeKey("BOM", "MEL"): BOMToMELStopovers,
	routeKey("DEL", "CDG"): DELToCDGStopovers,
	routeKey("BOM", "CDG"): BOMToCDGStopovers,
	routeKey("DEL", "ICN"): DELToICNStopovers,
	routeKey("BOM", "ICN"): BOMToICNStopovers,
	routeKey("DEL", "HKG"): DELToHKGStopovers,
	routeKey("BOM", "HKG"): BOMToHKGStopovers,
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
// origin-destination pair. It checks the forward direction first, then
// the reverse (dest→origin), and falls back to global hubs. Reverse
// results are filtered to exclude origin/destination airports.
func StopoversForRoute(origin, destination string) []StopoverCity {
	if route := stopoversMap[routeKey(origin, destination)]; route != nil {
		return route
	}
	// Try reverse direction — same stopovers work both ways.
	if route := stopoversMap[routeKey(destination, origin)]; route != nil {
		var filtered []StopoverCity
		for _, s := range route {
			if s.Airport != origin && s.Airport != destination {
				filtered = append(filtered, s)
			}
		}
		return filtered
	}
	var hubs []StopoverCity
	for _, h := range GlobalFallbackHubs {
		if h.Airport != origin && h.Airport != destination {
			hubs = append(hubs, h)
		}
	}
	return hubs
}
