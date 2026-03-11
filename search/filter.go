package search

import (
	"fmt"
	"sort"
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

// FilterByMaxPrice keeps only flights at or below the given price ceiling.
// A maxPrice of 0 means no limit.
func FilterByMaxPrice(flights []types.Flight, maxPrice float64) []types.Flight {
	if maxPrice <= 0 {
		return flights
	}
	filtered := make([]types.Flight, 0, len(flights))
	for _, f := range flights {
		if f.Price.Amount <= maxPrice {
			filtered = append(filtered, f)
		}
	}
	return filtered
}

// FilterByAlliance keeps only flights where at least one outbound segment's
// airline belongs to the preferred alliance. Empty preference keeps all flights.
func FilterByAlliance(flights []types.Flight, alliance string) []types.Flight {
	if alliance == "" {
		return flights
	}
	filtered := make([]types.Flight, 0, len(flights))
	for _, f := range flights {
		if flightMatchesAlliance(f, alliance) {
			filtered = append(filtered, f)
		}
	}
	return filtered
}

// FilterByDepartureTime keeps only flights whose first outbound segment departs
// within the given time-of-day range [after, before]. Both are "HH:MM" strings.
// Empty strings mean no constraint on that end. Invalid formats return all flights.
func FilterByDepartureTime(flights []types.Flight, after, before string) []types.Flight {
	if after == "" && before == "" {
		return flights
	}
	afterMin, afterOK := parseHHMM(after)
	beforeMin, beforeOK := parseHHMM(before)
	if (after != "" && !afterOK) || (before != "" && !beforeOK) {
		return flights
	}
	filtered := make([]types.Flight, 0, len(flights))
	for _, f := range flights {
		if len(f.Outbound) == 0 {
			continue
		}
		dep := f.Outbound[0].DepartureTime
		depMin := dep.Hour()*60 + dep.Minute()
		if afterOK && depMin < afterMin {
			continue
		}
		if beforeOK && depMin > beforeMin {
			continue
		}
		filtered = append(filtered, f)
	}
	return filtered
}

// parseHHMM parses "HH:MM" into minutes-since-midnight.
func parseHHMM(s string) (int, bool) {
	if s == "" {
		return 0, false
	}
	var h, m int
	if _, err := fmt.Sscanf(s, "%d:%d", &h, &m); err != nil {
		return 0, false
	}
	if h < 0 || h > 23 || m < 0 || m > 59 {
		return 0, false
	}
	return h*60 + m, true
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

// SortResults sorts itineraries in place by the given mode.
// Supported modes: "price" (default), "duration", "departure".
// Unknown modes default to price.
func SortResults(itins []Itinerary, sortBy string) {
	if len(itins) < 2 {
		return
	}
	switch sortBy {
	case "duration":
		sort.Slice(itins, func(i, j int) bool {
			return itins[i].TotalTravel < itins[j].TotalTravel
		})
	case "departure":
		sort.Slice(itins, func(i, j int) bool {
			return firstDeparture(itins[i]).Before(firstDeparture(itins[j]))
		})
	default: // "price" or unknown
		sort.Slice(itins, func(i, j int) bool {
			return itins[i].TotalPrice.Amount < itins[j].TotalPrice.Amount
		})
	}
}

// firstDeparture returns the departure time of the first segment of the first leg.
func firstDeparture(itin Itinerary) time.Time {
	if len(itin.Legs) > 0 && len(itin.Legs[0].Flight.Outbound) > 0 {
		return itin.Legs[0].Flight.Outbound[0].DepartureTime
	}
	return time.Time{}
}

func flightMatchesAlliance(f types.Flight, alliance string) bool {
	for _, seg := range f.Outbound {
		if Alliance(seg.Airline) == alliance || Alliance(seg.OperatingCarrier) == alliance {
			return true
		}
	}
	return false
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
