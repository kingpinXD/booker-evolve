package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"time"

	"booker/currency"
	"booker/search"

	"github.com/jedib0t/go-pretty/v6/table"
)

// isMultiLeg returns true if any itinerary has more than one leg.
func isMultiLeg(itineraries []search.Itinerary) bool {
	for _, itin := range itineraries {
		if len(itin.Legs) > 1 {
			return true
		}
	}
	return false
}

// hasScores returns true if any itinerary has a non-zero score.
func hasScores(itineraries []search.Itinerary) bool {
	for _, itin := range itineraries {
		if itin.Score != 0 {
			return true
		}
	}
	return false
}

// truncateText truncates s to maxLen characters, appending "..." if truncated.
func truncateText(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

const reasonMaxLen = 50

func printTable(w io.Writer, itineraries []search.Itinerary, cur string) {
	t := table.NewWriter()
	t.SetOutputMirror(w)
	t.SetStyle(table.StyleRounded)

	multiLeg := isMultiLeg(itineraries)
	scored := hasScores(itineraries)

	// Build header dynamically based on layout and whether scores exist.
	var header table.Row
	if multiLeg {
		header = table.Row{"#"}
		if scored {
			header = append(header, "Score")
		}
		header = append(header, "Price", "Route",
			"L1 Airlines", "L2 Airlines", "L1 Cabin", "L2 Cabin",
			"L1 Depart", "L1 Arrive",
			"L2 Depart", "L2 Arrive",
			"Stopover", "Stops", "Duration", "L1 CO2", "L2 CO2")
		if scored {
			header = append(header, "Reason")
		}
	} else {
		header = table.Row{"#"}
		if scored {
			header = append(header, "Score")
		}
		header = append(header, "Price", "Route",
			"Airlines", "Cabin", "Departure", "Arrival", "Stops", "Duration", "CO2")
		if scored {
			header = append(header, "Reason")
		}
	}
	t.AppendHeader(header)

	for i, itin := range itineraries {
		converted, _ := currency.Convert(itin.TotalPrice, cur)
		dur := formatDuration(itin.TotalTravel)
		stops := formatStops(itin)

		var row table.Row
		if multiLeg {
			row = table.Row{i + 1}
			if scored {
				row = append(row, fmt.Sprintf("%.0f", itin.Score))
			}
			row = append(row,
				fmt.Sprintf("%s%.0f", currencySymbol(cur), converted.Amount),
				routeString(itin),
				legAirlines(itin, 0),
				legAirlines(itin, 1),
				legCabin(itin, 0),
				legCabin(itin, 1),
				legDeparture(itin, 0),
				legArrival(itin, 0),
				legDeparture(itin, 1),
				legArrival(itin, 1),
				stopoverString(itin),
				stops,
				dur,
				legCarbon(itin, 0),
				legCarbon(itin, 1))
			if scored {
				row = append(row, truncateText(itin.Reasoning, reasonMaxLen))
			}
		} else {
			row = table.Row{i + 1}
			if scored {
				row = append(row, fmt.Sprintf("%.0f", itin.Score))
			}
			row = append(row,
				fmt.Sprintf("%s%.0f", currencySymbol(cur), converted.Amount),
				routeString(itin),
				legAirlines(itin, 0),
				legCabin(itin, 0),
				legDeparture(itin, 0),
				legArrival(itin, 0),
				stops,
				dur,
				legCarbon(itin, 0))
			if scored {
				row = append(row, truncateText(itin.Reasoning, reasonMaxLen))
			}
		}
		t.AppendRow(row)
	}

	_, _ = fmt.Fprintln(w)
	t.Render()
	if s := priceSummary(itineraries, cur); s != "" {
		_, _ = fmt.Fprintln(w, s)
	}
	_, _ = fmt.Fprintln(w)
}

// printBulletResults renders itineraries as concise numbered bullets.
// Multi-leg itineraries show per-leg sub-bullets.
func printBulletResults(w io.Writer, itineraries []search.Itinerary, cur string) {
	if len(itineraries) == 0 {
		return
	}
	scored := hasScores(itineraries)
	sym := currencySymbol(cur)
	for i, itin := range itineraries {
		converted, _ := currency.Convert(itin.TotalPrice, cur)
		dur := formatDuration(itin.TotalTravel)
		stops := formatStops(itin)

		var score string
		if scored && itin.Score != 0 {
			score = fmt.Sprintf(" [Score: %.0f]", itin.Score)
		}

		if len(itin.Legs) <= 1 {
			_, _ = fmt.Fprintf(w, "%d. %s | %s | %s | %s stops | %s%.0f%s\n",
				i+1, legAirlines(itin, 0), routeString(itin), dur, stops, sym, converted.Amount, score)
		} else {
			_, _ = fmt.Fprintf(w, "%d. %s | %s%.0f%s\n",
				i+1, routeString(itin), sym, converted.Amount, score)
			for j, leg := range itin.Legs {
				airline := legAirlines(itin, j)
				legDur := formatDuration(leg.Flight.TotalDuration)
				legStops := leg.Flight.Stops()
				_, _ = fmt.Fprintf(w, "   Leg %d: %s | %s | %s stops\n",
					j+1, airline, legDur, fmt.Sprint(legStops))
			}
		}
	}
}

// priceSummary returns a one-line summary of price range and result count.
// For multi-leg itineraries with TotalTrip data, appends trip duration range.
func priceSummary(itineraries []search.Itinerary, cur string) string {
	if len(itineraries) == 0 {
		return ""
	}
	sym := currencySymbol(cur)
	minPrice, maxPrice := itineraries[0].TotalPrice, itineraries[0].TotalPrice
	for _, itin := range itineraries[1:] {
		if itin.TotalPrice.Amount < minPrice.Amount {
			minPrice = itin.TotalPrice
		}
		if itin.TotalPrice.Amount > maxPrice.Amount {
			maxPrice = itin.TotalPrice
		}
	}
	minC, _ := currency.Convert(minPrice, cur)
	maxC, _ := currency.Convert(maxPrice, cur)

	noun := "results"
	if len(itineraries) == 1 {
		noun = "result"
	}
	var summary string
	if minC.Amount == maxC.Amount {
		summary = fmt.Sprintf("%d %s | %s%.0f", len(itineraries), noun, sym, minC.Amount)
	} else {
		summary = fmt.Sprintf("%d %s | %s%.0f - %s%.0f", len(itineraries), noun, sym, minC.Amount, sym, maxC.Amount)
	}

	// For multi-leg itineraries, append trip duration range.
	if len(itineraries[0].Legs) > 1 {
		minTrip, maxTrip := itineraries[0].TotalTrip, itineraries[0].TotalTrip
		hasTripData := itineraries[0].TotalTrip > 0
		for _, itin := range itineraries[1:] {
			if itin.TotalTrip <= 0 {
				continue
			}
			hasTripData = true
			if itin.TotalTrip < minTrip || minTrip == 0 {
				minTrip = itin.TotalTrip
			}
			if itin.TotalTrip > maxTrip {
				maxTrip = itin.TotalTrip
			}
		}
		if hasTripData {
			if minTrip == maxTrip {
				summary += fmt.Sprintf(" | %s total", formatTripDuration(minTrip))
			} else {
				summary += fmt.Sprintf(" | %s - %s total", formatTripDuration(minTrip), formatTripDuration(maxTrip))
			}
		}
	}

	return summary
}

// formatTripDuration formats a duration as "Xd Yh" for compact display.
func formatTripDuration(d time.Duration) string {
	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	return fmt.Sprintf("%dd %dh", days, hours)
}

func routeString(itin search.Itinerary) string {
	if len(itin.Legs) == 0 {
		return ""
	}
	route := ""
	for _, leg := range itin.Legs {
		for _, seg := range leg.Flight.Outbound {
			if route != "" {
				route += "→"
			}
			route += seg.Origin
		}
	}
	// Append final destination from last segment.
	lastLeg := itin.Legs[len(itin.Legs)-1].Flight.Outbound
	if len(lastLeg) > 0 {
		route += "→" + lastLeg[len(lastLeg)-1].Destination
	}
	return route
}

func legCabin(itin search.Itinerary, legIdx int) string {
	if legIdx >= len(itin.Legs) {
		return ""
	}
	segs := itin.Legs[legIdx].Flight.Outbound
	if len(segs) == 0 {
		return ""
	}
	return string(segs[0].CabinClass)
}

// legAircraft returns the aircraft type from the first segment of the given leg.
func legAircraft(itin search.Itinerary, legIdx int) string {
	if legIdx >= len(itin.Legs) {
		return ""
	}
	segs := itin.Legs[legIdx].Flight.Outbound
	if len(segs) == 0 {
		return ""
	}
	return segs[0].Aircraft
}

// legLegroom returns the legroom from the first segment of the given leg.
func legLegroom(itin search.Itinerary, legIdx int) string {
	if legIdx >= len(itin.Legs) {
		return ""
	}
	segs := itin.Legs[legIdx].Flight.Outbound
	if len(segs) == 0 {
		return ""
	}
	return segs[0].Legroom
}

// legSeatsLeft returns the minimum SeatsLeft across all segments of the given leg.
// Returns 0 if no segment has seat data.
func legSeatsLeft(itin search.Itinerary, legIdx int) int {
	if legIdx >= len(itin.Legs) {
		return 0
	}
	minSeats := 0
	for _, seg := range itin.Legs[legIdx].Flight.Outbound {
		if seg.SeatsLeft > 0 {
			if minSeats == 0 || seg.SeatsLeft < minSeats {
				minSeats = seg.SeatsLeft
			}
		}
	}
	return minSeats
}

func legAirlines(itin search.Itinerary, legIdx int) string {
	if legIdx >= len(itin.Legs) {
		return ""
	}
	seen := map[string]bool{}
	result := ""
	for _, seg := range itin.Legs[legIdx].Flight.Outbound {
		name := seg.AirlineName
		if name == "" {
			name = seg.Airline
		}
		// Append codeshare indicator when operating carrier differs.
		if seg.OperatingCarrier != "" && seg.OperatingCarrier != seg.Airline {
			name += " (op. " + seg.OperatingCarrier + ")"
		}
		if seen[name] {
			continue
		}
		seen[name] = true
		if result != "" {
			result += ", "
		}
		result += name
	}
	return result
}

func legDeparture(itin search.Itinerary, legIdx int) string {
	if legIdx >= len(itin.Legs) {
		return ""
	}
	segs := itin.Legs[legIdx].Flight.Outbound
	if len(segs) == 0 {
		return ""
	}
	return segs[0].DepartureTime.Format(outputDateTimeFmt)
}

// isNextDay returns true if the arrival date is after the departure date.
func isNextDay(dep, arr time.Time) bool {
	return arr.YearDay() != dep.YearDay() || arr.Year() != dep.Year()
}

func legArrival(itin search.Itinerary, legIdx int) string {
	if legIdx >= len(itin.Legs) {
		return ""
	}
	segs := itin.Legs[legIdx].Flight.Outbound
	if len(segs) == 0 {
		return ""
	}
	dep := segs[0].DepartureTime
	arr := segs[len(segs)-1].ArrivalTime
	s := arr.Format(outputDateTimeFmt)
	if isNextDay(dep, arr) {
		days := (arr.YearDay() - dep.YearDay())
		if arr.Year() != dep.Year() {
			days += 365 // approximate; good enough for display
		}
		s += fmt.Sprintf(" (+%d)", days)
	}
	return s
}

func legCarbon(itin search.Itinerary, legIdx int) string {
	if legIdx >= len(itin.Legs) || itin.Legs[legIdx].Flight.CarbonKg == 0 {
		return ""
	}
	f := itin.Legs[legIdx].Flight
	if f.CarbonDiffPct > 0 {
		return fmt.Sprintf("%dkg (+%d%%)", f.CarbonKg, f.CarbonDiffPct)
	}
	if f.CarbonDiffPct < 0 {
		return fmt.Sprintf("%dkg (%d%%)", f.CarbonKg, f.CarbonDiffPct)
	}
	return fmt.Sprintf("%dkg", f.CarbonKg)
}

func legBookingURL(itin search.Itinerary, legIdx int) string {
	if legIdx >= len(itin.Legs) {
		return ""
	}
	return itin.Legs[legIdx].Flight.BookingURL
}

func stopoverString(itin search.Itinerary) string {
	if len(itin.Legs) == 0 || itin.Legs[0].Stopover == nil {
		return ""
	}
	s := itin.Legs[0].Stopover
	return fmt.Sprintf("%s (%s)", s.City, formatDuration(s.Duration))
}

// itineraryStops returns the total number of connections across all legs.
func itineraryStops(itin search.Itinerary) int {
	n := 0
	for _, leg := range itin.Legs {
		n += leg.Flight.Stops()
	}
	return n
}

// formatStops returns a display string for total stops across all legs.
// If stops > 0 and layover data is available, it includes total layover time
// (e.g. "1 (2h 30m)"). Otherwise returns just the count (e.g. "0" or "1").
func formatStops(itin search.Itinerary) string {
	stops := itineraryStops(itin)
	if stops == 0 {
		return "0"
	}
	var totalLayover time.Duration
	for _, leg := range itin.Legs {
		for _, seg := range leg.Flight.Outbound {
			totalLayover += seg.LayoverDuration
		}
	}
	if totalLayover == 0 {
		return fmt.Sprintf("%d", stops)
	}
	return fmt.Sprintf("%d (%s)", stops, formatDuration(totalLayover))
}

func currencySymbol(cur string) string {
	switch cur {
	case "CAD":
		return "C$"
	case "USD":
		return "$"
	case "EUR":
		return "€"
	case "GBP":
		return "£"
	case "INR":
		return "₹"
	default:
		return cur + " "
	}
}

func formatDuration(d time.Duration) string {
	hours := int(d.Hours())
	mins := int(d.Minutes()) % 60
	switch {
	case hours >= 24:
		return fmt.Sprintf("%dd %dh", hours/24, hours%24)
	case mins == 0:
		return fmt.Sprintf("%dh", hours)
	default:
		return fmt.Sprintf("%dh %dm", hours, mins)
	}
}

// jsonSegment is the JSON representation of a single non-stop flight segment.
type jsonSegment struct {
	Airline         string `json:"airline"`
	FlightNumber    string `json:"flight_number,omitempty"`
	Origin          string `json:"origin"`
	Destination     string `json:"destination"`
	Departure       string `json:"departure"`
	Arrival         string `json:"arrival"`
	Duration        string `json:"duration"`
	Aircraft        string `json:"aircraft,omitempty"`
	Legroom         string `json:"legroom,omitempty"`
	LayoverDuration string `json:"layover_duration,omitempty"`
	Overnight       bool   `json:"overnight,omitempty"`
}

// jsonLeg is the JSON representation of a single flight leg.
type jsonLeg struct {
	Airlines         string        `json:"airlines"`
	AirlineCode      string        `json:"airline_code,omitempty"`
	FlightNumber     string        `json:"flight_number,omitempty"`
	OperatingCarrier string        `json:"operating_carrier,omitempty"`
	CabinClass       string        `json:"cabin_class,omitempty"`
	Origin           string        `json:"origin"`
	OriginCity       string        `json:"origin_city,omitempty"`
	OriginName       string        `json:"origin_name,omitempty"`
	Dest             string        `json:"destination"`
	DestinationCity  string        `json:"destination_city,omitempty"`
	DestinationName  string        `json:"destination_name,omitempty"`
	Departure        string        `json:"departure"`
	Arrival          string        `json:"arrival,omitempty"`
	Duration         string        `json:"duration"`
	Stops            int           `json:"stops"`
	CarbonKg         int           `json:"carbon_kg,omitempty"`
	TypicalCarbonKg  int           `json:"typical_carbon_kg,omitempty"`
	CarbonDiffPct    int           `json:"carbon_diff_percent,omitempty"`
	BookingURL       string        `json:"booking_url,omitempty"`
	Aircraft         string        `json:"aircraft,omitempty"`
	Legroom          string        `json:"legroom,omitempty"`
	SeatsLeft        int           `json:"seats_left,omitempty"`
	ArrivalNextDay   bool          `json:"arrival_next_day,omitempty"`
	Segments         []jsonSegment `json:"segments,omitempty"`
}

// jsonItinerary is the JSON representation of a search result.
type jsonItinerary struct {
	Rank      int       `json:"rank"`
	Score     float64   `json:"score,omitempty"`
	Reasoning string    `json:"reasoning,omitempty"`
	Price     float64   `json:"price"`
	Currency  string    `json:"currency"`
	Route     string    `json:"route"`
	Duration  string    `json:"duration"`
	TotalTrip string    `json:"total_trip,omitempty"`
	Legs      []jsonLeg `json:"legs"`
	Stopover  string    `json:"stopover,omitempty"`
}

// formatPriceInsights returns a one-line summary of price insights, or empty
// if no meaningful data is available.
func formatPriceInsights(pi search.PriceInsights) string {
	if pi.PriceLevel == "" {
		return ""
	}
	low, high := pi.TypicalPriceRange[0], pi.TypicalPriceRange[1]
	if low == 0 && high == 0 {
		return fmt.Sprintf("Price level: %s", pi.PriceLevel)
	}
	return fmt.Sprintf("Price level: %s | Typical: $%.0f - $%.0f", pi.PriceLevel, low, high)
}

// jsonPriceInsights is the JSON representation of price insights.
type jsonPriceInsights struct {
	PriceLevel        string     `json:"price_level,omitempty"`
	LowestPrice       float64    `json:"lowest_price,omitempty"`
	TypicalPriceRange [2]float64 `json:"typical_price_range,omitempty"`
}

func printJSONWithInsights(w io.Writer, itineraries []search.Itinerary, cur string, pi search.PriceInsights) error {
	type jsonOutput struct {
		Results       []jsonItinerary    `json:"results"`
		PriceInsights *jsonPriceInsights `json:"price_insights,omitempty"`
	}

	results := buildJSONItineraries(itineraries, cur)
	out := jsonOutput{Results: results}
	if pi.PriceLevel != "" {
		out.PriceInsights = &jsonPriceInsights{
			PriceLevel:        pi.PriceLevel,
			LowestPrice:       pi.LowestPrice,
			TypicalPriceRange: pi.TypicalPriceRange,
		}
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(out)
}

func buildJSONItineraries(itineraries []search.Itinerary, cur string) []jsonItinerary {
	out := make([]jsonItinerary, len(itineraries))
	for i, itin := range itineraries {
		converted, _ := currency.Convert(itin.TotalPrice, cur)

		var legs []jsonLeg
		for idx, leg := range itin.Legs {
			segs := leg.Flight.Outbound
			origin, dest, dep, arr := "", "", "", ""
			airlineCode, flightNum, opCarrier, originCity, destCity, originName, destName := "", "", "", "", "", "", ""
			if len(segs) > 0 {
				origin = segs[0].Origin
				dest = segs[len(segs)-1].Destination
				dep = segs[0].DepartureTime.Format(time.RFC3339)
				arr = segs[len(segs)-1].ArrivalTime.Format(time.RFC3339)
				airlineCode = segs[0].Airline
				flightNum = segs[0].FlightNumber
				opCarrier = segs[0].OperatingCarrier
				originCity = segs[0].OriginCity
				originName = segs[0].OriginName
				destCity = segs[len(segs)-1].DestinationCity
				destName = segs[len(segs)-1].DestinationName
			}
			nextDay := false
			if len(segs) > 0 {
				nextDay = isNextDay(segs[0].DepartureTime, segs[len(segs)-1].ArrivalTime)
			}
			var jSegs []jsonSegment
			for _, seg := range segs {
				js := jsonSegment{
					Airline:      seg.Airline,
					FlightNumber: seg.FlightNumber,
					Origin:       seg.Origin,
					Destination:  seg.Destination,
					Departure:    seg.DepartureTime.Format(time.RFC3339),
					Arrival:      seg.ArrivalTime.Format(time.RFC3339),
					Duration:     formatDuration(seg.Duration),
					Aircraft:     seg.Aircraft,
					Legroom:      seg.Legroom,
					Overnight:    seg.Overnight,
				}
				if seg.LayoverDuration > 0 {
					js.LayoverDuration = formatDuration(seg.LayoverDuration)
				}
				jSegs = append(jSegs, js)
			}
			legs = append(legs, jsonLeg{
				Airlines:         legAirlines(itin, idx),
				AirlineCode:      airlineCode,
				FlightNumber:     flightNum,
				OperatingCarrier: opCarrier,
				CabinClass:       legCabin(itin, idx),
				Origin:           origin,
				OriginCity:       originCity,
				OriginName:       originName,
				Dest:             dest,
				DestinationCity:  destCity,
				DestinationName:  destName,
				Departure:        dep,
				Arrival:          arr,
				Duration:         formatDuration(leg.Flight.TotalDuration),
				Stops:            leg.Flight.Stops(),
				CarbonKg:         leg.Flight.CarbonKg,
				TypicalCarbonKg:  leg.Flight.TypicalCarbonKg,
				CarbonDiffPct:    leg.Flight.CarbonDiffPct,
				BookingURL:       leg.Flight.BookingURL,
				Aircraft:         legAircraft(itin, idx),
				Legroom:          legLegroom(itin, idx),
				SeatsLeft:        legSeatsLeft(itin, idx),
				ArrivalNextDay:   nextDay,
				Segments:         jSegs,
			})
		}

		var totalTrip string
		if itin.TotalTrip > 0 {
			totalTrip = formatTripDuration(itin.TotalTrip)
		}

		out[i] = jsonItinerary{
			Rank:      i + 1,
			Score:     itin.Score,
			Reasoning: itin.Reasoning,
			Price:     converted.Amount,
			Currency:  cur,
			Route:     routeString(itin),
			Duration:  formatDuration(itin.TotalTravel),
			TotalTrip: totalTrip,
			Legs:      legs,
			Stopover:  stopoverString(itin),
		}
	}
	return out
}

func printJSON(w io.Writer, itineraries []search.Itinerary, cur string) error {
	out := buildJSONItineraries(itineraries, cur)
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(out)
}

// formatFareTrend returns a human-readable fare trend summary for flex-date
// searches. Returns empty string when the trend has no data or a single date.
func formatFareTrend(ft search.FareTrend) string {
	if ft.CheapestDate == "" || ft.CheapestDate == ft.PriciestDate {
		return ""
	}
	cheapDay := formatDateShort(ft.CheapestDate)
	priceDay := formatDateShort(ft.PriciestDate)
	diff := ft.MaxPrice - ft.MinPrice
	return fmt.Sprintf("Fare trend: %s is cheapest ($%.0f), %s is most expensive ($%.0f) — $%.0f difference.",
		cheapDay, ft.MinPrice, priceDay, ft.MaxPrice, diff)
}

// formatDateShort converts "2026-03-16" to "Mar 16".
func formatDateShort(dateStr string) string {
	t, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return dateStr
	}
	return t.Format("Jan 2")
}
