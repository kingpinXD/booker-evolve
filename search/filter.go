package search

import (
	"fmt"
	"sort"
	"strings"
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
// [earliest, latest]. When flex-date search is enabled, the provider may
// return results across a range of dates. We apply this post-fetch to
// narrow results to the user's travel window.
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
	return filterByTimeOfDay(flights, after, before, func(f types.Flight) time.Time {
		return f.Outbound[0].DepartureTime
	})
}

// FilterByArrivalTime keeps only flights whose last outbound segment arrives
// within the given time-of-day range [after, before]. Both are "HH:MM" strings.
// Empty strings mean no constraint on that end. Invalid formats return all flights.
func FilterByArrivalTime(flights []types.Flight, after, before string) []types.Flight {
	return filterByTimeOfDay(flights, after, before, func(f types.Flight) time.Time {
		return f.Outbound[len(f.Outbound)-1].ArrivalTime
	})
}

// filterByTimeOfDay is the shared implementation for departure/arrival time
// filters. The extractTime function selects which time to check from a flight
// (caller guarantees len(f.Outbound) > 0).
func filterByTimeOfDay(flights []types.Flight, after, before string, extractTime func(types.Flight) time.Time) []types.Flight {
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
		t := extractTime(f)
		mins := t.Hour()*60 + t.Minute()
		if afterOK && mins < afterMin {
			continue
		}
		if beforeOK && mins > beforeMin {
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

// FilterByMaxDuration keeps only flights whose TotalDuration is at most maxDur.
// A zero maxDur means no limit.
func FilterByMaxDuration(flights []types.Flight, maxDur time.Duration) []types.Flight {
	if maxDur <= 0 {
		return flights
	}
	filtered := make([]types.Flight, 0, len(flights))
	for _, f := range flights {
		if f.TotalDuration <= maxDur {
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

// FilterByAvoidAirlines removes flights where any segment's Airline or
// OperatingCarrier matches the comma-separated avoid list. Empty string keeps all.
func FilterByAvoidAirlines(flights []types.Flight, avoid string) []types.Flight {
	if avoid == "" {
		return flights
	}
	codes := parseAirlineCodes(avoid)
	if len(codes) == 0 {
		return flights
	}
	filtered := make([]types.Flight, 0, len(flights))
	for _, f := range flights {
		if !flightMatchesAvoid(f, codes) {
			filtered = append(filtered, f)
		}
	}
	return filtered
}

// FilterByPreferredAirlines keeps only flights where at least one segment's
// Airline or OperatingCarrier matches the comma-separated preferred list.
// Empty string keeps all flights.
func FilterByPreferredAirlines(flights []types.Flight, preferred string) []types.Flight {
	if preferred == "" {
		return flights
	}
	codes := parseAirlineCodes(preferred)
	if len(codes) == 0 {
		return flights
	}
	filtered := make([]types.Flight, 0, len(flights))
	for _, f := range flights {
		if flightMatchesPreferred(f, codes) {
			filtered = append(filtered, f)
		}
	}
	return filtered
}

func flightMatchesPreferred(f types.Flight, codes map[string]bool) bool {
	for _, seg := range f.Outbound {
		if codes[seg.Airline] || codes[seg.OperatingCarrier] {
			return true
		}
	}
	return false
}

func flightMatchesAvoid(f types.Flight, codes map[string]bool) bool {
	for _, segments := range [][]types.Segment{f.Outbound, f.Return} {
		for _, seg := range segments {
			if codes[seg.Airline] || codes[seg.OperatingCarrier] {
				return true
			}
		}
	}
	return false
}

// SortResults sorts itineraries in place by the given mode.
// Supported modes: "price" (default), "duration", "departure", "score".
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
	case "score":
		sort.Slice(itins, func(i, j int) bool {
			return itins[i].Score > itins[j].Score // descending
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

// ==========================================================================
// SINGLE-FLIGHT PREDICATES
// ==========================================================================
//
// These return true if a single flight passes the given filter criterion.
// Used by multicity.passesAllFilters to avoid wrapping flights in slices.

// FlightPassesBlocked returns true if the flight does not use blocked airlines or hubs.
func FlightPassesBlocked(f types.Flight) bool {
	return !isFlightBlocked(f)
}

// FlightPassesAlliance returns true if the flight matches the preferred alliance
// (or if alliance is empty, meaning no preference).
func FlightPassesAlliance(f types.Flight, alliance string) bool {
	return alliance == "" || flightMatchesAlliance(f, alliance)
}

// FlightPassesDepartureTime returns true if the flight's first departure is
// within the [after, before] time-of-day range. Empty strings mean no constraint.
func FlightPassesDepartureTime(f types.Flight, after, before string) bool {
	return flightPassesTimeOfDay(f, after, before, func(fl types.Flight) time.Time {
		return fl.Outbound[0].DepartureTime
	})
}

// FlightPassesArrivalTime returns true if the flight's last arrival is within
// the [after, before] time-of-day range. Empty strings mean no constraint.
func FlightPassesArrivalTime(f types.Flight, after, before string) bool {
	return flightPassesTimeOfDay(f, after, before, func(fl types.Flight) time.Time {
		return fl.Outbound[len(fl.Outbound)-1].ArrivalTime
	})
}

// FlightPassesMaxDuration returns true if the flight's TotalDuration is at most
// maxDur. Zero maxDur means no limit.
func FlightPassesMaxDuration(f types.Flight, maxDur time.Duration) bool {
	return maxDur <= 0 || f.TotalDuration <= maxDur
}

// FlightPassesAvoidAirlines returns true if no segment matches the avoid list.
func FlightPassesAvoidAirlines(f types.Flight, avoid string) bool {
	if avoid == "" {
		return true
	}
	codes := parseAirlineCodes(avoid)
	return len(codes) == 0 || !flightMatchesAvoid(f, codes)
}

// FlightPassesPreferredAirlines returns true if at least one segment matches
// the preferred list (or if preferred is empty).
func FlightPassesPreferredAirlines(f types.Flight, preferred string) bool {
	if preferred == "" {
		return true
	}
	codes := parseAirlineCodes(preferred)
	return len(codes) == 0 || flightMatchesPreferred(f, codes)
}

// flightPassesTimeOfDay checks a single flight's time against the range.
func flightPassesTimeOfDay(f types.Flight, after, before string, extractTime func(types.Flight) time.Time) bool {
	if after == "" && before == "" {
		return true
	}
	if len(f.Outbound) == 0 {
		return false
	}
	afterMin, afterOK := parseHHMM(after)
	beforeMin, beforeOK := parseHHMM(before)
	if (after != "" && !afterOK) || (before != "" && !beforeOK) {
		return true // invalid format = no constraint
	}
	t := extractTime(f)
	mins := t.Hour()*60 + t.Minute()
	if afterOK && mins < afterMin {
		return false
	}
	if beforeOK && mins > beforeMin {
		return false
	}
	return true
}

// parseAirlineCodes splits a comma-separated string into a set of airline codes.
func parseAirlineCodes(csv string) map[string]bool {
	codes := make(map[string]bool)
	for _, c := range strings.Split(csv, ",") {
		if c = strings.TrimSpace(c); c != "" {
			codes[c] = true
		}
	}
	return codes
}
