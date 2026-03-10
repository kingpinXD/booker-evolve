package search

import (
	"time"

	"booker/types"
)

// ==========================================================================
// BLOCKED AIRLINES & HUBS
// ==========================================================================
//
// These lists are driven by the March 2026 Middle East airspace closures
// caused by the Iran-Israel-US conflict.
//
// Status as of March 6, 2026:
//   - UAE airspace: limited reopening, most flights still cancelled
//   - Qatar airspace: fully closed, Qatar Airways suspended (6th extension)
//   - Bahrain, Kuwait: closed
//   - Oman: partially open but unreliable
//   - Saudi Arabia: Jeddah/Riyadh operational but adjacent to conflict zone,
//     routing reliability uncertain
//   - Iran, Iraq, Jordan, Israel: closed
//
// HOW TO UPDATE: When airspace reopens, remove entries from these maps.
// When new closures happen, add entries. Check https://safeairspace.net/
// for current status.
//
// TODO(iterate): Make this dynamic — pull from a news/status API or
// allow override via config. Airspace may reopen at any time.

var BlockedAirlines = map[string]string{
	// UAE carriers
	"EK": "Emirates",
	"EY": "Etihad Airways",
	"FZ": "Flydubai",
	"G9": "Air Arabia",

	// Qatar
	"QR": "Qatar Airways",

	// Bahrain
	"GF": "Gulf Air",

	// Kuwait
	"KU": "Kuwait Airways",

	// Oman (partially open but routing unreliable)
	"WY": "Oman Air",

	// Saudi Arabia (adjacent to conflict zone, unreliable routing)
	"SV": "Saudi Arabian Airlines",

	// Iran
	"IR": "Iran Air",
	"W5": "Mahan Air",
	"EP": "Iran Aseman Airlines",

	// Iraq
	"IA": "Iraqi Airways",

	// Israel (conflict zone)
	"LY": "El Al",

	// Jordan (airspace closed)
	"RJ": "Royal Jordanian",
}

var BlockedHubs = map[string]string{
	// UAE
	"DXB": "Dubai International",
	"AUH": "Abu Dhabi Zayed International",
	"SHJ": "Sharjah International",
	"DWC": "Dubai World Central",

	// Qatar
	"DOH": "Hamad International",

	// Bahrain
	"BAH": "Bahrain International",

	// Kuwait
	"KWI": "Kuwait International",

	// Saudi Arabia (adjacent to conflict zone)
	"JED": "Jeddah King Abdulaziz International",
	"RUH": "Riyadh King Khalid International",
	"DMM": "Dammam King Fahd International",

	// Iran
	"IKA": "Tehran Imam Khomeini",
	"THR": "Tehran Mehrabad",

	// Iraq
	"BGW": "Baghdad International",
	"EBL": "Erbil International",

	// Israel
	"TLV": "Ben Gurion",

	// Jordan
	"AMM": "Queen Alia International",
}

// ==========================================================================
// FILTER FUNCTIONS
// ==========================================================================

// IsAirlineBlocked returns true if the given IATA carrier code is blocked.
func IsAirlineBlocked(iataCode string) bool {
	_, blocked := BlockedAirlines[iataCode]
	return blocked
}

// IsHubBlocked returns true if the given airport IATA code is blocked.
func IsHubBlocked(iataCode string) bool {
	_, blocked := BlockedHubs[iataCode]
	return blocked
}

// FilterFlights removes flights that use blocked airlines or route through
// blocked hubs. It checks every segment of each flight.
func FilterFlights(flights []types.Flight) []types.Flight {
	filtered := make([]types.Flight, 0, len(flights))
	for _, f := range flights {
		if isFlightBlocked(f) {
			continue
		}
		filtered = append(filtered, f)
	}
	return filtered
}

// FilterByDateRange keeps only flights whose first departure falls within
// [earliest, latest]. This is needed because the Kiwi one-way API returns
// results across ALL future dates, not filtered by a specific date.
//
// We apply this post-fetch to narrow results to the user's travel window.
func FilterByDateRange(flights []types.Flight, earliest, latest time.Time) []types.Flight {
	filtered := make([]types.Flight, 0, len(flights))
	for _, f := range flights {
		if len(f.Outbound) == 0 {
			continue
		}
		dep := f.Outbound[0].DepartureTime
		if !dep.Before(earliest) && !dep.After(latest) {
			filtered = append(filtered, f)
		}
	}
	return filtered
}

// FilterByMaxStops keeps only flights with at most maxStops connections.
// A direct flight has 0 stops, a one-stop has 1, etc.
// A negative maxStops means no limit.
func FilterByMaxStops(flights []types.Flight, maxStops int) []types.Flight {
	if maxStops < 0 {
		return flights
	}
	filtered := make([]types.Flight, 0, len(flights))
	for _, f := range flights {
		if f.Stops() <= maxStops {
			filtered = append(filtered, f)
		}
	}
	return filtered
}

// FilterZeroPrices removes flights with a $0 price — these are data artifacts
// from providers (e.g. Google Flights returning incomplete pricing).
func FilterZeroPrices(flights []types.Flight) []types.Flight {
	filtered := make([]types.Flight, 0, len(flights))
	for _, f := range flights {
		if f.Price.Amount > 0 {
			filtered = append(filtered, f)
		}
	}
	return filtered
}

func isFlightBlocked(f types.Flight) bool {
	for _, segments := range [][]types.Segment{f.Outbound, f.Return} {
		for _, seg := range segments {
			if IsAirlineBlocked(seg.Airline) || IsAirlineBlocked(seg.OperatingCarrier) {
				return true
			}
			if IsHubBlocked(seg.Origin) || IsHubBlocked(seg.Destination) {
				return true
			}
		}
	}
	return false
}
